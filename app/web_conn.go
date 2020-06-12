// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	goi18n "github.com/mattermost/go-i18n/i18n"
	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
)

const (
	SEND_QUEUE_SIZE           = 256
	SEND_SLOW_WARN            = (SEND_QUEUE_SIZE * 50) / 100
	SEND_DEADLOCK_WARN        = (SEND_QUEUE_SIZE * 95) / 100
	WRITE_WAIT                = 30 * time.Second
	PONG_WAIT                 = 100 * time.Second
	PING_PERIOD               = (PONG_WAIT * 6) / 10
	AUTH_TIMEOUT              = 5 * time.Second
	WEBCONN_MEMBER_CACHE_TIME = 1000 * 60 * 30 // 30 minutes
)

type WebConn struct {
	// 过期时间
	sessionExpiresAt          int64 // This should stay at the top for 64-bit alignment of 64-bit words accessed atomically

	// 归属 App
	App                       *App

	// 底层 socket
	WebSocket                 *websocket.Conn

	// 消息发送管道
	Send                      chan model.WebSocketMessage

	// 鉴权参数
	sessionToken              atomic.Value
	session                   atomic.Value

	// 最后活跃时间
	LastUserActivityAt        int64


	// 用户信息
	UserId                    string
	T                         goi18n.TranslateFunc
	Locale                    string
	AllChannelMembers         map[string]string
	LastAllChannelMembersTime int64

	Sequence                  int64

	closeOnce                 sync.Once
	endWritePump              chan struct{}
	pumpFinished              chan struct{}
}


//
//
//
//
//

func (a *App) NewWebConn(ws *websocket.Conn, session model.Session, t goi18n.TranslateFunc, locale string) *WebConn {


	if len(session.UserId) > 0 {
		// 启动一个 goroutine 执行 f()
		a.Srv.Go(func() {
			// 设置登陆状态
			a.SetStatusOnline(session.UserId, false)
			// 更新 session 最后活跃时间
			a.UpdateLastActivityAtIfNeeded(session)
		})
	}


	//
	wc := &WebConn{
		App:                a,
		Send:               make(chan model.WebSocketMessage, SEND_QUEUE_SIZE),
		WebSocket:          ws,
		LastUserActivityAt: model.GetMillis(),
		UserId:             session.UserId,
		T:                  t,
		Locale:             locale,
		endWritePump:       make(chan struct{}),
		pumpFinished:       make(chan struct{}),
	}

	wc.SetSession(&session)
	wc.SetSessionToken(session.Token)
	wc.SetSessionExpiresAt(session.ExpiresAt)

	return wc
}

func (wc *WebConn) Close() {

	// 关闭底层连接，注意，这里也没有检查出错信息
	wc.WebSocket.Close()

	// 关闭 c.endWritePump 管道，重复关闭可能会 panic，所以用 once.Do()。
	// 关闭 c.endWritePump 管道，会主动通知 writePump 退出，writePump 退出后触发 readPump 退出
	wc.closeOnce.Do(func() {
		close(wc.endWritePump)
	})

	// 阻塞等待 readPump 和 writePump 两个协程均退出，二者都退出后会关闭 c.pumpFinished 管道，Close 便退出阻塞。
	<-wc.pumpFinished
}

func (c *WebConn) GetSessionExpiresAt() int64 {
	return atomic.LoadInt64(&c.sessionExpiresAt)
}

func (c *WebConn) SetSessionExpiresAt(v int64) {
	atomic.StoreInt64(&c.sessionExpiresAt, v)
}

func (c *WebConn) GetSessionToken() string {
	return c.sessionToken.Load().(string)
}

func (c *WebConn) SetSessionToken(v string) {
	c.sessionToken.Store(v)
}

func (c *WebConn) GetSession() *model.Session {
	return c.session.Load().(*model.Session)
}


func (c *WebConn) SetSession(v *model.Session) {

	// 深拷贝
	if v != nil {
		v = v.DeepCopy()
	}
	// 原子赋值
	c.session.Store(v)
}


// 启动 writePump 和 readPump 两个协程
func (c *WebConn) Pump() {

	ch := make(chan struct{})

	go func() {
		// 阻塞式的 writePump
		c.writePump()
		// writePump 退出后会关闭 ch 管道
		close(ch)
	}()

	// 阻塞式的 readPump
	c.readPump()

	// 退出时先关闭 c.endWritePump，主动通知 writePump 退出，它退出后会关闭 ch 管道
	c.closeOnce.Do(func() {
		close(c.endWritePump)
	})

	// 等待 writePump 退出
	<-ch

	// 两个协程都退出了，取消注册
	c.App.HubUnregister(c)

	// 关闭 c.pumpFinished 管道，以通知 close 退出阻塞。
	close(c.pumpFinished)

}


// 不断从 socket 中读取 req 消息， req 被处理后生成 rsp 写入到 c.Send 管道中。
func (c *WebConn) readPump() {


	defer func() {
		c.WebSocket.Close()
	}()

	// 设置读超时
	c.WebSocket.SetReadLimit(model.SOCKET_MAX_MESSAGE_SIZE_KB)
	c.WebSocket.SetReadDeadline(time.Now().Add(PONG_WAIT))

	// 收到 Pong 则更新一下 user 状态
	c.WebSocket.SetPongHandler(func(string) error {
		c.WebSocket.SetReadDeadline(time.Now().Add(PONG_WAIT))
		if c.IsAuthenticated() {
			c.App.Srv.Go(func() {
				c.App.SetStatusAwayIfNeeded(c.UserId, false)
			})
		}
		return nil
	})


	// 不断从 socket 中读取 req 消息，然后调用 Router.ServeWebSocket() 处理它。
	for {
		var req model.WebSocketRequest
		if err := c.WebSocket.ReadJSON(&req); err != nil {
			// browsers will appear as CloseNoStatusReceived
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				mlog.Debug(fmt.Sprintf("websocket.read: client side closed socket userId=%v", c.UserId))
			} else {
				mlog.Debug(fmt.Sprintf("websocket.read: closing websocket for userId=%v error=%v", c.UserId, err.Error()))
			}
			return
		}

		// 根据 action 查找已注册的 handler，用该 handler 处理请求 req 得到 rsp，并将 rsp 发送到 c.Send 管道中留待发送给对端。
		c.App.Srv.WebSocketRouter.ServeWebSocket(c, &req)
	}
}



// 不断从 c.Send 管道中读取 rsp 写入到 socket 中。
func (c *WebConn) writePump() {
	ticker := time.NewTicker(PING_PERIOD)
	authTicker := time.NewTicker(AUTH_TIMEOUT)

	defer func() {
		ticker.Stop()
		authTicker.Stop()
		c.WebSocket.Close()
	}()

	for {
		select {

		// 从 c.Send 管道中读取 rsp 消息并写入到 socket 中。
		case msg, ok := <-c.Send:

			// 如果 c.Send 管道被关闭了，就写一条 CloseMessage 消息后关闭 socket 。
			if !ok {
				c.WebSocket.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
				c.WebSocket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 消息类型强转，
			evt, evtOk := msg.(*model.WebSocketEvent)

			// 如果 c.Send 管道中消息有积压，则根据当前消息 msg 的类型，决定要不要丢弃它。
			skipSend := false
			if len(c.Send) >= SEND_SLOW_WARN {
				// When the pump starts to get slow we'll drop non-critical messages
				if  msg.EventType() == model.WEBSOCKET_EVENT_TYPING ||
					msg.EventType() == model.WEBSOCKET_EVENT_STATUS_CHANGE ||
					msg.EventType() == model.WEBSOCKET_EVENT_CHANNEL_VIEWED {
					mlog.Info(fmt.Sprintf("websocket.slow: dropping message userId=%v type=%v channelId=%v", c.UserId, msg.EventType(), evt.Broadcast.ChannelId))
					skipSend = true
				}
			}


			// 如果不能丢弃当前消息 msg ，则发送它，否则直接忽视。
			if !skipSend {

				// 消息字节数据
				var msgBytes []byte

				// 根据消息类型确定序列化方式，得到 msgBytes
				if evtOk {
					cpyEvt := &model.WebSocketEvent{}
					*cpyEvt = *evt
					cpyEvt.Sequence = c.Sequence
					msgBytes = []byte(cpyEvt.ToJson())
					c.Sequence++
				} else {
					msgBytes = []byte(msg.ToJson())
				}

				// ？？？
				if len(c.Send) >= SEND_DEADLOCK_WARN {
					if evtOk {
						mlog.Warn(fmt.Sprintf("websocket.full: message userId=%v type=%v channelId=%v size=%v", c.UserId, msg.EventType(), evt.Broadcast.ChannelId, len(msg.ToJson())))
					} else {
						mlog.Warn(fmt.Sprintf("websocket.full: message userId=%v type=%v size=%v", c.UserId, msg.EventType(), len(msg.ToJson())))
					}
				}

				// 将消息写入到 socket
				c.WebSocket.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
				if err := c.WebSocket.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
					// browsers will appear as CloseNoStatusReceived
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
						mlog.Debug(fmt.Sprintf("websocket.send: client side closed socket userId=%v", c.UserId))
					} else {
						mlog.Debug(fmt.Sprintf("websocket.send: closing websocket for userId=%v, error=%v", c.UserId, err.Error()))
					}
					return
				}

				// metric 数据统计
				if c.App.Metrics != nil {
					c.App.Srv.Go(func() {
						c.App.Metrics.IncrementWebSocketBroadcast(msg.EventType())
					})
				}
			}

		// 定时发送 Ping 心跳消息
		case <-ticker.C:
			c.WebSocket.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.WebSocket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				// browsers will appear as CloseNoStatusReceived
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
					mlog.Debug(fmt.Sprintf("websocket.ticker: client side closed socket userId=%v", c.UserId))
				} else {
					mlog.Debug(fmt.Sprintf("websocket.ticker: closing websocket for userId=%v error=%v", c.UserId, err.Error()))
				}
				return
			}
		// 监听退出信号
		case <-c.endWritePump:
			return

		// 定时检查鉴权信息
		case <-authTicker.C:
			// 鉴权未通过，则 return 退出
			if c.GetSessionToken() == "" {
				mlog.Debug(fmt.Sprintf("websocket.authTicker: did not authenticate ip=%v", c.WebSocket.RemoteAddr()))
				return
			}
			// 鉴权通过则关闭这个定时器
			authTicker.Stop()
		}
	}
}

func (webCon *WebConn) InvalidateCache() {
	webCon.AllChannelMembers = nil
	webCon.LastAllChannelMembersTime = 0
	webCon.SetSession(nil)
	webCon.SetSessionExpiresAt(0)
}

// 鉴权检查
func (webCon *WebConn) IsAuthenticated() bool {

	// Check the expiry to see if we need to check for a new session

	// 如果 `过期时间` 已经到达
	if webCon.GetSessionExpiresAt() < model.GetMillis() {

		// 检查鉴权参数 token 非空
		if webCon.GetSessionToken() == "" {
			return false
		}

		// 根据 token 获取 session 信息
		session, err := webCon.App.GetSession(webCon.GetSessionToken())
		if err != nil {
			mlog.Error(fmt.Sprintf("Invalid session err=%v", err.Error()))

			// 若 session 已经失效（不存在），则重置 webCon 所绑定的 session 信息
			webCon.SetSessionToken("")
			webCon.SetSession(nil)
			webCon.SetSessionExpiresAt(0)
			return false
		}

		// 更新 session 信息
		webCon.SetSession(session)

		// 重置过期事件
		webCon.SetSessionExpiresAt(session.ExpiresAt)
	}

	return true
}

// 写入 hello 消息到 webCon.Send 管道
func (webCon *WebConn) SendHello() {
	msg := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_HELLO, "", "", webCon.UserId, nil)
	msg.Add("server_version", fmt.Sprintf("%v.%v.%v.%v", model.CurrentVersion, model.BuildNumber, webCon.App.ClientConfigHash(), webCon.App.License() != nil))
	webCon.Send <- msg
}


//
func (webCon *WebConn) ShouldSendEvent(msg *model.WebSocketEvent) bool {
	// IMPORTANT: Do not send event if WebConn does not have a session


	// 鉴权检查
	if !webCon.IsAuthenticated() {
		return false
	}


	// If the event contains sanitized data, only send to users that don't have permission to see sensitive data.
	// Prevents admin clients from receiving events with bad data.

	var hasReadPrivateDataPermission *bool


	if msg.Broadcast.ContainsSanitizedData {
		hasReadPrivateDataPermission = model.NewBool(webCon.App.RolesGrantPermission(webCon.GetSession().GetUserRoles(), model.PERMISSION_MANAGE_SYSTEM.Id))
		if *hasReadPrivateDataPermission {
			return false
		}
	}


	// If the event contains sensitive data, only send to users with permission to see it
	if msg.Broadcast.ContainsSensitiveData {
		if hasReadPrivateDataPermission == nil {
			hasReadPrivateDataPermission = model.NewBool(webCon.App.RolesGrantPermission(webCon.GetSession().GetUserRoles(), model.PERMISSION_MANAGE_SYSTEM.Id))
		}

		if !*hasReadPrivateDataPermission {
			return false
		}
	}

	// If the event is destined to a specific user
	if len(msg.Broadcast.UserId) > 0 {
		if webCon.UserId == msg.Broadcast.UserId {
			return true
		}
		return false
	}

	// if the user is omitted don't send the message
	if len(msg.Broadcast.OmitUsers) > 0 {
		if _, ok := msg.Broadcast.OmitUsers[webCon.UserId]; ok {
			return false
		}
	}

	// Only report events to users who are in the channel for the event
	if len(msg.Broadcast.ChannelId) > 0 {


		if model.GetMillis()-webCon.LastAllChannelMembersTime > WEBCONN_MEMBER_CACHE_TIME {
			webCon.AllChannelMembers = nil
			webCon.LastAllChannelMembersTime = 0
		}


		if webCon.AllChannelMembers == nil {
			result, err := webCon.App.Srv.Store.Channel().GetAllChannelMembersForUser(webCon.UserId, true, false)
			if err != nil {
				mlog.Error("webhub.shouldSendEvent: " + err.Error())
				return false
			}
			webCon.AllChannelMembers = result
			webCon.LastAllChannelMembersTime = model.GetMillis()
		}


		if _, ok := webCon.AllChannelMembers[msg.Broadcast.ChannelId]; ok {
			return true
		}
		return false
	}



	// Only report events to users who are in the team for the event
	if len(msg.Broadcast.TeamId) > 0 {
		return webCon.IsMemberOfTeam(msg.Broadcast.TeamId)
	}


	if msg.Event == model.WEBSOCKET_EVENT_USER_UPDATED && webCon.GetSession().Props[model.SESSION_PROP_IS_GUEST] == "true" {
		canSee, err := webCon.App.UserCanSeeOtherUser(webCon.UserId, msg.Data["user"].(*model.User).Id)
		if err != nil {
			mlog.Error("webhub.shouldSendEvent: " + err.Error())
			return false
		}
		return canSee
	}

	return true
}

func (webCon *WebConn) IsMemberOfTeam(teamId string) bool {


	// 获取 webCon 归属 session
	currentSession := webCon.GetSession()

	// session 不存在则重新获取
	if currentSession == nil || len(currentSession.Token) == 0 {
		session, err := webCon.App.GetSession(webCon.GetSessionToken())
		if err != nil {
			mlog.Error(fmt.Sprintf("Invalid session err=%v", err.Error()))
			return false
		}
		webCon.SetSession(session)
		currentSession = session
	}

	// 获取 session 下 teamId 对应的 member
	member := currentSession.GetTeamByTeamId(teamId)

	if member != nil {
		return true
	}
	return false
}

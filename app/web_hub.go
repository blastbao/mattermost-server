// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"
	"hash/fnv"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
)

const (
	BROADCAST_QUEUE_SIZE = 4096
	DEADLOCK_TICKER      = 15 * time.Second                  // check every 15 seconds
	DEADLOCK_WARN        = (BROADCAST_QUEUE_SIZE * 99) / 100 // number of buffered messages before printing stack trace
)

type WebConnActivityMessage struct {
	UserId       string
	SessionToken string
	ActivityAt   int64
}

type Hub struct {
	// connectionCount should be kept first.
	// See https://github.com/blastbao/mattermost-server/pull/7281
	connectionCount int64
	app             *App
	connectionIndex int

	register        chan *WebConn
	unregister      chan *WebConn

	broadcast       chan *model.WebSocketEvent

	// 触发关闭的管道
	stop            chan struct{}
	// 关闭完成的信号管道
	didStop         chan struct{}

	invalidateUser  chan string
	activity        chan *WebConnActivityMessage

	// 强制退出标识
	ExplicitStop    bool

	//
	goroutineId     int
}


func (a *App) NewWebHub() *Hub {
	return &Hub{
		app:            a,
		register:       make(chan *WebConn, 1),
		unregister:     make(chan *WebConn, 1),
		broadcast:      make(chan *model.WebSocketEvent, BROADCAST_QUEUE_SIZE),
		stop:           make(chan struct{}),
		didStop:        make(chan struct{}),
		invalidateUser: make(chan string),
		activity:       make(chan *WebConnActivityMessage),
		ExplicitStop:   false,
	}
}

func (a *App) TotalWebsocketConnections() int {
	count := int64(0)
	for _, hub := range a.Srv.Hubs {
		count = count + atomic.LoadInt64(&hub.connectionCount)
	}

	return int(count)
}

func (a *App) HubStart() {

	// Total number of hubs is twice the number of CPUs.
	numberOfHubs := runtime.NumCPU() * 2
	mlog.Info(fmt.Sprintf("Starting %v websocket hubs", numberOfHubs))

	a.Srv.Hubs = make([]*Hub, numberOfHubs)
	a.Srv.HubsStopCheckingForDeadlock = make(chan bool, 1)

	for i := 0; i < len(a.Srv.Hubs); i++ {
		a.Srv.Hubs[i] = a.NewWebHub()
		a.Srv.Hubs[i].connectionIndex = i
		a.Srv.Hubs[i].Start()
	}


	go func() {
		ticker := time.NewTicker(DEADLOCK_TICKER)

		defer func() {
			ticker.Stop()
		}()

		for {
			select {

			// 定时检查
			case <-ticker.C:

				// 遍历所有 hubs
				for _, hub := range a.Srv.Hubs {

					// 死锁预警？
					if len(hub.broadcast) >= DEADLOCK_WARN {

						mlog.Error(fmt.Sprintf("Hub processing might be deadlock on hub %v goroutine %v with %v events in the buffer", hub.connectionIndex, hub.goroutineId, len(hub.broadcast)))

						// 获取堆栈，追踪可能死锁的协程
						buf := make([]byte, 1<<16)
						runtime.Stack(buf, true)
						output := fmt.Sprintf("%s", buf)
						splits := strings.Split(output, "goroutine ")

						for _, part := range splits {
							if strings.Contains(part, fmt.Sprintf("%v", hub.goroutineId)) {
								mlog.Error(fmt.Sprintf("Trace for possible deadlock goroutine %v", part))
							}
						}
					}
				}
			// 退出信号
			case <-a.Srv.HubsStopCheckingForDeadlock:
				return
			}
		}
	}()
}

func (a *App) HubStop() {
	mlog.Info("stopping websocket hub connections")

	select {
	case a.Srv.HubsStopCheckingForDeadlock <- true:
	default:
		mlog.Warn("We appear to have already sent the stop checking for deadlocks command")
	}

	// 逐个停止 hubs
	for _, hub := range a.Srv.Hubs {
		hub.Stop()
	}

	// 清空
	a.Srv.Hubs = []*Hub{}
}



// 根据 userId 哈希值计算他属于哪个 hub
func (a *App) GetHubForUserId(userId string) *Hub {

	if len(a.Srv.Hubs) == 0 {
		return nil
	}

	// 计算 hash 值
	hash := fnv.New32a()
	hash.Write([]byte(userId))

	// idx = hash % len(Hubs)
	index := hash.Sum32() % uint32(len(a.Srv.Hubs))
	return a.Srv.Hubs[index]
}

func (a *App) HubRegister(webConn *WebConn) {

	// 根据 userId 哈希值计算 webConn 属于哪个 hub
	hub := a.GetHubForUserId(webConn.UserId)

	// 把 conn 注册到该 hub 中
	if hub != nil {
		hub.Register(webConn)
	}
}

func (a *App) HubUnregister(webConn *WebConn) {

	// 根据 userId 哈希值计算 webConn 属于哪个 hub
	hub := a.GetHubForUserId(webConn.UserId)

	// 把 conn 从 hub 中删除
	if hub != nil {
		hub.Unregister(webConn)
	}
}


func (a *App) Publish(message *model.WebSocketEvent) {


	// 统计上报
	if metrics := a.Metrics; metrics != nil {
		metrics.IncrementWebsocketEvent(message.Event)
	}

	// 广播消息：将 message 发送到每一个 hub 中
	a.PublishSkipClusterSend(message)

	// 发送集群消息
	if a.Cluster != nil {

		// 构造集群消息
		cm := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_PUBLISH,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     message.ToJson(),
		}

		// 确定消息的发送类型
		if  message.Event == model.WEBSOCKET_EVENT_POSTED ||
			message.Event == model.WEBSOCKET_EVENT_POST_EDITED ||
			message.Event == model.WEBSOCKET_EVENT_DIRECT_ADDED ||
			message.Event == model.WEBSOCKET_EVENT_GROUP_ADDED ||
			message.Event == model.WEBSOCKET_EVENT_ADDED_TO_TEAM {
			cm.SendType = model.CLUSTER_SEND_RELIABLE
		}

		// 发送集群消息
		a.Cluster.SendClusterMessage(cm)
	}


	// ...
}


// 发送广播消息
func (a *App) PublishSkipClusterSend(message *model.WebSocketEvent) {

	// 如果广播消息指定了发送的目标 user
	if message.Broadcast.UserId != "" {
		// 获取 user 所在的 hub，把消息 message 推送到该 hub 中
		hub := a.GetHubForUserId(message.Broadcast.UserId)
		if hub != nil {
			hub.Broadcast(message)
		}
	// 否则
	} else {
		// 把消息 message 发送到每一个 hub 中
		for _, hub := range a.Srv.Hubs {
			hub.Broadcast(message)
		}
	}
}

func (a *App) InvalidateCacheForChannel(channel *model.Channel) {
	a.InvalidateCacheForChannelSkipClusterSend(channel.Id)
	a.InvalidateCacheForChannelByNameSkipClusterSend(channel.TeamId, channel.Name)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     channel.Id,
		}

		a.Cluster.SendClusterMessage(msg)

		nameMsg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_BY_NAME,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Props:    make(map[string]string),
		}

		nameMsg.Props["name"] = channel.Name
		if channel.TeamId == "" {
			nameMsg.Props["id"] = "dm"
		} else {
			nameMsg.Props["id"] = channel.TeamId
		}

		a.Cluster.SendClusterMessage(nameMsg)
	}
}

func (a *App) InvalidateCacheForChannelSkipClusterSend(channelId string) {
	a.Srv.Store.Channel().InvalidateChannel(channelId)
}

func (a *App) InvalidateCacheForChannelMembers(channelId string) {
	a.InvalidateCacheForChannelMembersSkipClusterSend(channelId)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_MEMBERS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     channelId,
		}
		a.Cluster.SendClusterMessage(msg)
	}
}

func (a *App) InvalidateCacheForChannelMembersSkipClusterSend(channelId string) {
	a.Srv.Store.User().InvalidateProfilesInChannelCache(channelId)
	a.Srv.Store.Channel().InvalidateMemberCount(channelId)
	a.Srv.Store.Channel().InvalidateGuestCount(channelId)
}

func (a *App) InvalidateCacheForChannelMembersNotifyProps(channelId string) {
	a.InvalidateCacheForChannelMembersNotifyPropsSkipClusterSend(channelId)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_MEMBERS_NOTIFY_PROPS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     channelId,
		}
		a.Cluster.SendClusterMessage(msg)
	}
}

func (a *App) InvalidateCacheForChannelMembersNotifyPropsSkipClusterSend(channelId string) {
	a.Srv.Store.Channel().InvalidateCacheForChannelMembersNotifyProps(channelId)
}

func (a *App) InvalidateCacheForChannelByNameSkipClusterSend(teamId, name string) {
	if teamId == "" {
		teamId = "dm"
	}

	a.Srv.Store.Channel().InvalidateChannelByName(teamId, name)
}

func (a *App) InvalidateCacheForChannelPosts(channelId string) {
	a.InvalidateCacheForChannelPostsSkipClusterSend(channelId)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_POSTS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     channelId,
		}
		a.Cluster.SendClusterMessage(msg)
	}
}

func (a *App) InvalidateCacheForChannelPostsSkipClusterSend(channelId string) {
	a.Srv.Store.Post().InvalidateLastPostTimeCache(channelId)
	a.Srv.Store.Channel().InvalidatePinnedPostCount(channelId)
}

func (a *App) InvalidateCacheForUser(userId string) {
	a.InvalidateCacheForUserSkipClusterSend(userId)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_USER,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     userId,
		}
		a.Cluster.SendClusterMessage(msg)
	}
}

func (a *App) InvalidateCacheForUserTeams(userId string) {
	a.InvalidateCacheForUserTeamsSkipClusterSend(userId)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_USER_TEAMS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     userId,
		}
		a.Cluster.SendClusterMessage(msg)
	}
}

func (a *App) InvalidateCacheForUserSkipClusterSend(userId string) {
	a.Srv.Store.Channel().InvalidateAllChannelMembersForUser(userId)
	a.Srv.Store.User().InvalidateProfilesInChannelCacheByUser(userId)
	a.Srv.Store.User().InvalidatProfileCacheForUser(userId)

	hub := a.GetHubForUserId(userId)
	if hub != nil {
		hub.InvalidateUser(userId)
	}
}

func (a *App) InvalidateCacheForUserTeamsSkipClusterSend(userId string) {
	a.Srv.Store.Team().InvalidateAllTeamIdsForUser(userId)

	hub := a.GetHubForUserId(userId)
	if hub != nil {
		hub.InvalidateUser(userId)
	}
}

func (a *App) InvalidateCacheForWebhook(webhookId string) {
	a.InvalidateCacheForWebhookSkipClusterSend(webhookId)

	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_WEBHOOK,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     webhookId,
		}
		a.Cluster.SendClusterMessage(msg)
	}
}

func (a *App) InvalidateCacheForWebhookSkipClusterSend(webhookId string) {
	a.Srv.Store.Webhook().InvalidateWebhookCache(webhookId)
}

func (a *App) InvalidateWebConnSessionCacheForUser(userId string) {

	// 根据 userId 哈希值计算他属于哪个 hub
	hub := a.GetHubForUserId(userId)

	// 使用户登陆态失效
	if hub != nil {
		hub.InvalidateUser(userId)
	}
}

// 更新用户连接活跃时间
func (a *App) UpdateWebConnUserActivity(session model.Session, activityAt int64) {
	hub := a.GetHubForUserId(session.UserId)
	if hub != nil {
		hub.UpdateActivity(session.UserId, session.Token, activityAt)
	}
}

// 注册连接
func (h *Hub) Register(webConn *WebConn) {
	select {
	case h.register <- webConn:
	case <-h.didStop:
	}

	if webConn.IsAuthenticated() {
		webConn.SendHello()
	}
}

// 注销连接
func (h *Hub) Unregister(webConn *WebConn) {
	select {
	case h.unregister <- webConn:
	case <-h.stop:
	}
}

// 广播消息
func (h *Hub) Broadcast(message *model.WebSocketEvent) {

	if h != nil && h.broadcast != nil && message != nil {
		select {
		case h.broadcast <- message: //把广播消息写入管道，Start() 协程中会接收和处理。
		case <-h.didStop:
		}
	}
}

// 使用户登陆态失效
func (h *Hub) InvalidateUser(userId string) {
	select {
	case h.invalidateUser <- userId:
	case <-h.didStop:
	}
}

// 更新连接活跃时间
func (h *Hub) UpdateActivity(userId, sessionToken string, activityAt int64) {
	select {
	case h.activity <- &WebConnActivityMessage{UserId: userId, SessionToken: sessionToken, ActivityAt: activityAt}:
	case <-h.didStop:
	}
}

// 获取协程 ID
func getGoroutineId() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		id = -1
	}
	return id
}

func (h *Hub) Stop() {
	// 发送关闭信号，触发 Start 协程退出
	close(h.stop)
	// 阻塞等待 Start 协程完成退出
	<-h.didStop
}

func (h *Hub) Start() {
	var doStart func()
	var doRecoverableStart func()
	var doRecover func()

	//1. 注册新连接
	//2. 注销连接
	//3. 用户登陆态失效
	//4. 更新 user 最近活跃时间
	//5. 广播消息
	//6. 监听退出信号

	doStart = func() {

		// 获取当前协程的 goroutine ID
		h.goroutineId = getGoroutineId()
		mlog.Debug(fmt.Sprintf("Hub for index %v is starting with goroutine %v", h.connectionIndex, h.goroutineId))

		// 创建 connections 存储所有的连接
		connections := newHubConnectionIndex()

		for {
			select {

			// 注册新连接
			case webCon := <-h.register:
				// 添加 webCon 到 connections 中
				connections.Add(webCon)
				// 更新当前连接总数
				atomic.StoreInt64(&h.connectionCount, int64(len(connections.All())))

			// 注销连接
			case webCon := <-h.unregister:

				// 从 connections 中移除 webCon
				connections.Remove(webCon)

				// 更新当前连接总数
				atomic.StoreInt64(&h.connectionCount, int64(len(connections.All())))

				// 检查 userId 是否为空
				if len(webCon.UserId) == 0 {
					continue
				}

				// 获取 userId 的所有连接 []*WebConn
				conns := connections.ForUser(webCon.UserId)

				// 如果此时 userId 不存在任何有效连接
				if len(conns) == 0 {
					// 设置该 user 为下线状态
					h.app.Srv.Go(func() {
						h.app.SetStatusOffline(webCon.UserId, false)
					})

				// 如果此时 userId 仍有其它连接存在
				} else {

					// 遍历所有连接，得到该 user 最近活跃的时间
					var latestActivity int64 = 0
					for _, conn := range conns {
						if conn.LastUserActivityAt > latestActivity {
							latestActivity = conn.LastUserActivityAt
						}
					}

					// 如果最近活跃的时间已经超过了在线阈值，则可能已经离开了，设置该 user 为离开状态 =》类似于 skype 的 "上一次活跃是在XXX时间"
					if h.app.IsUserAway(latestActivity) {
						h.app.Srv.Go(func() {
							h.app.SetStatusLastActivityAt(webCon.UserId, latestActivity)
						})
					}

				}

			// 用户登陆态失效
			case userId := <-h.invalidateUser:
				// 遍历 userId 的所有连接，重置每个连接的缓存信息。
				for _, webCon := range connections.ForUser(userId) {
					webCon.InvalidateCache()
				}

			// 更新 user 最近活跃时间
			case activity := <-h.activity:
				// 遍历 userId 的所有连接，得到事件对应的连接，更新该连接的最近活跃时间
				for _, webCon := range connections.ForUser(activity.UserId) {
					if webCon.GetSessionToken() == activity.SessionToken {
						webCon.LastUserActivityAt = activity.ActivityAt
					}
				}

			// 广播消息
			case msg := <-h.broadcast:

				// 获取所有连接
				candidates := connections.All()

				// 检查是否指定了发送的目标用户
				if msg.Broadcast.UserId != "" {
					candidates = connections.ForUser(msg.Broadcast.UserId)
				}

				// 提前将 msg 序列化
				msg.PrecomputeJSON()

				// 遍历所有连接 webCons 逐个将消息发送给每个连接
				for _, webCon := range candidates {
					// 检查消息应不应该被发送给当前连接 webCon（鉴黄，敏感信息...）
					if webCon.ShouldSendEvent(msg) {
						//【重要】将消息写入到 webCon.Send 待发送管道: 写入失败的原因是管道满。
						select {
						case webCon.Send <- msg:
						default:
							mlog.Error(fmt.Sprintf("webhub.broadcast: cannot send, closing websocket for userId=%v", webCon.UserId))
							// 关闭管道，注意，关闭一个已经关闭的管道会 panic，所有这里 webCon.Send 不可能已经是关闭的
							close(webCon.Send)
							// 从 webCons 集合中删除当前 webCon 连接，不再发消息给他
							connections.Remove(webCon)
						}
					}
				}

			// 退出监听
			case <-h.stop:

				// 遍历所有的 webConns 连接，记录每个连接的 userIds
				userIds := make(map[string]bool)
				for _, webCon := range connections.All() {
					userIds[webCon.UserId] = true // 保存 userId 到 userIds 中，这里使用 map 是为了去重
					webCon.Close() // 关闭连接
				}

				// 设置 users 为下线状态
				for userId := range userIds {
					h.app.SetStatusOffline(userId, false)
				}

				// 设置关闭标识
				h.ExplicitStop = true

				// 发送关闭完成的信号
				close(h.didStop)

				return
			}
		}
	}




	doRecoverableStart = func() {
		defer doRecover()
		doStart()
	}

	doRecover = func() {
		// 如果不是强制退出，则发生异常时需要重启任务
		if !h.ExplicitStop {
			// 异常捕获
			if r := recover(); r != nil {
				mlog.Error(fmt.Sprintf("Recovering from Hub panic. Panic was: %v", r))
			} else {
				mlog.Error("Webhub stopped unexpectedly. Recovering.")
			}

			mlog.Error(string(debug.Stack()))
			// 重新调用 doRecoverableStart
			go doRecoverableStart()
		}
	}

	// 启动 doStart()
	go doRecoverableStart()
}





// 索引结构
type hubConnectionIndexIndexes struct {
	connections         int
	connectionsByUserId int
}

// hubConnectionIndex provides fast addition, removal, and iteration of web connections.
type hubConnectionIndex struct {

	// 所有连接
	connections         []*WebConn

	// 按 userId 做索引
	connectionsByUserId map[string][]*WebConn

	// 保存 WebConn 在上面两个数组中的下标
	connectionIndexes   map[*WebConn]*hubConnectionIndexIndexes
}



func newHubConnectionIndex() *hubConnectionIndex {
	return &hubConnectionIndex{
		connections:         make([]*WebConn, 0, model.SESSION_CACHE_SIZE),
		connectionsByUserId: make(map[string][]*WebConn),
		connectionIndexes:   make(map[*WebConn]*hubConnectionIndexIndexes),
	}
}

func (i *hubConnectionIndex) Add(wc *WebConn) {

	// 添加新连接 wc 到 connections 数组
	i.connections = append(i.connections, wc)

	// 添加新连接 wc 到 UserId 的连接数组
	i.connectionsByUserId[wc.UserId] = append(i.connectionsByUserId[wc.UserId], wc)

	// 保存 wc 在上面两个数组中的索引下标
	i.connectionIndexes[wc] = &hubConnectionIndexIndexes{
		connections:         len(i.connections) - 1, 					// connections 数组下标
		connectionsByUserId: len(i.connectionsByUserId[wc.UserId]) - 1, // user 连接数组下标
	}
}

func (i *hubConnectionIndex) Remove(wc *WebConn) {

	// 检查 wc 是否存在 ？若存在则获取对应的存储下标。
	indexes, ok := i.connectionIndexes[wc]
	if !ok {
		return
	}

	// 取末尾元素
	last := i.connections[len(i.connections)-1]

	// 把末尾元素移动到 wc 的位置，需要更新关联的三个数据结构

	// 1
	i.connections[indexes.connections] = last
	i.connections = i.connections[:len(i.connections)-1]
	i.connectionIndexes[last].connections = indexes.connections

	// 2
	userConnections := i.connectionsByUserId[wc.UserId]
	last = userConnections[len(userConnections)-1]
	userConnections[indexes.connectionsByUserId] = last
	i.connectionsByUserId[wc.UserId] = userConnections[:len(userConnections)-1]

	// 3
	i.connectionIndexes[last].connectionsByUserId = indexes.connectionsByUserId

	// 从 i.connectionIndexes 中删除当前连接
	delete(i.connectionIndexes, wc)
}

// 获取 userId 的所有连接 []*WebConn
func (i *hubConnectionIndex) ForUser(id string) []*WebConn {
	return i.connectionsByUserId[id]
}

// 返回所有连接
func (i *hubConnectionIndex) All() []*WebConn {
	return i.connections
}

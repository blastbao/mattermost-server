// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	SOCKET_MAX_MESSAGE_SIZE_KB  = 8 * 1024  // 8KB
	PING_TIMEOUT_BUFFER_SECONDS = 5 		//
)

type WebSocketClient struct {
	Url                string          // The location of the server like "ws://localhost:8065"
	ApiUrl             string          // The api location of the server like "ws://localhost:8065/api/v3"
	ConnectUrl         string          // The websocket URL to connect to like "ws://localhost:8065/api/v3/path/to/websocket"
	Conn               *websocket.Conn // The WebSocket connection
	AuthToken          string          // The token used to open the WebSocket
	// 不断自增的序号
	Sequence           int64           // The ever-incrementing sequence attached to each WebSocket action
	PingTimeoutChannel chan bool       // The channel used to signal ping timeouts

	// inChan
	EventChannel       chan *WebSocketEvent
	// outChan
	ResponseChannel    chan *WebSocketResponse

	ListenError        *AppError
	pingTimeoutTimer   *time.Timer
}

// NewWebSocketClient constructs a new WebSocket client with convenience
// methods for talking to the server.
func NewWebSocketClient(url, authToken string) (*WebSocketClient, *AppError) {
	return NewWebSocketClientWithDialer(websocket.DefaultDialer, url, authToken)
}

// NewWebSocketClientWithDialer constructs a new WebSocket client with convenience
// methods for talking to the server using a custom dialer.
func NewWebSocketClientWithDialer(dialer *websocket.Dialer, url, authToken string) (*WebSocketClient, *AppError) {
	// 创建 websocket 连接
	conn, _, err := dialer.Dial(url+API_URL_SUFFIX+"/websocket", nil)
	if err != nil {
		return nil, NewAppError("NewWebSocketClient", "model.websocket_client.connect_fail.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	// 构造 client
	client := &WebSocketClient{
		url,
		url + API_URL_SUFFIX,
		url + API_URL_SUFFIX + "/websocket",
		conn,
		authToken,
		1,
		make(chan bool, 1),
		make(chan *WebSocketEvent, 100),
		make(chan *WebSocketResponse, 100),
		nil,
		nil,
	}
	// 设置 PingHandler 函数，当 server 发送 "ping" 时会自动调用
	client.configurePingHandling()
	// 发送 token 消息
	client.SendMessage(WEBSOCKET_AUTHENTICATION_CHALLENGE, map[string]interface{}{"token": authToken})
	return client, nil
}

// NewWebSocketClient4 constructs a new WebSocket client with convenience
// methods for talking to the server. Uses the v4 endpoint.
func NewWebSocketClient4(url, authToken string) (*WebSocketClient, *AppError) {
	return NewWebSocketClient4WithDialer(websocket.DefaultDialer, url, authToken)
}

// NewWebSocketClient4WithDialer constructs a new WebSocket client with convenience
// methods for talking to the server using a custom dialer. Uses the v4 endpoint.
func NewWebSocketClient4WithDialer(dialer *websocket.Dialer, url, authToken string) (*WebSocketClient, *AppError) {
	return NewWebSocketClientWithDialer(dialer, url, authToken)
}

func (wsc *WebSocketClient) Connect() *AppError {
	return wsc.ConnectWithDialer(websocket.DefaultDialer)
}

func (wsc *WebSocketClient) ConnectWithDialer(dialer *websocket.Dialer) *AppError {
	// 建立连接
	var err error
	wsc.Conn, _, err = dialer.Dial(wsc.ConnectUrl, nil)
	if err != nil {
		return NewAppError("Connect", "model.websocket_client.connect_fail.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	// 配置 PingPong
	wsc.configurePingHandling()
	// 初始化消息读取/发送的管道
	wsc.EventChannel = make(chan *WebSocketEvent, 100)
	wsc.ResponseChannel = make(chan *WebSocketResponse, 100)
	// 发送 auth 消息给对端（server）
	wsc.SendMessage(WEBSOCKET_AUTHENTICATION_CHALLENGE, map[string]interface{}{"token": wsc.AuthToken})
	return nil
}

func (wsc *WebSocketClient) Close() {
	wsc.Conn.Close()
}

func (wsc *WebSocketClient) Listen() {
	go func() {
		defer func() {
			wsc.Conn.Close()
			close(wsc.EventChannel)
			close(wsc.ResponseChannel)
		}()

		for {

			// 读取消息
			var rawMsg json.RawMessage
			var err error
			if _, rawMsg, err = wsc.Conn.ReadMessage(); err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
					wsc.ListenError = NewAppError("NewWebSocketClient", "model.websocket_client.connect_fail.app_error", nil, err.Error(), http.StatusInternalServerError)
				}

				return
			}

			// 解析消息
			var event WebSocketEvent
			if err := json.Unmarshal(rawMsg, &event); err == nil && event.IsValid() {
				wsc.EventChannel <- &event
				continue
			}

			// 构造返回消息，写入响应管道
			var response WebSocketResponse
			if err := json.Unmarshal(rawMsg, &response); err == nil && response.IsValid() {
				wsc.ResponseChannel <- &response
				continue
			}

		}
	}()
}

func (wsc *WebSocketClient) SendMessage(action string, data map[string]interface{}) {
	// 构造请求
	req := &WebSocketRequest{}
	req.Seq = wsc.Sequence
	req.Action = action
	req.Data = data
	// 序号自增
	wsc.Sequence++
	// 发送消息
	wsc.Conn.WriteJSON(req)
}

// UserTyping will push a user_typing event out to all connected users
// who are in the specified channel
//
// UserTyping 将向指定 channel 中的所有已连接用户推送一个 user_typing 事件。
//
func (wsc *WebSocketClient) UserTyping(channelId, parentId string) {
	data := map[string]interface{}{
		"channel_id": channelId,
		"parent_id":  parentId,
	}
	wsc.SendMessage("user_typing", data)
}

// GetStatuses will return a map of string statuses using user id as the key
//
// GetStatuses 将返回一个以 userId 为 key 的 map[string]*status 对象。
//
func (wsc *WebSocketClient) GetStatuses() {
	wsc.SendMessage("get_statuses", nil)
}


// GetStatusesByIds will fetch certain user statuses based on ids and return
// a map of string statuses using user id as the key
//
//
func (wsc *WebSocketClient) GetStatusesByIds(userIds []string) {
	data := map[string]interface{}{
		"user_ids": userIds,
	}
	wsc.SendMessage("get_statuses_by_ids", data)
}


func (wsc *WebSocketClient) configurePingHandling() {
	wsc.Conn.SetPingHandler(wsc.pingHandler)
	wsc.pingTimeoutTimer = time.NewTimer(time.Second * (60 + PING_TIMEOUT_BUFFER_SECONDS))
	go wsc.pingWatchdog()
}


// 每当收到 "ping" 会自动调用
func (wsc *WebSocketClient) pingHandler(appData string) error {
	// 停止超时计时器
	if !wsc.pingTimeoutTimer.Stop() {
		<-wsc.pingTimeoutTimer.C
	}
	// 重置超时计时器（60+5秒）
	wsc.pingTimeoutTimer.Reset(time.Second * (60 + PING_TIMEOUT_BUFFER_SECONDS))
	// 回复 pong
	wsc.Conn.WriteMessage(websocket.PongMessage, []byte{})
	return nil
}


func (wsc *WebSocketClient) pingWatchdog() {
	<-wsc.pingTimeoutTimer.C
	wsc.PingTimeoutChannel <- true
}

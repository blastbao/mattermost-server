// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"
	"net/http"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
	"github.com/blastbao/mattermost-server/utils"
)

type webSocketHandler interface {
	ServeWebSocket(*WebConn, *model.WebSocketRequest)
}

type WebSocketRouter struct {
	app      *App
	handlers map[string]webSocketHandler
}

// 注册
func (wr *WebSocketRouter) Handle(action string, handler webSocketHandler) {
	wr.handlers[action] = handler
}

// 请求处理
func (wr *WebSocketRouter) ServeWebSocket(conn *WebConn, r *model.WebSocketRequest) {

	// 检查 action 合法性
	if r.Action == "" {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.no_action.app_error", nil, "", http.StatusBadRequest)
		ReturnWebSocketError(conn, r, err)
		return
	}

	// 检查 seq 序号合法性
	if r.Seq <= 0 {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.bad_seq.app_error", nil, "", http.StatusBadRequest)
		ReturnWebSocketError(conn, r, err)
		return
	}

	// ???
	if r.Action == model.WEBSOCKET_AUTHENTICATION_CHALLENGE {

		if conn.GetSessionToken() != "" {
			return
		}

		token, ok := r.Data["token"].(string)
		if !ok {
			conn.WebSocket.Close()
			return
		}

		session, err := wr.app.GetSession(token)
		if err != nil {
			conn.WebSocket.Close()
			return
		}

		wr.app.Srv.Go(func() {
			wr.app.SetStatusOnline(session.UserId, false)
			wr.app.UpdateLastActivityAtIfNeeded(*session)
		})

		conn.SetSession(session)
		conn.SetSessionToken(session.Token)
		conn.UserId = session.UserId

		wr.app.HubRegister(conn)

		resp := model.NewWebSocketResponse(model.STATUS_OK, r.Seq, nil)
		conn.Send <- resp

		return
	}

	// 检查是否通过鉴权
	if !conn.IsAuthenticated() {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.not_authenticated.app_error", nil, "", http.StatusUnauthorized)
		ReturnWebSocketError(conn, r, err)
		return
	}

	// 根据 action 查找已注册的 handler
	handler, ok := wr.handlers[r.Action]
	if !ok {
		err := model.NewAppError("ServeWebSocket", "api.web_socket_router.bad_action.app_error", nil, "", http.StatusInternalServerError)
		ReturnWebSocketError(conn, r, err)
		return
	}

	// 执行 handler 处理请求 r
	handler.ServeWebSocket(conn, r)
}

func ReturnWebSocketError(conn *WebConn, r *model.WebSocketRequest, err *model.AppError) {

	mlog.Error(fmt.Sprintf("websocket routing error: seq=%v uid=%v %v [details: %v]", r.Seq, conn.UserId, err.SystemMessage(utils.T), err.DetailedError))

	err.DetailedError = ""
	errorResp := model.NewWebSocketError(r.Seq, err)

	conn.Send <- errorResp
}

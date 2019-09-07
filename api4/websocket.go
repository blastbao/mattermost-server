// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
)

func (api *API) InitWebSocket() {
	api.BaseRoutes.ApiRoot.Handle("/websocket", api.ApiHandlerTrustRequester(connectWebSocket)).Methods("GET")
}

func connectWebSocket(c *Context, w http.ResponseWriter, r *http.Request) {

	// 协议升级
	upgrader := websocket.Upgrader{
		ReadBufferSize:  model.SOCKET_MAX_MESSAGE_SIZE_KB,
		WriteBufferSize: model.SOCKET_MAX_MESSAGE_SIZE_KB,
		CheckOrigin:     c.App.OriginChecker(),
	}

	// 协议升级
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		mlog.Error(fmt.Sprintf("websocket connect err: %v", err))
		c.Err = model.NewAppError("connect", "api.web_socket.connect.upgrade.app_error", nil, "", http.StatusInternalServerError)
		return
	}

	// 构造 conn 结构体
	wc := c.App.NewWebConn(ws, c.App.Session, c.App.T, "")

	// 把连接注册到 Hub 中
	if len(c.App.Session.UserId) > 0 {
		c.App.HubRegister(wc)
	}

	// 开启 wc 的消息读、处理、写等协程
	wc.Pump()
}

// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package wsapi

import (
	"fmt"
	"net/http"

	"github.com/blastbao/mattermost-server/app"
	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
	"github.com/blastbao/mattermost-server/utils"
)

func (api *API) ApiWebSocketHandler(wh func(*model.WebSocketRequest) (map[string]interface{}, *model.AppError)) webSocketHandler {
	return webSocketHandler{api.App, wh}
}

type webSocketHandler struct {
	app         *app.App
	handlerFunc func(*model.WebSocketRequest) (map[string]interface{}, *model.AppError)
}

// 请求处理主函数
func (wh webSocketHandler) ServeWebSocket(conn *app.WebConn, r *model.WebSocketRequest) {

	mlog.Debug(fmt.Sprintf("websocket: %s", r.Action))

	// 根据 token 获取 session
	session, sessionErr := wh.app.GetSession(conn.GetSessionToken())
	if sessionErr != nil {
		mlog.Error(fmt.Sprintf("%v:%v seq=%v uid=%v %v [details: %v]", "websocket", r.Action, r.Seq, conn.UserId, sessionErr.SystemMessage(utils.T), sessionErr.Error()))
		sessionErr.DetailedError = ""
		errResp := model.NewWebSocketError(r.Seq, sessionErr)
		// 回包
		conn.Send <- errResp
		return
	}

	// 填充 request
	r.Session = *session
	r.Locale = conn.Locale
	r.T = conn.T

	// 处理请求
	var data map[string]interface{}
	var err *model.AppError
	if data, err = wh.handlerFunc(r); err != nil {
		mlog.Error(fmt.Sprintf("%v:%v seq=%v uid=%v %v [details: %v]", "websocket", r.Action, r.Seq, r.Session.UserId, err.SystemMessage(utils.T), err.DetailedError))
		err.DetailedError = ""
		errResp := model.NewWebSocketError(r.Seq, err)
		conn.Send <- errResp
		return
	}

	// 构造 response：需要填充 r.Seq 来指明回复的哪个请求
	resp := model.NewWebSocketResponse(model.STATUS_OK, r.Seq, data)
	conn.Send <- resp // 回包
}

func NewInvalidWebSocketParamError(action string, name string) *model.AppError {
	return model.NewAppError("websocket: "+action, "api.websocket_handler.invalid_param.app_error", map[string]interface{}{"Name": name}, "", http.StatusBadRequest)
}

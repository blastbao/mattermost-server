// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package wsapi

import (
	"github.com/blastbao/mattermost-server/model"
)

// 注册处理函数
func (api *API) InitUser() {
	api.Router.Handle("user_typing", api.ApiWebSocketHandler(api.userTyping))
	api.Router.Handle("user_update_active_status", api.ApiWebSocketHandler(api.userUpdateActiveStatus))
}


// 用户正在输入中
func (api *API) userTyping(req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var channelId string

	// 从 req.Data 中取出 "channel_id"、"parent_id" 参数，用来构造广播事件
	if channelId, ok = req.Data["channel_id"].(string); !ok || len(channelId) != 26 {
		return nil, NewInvalidWebSocketParamError(req.Action, "channel_id")
	}
	var parentId string
	if parentId, ok = req.Data["parent_id"].(string); !ok {
		parentId = ""
	}

	// 把 UserId 添加到 `忽略 users 列表` 中，他将不会收到自己触发的本消息
	omitUsers := make(map[string]bool, 1)
	omitUsers[req.Session.UserId] = true

	// 构造 `用户正在输入中` 的事件，并添加附加参数
	event := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_TYPING, "", channelId, "", omitUsers)
	event.Add("parent_id", parentId)
	event.Add("user_id", req.Session.UserId)

	// 执行广播
	api.App.Publish(event)

	return nil, nil
}


// 用户更新活跃状态
func (api *API) userUpdateActiveStatus(req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {

	var ok bool
	// 是否为 `活跃状态`
	var userIsActive bool
	if userIsActive, ok = req.Data["user_is_active"].(bool); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "user_is_active")
	}
	// 是否为 `手动设置状态`
	var manual bool
	if manual, ok = req.Data["manual"].(bool); !ok {
		manual = false
	}

	if userIsActive {
		// 设置用户在线状态
		api.App.SetStatusOnline(req.Session.UserId, manual)
	} else {
		// 设置用户离线状态
		api.App.SetStatusAwayIfNeeded(req.Session.UserId, manual)
	}

	return nil, nil
}

// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package wsapi

import (
	"github.com/blastbao/mattermost-server/model"
)

func (api *API) InitSystem() {
	api.Router.Handle("ping", api.ApiWebSocketHandler(ping))
}

// 收到 `ping` 时回复 `pong`，并附带一组服务状态信息。
func ping(req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	data := map[string]interface{}{}
	data["text"] = "pong"
	data["version"] = model.CurrentVersion
	data["server_time"] = model.GetMillis()
	data["node_id"] = ""
	return data, nil
}

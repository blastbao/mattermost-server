// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package wsapi

import (
	"github.com/blastbao/mattermost-server/app"
)

type API struct {
	App    *app.App
	Router *app.WebSocketRouter
}

func Init(a *app.App, router *app.WebSocketRouter) {
	api := &API{
		App:    a,
		Router: router,
	}

	api.InitUser()
	api.InitSystem()
	api.InitStatus()

	// 创建&启动 n 个 hubs ， n 默认为 cpu 核数的两倍
	a.HubStart()
}

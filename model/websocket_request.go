// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"

	goi18n "github.com/mattermost/go-i18n/i18n"
)

type WebSocketRequest struct {
	// Client-provided fields
	Seq    int64                  `json:"seq"`		// 序号，回包会带回此序号
	Action string                 `json:"action"`	// 方法(method)
	Data   map[string]interface{} `json:"data"`		// 参数

	// Server-provided fields
	Session Session              `json:"-"`			//
	T       goi18n.TranslateFunc `json:"-"`			//
	Locale  string               `json:"-"`			//
}

func (o *WebSocketRequest) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func WebSocketRequestFromJson(data io.Reader) *WebSocketRequest {
	var o *WebSocketRequest
	json.NewDecoder(data).Decode(&o)
	return o
}

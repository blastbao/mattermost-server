// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storetest

import (
	"github.com/blastbao/mattermost-server/model"
)

func MakeEmail() string {
	return "success_" + model.NewId() + "@simulator.amazonses.com"
}

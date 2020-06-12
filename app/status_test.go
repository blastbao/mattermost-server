// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"testing"

	"github.com/blastbao/mattermost-server/model"
)

func TestSaveStatus(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user := th.BasicUser

	for _, statusString := range []string{
		model.STATUS_ONLINE, 	// 在线
		model.STATUS_AWAY,		// 离开
		model.STATUS_DND,		// 不要打扰
		model.STATUS_OFFLINE,	// 离线
	} {
		t.Run(statusString, func(t *testing.T) {

			status := &model.Status{
				UserId: user.Id,
				Status: statusString,
			}

			th.App.SaveAndBroadcastStatus(status)

			after, err := th.App.GetStatus(user.Id)
			if err != nil {
				t.Fatalf("failed to get status after save: %v", err)
			} else if after.Status != statusString {
				t.Fatalf("failed to save status, got %v, expected %v", after.Status, statusString)
			}
		})
	}
}

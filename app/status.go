// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
	"github.com/blastbao/mattermost-server/utils"
)

var statusCache *utils.Cache = utils.NewLru(model.STATUS_CACHE_SIZE)

func ClearStatusCache() {
	statusCache.Purge()
}


// 保存状态到 Cache 中
func (a *App) AddStatusCacheSkipClusterSend(status *model.Status) {
	statusCache.Add(status.UserId, status)
}

// 保存状态到 Cache 中
func (a *App) AddStatusCache(status *model.Status) {

	// 保存状态到 Cache 中
	a.AddStatusCacheSkipClusterSend(status)

	// ???
	if a.Cluster != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_UPDATE_STATUS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     status.ToClusterJson(),
		}
		a.Cluster.SendClusterMessage(msg)
	}

}

func (a *App) GetAllStatuses() map[string]*model.Status {
	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return map[string]*model.Status{}
	}

	userIds := statusCache.Keys()
	statusMap := map[string]*model.Status{}

	for _, userId := range userIds {
		if id, ok := userId.(string); ok {
			status := GetStatusFromCache(id)
			if status != nil {
				statusMap[id] = status
			}
		}
	}

	return statusMap
}

func (a *App) GetStatusesByIds(userIds []string) (map[string]interface{}, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return map[string]interface{}{}, nil
	}

	statusMap := map[string]interface{}{}
	metrics := a.Metrics

	missingUserIds := []string{}
	for _, userId := range userIds {
		if result, ok := statusCache.Get(userId); ok {
			statusMap[userId] = result.(*model.Status).Status
			if metrics != nil {
				metrics.IncrementMemCacheHitCounter("Status")
			}
		} else {
			missingUserIds = append(missingUserIds, userId)
			if metrics != nil {
				metrics.IncrementMemCacheMissCounter("Status")
			}
		}
	}

	if len(missingUserIds) > 0 {
		statuses, err := a.Srv.Store.Status().GetByIds(missingUserIds)
		if err != nil {
			return nil, err
		}

		for _, s := range statuses {
			a.AddStatusCacheSkipClusterSend(s)
			statusMap[s.UserId] = s.Status
		}

	}

	// For the case where the user does not have a row in the Status table and cache
	for _, userId := range missingUserIds {
		if _, ok := statusMap[userId]; !ok {
			statusMap[userId] = model.STATUS_OFFLINE
		}
	}

	return statusMap, nil
}

// GetUserStatusesByIds used by apiV4
func (a *App) GetUserStatusesByIds(userIds []string) ([]*model.Status, *model.AppError) {


	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return []*model.Status{}, nil
	}

	var statusMap []*model.Status
	metrics := a.Metrics

	missingUserIds := []string{}
	for _, userId := range userIds {
		if result, ok := statusCache.Get(userId); ok {
			statusMap = append(statusMap, result.(*model.Status))
			if metrics != nil {
				metrics.IncrementMemCacheHitCounter("Status")
			}
		} else {
			missingUserIds = append(missingUserIds, userId)
			if metrics != nil {
				metrics.IncrementMemCacheMissCounter("Status")
			}
		}
	}

	if len(missingUserIds) > 0 {
		statuses, err := a.Srv.Store.Status().GetByIds(missingUserIds)
		if err != nil {
			return nil, err
		}

		for _, s := range statuses {
			a.AddStatusCacheSkipClusterSend(s)
		}

		statusMap = append(statusMap, statuses...)

	}

	// For the case where the user does not have a row in the Status table and cache
	// remove the existing ids from missingUserIds and then create a offline state for the missing ones
	// This also return the status offline for the non-existing Ids in the system
	for i := 0; i < len(missingUserIds); i++ {
		missingUserId := missingUserIds[i]
		for _, userMap := range statusMap {
			if missingUserId == userMap.UserId {
				missingUserIds = append(missingUserIds[:i], missingUserIds[i+1:]...)
				i--
				break
			}
		}
	}
	for _, userId := range missingUserIds {
		statusMap = append(statusMap, &model.Status{UserId: userId, Status: "offline"})
	}

	return statusMap, nil
}

// SetStatusLastActivityAt sets the last activity at for a user on the local app server and updates
// status to away if needed. Used by the WS to set status to away if an 'online' device disconnects
// while an 'away' device is still connected
//
// SetStatusLastActivityAt 设置用户在本地应用服务器上的最后一次活动，并在需要时将状态更新为 away 。
// 如果 "online" 设备断开连接，而 "away" 设备仍在连接，WS 将使用此功能将状态设置为 "away" 。
//
func (a *App) SetStatusLastActivityAt(userId string, activityAt int64) {
	var status *model.Status
	var err *model.AppError
	if status, err = a.GetStatus(userId); err != nil {
		return
	}
	// 更新上次活跃的时间戳
	status.LastActivityAt = activityAt
	// 保存状态到 Cache 中
	a.AddStatusCacheSkipClusterSend(status)
	// 设置为 "away" 状态（如果满足条件的话）
	a.SetStatusAwayIfNeeded(userId, false)
}

// 在线
func (a *App) SetStatusOnline(userId string, manual bool) {


	// 是否允许显示用户状态，若否则直接返回
	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return
	}

	var (
		// 如果 user 是从 `离线` 变成 `在线`，则需要广播
		broadcast = false

		// 旧状态
		oldStatus = model.STATUS_OFFLINE
		oldTime int64
		oldManual bool

		// 新状态
		status *model.Status

		// 错误信息
		err *model.AppError
	)

	// 查询用户状态
	//
	// 如果查询出错，则默认用户此前为离线状态，此时变为在线状态。
	if status, err = a.GetStatus(userId); err != nil {
		status = &model.Status{
			UserId: userId,
			Status: model.STATUS_ONLINE,
			Manual: false,
			LastActivityAt: model.GetMillis(),
			ActiveChannel: "",
		}
		broadcast = true

	} else {

		// 若目前状态是之前手动设置的，而本次设置非手动行为，则忽略本次行为。
		if status.Manual && !manual {
			// manually set status always overrides non-manual one
			return
		}

		// 如果是从 `离线` 变成 `在线`，则需要广播
 		if status.Status != model.STATUS_ONLINE {
			broadcast = true
		}

		// 保存旧的状态
		oldStatus = status.Status
		oldTime = status.LastActivityAt
		oldManual = status.Manual

		// 设置新的状态
		status.Status = model.STATUS_ONLINE
		status.Manual = false // for "online" there's no manual setting
		status.LastActivityAt = model.GetMillis()
	}

	// 保存新状态到 Cache 中
	a.AddStatusCache(status)


	// Only update the database if the status has changed,
	// the status has been manually set, or enough time has passed since the previous action
	//
	//
	// 只有在 状态发生变化、状态被手动设置 或 自上次操作以来已超过一定时间(2min)时，才更新数据库。
	if status.Status != oldStatus || status.Manual != oldManual || status.LastActivityAt-oldTime > model.STATUS_MIN_UPDATE_TIME {
		// 如果是从 `离线` 变成 `在线`，则更新状态到 Store 中
		if broadcast {
			if err := a.Srv.Store.Status().SaveOrUpdate(status); err != nil {
				mlog.Error(fmt.Sprintf("Failed to save status for user_id=%v, err=%v", userId, err), mlog.String("user_id", userId))
			}
		// 否则，只更新 "LastActivityAt" 到 Store 中
		} else {

			if err := a.Srv.Store.Status().UpdateLastActivityAt(status.UserId, status.LastActivityAt); err != nil {
				mlog.Error(fmt.Sprintf("Failed to save status for user_id=%v, err=%v", userId, err), mlog.String("user_id", userId))
			}

		}

	}

	// 广播 `用户状态变化` 事件
	if broadcast {
		a.BroadcastStatus(status)
	}
}

func (a *App) BroadcastStatus(status *model.Status) {
	// 用户状态变化
	event := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_STATUS_CHANGE, "", "", status.UserId, nil)
	event.Add("status", status.Status)
	event.Add("user_id", status.UserId)
	// 广播
	a.Publish(event)
}

// 离线
func (a *App) SetStatusOffline(userId string, manual bool) {
	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return
	}

	status, err := a.GetStatus(userId)
	if err == nil && status.Manual && !manual {
		return // manually set status always overrides non-manual one
	}

	status = &model.Status{UserId: userId, Status: model.STATUS_OFFLINE, Manual: manual, LastActivityAt: model.GetMillis(), ActiveChannel: ""}

	a.SaveAndBroadcastStatus(status)
}

// 离开
func (a *App) SetStatusAwayIfNeeded(userId string, manual bool) {


	// 是否允许显示用户状态，若否则直接返回
	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return
	}

	status, err := a.GetStatus(userId)
	if err != nil {
		status = &model.Status{
			UserId: userId,
			Status: model.STATUS_OFFLINE,
			Manual: manual,
			LastActivityAt: 0,
			ActiveChannel: "",
		}
	}

	// 手动设置的状态不被覆盖
	if !manual && status.Manual {
		return // manually set status always overrides non-manual one
	}

	if !manual {
		if status.Status == model.STATUS_AWAY {
			return
		}

		if !a.IsUserAway(status.LastActivityAt) {
			return
		}
	}

	status.Status = model.STATUS_AWAY
	status.Manual = manual
	status.ActiveChannel = ""

	a.SaveAndBroadcastStatus(status)
}

// 不要打扰
func (a *App) SetStatusDoNotDisturb(userId string) {

	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return
	}

	status, err := a.GetStatus(userId)
	if err != nil {
		status = &model.Status{
			UserId: userId,
			Status: model.STATUS_OFFLINE,
			Manual: false,
			LastActivityAt: 0,
			ActiveChannel: "",
		}
	}

	status.Status = model.STATUS_DND
	status.Manual = true // 手动设置
	a.SaveAndBroadcastStatus(status)
}

func (a *App) SaveAndBroadcastStatus(status *model.Status) {

	// 保存状态到 Cache 中
	a.AddStatusCache(status)

	// 保存状态到 Store 中
	if err := a.Srv.Store.Status().SaveOrUpdate(status); err != nil {
		mlog.Error(fmt.Sprintf("Failed to save status for user_id=%v, err=%v", status.UserId, err))
	}

	// 广播状态变化事件
	a.BroadcastStatus(status)
}

// 休假
func (a *App) SetStatusOutOfOffice(userId string) {

	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return
	}

	status, err := a.GetStatus(userId)
	if err != nil {
		status = &model.Status{
			UserId: userId,
			Status: model.STATUS_OUT_OF_OFFICE,
			Manual: false,
			LastActivityAt: 0,
			ActiveChannel: "",
		}
	}
	status.Status = model.STATUS_OUT_OF_OFFICE
	status.Manual = true // 手动设置
	a.SaveAndBroadcastStatus(status)
}


// 从 Cache 中查询用户状态
func GetStatusFromCache(userId string) *model.Status {

	if result, ok := statusCache.Get(userId); ok {
		status := result.(*model.Status)
		statusCopy := &model.Status{}
		*statusCopy = *status
		return statusCopy
	}

	return nil
}

// 查询用户状态
func (a *App) GetStatus(userId string) (*model.Status, *model.AppError) {

	// 配置检查
	if !*a.Config().ServiceSettings.EnableUserStatuses {
		return &model.Status{}, nil
	}

	// 从 Cache 中查询用户状态
	status := GetStatusFromCache(userId)
	if status != nil {
		return status, nil
	}
	// 从 Store 中查询用户状态
	return a.Srv.Store.Status().Get(userId)
}


// 是否已经离开超过 xxx 毫秒
func (a *App) IsUserAway(lastActivityAt int64) bool {
	return model.GetMillis()-lastActivityAt >= *a.Config().TeamSettings.UserStatusAwayTimeout*1000
}

// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/blastbao/mattermost-server/mlog"
	"github.com/blastbao/mattermost-server/model"
	"github.com/blastbao/mattermost-server/store"
	"github.com/blastbao/mattermost-server/utils"
	"github.com/blastbao/mattermost-server/utils/markdown"
)

const (
	THREAD_ANY  = "any"
	THREAD_ROOT = "root"
)


func (a *App) SendNotifications(post *model.Post, team *model.Team, channel *model.Channel, sender *model.User, parentPostList *model.PostList) ([]string, error) {

	// Do not send notifications in archived channels
	// 不要往已存档（删除）的频道发送通知
	if channel.DeleteAt > 0 {
		return []string{}, nil
	}

	// 拉取频道中所有用户的个人资料
	pchan := make(chan store.StoreResult, 1)
	go func() {
		props, err := a.Srv.Store.User().GetAllProfilesInChannel(channel.Id, true)
		pchan <- store.StoreResult{ Data: props, Err: err }
		close(pchan)
	}()

	// 拉取频道中所有用户的通知配置
	cmnchan := make(chan store.StoreResult, 1)
	go func() {
		props, err := a.Srv.Store.Channel().GetAllChannelMembersNotifyPropsForChannel(channel.Id, true)
		cmnchan <- store.StoreResult{ Data: props, Err: err }
		close(cmnchan)
	}()

	// 如果当前 post 有附件文件，则拉取这些文件的信息
	var fchan chan store.StoreResult
	if len(post.FileIds) != 0 {
		fchan = make(chan store.StoreResult, 1)
		go func() {
			fileInfos, err := a.Srv.Store.FileInfo().GetForPost(post.Id, true, false, true)
			fchan <- store.StoreResult{ Data: fileInfos, Err: err }
			close(fchan)
		}()
	}

	result := <-pchan
	if result.Err != nil {
		return nil, result.Err
	}
	profileMap := result.Data.(map[string]*model.User)

	result = <-cmnchan
	if result.Err != nil {
		return nil, result.Err
	}
	channelMemberNotifyPropsMap := result.Data.(map[string]model.StringMap)

	// 被 @ 的用户列表
	mentionedUserIds := make(map[string]bool)
	// ?
	threadMentionedUserIds := make(map[string]string)
	// 接收所有通知的用户列表
	allActivityPushUserIds := []string{}
	// @here 即 @ 频道中所有在线用户
	hereNotification := false
	// @channel 即 @ 频道中所有用户
	channelNotification := false
	// @all 即 @ 频道中所有用户
	allNotification := false
	// 用于接收 `更新用户被 @ 次数` 结果
	updateMentionChans := []chan *model.AppError{}

	// 如果频道是 Direct 类型，则取出对方的 userId 添加到 mentionedUserIds[] 中。
	if channel.Type == model.CHANNEL_DIRECT {
		otherUserId := channel.GetOtherUserIdForDM(post.UserId)
		_, ok := profileMap[otherUserId]
		if ok {
			mentionedUserIds[otherUserId] = true
		}
		if post.Props["from_webhook"] == "true" {
			mentionedUserIds[post.UserId] = true
		}
	} else {
		keywords := a.getMentionKeywordsInChannel(profileMap, post.Type != model.POST_HEADER_CHANGE && post.Type != model.POST_PURPOSE_CHANGE, channelMemberNotifyPropsMap)
		m := getExplicitMentions(post, keywords)

		// Add an implicit mention when a user is added to a channel
		// even if the user has set 'username mentions' to false in account settings.
		if post.Type == model.POST_ADD_TO_CHANNEL {
			val := post.Props[model.POST_PROPS_ADDED_USER_ID]
			if val != nil {
				uid := val.(string)
				m.MentionedUserIds[uid] = true
			}
		}
		mentionedUserIds = m.MentionedUserIds
		hereNotification = m.HereMentioned
		channelNotification = m.ChannelMentioned
		allNotification = m.AllMentioned

		// get users that have comment thread mentions enabled
		if len(post.RootId) > 0 && parentPostList != nil {
			for _, threadPost := range parentPostList.Posts {
				profile := profileMap[threadPost.UserId]
				if profile != nil && (profile.NotifyProps[model.COMMENTS_NOTIFY_PROP] == THREAD_ANY || (profile.NotifyProps[model.COMMENTS_NOTIFY_PROP] == THREAD_ROOT && threadPost.Id == parentPostList.Order[0])) {
					if threadPost.Id == parentPostList.Order[0] {
						threadMentionedUserIds[threadPost.UserId] = THREAD_ROOT
					} else {
						threadMentionedUserIds[threadPost.UserId] = THREAD_ANY
					}
					if _, ok := mentionedUserIds[threadPost.UserId]; !ok {
						mentionedUserIds[threadPost.UserId] = false
					}
				}
			}
		}

		// prevent the user from mentioning themselves
		if post.Props["from_webhook"] != "true" {
			delete(mentionedUserIds, post.UserId)
		}

		go func() {
			_, err := a.sendOutOfChannelMentions(sender, post, channel, m.OtherPotentialMentions)
			if err != nil {
				mlog.Error("Failed to send warning for out of channel mentions", mlog.String("user_id", sender.Id), mlog.String("post_id", post.Id), mlog.Err(err))
			}
		}()

		// find which users in the channel are set up to always receive mobile notifications
		// 查找频道中哪些用户被设置为总是接收移动通知。
		for _, profile := range profileMap {
			if (profile.NotifyProps[model.PUSH_NOTIFY_PROP] == model.USER_NOTIFY_ALL ||
				channelMemberNotifyPropsMap[profile.Id][model.PUSH_NOTIFY_PROP] == model.CHANNEL_NOTIFY_ALL) &&
				(post.UserId != profile.Id || post.Props["from_webhook"] == "true") &&
				!post.IsSystemMessage() {
				allActivityPushUserIds = append(allActivityPushUserIds, profile.Id)
			}
		}
	}

	// @ 用户列表
	mentionedUsersList := make([]string, 0, len(mentionedUserIds))
	for id := range mentionedUserIds {
		mentionedUsersList = append(mentionedUsersList, id)
		umc := make(chan *model.AppError, 1)
		go func(userId string) {
			// 增加用户 userId 的被 @ 数
			umc <- a.Srv.Store.Channel().IncrementMentionCount(post.ChannelId, userId)
			close(umc)
		}(id)
		updateMentionChans = append(updateMentionChans, umc)
	}

	notification := &postNotification{
		post:       post,
		channel:    channel,
		profileMap: profileMap,
		sender:     sender,
	}

	// 发送 email 给被 @ 的 user
	if *a.Config().EmailSettings.SendEmailNotifications {

		for _, id := range mentionedUsersList {

			// 拉取用户详情
			if profileMap[id] == nil {
				continue
			}

			// 检查1: 用户是否开启全局通知）
			userAllowsEmails := profileMap[id].NotifyProps[model.EMAIL_NOTIFY_PROP] != "false"
			// 检查2: 用户是否开启频道通知
			if channelEmail, ok := channelMemberNotifyPropsMap[id][model.EMAIL_NOTIFY_PROP]; ok {
				if channelEmail != model.CHANNEL_NOTIFY_DEFAULT {
					userAllowsEmails = channelEmail != "false"
				}
			}
			// 检查3: 用户是否静音了频道通知
			// Remove the user as recipient when the user has muted the channel.
			if channelMuted, ok := channelMemberNotifyPropsMap[id][model.MARK_UNREAD_NOTIFY_PROP]; ok {
				if channelMuted == model.CHANNEL_MARK_UNREAD_MENTION {
					mlog.Debug(fmt.Sprintf("Channel muted for user_id %v, channel_mute %v", id, channelMuted))
					userAllowsEmails = false
				}
			}
			// 检查4: 用户的 email 是否已经认证
			//If email verification is required and user email is not verified don't send email.
			if *a.Config().EmailSettings.RequireEmailVerification && !profileMap[id].EmailVerified {
				mlog.Error(fmt.Sprintf("Skipped sending notification email to %v, address not verified. [details: user_id=%v]", profileMap[id].Email, id))
				continue
			}

			// 拉取用户的在线状态
			var status *model.Status
			var err *model.AppError
			if status, err = a.GetStatus(id); err != nil {
				status = &model.Status{
					UserId:         id,
					Status:         model.STATUS_OFFLINE,
					Manual:         false,
					LastActivityAt: 0,
					ActiveChannel:  "",
				}
			}

			// 检查5: 用户是否为 `休假` 状态
			autoResponderRelated := status.Status == model.STATUS_OUT_OF_OFFICE || post.Type == model.POST_AUTO_RESPONDER
			// 检查6: 用户未在线、用户未注销
			if userAllowsEmails && status.Status != model.STATUS_ONLINE && profileMap[id].DeleteAt == 0 && !autoResponderRelated {
				// 发送 Email
				a.sendNotificationEmail(notification, profileMap[id], team)
			}
		}
	}

	T := utils.GetUserTranslations(sender.Locale)

	// If the channel has more than 1K users then @here is disabled
	// 如果频道的用户数超过 1K ，那么 @here 会被禁用。（注：@here 即 @ 频道中所有在线用户）
	if hereNotification && int64(len(profileMap)) > *a.Config().TeamSettings.MaxNotificationsPerChannel {
		hereNotification = false
		// 发送临时消息给 userId，告知其 @here 的人数超过限制。
		a.SendEphemeralPost(
			post.UserId,
			&model.Post{
				ChannelId: post.ChannelId,
				Message:   T("api.post.disabled_here", map[string]interface{}{"Users": *a.Config().TeamSettings.MaxNotificationsPerChannel}),
				CreateAt:  post.CreateAt + 1,
			},
		)

	}

	// If the channel has more than 1K users then @channel is disabled
	// 如果频道的用户数超过 1K ，那么 @channel 会被禁用。（注：@channel 即 @ 频道中所有用户）
	if channelNotification && int64(len(profileMap)) > *a.Config().TeamSettings.MaxNotificationsPerChannel {
		// 发送临时消息给 userId，告知其 @channel 的人数超过限制。
		a.SendEphemeralPost(
			post.UserId,
			&model.Post{
				ChannelId: post.ChannelId,
				Message:   T("api.post.disabled_channel", map[string]interface{}{"Users": *a.Config().TeamSettings.MaxNotificationsPerChannel}),
				CreateAt:  post.CreateAt + 1,
			},
		)
	}

	// If the channel has more than 1K users then @all is disabled
	// 如果频道的用户数超过 1K ，那么 @all 会被禁用。
	if allNotification && int64(len(profileMap)) > *a.Config().TeamSettings.MaxNotificationsPerChannel {
		// 发送临时消息给 userId，告知其 @all 的人数超过限制。
		a.SendEphemeralPost(
			post.UserId,
			&model.Post{
				ChannelId: post.ChannelId,
				Message:   T("api.post.disabled_all", map[string]interface{}{"Users": *a.Config().TeamSettings.MaxNotificationsPerChannel}),
				CreateAt:  post.CreateAt + 1,
			},
		)
	}

	// Make sure all mention updates are complete to prevent race
	// Probably better to batch these DB updates in the future
	// MUST be completed before push notifications send
	for _, umc := range updateMentionChans {
		if err := <-umc; err != nil {
			mlog.Warn(fmt.Sprintf("Failed to update mention count, post_id=%v channel_id=%v err=%v", post.Id, post.ChannelId, result.Err), mlog.String("post_id", post.Id))
		}
	}

	// 是否需要发送通知
	sendPushNotifications := false
	if *a.Config().EmailSettings.SendPushNotifications {
		pushServer := *a.Config().EmailSettings.PushNotificationServer
		license := a.License()
		if pushServer == model.MHPNS && (license == nil || !*license.Features.MHPNS) {
			mlog.Warn("Push notifications are disabled. Go to System Console > Notifications > Mobile Push to enable them.")
			sendPushNotifications = false
		} else {
			sendPushNotifications = true
		}
	}

	if sendPushNotifications {

		// 遍历被 @ 用户列表
		for _, id := range mentionedUsersList {
			if profileMap[id] == nil {
				continue
			}
			var status *model.Status
			var err *model.AppError
			// 查询用户在线状态，查询失败则默认其为 `离线`
			if status, err = a.GetStatus(id); err != nil {
				status = &model.Status{
					UserId: id,
					Status: model.STATUS_OFFLINE,
					Manual: false,
					LastActivityAt: 0,
					ActiveChannel: "",
				}
			}
			// 是否应该推送通知
			if ShouldSendPushNotification(profileMap[id], channelMemberNotifyPropsMap[id], true, status, post) {
				replyToThreadType := ""
				if value, ok := threadMentionedUserIds[id]; ok {
					replyToThreadType = value
				}
				// 发送通知给 user
				a.sendPushNotification(
					notification,
					profileMap[id],
					mentionedUserIds[id],
					(channelNotification || hereNotification || allNotification),
					replyToThreadType,
				)
			} else {
				// register that a notification was not sent
				a.NotificationsLog.Warn("Notification not sent",
					mlog.String("ackId", ""),
					mlog.String("type", model.PUSH_TYPE_MESSAGE),
					mlog.String("userId", id),
					mlog.String("postId", post.Id),
					mlog.String("status", model.PUSH_NOT_SENT),
				)
			}
		}


		for _, id := range allActivityPushUserIds {
			if profileMap[id] == nil {
				continue
			}
			if _, ok := mentionedUserIds[id]; !ok {
				var status *model.Status
				var err *model.AppError
				if status, err = a.GetStatus(id); err != nil {
					status = &model.Status{UserId: id, Status: model.STATUS_OFFLINE, Manual: false, LastActivityAt: 0, ActiveChannel: ""}
				}
				if ShouldSendPushNotification(profileMap[id], channelMemberNotifyPropsMap[id], false, status, post) {
					// 发送通知给 user
					a.sendPushNotification(
						notification,
						profileMap[id],
						false,
						false,
						"",
					)
				} else {
					// register that a notification was not sent
					a.NotificationsLog.Warn("Notification not sent",
						mlog.String("ackId", ""),
						mlog.String("type", model.PUSH_TYPE_MESSAGE),
						mlog.String("userId", id),
						mlog.String("postId", post.Id),
						mlog.String("status", model.PUSH_NOT_SENT),
					)
				}
			}
		}
	}


	// Note that PreparePostForClient should've already been called by this point
	//
	// 注意，此时 PreparePostForClient 应该已经被调用了。
	//

	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_POSTED, "", post.ChannelId, "", nil)
	message.Add("post", post.ToJson())
	message.Add("channel_type", channel.Type)
	message.Add("channel_display_name", notification.GetChannelName(model.SHOW_USERNAME, ""))
	message.Add("channel_name", channel.Name)
	message.Add("sender_name", notification.GetSenderName(model.SHOW_USERNAME, *a.Config().ServiceSettings.EnablePostUsernameOverride))
	message.Add("team_id", team.Id)

	// 添加 附件信息
	if len(post.FileIds) != 0 && fchan != nil {
		message.Add("otherFile", "true")
		var infos []*model.FileInfo
		if result := <-fchan; result.Err != nil {
			mlog.Warn(fmt.Sprint("Unable to get fileInfo for push notifications.", post.Id, result.Err), mlog.String("post_id", post.Id))
		} else {
			infos = result.Data.([]*model.FileInfo)
		}
		for _, info := range infos {
			if info.IsImage() {
				message.Add("image", "true")
				break
			}
		}
	}

	// 添加 @ 用户列表
	if len(mentionedUsersList) != 0 {
		message.Add("mentions", model.ArrayToJson(mentionedUsersList))
	}

	// 广播消息到频道中
	a.Publish(message)

	return mentionedUsersList, nil
}

// sendOutOfChannelMentions sends an ephemeral post to the sender of a post if any of the given potential mentions
// are outside of the post's channel. Returns whether or not an ephemeral post was sent.
func (a *App) sendOutOfChannelMentions(sender *model.User, post *model.Post, channel *model.Channel, potentialMentions []string) (bool, error) {
	outOfChannelUsers, outOfGroupsUsers, err := a.filterOutOfChannelMentions(sender, post, channel, potentialMentions)
	if err != nil {
		return false, err
	}

	if len(outOfChannelUsers) == 0 && len(outOfGroupsUsers) == 0 {
		return false, nil
	}

	a.SendEphemeralPost(post.UserId, makeOutOfChannelMentionPost(sender, post, outOfChannelUsers, outOfGroupsUsers))

	return true, nil
}

func (a *App) filterOutOfChannelMentions(sender *model.User, post *model.Post, channel *model.Channel, potentialMentions []string) ([]*model.User, []*model.User, error) {
	if post.IsSystemMessage() {
		return nil, nil, nil
	}

	if channel.TeamId == "" || channel.Type == model.CHANNEL_DIRECT || channel.Type == model.CHANNEL_GROUP {
		return nil, nil, nil
	}

	if len(potentialMentions) == 0 {
		return nil, nil, nil
	}

	users, err := a.Srv.Store.User().GetProfilesByUsernames(potentialMentions, &model.ViewUsersRestrictions{Teams: []string{channel.TeamId}})
	if err != nil {
		return nil, nil, err
	}

	// Filter out inactive users and bots
	allUsers := model.UserSlice(users).FilterByActive(true)
	allUsers = allUsers.FilterWithoutBots()

	if len(allUsers) == 0 {
		return nil, nil, nil
	}

	// Differentiate between users who can and can't be added to the channel
	var outOfChannelUsers model.UserSlice
	var outOfGroupsUsers model.UserSlice
	if channel.IsGroupConstrained() {
		nonMemberIDs, err := a.FilterNonGroupChannelMembers(allUsers.IDs(), channel)
		if err != nil {
			return nil, nil, err
		}

		outOfChannelUsers = allUsers.FilterWithoutID(nonMemberIDs)
		outOfGroupsUsers = allUsers.FilterByID(nonMemberIDs)
	} else {
		outOfChannelUsers = users
	}

	return outOfChannelUsers, outOfGroupsUsers, nil
}

func makeOutOfChannelMentionPost(sender *model.User, post *model.Post, outOfChannelUsers, outOfGroupsUsers []*model.User) *model.Post {
	allUsers := model.UserSlice(append(outOfChannelUsers, outOfGroupsUsers...))

	ocUsers := model.UserSlice(outOfChannelUsers)
	ocUsernames := ocUsers.Usernames()
	ocUserIDs := ocUsers.IDs()

	ogUsers := model.UserSlice(outOfGroupsUsers)
	ogUsernames := ogUsers.Usernames()

	T := utils.GetUserTranslations(sender.Locale)

	ephemeralPostId := model.NewId()
	var message string
	if len(outOfChannelUsers) == 1 {
		message = T("api.post.check_for_out_of_channel_mentions.message.one", map[string]interface{}{
			"Username": ocUsernames[0],
		})
	} else if len(outOfChannelUsers) > 1 {
		preliminary, final := splitAtFinal(ocUsernames)

		message = T("api.post.check_for_out_of_channel_mentions.message.multiple", map[string]interface{}{
			"Usernames":    strings.Join(preliminary, ", @"),
			"LastUsername": final,
		})
	}

	if len(outOfGroupsUsers) == 1 {
		if len(message) > 0 {
			message += "\n"
		}

		message += T("api.post.check_for_out_of_channel_groups_mentions.message.one", map[string]interface{}{
			"Username": ogUsernames[0],
		})
	} else if len(outOfGroupsUsers) > 1 {
		preliminary, final := splitAtFinal(ogUsernames)

		if len(message) > 0 {
			message += "\n"
		}

		message += T("api.post.check_for_out_of_channel_groups_mentions.message.multiple", map[string]interface{}{
			"Usernames":    strings.Join(preliminary, ", @"),
			"LastUsername": final,
		})
	}

	props := model.StringInterface{
		model.PROPS_ADD_CHANNEL_MEMBER: model.StringInterface{
			"post_id": ephemeralPostId,

			"usernames":                allUsers.Usernames(), // Kept for backwards compatibility of mobile app.
			"not_in_channel_usernames": ocUsernames,

			"user_ids":                allUsers.IDs(), // Kept for backwards compatibility of mobile app.
			"not_in_channel_user_ids": ocUserIDs,

			"not_in_groups_usernames": ogUsernames,
			"not_in_groups_user_ids":  ogUsers.IDs(),
		},
	}

	return &model.Post{
		Id:        ephemeralPostId,
		RootId:    post.RootId,
		ChannelId: post.ChannelId,
		Message:   message,
		CreateAt:  post.CreateAt + 1,
		Props:     props,
	}
}

func splitAtFinal(items []string) (preliminary []string, final string) {
	if len(items) == 0 {
		return
	}
	preliminary = items[:len(items)-1]
	final = items[len(items)-1]
	return
}

type ExplicitMentions struct {

	// MentionedUserIds contains a key for each user mentioned by keyword.
	MentionedUserIds map[string]bool

	// OtherPotentialMentions contains a list of strings that looked like mentions, but didn't have
	// a corresponding keyword.
	OtherPotentialMentions []string

	// HereMentioned is true if the message contained @here.
	HereMentioned bool

	// AllMentioned is true if the message contained @all.
	AllMentioned bool

	// ChannelMentioned is true if the message contained @channel.
	ChannelMentioned bool
}


// Given a message and a map mapping mention keywords to the users who use them, returns a map of mentioned
// users and a slice of potential mention users not in the channel and whether or not @here was mentioned.
//
//
//
func getExplicitMentions(post *model.Post, keywords map[string][]string) *ExplicitMentions {

	ret := &ExplicitMentions{
		MentionedUserIds: make(map[string]bool),
	}

	buf := ""
	mentionsEnabledFields := getMentionsEnabledFields(post)
	for _, message := range mentionsEnabledFields {
		markdown.Inspect(message, func(node interface{}) bool {
			text, ok := node.(*markdown.Text)
			if !ok {
				ret.processText(buf, keywords)
				buf = ""
				return true
			}
			buf += text.Text
			return false
		})
	}
	ret.processText(buf, keywords)

	return ret
}

// Given a post returns the values of the fields in which mentions are possible.
// post.message, preText and text in the attachment are enabled.
func getMentionsEnabledFields(post *model.Post) model.StringArray {
	ret := []string{}

	ret = append(ret, post.Message)
	for _, attachment := range post.Attachments() {

		if len(attachment.Pretext) != 0 {
			ret = append(ret, attachment.Pretext)
		}
		if len(attachment.Text) != 0 {
			ret = append(ret, attachment.Text)
		}
	}
	return ret
}

// Given a map of user IDs to profiles, returns a list of mention
// keywords for all users in the channel.
func (a *App) getMentionKeywordsInChannel(profiles map[string]*model.User, lookForSpecialMentions bool, channelMemberNotifyPropsMap map[string]model.StringMap) map[string][]string {



	keywords := make(map[string][]string)



	for id, profile := range profiles {


		userMention := "@" + strings.ToLower(profile.Username)
		keywords[userMention] = append(keywords[userMention], id)

		if len(profile.NotifyProps[model.MENTION_KEYS_NOTIFY_PROP]) > 0 {
			// Add all the user's mention keys
			splitKeys := strings.Split(profile.NotifyProps[model.MENTION_KEYS_NOTIFY_PROP], ",")
			for _, k := range splitKeys {
				// note that these are made lower case so that we can do a case insensitive check for them
				key := strings.ToLower(k)
				if key != "" {
					keywords[key] = append(keywords[key], id)
				}
			}
		}

		// If turned on, add the user's case sensitive first name
		if profile.NotifyProps[model.FIRST_NAME_NOTIFY_PROP] == "true" {
			keywords[profile.FirstName] = append(keywords[profile.FirstName], profile.Id)
		}

		ignoreChannelMentions := false
		if ignoreChannelMentionsNotifyProp, ok := channelMemberNotifyPropsMap[profile.Id][model.IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP]; ok {
			if ignoreChannelMentionsNotifyProp == model.IGNORE_CHANNEL_MENTIONS_ON {
				ignoreChannelMentions = true
			}
		}

		// Add @channel and @all to keywords if user has them turned on
		if lookForSpecialMentions {
			if int64(len(profiles)) <= *a.Config().TeamSettings.MaxNotificationsPerChannel && profile.NotifyProps[model.CHANNEL_MENTIONS_NOTIFY_PROP] == "true" && !ignoreChannelMentions {
				keywords["@channel"] = append(keywords["@channel"], profile.Id)
				keywords["@all"] = append(keywords["@all"], profile.Id)

				status := GetStatusFromCache(profile.Id)
				if status != nil && status.Status == model.STATUS_ONLINE {
					keywords["@here"] = append(keywords["@here"], profile.Id)
				}
			}
		}
	}

	return keywords
}

// Represents either an email or push notification and contains the fields required to send it to any user.
type postNotification struct {
	channel    *model.Channel
	post       *model.Post
	profileMap map[string]*model.User
	sender     *model.User
}

// Returns the name of the channel for this notification. For direct messages, this is the sender's name
// preceeded by an at sign. For group messages, this is a comma-separated list of the members of the
// channel, with an option to exclude the recipient of the message from that list.
//
// 获取通知的名称：
// 	对于直接消息，是在发件人的名字前加一个 at 符号。
// 	对于群组消息，是一个以逗号分隔的频道成员列表，并且将 excludeId 从该列表中排除。
//
func (n *postNotification) GetChannelName(userNameFormat string, excludeId string) string {

	switch n.channel.Type {
	case model.CHANNEL_DIRECT:
		return n.sender.GetDisplayNameWithPrefix(userNameFormat, "@")
	case model.CHANNEL_GROUP:
		names := []string{}
		for _, user := range n.profileMap {
			if user.Id != excludeId {
				names = append(names, user.GetDisplayName(userNameFormat))
			}
		}
		sort.Strings(names)

		return strings.Join(names, ", ")
	default:
		return n.channel.DisplayName
	}
}

// Returns the name of the sender of this notification, accounting for things like system messages
// and whether or not the username has been overridden by an integration.
//
// 返回该通知的发送者的名称，考虑到系统消息等因素，以及用户名是否已被重写。
func (n *postNotification) GetSenderName(userNameFormat string, overridesAllowed bool) string {
	// 系统消息
	if n.post.IsSystemMessage() {
		return utils.T("system.message.name")
	}
	if overridesAllowed && n.channel.Type != model.CHANNEL_DIRECT {
		if value, ok := n.post.Props["override_username"]; ok && n.post.Props["from_webhook"] == "true" {
			return value.(string)
		}
	}
	return n.sender.GetDisplayNameWithPrefix(userNameFormat, "@")
}

// addMentionedUsers will add the mentioned user id in the struct's list for mentioned users
func (e *ExplicitMentions) addMentionedUsers(ids []string) {
	for _, id := range ids {
		e.MentionedUserIds[id] = true
	}
}

// checkForMention checks if there is a mention to a specific user or to the keywords here / channel / all
func (e *ExplicitMentions) checkForMention(word string, keywords map[string][]string) bool {
	isMention := false

	switch strings.ToLower(word) {
	case "@here":
		e.HereMentioned = true
	case "@channel":
		e.ChannelMentioned = true
	case "@all":
		e.AllMentioned = true
	}

	if ids, match := keywords[strings.ToLower(word)]; match {
		e.addMentionedUsers(ids)
		isMention = true
	}

	// Case-sensitive check for first name
	if ids, match := keywords[word]; match {
		e.addMentionedUsers(ids)
		isMention = true
	}

	return isMention
}

// isKeywordMultibyte checks if a word containing a multibyte character contains a multibyte keyword
func isKeywordMultibyte(keywords map[string][]string, word string) ([]string, bool) {
	ids := []string{}
	match := false
	var multibyteKeywords []string
	for keyword := range keywords {
		if len(keyword) != utf8.RuneCountInString(keyword) {
			multibyteKeywords = append(multibyteKeywords, keyword)
		}
	}

	if len(word) != utf8.RuneCountInString(word) {
		for _, key := range multibyteKeywords {
			if strings.Contains(word, key) {
				ids, match = keywords[key]
			}
		}
	}
	return ids, match
}

// Processes text to filter mentioned users and other potential mentions
func (e *ExplicitMentions) processText(text string, keywords map[string][]string) {
	systemMentions := map[string]bool{"@here": true, "@channel": true, "@all": true}

	for _, word := range strings.FieldsFunc(text, func(c rune) bool {
		// Split on any whitespace or punctuation that can't be part of an at mention or emoji pattern
		return !(c == ':' || c == '.' || c == '-' || c == '_' || c == '@' || unicode.IsLetter(c) || unicode.IsNumber(c))
	}) {
		// skip word with format ':word:' with an assumption that it is an emoji format only
		if word[0] == ':' && word[len(word)-1] == ':' {
			continue
		}

		word = strings.TrimLeft(word, ":.-_")

		if e.checkForMention(word, keywords) {
			continue
		}

		foundWithoutSuffix := false
		wordWithoutSuffix := word
		for len(wordWithoutSuffix) > 0 && strings.LastIndexAny(wordWithoutSuffix, ".-:_") == (len(wordWithoutSuffix)-1) {
			wordWithoutSuffix = wordWithoutSuffix[0 : len(wordWithoutSuffix)-1]

			if e.checkForMention(wordWithoutSuffix, keywords) {
				foundWithoutSuffix = true
				break
			}
		}

		if foundWithoutSuffix {
			continue
		}

		if _, ok := systemMentions[word]; !ok && strings.HasPrefix(word, "@") {
			e.OtherPotentialMentions = append(e.OtherPotentialMentions, word[1:])
		} else if strings.ContainsAny(word, ".-:") {
			// This word contains a character that may be the end of a sentence, so split further
			splitWords := strings.FieldsFunc(word, func(c rune) bool {
				return c == '.' || c == '-' || c == ':'
			})

			for _, splitWord := range splitWords {
				if e.checkForMention(splitWord, keywords) {
					continue
				}
				if _, ok := systemMentions[splitWord]; !ok && strings.HasPrefix(splitWord, "@") {
					e.OtherPotentialMentions = append(e.OtherPotentialMentions, splitWord[1:])
				}
			}
		}
		if ids, match := isKeywordMultibyte(keywords, word); match {
			e.addMentionedUsers(ids)
		}
	}
}

func (a *App) GetNotificationNameFormat(user *model.User) string {
	if !*a.Config().PrivacySettings.ShowFullName {
		return model.SHOW_USERNAME
	}

	data, err := a.Srv.Store.Preference().Get(user.Id, model.PREFERENCE_CATEGORY_DISPLAY_SETTINGS, model.PREFERENCE_NAME_NAME_FORMAT)
	if err != nil {
		return *a.Config().TeamSettings.TeammateNameDisplay
	}

	return data.Value
}

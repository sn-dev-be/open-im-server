// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package push

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/conversation"
	"github.com/OpenIMSDK/protocol/msggateway"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/protocol/user"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/common/i18n"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlineinfo"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/dummy"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/fcm"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/getui"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/gorush"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/jpush"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/localcache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/prommetrics"
	"github.com/openimsdk/open-im-server/v3/pkg/msgprocessor"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

type Pusher struct {
	database               controller.PushDatabase
	discov                 discoveryregistry.SvcDiscoveryRegistry
	offlinePusher          offlinepush.OfflinePusher
	groupLocalCache        *localcache.GroupLocalCache
	serverLocalCache       *localcache.ServerLocalCache
	conversationLocalCache *localcache.ConversationLocalCache
	msgRpcClient           *rpcclient.MessageRpcClient
	conversationRpcClient  *rpcclient.ConversationRpcClient
	groupRpcClient         *rpcclient.GroupRpcClient
	offlineInfoParse       *offlineinfo.OfflineInfoParse
	userRpcClient          *rpcclient.UserRpcClient
	clubRpcClient          *rpcclient.ClubRpcClient
}

var errNoOfflinePusher = errors.New("no offlinePusher is configured")

func NewPusher(discov discoveryregistry.SvcDiscoveryRegistry, offlinePusher offlinepush.OfflinePusher, database controller.PushDatabase,
	groupLocalCache *localcache.GroupLocalCache, conversationLocalCache *localcache.ConversationLocalCache, serverLocalCache *localcache.ServerLocalCache,
	conversationRpcClient *rpcclient.ConversationRpcClient, groupRpcClient *rpcclient.GroupRpcClient, msgRpcClient *rpcclient.MessageRpcClient,
	userRpcClient *rpcclient.UserRpcClient, clubRpcClient *rpcclient.ClubRpcClient,
) *Pusher {
	return &Pusher{
		discov:                 discov,
		database:               database,
		offlinePusher:          offlinePusher,
		groupLocalCache:        groupLocalCache,
		serverLocalCache:       serverLocalCache,
		conversationLocalCache: conversationLocalCache,
		msgRpcClient:           msgRpcClient,
		conversationRpcClient:  conversationRpcClient,
		groupRpcClient:         groupRpcClient,
		offlineInfoParse:       offlineinfo.NewOfflineInfoParse(groupRpcClient, clubRpcClient),
		userRpcClient:          userRpcClient,
		clubRpcClient:          clubRpcClient,
	}
}

func NewOfflinePusher(cache cache.MsgModel) offlinepush.OfflinePusher {
	var offlinePusher offlinepush.OfflinePusher
	switch config.Config.Push.Enable {
	case "getui":
		offlinePusher = getui.NewClient(cache)
	case "fcm":
		offlinePusher = fcm.NewClient(cache)
	case "jpush":
		offlinePusher = jpush.NewClient()
	case "gorush":
		offlinePusher = gorush.NewClient(cache)
	default:
		offlinePusher = dummy.NewClient()
	}
	return offlinePusher
}

func (p *Pusher) DeleteMemberAndSetConversationSeq(ctx context.Context, groupID string, userIDs []string) error {
	conevrsationID := msgprocessor.GetConversationIDBySessionType(constant.SuperGroupChatType, groupID)
	maxSeq, err := p.msgRpcClient.GetConversationMaxSeq(ctx, conevrsationID)
	if err != nil {
		return err
	}
	return p.conversationRpcClient.SetConversationMaxSeq(ctx, userIDs, conevrsationID, maxSeq)
}

func (p *Pusher) Push2User(ctx context.Context, userIDs []string, msg *sdkws.MsgData) error {
	log.ZDebug(ctx, "Get msg from msg_transfer And push msg", "userIDs", userIDs, "msg", msg.String())
	// callback
	if err := callbackOnlinePush(ctx, userIDs, msg); err != nil {
		return err
	}

	// push
	wsResults, err := p.GetConnsAndOnlinePush(ctx, msg, userIDs)
	if err != nil {
		return err
	}

	isOfflinePush := utils.GetSwitchFromOptions(msg.Options, constant.IsOfflinePush)
	log.ZDebug(ctx, "push_result", "ws push result", wsResults, "sendData", msg, "isOfflinePush", isOfflinePush, "push_to_userID", userIDs)

	if !isOfflinePush {
		return nil
	}

	for _, v := range wsResults {
		if msg.SendID != v.UserID && (!v.OnlinePush) {
			if err = callbackOfflinePush(ctx, userIDs, msg, &[]string{}); err != nil {
				return err
			}

			//newMsgPushMode
			//p.userRpcClient.Client.GetUserSetting()
			err = p.offlinePushMsg(ctx, msg.SendID, msg, []string{v.UserID})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Pusher) UnmarshalNotificationElem(bytes []byte, t interface{}) error {
	var notification sdkws.NotificationElem
	if err := json.Unmarshal(bytes, &notification); err != nil {
		return err
	}

	return json.Unmarshal([]byte(notification.Detail), t)
}

func (p *Pusher) Push2SuperGroup(ctx context.Context, groupID string, msg *sdkws.MsgData) (err error) {
	log.ZDebug(ctx, "Get super group msg from msg_transfer and push msg", "msg", msg.String(), "groupID", groupID)
	var pushToUserIDs []string
	if err = callbackBeforeSuperGroupOnlinePush(ctx, groupID, msg, &pushToUserIDs); err != nil {
		return err
	}
	if len(msg.RecvIDList) > 0 {
		pushToUserIDs = msg.RecvIDList
	}
	if len(pushToUserIDs) == 0 {
		pushToUserIDs, err = p.groupLocalCache.GetGroupMemberIDs(ctx, groupID)
		if err != nil {
			return err
		}

		switch msg.ContentType {
		case constant.MemberQuitNotification:
			var tips sdkws.MemberQuitTips
			if p.UnmarshalNotificationElem(msg.Content, &tips) != nil {
				return err
			}
			defer func(groupID string, userIDs []string) {
				if err = p.DeleteMemberAndSetConversationSeq(ctx, groupID, userIDs); err != nil {
					log.ZError(ctx, "MemberQuitNotification DeleteMemberAndSetConversationSeq", err, "groupID", groupID, "userIDs", userIDs)
				}
			}(groupID, []string{tips.QuitUser.UserID})
			pushToUserIDs = append(pushToUserIDs, tips.QuitUser.UserID)
		case constant.MemberKickedNotification:
			var tips sdkws.MemberKickedTips
			if p.UnmarshalNotificationElem(msg.Content, &tips) != nil {
				return err
			}
			kickedUsers := utils.Slice(tips.KickedUserList, func(e *sdkws.GroupMemberFullInfo) string { return e.UserID })
			defer func(groupID string, userIDs []string) {
				if err = p.DeleteMemberAndSetConversationSeq(ctx, groupID, userIDs); err != nil {
					log.ZError(ctx, "MemberKickedNotification DeleteMemberAndSetConversationSeq", err, "groupID", groupID, "userIDs", userIDs)
				}
			}(groupID, kickedUsers)
			pushToUserIDs = append(pushToUserIDs, kickedUsers...)
		case constant.GroupDismissedNotification:
			if msgprocessor.IsNotification(msgprocessor.GetConversationIDByMsg(msg)) { // 消息先到,通知后到
				var tips sdkws.GroupDismissedTips
				if p.UnmarshalNotificationElem(msg.Content, &tips) != nil {
					return err
				}
				log.ZInfo(ctx, "GroupDismissedNotificationInfo****", "groupID", groupID, "num", len(pushToUserIDs), "list", pushToUserIDs)
				if len(config.Config.Manager.UserID) > 0 {
					ctx = mcontext.WithOpUserIDContext(ctx, config.Config.Manager.UserID[0])
				}
				defer func(groupID string) {
					if err = p.groupRpcClient.DismissGroup(ctx, groupID); err != nil {
						log.ZError(ctx, "DismissGroup Notification clear members", err, "groupID", groupID)
					}
				}(groupID)
			}
		}
	}

	wsResults, err := p.GetConnsAndOnlinePush(ctx, msg, pushToUserIDs)
	if err != nil {
		return err
	}

	log.ZDebug(ctx, "get conn and online push success", "result", wsResults, "msg", msg)
	isOfflinePush := utils.GetSwitchFromOptions(msg.Options, constant.IsOfflinePush)
	if isOfflinePush {
		var (
			onlineSuccessUserIDs      = []string{msg.SendID}
			webAndPcBackgroundUserIDs []string
		)

		for _, v := range wsResults {
			if v.OnlinePush && v.UserID != msg.SendID {
				onlineSuccessUserIDs = append(onlineSuccessUserIDs, v.UserID)
			}

			if v.OnlinePush {
				continue
			}

			if len(v.Resp) == 0 {
				continue
			}

			for _, singleResult := range v.Resp {
				if singleResult.ResultCode != -2 {
					continue
				}

				isPC := constant.PlatformIDToName(int(singleResult.RecvPlatFormID)) == constant.TerminalPC
				isWebID := singleResult.RecvPlatFormID == constant.WebPlatformID

				if isPC || isWebID {
					webAndPcBackgroundUserIDs = append(webAndPcBackgroundUserIDs, v.UserID)
				}
			}
		}

		needOfflinePushUserIDs := utils.DifferenceString(onlineSuccessUserIDs, pushToUserIDs)
		if msg.ContentType != constant.SignalingNotification {
			notNotificationUserIDs, err := p.conversationLocalCache.GetRecvMsgNotNotifyUserIDs(ctx, groupID)
			if err != nil {
				return err
			}
			needOfflinePushUserIDs = utils.SliceSub(needOfflinePushUserIDs, notNotificationUserIDs)
		}
		// Use offline push messaging
		if len(needOfflinePushUserIDs) > 0 {
			var offlinePushUserIDs []string
			err = callbackOfflinePush(ctx, needOfflinePushUserIDs, msg, &offlinePushUserIDs)
			if err != nil {
				return err
			}
			if len(offlinePushUserIDs) > 0 {
				needOfflinePushUserIDs = offlinePushUserIDs
			}
			resp, err := p.conversationRpcClient.Client.GetConversationOfflinePushUserIDs(
				ctx,
				&conversation.GetConversationOfflinePushUserIDsReq{ConversationID: utils.GenGroupConversationID(groupID), UserIDs: needOfflinePushUserIDs},
			)
			if err != nil {
				return err
			}
			if len(resp.UserIDs) > 0 {
				err = p.offlinePushMsg(ctx, groupID, msg, resp.UserIDs)
				if err != nil {
					log.ZError(ctx, "offlinePushMsg failed", err, "groupID", groupID, "msg", msg)
					return err
				}
				if _, err := p.GetConnsAndOnlinePush(ctx, msg, utils.IntersectString(resp.UserIDs, webAndPcBackgroundUserIDs)); err != nil {
					log.ZError(ctx, "offlinePushMsg failed", err, "groupID", groupID, "msg", msg, "userIDs", utils.IntersectString(needOfflinePushUserIDs, webAndPcBackgroundUserIDs))
					return err
				}
			}
		}
	}
	return nil
}

func (p *Pusher) Push2ServerGroup(ctx context.Context, groupID string, msg *sdkws.MsgData) (err error) {
	log.ZDebug(ctx, "Get server group msg from msg_transfer and push msg", "msg", msg.String(), "groupID", groupID)
	var pushToUserIDs []string
	if len(msg.RecvIDList) > 0 {
		pushToUserIDs = msg.RecvIDList
	}
	if len(pushToUserIDs) == 0 {
		pushToUserIDs, err = p.serverLocalCache.GetServerMemberIDs(ctx, groupID)
		if err != nil {
			return err
		}
	}
	wsResults, err := p.GetConnsAndOnlinePush(ctx, msg, pushToUserIDs)
	if err != nil {
		return err
	}
	log.ZDebug(ctx, "get conn and online push success", "result", wsResults, "msg", msg)
	isOfflinePush := utils.GetSwitchFromOptions(msg.Options, constant.IsOfflinePush)
	if isOfflinePush {
		var (
			onlineSuccessUserIDs      = []string{msg.SendID}
			webAndPcBackgroundUserIDs []string
		)

		for _, v := range wsResults {
			if v.OnlinePush && v.UserID != msg.SendID {
				onlineSuccessUserIDs = append(onlineSuccessUserIDs, v.UserID)
			}

			if v.OnlinePush {
				continue
			}

			if len(v.Resp) == 0 {
				continue
			}

			for _, singleResult := range v.Resp {
				if singleResult.ResultCode != -2 {
					continue
				}

				isPC := constant.PlatformIDToName(int(singleResult.RecvPlatFormID)) == constant.TerminalPC
				isWebID := singleResult.RecvPlatFormID == constant.WebPlatformID

				if isPC || isWebID {
					webAndPcBackgroundUserIDs = append(webAndPcBackgroundUserIDs, v.UserID)
				}
			}

		}
		needOfflinePushUserIDs := utils.DifferenceString(onlineSuccessUserIDs, pushToUserIDs)
		if msg.ContentType != constant.SignalingNotification {
			notNotificationUserIDs, err := p.conversationLocalCache.GetRecvMsgNotNotifyUserIDs(ctx, groupID)
			if err != nil {
				// log.ZError(ctx, "GetRecvMsgNotNotifyUserIDs failed", err, "groupID", groupID)
				return err
			}
			needOfflinePushUserIDs = utils.SliceSub(needOfflinePushUserIDs, notNotificationUserIDs)
		}

		//filter users who set newMsgPushMode false

		// Use offline push messaging
		if len(needOfflinePushUserIDs) > 0 {
			var offlinePushUserIDs []string
			err = callbackOfflinePush(ctx, needOfflinePushUserIDs, msg, &offlinePushUserIDs)
			if err != nil {
				return err
			}

			if len(offlinePushUserIDs) > 0 {
				needOfflinePushUserIDs = offlinePushUserIDs
			}
			resp, err := p.conversationRpcClient.Client.GetConversationOfflinePushUserIDs(
				ctx,
				&conversation.GetConversationOfflinePushUserIDsReq{ConversationID: utils.GenGroupConversationID(groupID), UserIDs: needOfflinePushUserIDs},
			)
			if err != nil {
				return err
			}
			if len(resp.UserIDs) > 0 {
				err = p.offlinePushMsg(ctx, groupID, msg, resp.UserIDs)
				if err != nil {
					log.ZError(ctx, "offlinePushMsg failed", err, "groupID", groupID, "msg", msg)
					return err
				}
				if _, err := p.GetConnsAndOnlinePush(ctx, msg, utils.IntersectString(resp.UserIDs, webAndPcBackgroundUserIDs)); err != nil {
					log.ZError(ctx, "offlinePushMsg failed", err, "groupID", groupID, "msg", msg, "userIDs", utils.IntersectString(needOfflinePushUserIDs, webAndPcBackgroundUserIDs))
					return err
				}
			}
		}
	}
	return nil
}

func (p *Pusher) GetConnsAndOnlinePush(ctx context.Context, msg *sdkws.MsgData, pushToUserIDs []string) (wsResults []*msggateway.SingleMsgToUserResults, err error) {
	conns, err := p.discov.GetConns(ctx, config.Config.RpcRegisterName.OpenImMessageGatewayName)
	log.ZDebug(ctx, "get gateway conn", "conn length", len(conns))
	if err != nil {
		return nil, err
	}

	var (
		mu         sync.Mutex
		wg         = errgroup.Group{}
		input      = &msggateway.OnlineBatchPushOneMsgReq{MsgData: msg, PushToUserIDs: pushToUserIDs}
		maxWorkers = config.Config.Push.MaxConcurrentWorkers
	)

	if maxWorkers < 3 {
		maxWorkers = 3
	}

	wg.SetLimit(maxWorkers)

	// Online push message
	for _, conn := range conns {
		conn := conn // loop var safe
		wg.Go(func() error {
			msgClient := msggateway.NewMsgGatewayClient(conn)
			reply, err := msgClient.SuperGroupOnlineBatchPushOneMsg(ctx, input)
			if err != nil {
				return nil
			}

			log.ZDebug(ctx, "push result", "reply", reply)
			if reply != nil && reply.SinglePushResult != nil {
				mu.Lock()
				wsResults = append(wsResults, reply.SinglePushResult...)
				mu.Unlock()
			}

			return nil
		})
	}

	_ = wg.Wait()

	// always return nil
	return wsResults, nil
}

func (p *Pusher) offlinePushMsg(ctx context.Context, conversationID string, msg *sdkws.MsgData, offlinePushUserIDs []string) error {

	for _, userID := range offlinePushUserIDs {
		p.offlinePushMsgToUser(ctx, conversationID, msg, userID)
	}
	return nil
}

func (p *Pusher) offlinePushMsgToUser(ctx context.Context, conversationID string, msg *sdkws.MsgData, userID string) error {

	userPushSetting, err := p.userRpcClient.Client.GetUserSetting(ctx, &user.GetUserSettingReq{UserID: userID})
	if err != nil {
		return err
	}

	//user not allow offline push
	if userPushSetting.GlobalRecvMsgOpt == constant.ReceiveNotPushMessage {
		return nil
	}

	title, content, opts, err := p.getOfflinePushInfos(ctx, conversationID, userID, msg, userPushSetting)
	if err != nil {
		return err
	}

	// pus msg content

	// if msg.ContentType == constant.AtText {
	// 	if utils.Contain(userID, msg.AtUserIDList...) {
	// 		content = constant.ContentType2PushContentI18n[constant.AtText]
	// 	}
	// 	err := p.offlinePusher.Push(ctx, []string{userID}, title, constant.ContentType2PushContentI18n[constant.Common], opts)

	// 	if err != nil {
	// 		prommetrics.MsgOfflinePushFailedCounter.Inc()
	// 		return err
	// 	}
	// 	return nil
	// }

	err = p.offlinePusher.Push(ctx, []string{userID}, title, content, opts)
	if err != nil {
		prommetrics.MsgOfflinePushFailedCounter.Inc()
		return err
	}
	return nil
}

func (p *Pusher) GetOfflinePushOpts(ctx context.Context, msg *sdkws.MsgData) (opts *offlinepush.Opts, err error) {
	opts = &offlinepush.Opts{
		Signal: &offlinepush.Signal{},
		Server: &offlinepush.Server{},
		Msg: &offlinepush.Msg{
			ConversationID: msgprocessor.GetConversationIDByMsg(msg),
			ContentType:    msg.ContentType,
		},
	}

	if msg.SessionType == constant.ServerGroupChatType {
		group, err := p.groupRpcClient.GetGroupInfo(ctx, msg.GroupID)
		if err == nil {
			opts.Server.ServerID = group.ServerID
		}
	}

	if msg.OfflinePushInfo != nil {
		opts.IOSBadgeCount = msg.OfflinePushInfo.IOSBadgeCount
		opts.IOSPushSound = msg.OfflinePushInfo.IOSPushSound
		opts.Ex = msg.OfflinePushInfo.Ex
	}
	return opts, nil
}

func (p *Pusher) getOfflinePushInfos(ctx context.Context, conversationID, userID string, msg *sdkws.MsgData, userPushSetting *user.GetUserSettingResp) (title, content string, opts *offlinepush.Opts, err error) {
	if p.offlinePusher == nil {
		err = errNoOfflinePusher
		return
	}

	opts, err = p.GetOfflinePushOpts(ctx, msg)
	if err != nil {
		return
	}

	if userPushSetting.AllowBeep == constant.NewMsgPushSettingAllowed {
		opts.IOSPushSound = "default"
	}
	if userPushSetting.AllowVibration == constant.NewMsgPushSettingAllowed {

	}

	if msg.OfflinePushInfo != nil {
		title = msg.OfflinePushInfo.Title
		content = msg.OfflinePushInfo.Desc
	}

	language := "zh_CN"
	if userPushSetting.Language != "" {
		language = userPushSetting.Language
	}

	if title == "" {
		var offlineMsg *offlineinfo.OfflineMsg
		lang := i18n.Language(language)
		offlineMsg, err = p.offlineInfoParse.GetOfflineInfo(ctx, msg, userPushSetting.AllowPushContent, lang)
		if err != nil {
			log.ZError(ctx, "getOfflineInfo failed", err, "contentType", msg.ContentType)
			return
		}
		title = offlineMsg.Title
		content = offlineMsg.Content
	}
	return
}

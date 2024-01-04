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

package rpcclient

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	// "google.golang.org/protobuf/proto".
)

const (
	SuperGroupDesignateMaxRecvID = 5
)

func newContentTypeConf() map[int32]config.NotificationConf {
	return map[int32]config.NotificationConf{
		// group
		constant.GroupCreatedNotification:                 config.Config.Notification.GroupCreated,
		constant.GroupInfoSetNotification:                 config.Config.Notification.GroupInfoSet,
		constant.JoinGroupApplicationNotification:         config.Config.Notification.JoinGroupApplication,
		constant.MemberQuitNotification:                   config.Config.Notification.MemberQuit,
		constant.GroupApplicationAcceptedNotification:     config.Config.Notification.GroupApplicationAccepted,
		constant.GroupApplicationRejectedNotification:     config.Config.Notification.GroupApplicationRejected,
		constant.GroupOwnerTransferredNotification:        config.Config.Notification.GroupOwnerTransferred,
		constant.MemberKickedNotification:                 config.Config.Notification.MemberKicked,
		constant.MemberInvitedNotification:                config.Config.Notification.MemberInvited,
		constant.MemberEnterNotification:                  config.Config.Notification.MemberEnter,
		constant.GroupDismissedNotification:               config.Config.Notification.GroupDismissed,
		constant.GroupMutedNotification:                   config.Config.Notification.GroupMuted,
		constant.GroupCancelMutedNotification:             config.Config.Notification.GroupCancelMuted,
		constant.GroupMemberMutedNotification:             config.Config.Notification.GroupMemberMuted,
		constant.GroupMemberCancelMutedNotification:       config.Config.Notification.GroupMemberCancelMuted,
		constant.GroupMemberInfoSetNotification:           config.Config.Notification.GroupMemberInfoSet,
		constant.GroupMemberSetToAdminNotification:        config.Config.Notification.GroupMemberSetToAdmin,
		constant.GroupMemberSetToOrdinaryUserNotification: config.Config.Notification.GroupMemberSetToOrdinary,
		constant.GroupInfoSetAnnouncementNotification:     config.Config.Notification.GroupInfoSetAnnouncement,
		constant.GroupInfoSetNameNotification:             config.Config.Notification.GroupInfoSetName,
		// user
		constant.UserInfoUpdatedNotification:  config.Config.Notification.UserInfoUpdated,
		constant.UserStatusChangeNotification: config.Config.Notification.UserStatusChanged,
		// friend
		constant.FriendApplicationNotification:         config.Config.Notification.FriendApplicationAdded,
		constant.FriendApplicationApprovedNotification: config.Config.Notification.FriendApplicationApproved,
		constant.FriendApplicationRejectedNotification: config.Config.Notification.FriendApplicationRejected,
		constant.FriendAddedNotification:               config.Config.Notification.FriendAdded,
		constant.FriendDeletedNotification:             config.Config.Notification.FriendDeleted,
		constant.FriendRemarkSetNotification:           config.Config.Notification.FriendRemarkSet,
		constant.BlackAddedNotification:                config.Config.Notification.BlackAdded,
		constant.BlackDeletedNotification:              config.Config.Notification.BlackDeleted,
		constant.FriendInfoUpdatedNotification:         config.Config.Notification.FriendInfoUpdated,
		// conversation
		constant.ConversationChangeNotification:      config.Config.Notification.ConversationChanged,
		constant.ConversationUnreadNotification:      config.Config.Notification.ConversationChanged,
		constant.ConversationPrivateChatNotification: config.Config.Notification.ConversationSetPrivate,
		// msg
		constant.MsgRevokeNotification:  {IsSendMsg: false, ReliabilityLevel: constant.ReliableNotificationNoMsg},
		constant.HasReadReceipt:         {IsSendMsg: false, ReliabilityLevel: constant.ReliableNotificationNoMsg},
		constant.DeleteMsgsNotification: {IsSendMsg: false, ReliabilityLevel: constant.ReliableNotificationNoMsg},
		// signaling
		constant.SignalingInvitedNotification:               config.Config.Notification.SignalingInvited,
		constant.SignalingGroupInvitedNotification:          {IsSendMsg: true, ReliabilityLevel: constant.UnreliableNotification},
		constant.SignalingAcceptedNotification:              config.Config.Notification.SignalingAccepted,
		constant.SignalingRejectedNotification:              config.Config.Notification.SignalingRejected,
		constant.SignalingSingleChatRejectedNotification:    {IsSendMsg: true, ReliabilityLevel: constant.UnreliableNotification},
		constant.SignalingJoinedNotification:                config.Config.Notification.SignalingJoined,
		constant.SignalingGroupJoinedNotification:           {IsSendMsg: true, ReliabilityLevel: constant.UnreliableNotification},
		constant.SignalingCanceledNotification:              config.Config.Notification.SignalingCanceled,
		constant.SignalingSingleChatCanceledNotification:    {IsSendMsg: true, ReliabilityLevel: constant.UnreliableNotification},
		constant.SignalingHungUpNotification:                config.Config.Notification.SignalingHungUp,
		constant.SignalingClosedNotification:                config.Config.Notification.SignalingClosed,
		constant.SignalingSingleChatClosedNotification:      {IsSendMsg: true, ReliabilityLevel: constant.UnreliableNotification},
		constant.SignalingMicphoneStatusChangedNotification: {IsSendMsg: false, ReliabilityLevel: constant.UnreliableNotification},
		constant.SignalingSpeakStatusChangedNotification:    {IsSendMsg: false, ReliabilityLevel: constant.UnreliableNotification},
		// redPacket
		constant.RedPacketClaimedByUserNotification: config.Config.Notification.RedPacketClaimedByUser,
		constant.RedPacketClaimedNotification:       {IsSendMsg: false, ReliabilityLevel: constant.ReliableNotificationNoMsg},
		constant.RedPacketExpiredNotification:       {IsSendMsg: false, ReliabilityLevel: constant.ReliableNotificationNoMsg},
		// server
		constant.JoinServerApplicationNotification:     config.Config.Notification.JoinServerApplication,
		constant.ServerApplicationAcceptedNotification: config.Config.Notification.ServerApplicationAccepted,
		constant.ServerApplicationRejectedNotification: config.Config.Notification.ServerApplicationRejected,
		constant.ServerCreatedNotification:             config.Config.Notification.ServerApplicationRejected,

		// modifyMsg
		constant.ModifyMessageNotification: {IsSendMsg: false, ReliabilityLevel: constant.ReliableNotificationNoMsg},

		// cron
		constant.CronMsgClearSetNotification: config.Config.Notification.CronMsgClearSet,
	}
}

func newSessionTypeConf() map[int32]int32 {
	return map[int32]int32{
		// group
		constant.GroupCreatedNotification:                 constant.SuperGroupChatType,
		constant.GroupInfoSetNotification:                 constant.SuperGroupChatType,
		constant.JoinGroupApplicationNotification:         constant.SingleChatType,
		constant.MemberQuitNotification:                   constant.SuperGroupChatType,
		constant.GroupApplicationAcceptedNotification:     constant.SingleChatType,
		constant.GroupApplicationRejectedNotification:     constant.SingleChatType,
		constant.GroupOwnerTransferredNotification:        constant.SuperGroupChatType,
		constant.MemberKickedNotification:                 constant.SuperGroupChatType,
		constant.MemberInvitedNotification:                constant.SuperGroupChatType,
		constant.MemberEnterNotification:                  constant.SuperGroupChatType,
		constant.GroupDismissedNotification:               constant.SuperGroupChatType,
		constant.GroupMutedNotification:                   constant.SuperGroupChatType,
		constant.GroupCancelMutedNotification:             constant.SuperGroupChatType,
		constant.GroupMemberMutedNotification:             constant.SuperGroupChatType,
		constant.GroupMemberCancelMutedNotification:       constant.SuperGroupChatType,
		constant.GroupMemberInfoSetNotification:           constant.SuperGroupChatType,
		constant.GroupMemberSetToAdminNotification:        constant.SuperGroupChatType,
		constant.GroupMemberSetToOrdinaryUserNotification: constant.SuperGroupChatType,
		constant.GroupInfoSetAnnouncementNotification:     constant.SuperGroupChatType,
		constant.GroupInfoSetNameNotification:             constant.SuperGroupChatType,
		// user
		constant.UserInfoUpdatedNotification:  constant.SingleChatType,
		constant.UserStatusChangeNotification: constant.SingleChatType,
		// friend
		constant.FriendApplicationNotification:         constant.SingleChatType,
		constant.FriendApplicationApprovedNotification: constant.SingleChatType,
		constant.FriendApplicationRejectedNotification: constant.SingleChatType,
		constant.FriendAddedNotification:               constant.SingleChatType,
		constant.FriendDeletedNotification:             constant.SingleChatType,
		constant.FriendRemarkSetNotification:           constant.SingleChatType,
		constant.BlackAddedNotification:                constant.SingleChatType,
		constant.BlackDeletedNotification:              constant.SingleChatType,
		constant.FriendInfoUpdatedNotification:         constant.SingleChatType,
		// conversation
		constant.ConversationChangeNotification:      constant.SingleChatType,
		constant.ConversationUnreadNotification:      constant.SingleChatType,
		constant.ConversationPrivateChatNotification: constant.SingleChatType,
		// delete
		constant.DeleteMsgsNotification: constant.SingleChatType,
		// server
		constant.JoinServerApplicationNotification:     constant.SingleChatType,
		constant.ServerApplicationAcceptedNotification: constant.SingleChatType,
		constant.ServerApplicationRejectedNotification: constant.SingleChatType,
		constant.ServerCreatedNotification:             constant.ServerGroupChatType,
	}
}

type Message struct {
	conn   grpc.ClientConnInterface
	Client msg.MsgClient
	discov discoveryregistry.SvcDiscoveryRegistry
}

func NewMessage(discov discoveryregistry.SvcDiscoveryRegistry) *Message {
	conn, err := discov.GetConn(context.Background(), config.Config.RpcRegisterName.OpenImMsgName)
	if err != nil {
		panic(err)
	}
	client := msg.NewMsgClient(conn)
	return &Message{discov: discov, conn: conn, Client: client}
}

type MessageRpcClient Message

func NewMessageRpcClient(discov discoveryregistry.SvcDiscoveryRegistry) MessageRpcClient {
	return MessageRpcClient(*NewMessage(discov))
}

func (m *MessageRpcClient) SendMsg(ctx context.Context, req *msg.SendMsgReq) (*msg.SendMsgResp, error) {
	resp, err := m.Client.SendMsg(ctx, req)
	return resp, err
}

func (m *MessageRpcClient) SendSignalMsg(ctx context.Context, req *msg.SendSignalMsgReq) (*msg.SendSignalMsgResp, error) {
	resp, err := m.Client.SendSignalMsg(ctx, req)
	return resp, err
}

func (m *MessageRpcClient) GetMaxSeq(ctx context.Context, req *sdkws.GetMaxSeqReq) (*sdkws.GetMaxSeqResp, error) {
	resp, err := m.Client.GetMaxSeq(ctx, req)
	return resp, err
}

func (m *MessageRpcClient) PullMessageBySeqList(ctx context.Context, req *sdkws.PullMessageBySeqsReq) (*sdkws.PullMessageBySeqsResp, error) {
	resp, err := m.Client.PullMessageBySeqs(ctx, req)
	return resp, err
}

func (m *MessageRpcClient) GetConversationMaxSeq(ctx context.Context, conversationID string) (int64, error) {
	resp, err := m.Client.GetConversationMaxSeq(ctx, &msg.GetConversationMaxSeqReq{ConversationID: conversationID})
	if err != nil {
		return 0, err
	}
	return resp.MaxSeq, nil
}

func (m *MessageRpcClient) ModifyMsg(ctx context.Context, req *msg.ModifyMsgReq) (*msg.ModifyMsgResp, error) {
	resp, err := m.Client.ModifyMsg(ctx, req)
	return resp, err
}

type NotificationSender struct {
	contentTypeConf map[int32]config.NotificationConf
	sessionTypeConf map[int32]int32
	sendMsg         func(ctx context.Context, req *msg.SendMsgReq) (*msg.SendMsgResp, error)
	getUserInfo     func(ctx context.Context, userID string) (*sdkws.UserInfo, error)
}

type NotificationSenderOptions func(*NotificationSender)

func WithLocalSendMsg(sendMsg func(ctx context.Context, req *msg.SendMsgReq) (*msg.SendMsgResp, error)) NotificationSenderOptions {
	return func(s *NotificationSender) {
		s.sendMsg = sendMsg
	}
}

func WithRpcClient(msgRpcClient *MessageRpcClient) NotificationSenderOptions {
	return func(s *NotificationSender) {
		s.sendMsg = msgRpcClient.SendMsg
	}
}

func WithUserRpcClient(userRpcClient *UserRpcClient) NotificationSenderOptions {
	return func(s *NotificationSender) {
		s.getUserInfo = userRpcClient.GetUserInfo
	}
}

func NewNotificationSender(opts ...NotificationSenderOptions) *NotificationSender {
	notificationSender := &NotificationSender{contentTypeConf: newContentTypeConf(), sessionTypeConf: newSessionTypeConf()}
	for _, opt := range opts {
		opt(notificationSender)
	}
	return notificationSender
}

type notificationOpt struct {
	WithRpcGetUsername  bool
	WithDesignateUserID []string
}

type NotificationOptions func(*notificationOpt)

func WithRpcGetUserName() NotificationOptions {
	return func(opt *notificationOpt) {
		opt.WithRpcGetUsername = true
	}
}

func WithDesignateUserID(userID ...string) NotificationOptions {
	return func(opt *notificationOpt) {
		opt.WithDesignateUserID = userID
	}
}

func (s *NotificationSender) NotificationWithSesstionType(ctx context.Context, sendID, recvID string, contentType, sesstionType int32, m proto.Message, opts ...NotificationOptions) (err error) {
	n := sdkws.NotificationElem{Detail: utils.StructToJsonString(m)}
	content, err := json.Marshal(&n)
	if err != nil {
		log.ZError(ctx, "MsgClient Notification json.Marshal failed", err, "sendID", sendID, "recvID", recvID, "contentType", contentType, "msg", m)
		return err
	}
	notificationOpt := &notificationOpt{}
	for _, opt := range opts {
		opt(notificationOpt)
	}
	var req msg.SendMsgReq
	var msg sdkws.MsgData
	var userInfo *sdkws.UserInfo
	if notificationOpt.WithRpcGetUsername && s.getUserInfo != nil {
		userInfo, err = s.getUserInfo(ctx, sendID)
		if err != nil {
			log.ZWarn(ctx, "getUserInfo failed", err, "sendID", sendID)
		} else {
			msg.SenderNickname = userInfo.Nickname
			msg.SenderFaceURL = userInfo.FaceURL
		}
	}
	if len(notificationOpt.WithDesignateUserID) < SuperGroupDesignateMaxRecvID {
		msg.RecvIDList = notificationOpt.WithDesignateUserID
	}
	var offlineInfo sdkws.OfflinePushInfo
	msg.SendID = sendID
	msg.RecvID = recvID
	msg.Content = content
	msg.MsgFrom = constant.SysMsgType
	msg.ContentType = contentType
	msg.SessionType = sesstionType
	if msg.SessionType == constant.SuperGroupChatType || msg.SessionType == constant.ServerGroupChatType {
		msg.GroupID = recvID
	}
	msg.CreateTime = utils.GetCurrentTimestampByMill()
	msg.ClientMsgID = utils.GetMsgID(sendID)
	optionsConfig := s.contentTypeConf[contentType]
	if sendID == recvID && contentType == constant.HasReadReceipt {
		optionsConfig.ReliabilityLevel = constant.UnreliableNotification
	}
	options := config.GetOptionsByNotification(optionsConfig)
	msg.Options = options
	if options.IsOfflinePush() {
		offlineInfo.Title = optionsConfig.OfflinePush.Title
		offlineInfo.Desc = optionsConfig.OfflinePush.Desc
		offlineInfo.Ex = optionsConfig.OfflinePush.Ext
		msg.OfflinePushInfo = &offlineInfo
	}
	req.MsgData = &msg
	_, err = s.sendMsg(ctx, &req)
	if err == nil {
		log.ZDebug(ctx, "MsgClient Notification SendMsg success", "req", &req)
	} else {
		log.ZError(ctx, "MsgClient Notification SendMsg failed", err, "req", &req)
	}
	return err
}

func (s *NotificationSender) Notification(ctx context.Context, sendID, recvID string, contentType int32, m proto.Message, opts ...NotificationOptions) error {
	return s.NotificationWithSesstionType(ctx, sendID, recvID, contentType, s.sessionTypeConf[contentType], m, opts...)
}

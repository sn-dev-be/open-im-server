package msg

import (
	"context"

	"github.com/OpenIMSDK/protocol/constant"
	pbmsg "github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"

	"google.golang.org/protobuf/proto"
)

func (m *msgServer) SendSignalMsg(
	ctx context.Context,
	req *pbmsg.SendSignalMsgReq,
) (resp *pbmsg.SendSignalMsgResp, error error) {
	if req.SignalData != nil {
		signal := req.SignalData
		return m.handleVoiceSignal(ctx, signal)
	} else {
		return nil, errs.ErrArgs.Wrap("signalData is nil")
	}
}

func (m *msgServer) handleVoiceSignal(
	ctx context.Context,
	signalMsg *sdkws.SignalData,
) (*pbmsg.SendSignalMsgResp, error) {
	req := sdkws.SignalVoiceReq{}
	err := utils.JsonStringToStruct(string(signalMsg.Content), &req)
	if err != nil {
		return nil, errs.ErrArgs.Wrap("signalVoiceReq format err")
	}
	switch signalMsg.SignalType {
	case constant.SignalingInviation:
		return m.invitationNotification(ctx, &req)
	case constant.SignalingAccept:
		return m.acceptNotification(ctx, &req)
	case constant.SignalingReject:
		return m.rejectNotification(ctx, &req)
	case constant.SignalingJoin:
		return m.joinNotification(ctx, &req)
	case constant.SignalingCancel:
		return m.cancelNotification(ctx, &req)
	case constant.SignalingHungUp:
		return m.hungUpNotification(ctx, &req)
	case constant.SignalingClose:
		return m.closeNotification(ctx, &req)
	case constant.SignalingMicphoneStatusChange:
		return m.micphoneStatusChangeNotification(ctx, &req)
	case constant.SignalingSpeakStatusChange:
		return m.speakStatusChangeNotification(ctx, &req)
	default:
		return nil, errs.ErrArgs.Wrap("unknown signalAction")
	}
}

func (m *msgServer) invitationNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	usersID := []string{req.FromUserID}
	inviteUsersID := append(usersID, req.InviteUsersID...)
	if err := m.MsgDatabase.CreateVoiceChannel(ctx, req.ChannelID, inviteUsersID); err != nil {
		return nil, err
	}
	if req.SessionType == constant.SuperGroupChatType {
		opUsers, err := m.User.GetPublicUserInfos(ctx, []string{req.FromUserID}, true)
		if err != nil {
			return nil, err
		}
		tips := sdkws.SignalGroupVoiceCardTips{
			ChannelID:  req.ChannelID,
			OpUsers:    opUsers,
			Status:     constant.VoiceChannelEnabled,
			CreateTime: utils.GetCurrentTimestampByMill(),
		}
		m.notificationSender.Notification(ctx, req.FromUserID, req.GroupID, constant.SignalingGroupInvitedNotification, &tips)
	}
	if err := m.Cron.SetCloseVoiceChannelJob(ctx, req.FromUserID, req.ChannelID, req.GroupID, req.SessionType); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingInvitedNotification)
}

func (m *msgServer) acceptNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingAcceptedNotification)
}

func (m *msgServer) rejectNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		tips := &sdkws.SignalVoiceSingleChatTips{
			OpUserID: req.FromUserID,
		}
		m.broadcastSingleChatNotification(ctx, req, constant.SignalingSingleChatRejectedNotification, tips)
	}
	if err := m.broadcastNotification(ctx, req, constant.SignalingRejectedNotification); err != nil {
		return nil, err
	}
	if err := m.MsgDatabase.RemoveUserFromVoiceChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, nil
}

func (m *msgServer) joinNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if err := m.MsgDatabase.AddUserToVoiceChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	if req.SessionType == constant.SuperGroupChatType {
		m.notificationSender.Notification(ctx, req.FromUserID, req.GroupID, constant.SignalingGroupJoinedNotification, nil)
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingJoinedNotification)
}

func (m *msgServer) cancelNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		tips := &sdkws.SignalVoiceSingleChatTips{
			OpUserID: req.FromUserID,
		}
		m.broadcastSingleChatNotification(ctx, req, constant.SignalingSingleChatCanceledNotification, tips)
	}
	if err := m.broadcastNotification(ctx, req, constant.SignalingCanceledNotification); err != nil {
		return nil, err
	}
	if err := m.MsgDatabase.RemoveUserFromVoiceChannel(ctx, req.ChannelID, req.InviteUsersID[0]); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, nil
}

func (m *msgServer) hungUpNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		_, elapsedSec, err := m.MsgDatabase.GetVoiceChannelDuration(ctx, req.ChannelID)
		if err != nil {
			return nil, err
		}
		tips := &sdkws.SignalVoiceSingleChatTips{
			ElapsedSec: int32(elapsedSec),
			OpUserID:   req.FromUserID,
		}
		m.broadcastSingleChatNotification(ctx, req, constant.SignalingSingleChatClosedNotification, tips)
	}
	if err := m.broadcastNotification(ctx, req, constant.SignalingHungUpNotification); err != nil {
		return nil, err
	}
	if err := m.MsgDatabase.RemoveUserFromVoiceChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, nil
}

func (m *msgServer) closeNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastSingleChatNotification(ctx, req, constant.SignalingSingleChatClosedNotification, nil)
	}
	if err := m.broadcastNotification(ctx, req, constant.SignalingClosedNotification); err != nil {
		return nil, err
	}
	if err := m.MsgDatabase.DelVoiceChannel(ctx, req.ChannelID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, nil
}

func (m *msgServer) micphoneStatusChangeNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	opUsers, err := m.User.GetPublicUserInfos(ctx, []string{req.FromUserID}, true)
	if err != nil {
		return nil, err
	}
	tips := &sdkws.SignalVoiceMicphoneStatusTips{
		ChannelID:      req.ChannelID,
		OpUser:         opUsers[0],
		MicphoneStatus: uint32(req.MicphoneStatus),
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotificationWithTips(ctx, req, constant.SignalingMicphoneStatusChangedNotification, tips)
}

func (m *msgServer) speakStatusChangeNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	opUsers, err := m.User.GetPublicUserInfos(ctx, []string{req.FromUserID}, true)
	if err != nil {
		return nil, err
	}
	tips := &sdkws.SignalVoiceSpeakStatusTips{
		ChannelID: req.ChannelID,
		OpUser:    opUsers[0],
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotificationWithTips(ctx, req, constant.SignalingSpeakStatusChangedNotification, tips)
}

func (m *msgServer) broadcastNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
	notificationType int32,
) error {
	tips, err := m.getSignalVoiceCommonTips(ctx, req.FromUserID, req.ChannelID, req.ConversationID)
	if err != nil {
		return err
	}
	return m.broadcastNotificationWithTips(ctx, req, notificationType, tips)
}

func (m *msgServer) broadcastNotificationWithTips(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
	notificationType int32,
	tips proto.Message,
) error {
	usersID, err := m.MsgDatabase.GetVoiceChannelUsersID(ctx, req.ChannelID, req.FromUserID)
	if err != nil {
		return err
	}
	for _, userID := range usersID {
		if err := m.notificationSender.NotificationWithSesstionType(
			ctx,
			req.FromUserID,
			userID,
			notificationType,
			constant.SingleChatType,
			tips,
		); err != nil {
			continue
		}
	}
	return nil
}

func (m *msgServer) broadcastSingleChatNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
	notificationType int32,
	tips proto.Message,
) error {
	createUserID, err := m.MsgDatabase.GetVoiceChannelCreateUserID(ctx, req.ChannelID)
	if err != nil {
		return err
	}
	usersID, err := m.MsgDatabase.GetVoiceChannelUsersID(ctx, req.ChannelID, createUserID)
	if err != nil {
		return err
	}
	if err := m.notificationSender.NotificationWithSesstionType(
		ctx,
		createUserID,
		usersID[0],
		notificationType,
		constant.SingleChatType,
		tips,
		rpcclient.WithRpcGetUserName(),
	); err != nil {
		return err
	}
	return nil
}

func (m *msgServer) getSignalVoiceCommonTips(
	ctx context.Context,
	fromUserID,
	channelID,
	conversationID string,
) (*sdkws.SignalVoiceTips, error) {
	opUsers, err := m.User.GetPublicUserInfos(ctx, []string{fromUserID}, true)
	if err != nil {
		return nil, err
	}
	remainingSec, elapsedSec, err := m.MsgDatabase.GetVoiceChannelDuration(ctx, channelID)
	if err != nil {
		return nil, err
	}
	userIDs, err := m.MsgDatabase.GetVoiceChannelUsersID(ctx, channelID, "")
	if err != nil {
		return nil, err
	}

	participants, err := m.User.GetPublicUserInfos(ctx, userIDs, true)
	if err != nil {
		return nil, err
	}
	tips := &sdkws.SignalVoiceTips{
		ChannelID:      channelID,
		OpUser:         opUsers[0],
		RemainingSec:   int32(remainingSec),
		ElapsedSec:     int32(elapsedSec),
		Participants:   participants,
		ConversationID: conversationID,
	}
	return tips, nil
}

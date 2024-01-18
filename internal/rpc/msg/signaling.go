package msg

import (
	"context"

	"github.com/OpenIMSDK/protocol/constant"
	pbmsg "github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/utils"

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
	inviteUsersID := append(req.InviteUsersID, req.FromUserID)
	err := m.MsgDatabase.CreateVoiceChannel(ctx, req.ChannelID, inviteUsersID)
	if err != nil {
		return nil, err
	}
	if req.SessionType == constant.SuperGroupChatType {
		user, err := m.getOperatorUserInfo(ctx, req.FromUserID)
		if err != nil {
			return nil, err
		}
		voiceTips := sdkws.SignalVoiceGroupTips{
			ChannelID:  req.ChannelID,
			OpUsers:    []*sdkws.PublicUserInfo{user},
			Status:     constant.VoiceCallRoomEnabled,
			CreateTime: utils.GetCurrentTimestampByMill(),
		}
		m.broadcastNotificationWithSessionType(
			ctx,
			req,
			constant.SuperGroupChatType,
			constant.SignalingGroupInvitedNotification,
			req.GroupID,
			&voiceTips,
		)
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
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatRejectedNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingRejectedNotification)

	if err := m.MsgDatabase.RemoveUserFromVoiceChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) joinNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if err := m.MsgDatabase.AddUserToVoiceChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	if req.SessionType == constant.SuperGroupChatType {
		m.broadcastNotificationWithSessionType(
			ctx,
			req,
			constant.SuperGroupChatType,
			constant.SignalingGroupJoinedNotification,
			req.GroupID,
			nil,
		)
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingJoinedNotification)
}

func (m *msgServer) cancelNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatCanceledNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingCanceledNotification)

	if err := m.MsgDatabase.RemoveUserFromVoiceChannel(ctx, req.ChannelID, req.InviteUsersID[0]); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) hungUpNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatClosedNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingHungUpNotification)

	if err := m.MsgDatabase.RemoveUserFromVoiceChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) closeNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatClosedNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingClosedNotification)

	if err := m.MsgDatabase.DelVoiceChannel(ctx, req.ChannelID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) micphoneStatusChangeNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingMicphoneStatusChangedNotification)
}

func (m *msgServer) speakStatusChangeNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
) (*pbmsg.SendSignalMsgResp, error) {
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingSpeakStatusChangedNotification)
}

func (m *msgServer) broadcastNotification(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
	notificationType int32,
) error {
	return m.broadcastNotificationWithSessionType(ctx, req, constant.SingleChatType, notificationType, "", nil)
}

func (m *msgServer) broadcastNotificationWithSessionType(
	ctx context.Context,
	req *sdkws.SignalVoiceReq,
	sessionType int32,
	notificationType int32,
	groupID string,
	tips proto.Message,
) error {
	voiceTips := &sdkws.SignalVoiceTips{}
	if tips == nil {
		tips, err := m.getSignalVoiceTips(ctx, req.FromUserID, req.ChannelID, req.MicphoneStatus)
		if err != nil {
			return err
		}
		voiceTips = tips
	}
	// 单聊给语音房除了操作者外都广播通知
	if sessionType == constant.SingleChatType {
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
				voiceTips,
			); err != nil {
				continue
			}
		}
	}
	if sessionType == constant.SuperGroupChatType {
		if err := m.notificationSender.NotificationWithSesstionType(
			ctx,
			req.FromUserID,
			groupID,
			notificationType,
			constant.GroupChatType,
			voiceTips,
		); err != nil {
			return err
		}
	}
	return nil
}

func (m *msgServer) getSignalVoiceTips(
	ctx context.Context,
	fromUserID,
	channelID string,
	micphoneStatus int32,
) (*sdkws.SignalVoiceTips, error) {
	user, err := m.getOperatorUserInfo(ctx, fromUserID)
	if err != nil {
		return nil, err
	}
	remainingSec, elapsedSec, err := m.MsgDatabase.GetVoiceChannelDuration(ctx, channelID)
	if err != nil {
		return nil, err
	}
	tips := sdkws.SignalVoiceTips{
		ChannelID:      channelID,
		User:           user,
		MicphoneStatus: micphoneStatus,
		RemainingSec:   int32(remainingSec),
		ElapsedSec:     int32(elapsedSec),
	}
	return &tips, nil
}

func (m *msgServer) getOperatorUserInfo(
	ctx context.Context,
	userID string,
) (*sdkws.PublicUserInfo, error) {
	user, err := m.User.GetPublicUserInfo(ctx, userID)
	if err != nil {
		log.ZError(ctx, "GetPublicUserInfo err", err, "userID", userID)
		return nil, err
	}
	return user, err
}

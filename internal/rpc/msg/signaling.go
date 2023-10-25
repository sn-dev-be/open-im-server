package msg

import (
	"context"
	"fmt"

	"github.com/OpenIMSDK/protocol/constant"
	pbmsg "github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/utils"

	"google.golang.org/protobuf/proto"
)

func (m *msgServer) SendSignalMsg(ctx context.Context, req *pbmsg.SendSignalMsgReq) (resp *pbmsg.SendSignalMsgResp, error error) {
	resp = &pbmsg.SendSignalMsgResp{}

	if req.SignalData != nil {
		signalMsg := req.SignalData
		signalReq := sdkws.SignalReq{}
		err := utils.JsonStringToStruct(string(signalMsg.Content), &signalReq)
		if err != nil {
			return nil, errs.ErrArgs.Wrap("signalReq format err")
		}
		switch signalMsg.SignalType {
		case constant.SignalingInviation:
			return m.invitationNotification(ctx, &signalReq)
		case constant.SignalingAccept:
			return m.acceptNotification(ctx, &signalReq)
		case constant.SignalingReject:
			return m.rejectNotification(ctx, &signalReq)
		case constant.SignalingJoin:
			return m.joinNotification(ctx, &signalReq)
		case constant.SignalingCancel:
			return m.cancelNotification(ctx, &signalReq)
		case constant.SignalingHungUp:
			return m.hungUpNotification(ctx, &signalReq)
		case constant.SignalingClose:
			return m.closeNotification(ctx, &signalReq)
		case constant.SignalingMicphoneStatusChange:
			return m.micphoneStatusChangeNotification(ctx, &signalReq)
		case constant.SignalingSpeakStatusChange:
			return m.speakStatusChangeNotification(ctx, &signalReq)
		default:
			return nil, errs.ErrArgs.Wrap("unknown signalAction")
		}
	} else {
		return nil, errs.ErrArgs.Wrap("signalData is nil")
	}
}

func (m *msgServer) invitationNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {

	inviteUsersID := append(req.InviteUsersID, req.FromUserID)
	err := m.MsgDatabase.CreateVoiceCallChannel(ctx, req.ChannelID, inviteUsersID)
	if err != nil {
		return nil, err
	}

	if req.SessionType == constant.SuperGroupChatType {
		user, err := m.getOperatorUserInfo(ctx, req.FromUserID)
		if err != nil {
			return nil, err
		}

		voiceCallElem := sdkws.SignalVoiceCallElem{
			ChannelID:  req.ChannelID,
			OpUsers:    []*sdkws.PublicUserInfo{user},
			Status:     constant.VoiceCallRoomEnabled,
			CreateTime: utils.GetCurrentTimestampByMill(),
		}
		m.broadcastNotificationWithSessionType(ctx, req, constant.SuperGroupChatType, constant.SignalingGroupInvitedNotification, req.GroupID, &voiceCallElem)
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingInvitedNotification)
}

func (m *msgServer) acceptNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingAcceptedNotification)
}

func (m *msgServer) rejectNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatRejectedNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingRejectedNotification)

	if err := m.MsgDatabase.DelUserFromVoiceCallChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) joinNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	if err := m.MsgDatabase.AddUserToVoiceCallChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}

	if req.SessionType == constant.SuperGroupChatType {
		m.broadcastNotificationWithSessionType(ctx, req, constant.SuperGroupChatType, constant.SignalingGroupJoinedNotification, req.GroupID, nil)
	}
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingJoinedNotification)
}

func (m *msgServer) cancelNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatCanceledNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingCanceledNotification)

	if err := m.MsgDatabase.DelUserFromVoiceCallChannel(ctx, req.ChannelID, req.InviteUsersID[0]); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) hungUpNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatClosedNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingHungUpNotification)

	if err := m.MsgDatabase.DelUserFromVoiceCallChannel(ctx, req.ChannelID, req.FromUserID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) closeNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	if req.SessionType == constant.SingleChatType {
		m.broadcastNotification(ctx, req, constant.SignalingSingleChatClosedNotification)
	}
	err := m.broadcastNotification(ctx, req, constant.SignalingClosedNotification)

	if err := m.MsgDatabase.DelVoiceCallChannel(ctx, req.ChannelID); err != nil {
		return nil, err
	}
	return &pbmsg.SendSignalMsgResp{}, err
}

func (m *msgServer) micphoneStatusChangeNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingMicphoneStatusChangedNotification)
}

func (m *msgServer) speakStatusChangeNotification(ctx context.Context, req *sdkws.SignalReq) (*pbmsg.SendSignalMsgResp, error) {
	return &pbmsg.SendSignalMsgResp{}, m.broadcastNotification(ctx, req, constant.SignalingSpeakStatusChangedNotification)
}

func (m *msgServer) broadcastNotification(
	ctx context.Context,
	req *sdkws.SignalReq,
	notificationType int32,
) error {
	return m.broadcastNotificationWithSessionType(ctx, req, constant.SingleChatType, notificationType, "", nil)
}

func (m *msgServer) broadcastNotificationWithSessionType(
	ctx context.Context,
	req *sdkws.SignalReq,
	sessionType int32,
	notificationType int32,
	groupID string,
	tips proto.Message,
) error {

	signalRespMsg := &sdkws.SignalResp{}

	if tips == nil {
		tips, err := m.getSignalRespMsg(ctx, req.FromUserID, req.ChannelID, req.MicphoneStatus)
		if err != nil {
			return err
		}
		signalRespMsg = tips
	}

	// 单聊给语音房除了操作者外都广播通知
	if sessionType == constant.SingleChatType {
		usersID, err := m.MsgDatabase.GetVoiceCallChannelUsersID(ctx, req.ChannelID, req.FromUserID)
		if err != nil {
			return err
		}

		for _, userID := range usersID {
			err := m.notificationSender.NotificationWithSesstionType(
				ctx,
				req.FromUserID,
				userID,
				notificationType,
				constant.SingleChatType,
				signalRespMsg,
			)
			if err != nil {
				return err
			}
		}
	}

	if sessionType == constant.SuperGroupChatType {
		err := m.notificationSender.NotificationWithSesstionType(
			ctx,
			req.FromUserID,
			groupID,
			notificationType,
			constant.GroupChatType,
			signalRespMsg,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *msgServer) getSignalRespMsg(ctx context.Context, fromUserID, channelID string, micphoneStatus int32) (*sdkws.SignalResp, error) {
	user, err := m.getOperatorUserInfo(ctx, fromUserID)
	if err != nil {
		return nil, err
	}

	remainingSeconds, elapsedSeconds, err := m.MsgDatabase.GetVoiceCallChannelDuration(ctx, channelID)
	if err != nil {
		return nil, err
	}

	tips := sdkws.SignalResp{
		ChannelID:        channelID,
		User:             user,
		MicphoneStatus:   micphoneStatus,
		RemainingSeconds: int32(remainingSeconds),
		ElapsedSeconds:   int32(elapsedSeconds),
	}
	return &tips, nil
}

func (m *msgServer) getOperatorUserInfo(ctx context.Context, userID string) (*sdkws.PublicUserInfo, error) {
	user, err := m.User.GetPublicUserInfo(ctx, userID)
	if err != nil {
		log.ZError(ctx, "GetPublicUserInfo err", err, "userID", userID)
		return nil, errs.ErrUserIDNotFound.Wrap(fmt.Sprintf("user %s not found", userID))
	}
	return user, err
}

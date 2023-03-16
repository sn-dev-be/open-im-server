package msg

import (
	"context"
	cbapi "github.com/OpenIMSDK/Open-IM-Server/pkg/callbackstruct"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/config"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/constant"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/http"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/tracelog"
	pbChat "github.com/OpenIMSDK/Open-IM-Server/pkg/proto/msg"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/utils"
)

func cbURL() string {
	return config.Config.Callback.CallbackUrl
}

func toCommonCallback(ctx context.Context, msg *pbChat.SendMsgReq, command string) cbapi.CommonCallbackReq {
	return cbapi.CommonCallbackReq{
		SendID:           msg.MsgData.SendID,
		ServerMsgID:      msg.MsgData.ServerMsgID,
		CallbackCommand:  command,
		ClientMsgID:      msg.MsgData.ClientMsgID,
		OperationID:      tracelog.GetOperationID(ctx),
		SenderPlatformID: msg.MsgData.SenderPlatformID,
		SenderNickname:   msg.MsgData.SenderNickname,
		SessionType:      msg.MsgData.SessionType,
		MsgFrom:          msg.MsgData.MsgFrom,
		ContentType:      msg.MsgData.ContentType,
		Status:           msg.MsgData.Status,
		CreateTime:       msg.MsgData.CreateTime,
		AtUserIDList:     msg.MsgData.AtUserIDList,
		SenderFaceURL:    msg.MsgData.SenderFaceURL,
		Content:          utils.GetContent(msg.MsgData),
		Seq:              uint32(msg.MsgData.Seq),
		Ex:               msg.MsgData.Ex,
	}
}

func CallbackBeforeSendSingleMsg(ctx context.Context, msg *pbChat.SendMsgReq) error {
	if !config.Config.Callback.CallbackBeforeSendSingleMsg.Enable {
		return nil
	}
	req := &cbapi.CallbackBeforeSendSingleMsgReq{
		CommonCallbackReq: toCommonCallback(ctx, msg, constant.CallbackBeforeSendSingleMsgCommand),
		RecvID:            msg.MsgData.RecvID,
	}
	resp := &cbapi.CallbackBeforeSendSingleMsgResp{}
	return http.CallBackPostReturn(cbURL(), req, resp, config.Config.Callback.CallbackBeforeSendSingleMsg)
}

func CallbackAfterSendSingleMsg(ctx context.Context, msg *pbChat.SendMsgReq) error {
	if !config.Config.Callback.CallbackAfterSendSingleMsg.Enable {
		return nil
	}
	req := &cbapi.CallbackAfterSendSingleMsgReq{
		CommonCallbackReq: toCommonCallback(ctx, msg, constant.CallbackBeforeSendSingleMsgCommand),
		RecvID:            msg.MsgData.RecvID,
	}
	resp := &cbapi.CallbackAfterSendSingleMsgResp{}
	return http.CallBackPostReturn(cbURL(), req, resp, config.Config.Callback.CallbackAfterSendSingleMsg)
}

func CallbackBeforeSendGroupMsg(ctx context.Context, msg *pbChat.SendMsgReq) error {
	if !config.Config.Callback.CallbackAfterSendSingleMsg.Enable {
		return nil
	}
	req := &cbapi.CallbackAfterSendGroupMsgReq{
		CommonCallbackReq: toCommonCallback(ctx, msg, constant.CallbackBeforeSendSingleMsgCommand),
		GroupID:           msg.MsgData.GroupID,
	}
	resp := &cbapi.CallbackBeforeSendGroupMsgResp{}
	return http.CallBackPostReturn(cbURL(), req, resp, config.Config.Callback.CallbackAfterSendSingleMsg)
}

func CallbackAfterSendGroupMsg(ctx context.Context, msg *pbChat.SendMsgReq) error {
	if !config.Config.Callback.CallbackAfterSendGroupMsg.Enable {
		return nil
	}
	req := &cbapi.CallbackAfterSendGroupMsgReq{
		CommonCallbackReq: toCommonCallback(ctx, msg, constant.CallbackAfterSendGroupMsgCommand),
		GroupID:           msg.MsgData.GroupID,
	}
	resp := &cbapi.CallbackAfterSendGroupMsgResp{}
	return http.CallBackPostReturn(cbURL(), req, resp, config.Config.Callback.CallbackAfterSendGroupMsg)
}

func CallbackMsgModify(ctx context.Context, msg *pbChat.SendMsgReq) error {
	if !config.Config.Callback.CallbackMsgModify.Enable || msg.MsgData.ContentType != constant.Text {
		return nil
	}
	req := &cbapi.CallbackMsgModifyCommandReq{
		CommonCallbackReq: toCommonCallback(ctx, msg, constant.CallbackMsgModifyCommand),
	}
	resp := &cbapi.CallbackMsgModifyCommandResp{}
	if err := http.CallBackPostReturn(cbURL(), req, resp, config.Config.Callback.CallbackAfterSendGroupMsg); err != nil {
		return err
	}
	if resp.Content != nil {
		msg.MsgData.Content = []byte(*resp.Content)
	}
	utils.NotNilReplace(msg.MsgData.OfflinePushInfo, resp.OfflinePushInfo)
	utils.NotNilReplace(&msg.MsgData.RecvID, resp.RecvID)
	utils.NotNilReplace(&msg.MsgData.GroupID, resp.GroupID)
	utils.NotNilReplace(&msg.MsgData.ClientMsgID, resp.ClientMsgID)
	utils.NotNilReplace(&msg.MsgData.ServerMsgID, resp.ServerMsgID)
	utils.NotNilReplace(&msg.MsgData.SenderPlatformID, resp.SenderPlatformID)
	utils.NotNilReplace(&msg.MsgData.SenderNickname, resp.SenderNickname)
	utils.NotNilReplace(&msg.MsgData.SenderFaceURL, resp.SenderFaceURL)
	utils.NotNilReplace(&msg.MsgData.SessionType, resp.SessionType)
	utils.NotNilReplace(&msg.MsgData.MsgFrom, resp.MsgFrom)
	utils.NotNilReplace(&msg.MsgData.ContentType, resp.ContentType)
	utils.NotNilReplace(&msg.MsgData.Status, resp.Status)
	utils.NotNilReplace(&msg.MsgData.Options, resp.Options)
	utils.NotNilReplace(&msg.MsgData.AtUserIDList, resp.AtUserIDList)
	utils.NotNilReplace(&msg.MsgData.MsgDataList, resp.MsgDataList)
	utils.NotNilReplace(&msg.MsgData.AttachedInfo, resp.AttachedInfo)
	utils.NotNilReplace(&msg.MsgData.Ex, resp.Ex)
	return nil
}

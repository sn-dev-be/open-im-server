package api

import (
	"OpenIM/internal/api/a2r"
	"OpenIM/internal/apiresp"
	"OpenIM/pkg/apistruct"
	"OpenIM/pkg/common/config"
	"OpenIM/pkg/common/constant"
	"OpenIM/pkg/common/log"
	"OpenIM/pkg/common/tracelog"
	"OpenIM/pkg/discoveryregistry"
	"OpenIM/pkg/errs"
	"OpenIM/pkg/proto/msg"
	"OpenIM/pkg/proto/sdkws"
	"OpenIM/pkg/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"
)

var _ context.Context // 解决goland编辑器bug

func NewMsg(c discoveryregistry.SvcDiscoveryRegistry) *Msg {
	return &Msg{c: c, validate: validator.New()}
}

type Msg struct {
	c        discoveryregistry.SvcDiscoveryRegistry
	validate *validator.Validate
}

func (Msg) SetOptions(options map[string]bool, value bool) {
	utils.SetSwitchFromOptions(options, constant.IsHistory, value)
	utils.SetSwitchFromOptions(options, constant.IsPersistent, value)
	utils.SetSwitchFromOptions(options, constant.IsSenderSync, value)
	utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, value)
}

func (m Msg) newUserSendMsgReq(c *gin.Context, params *apistruct.ManagementSendMsgReq) *msg.SendMsgReq {
	var newContent string
	var err error
	switch params.ContentType {
	case constant.Text:
		newContent = params.Content["text"].(string)
	case constant.Picture:
		fallthrough
	case constant.Custom:
		fallthrough
	case constant.Voice:
		fallthrough
	case constant.Video:
		fallthrough
	case constant.File:
		fallthrough
	case constant.CustomNotTriggerConversation:
		fallthrough
	case constant.CustomOnlineOnly:
		fallthrough
	case constant.AdvancedRevoke:
		newContent = utils.StructToJsonString(params.Content)
	case constant.Revoke:
		newContent = params.Content["revokeMsgClientID"].(string)
	default:
	}
	options := make(map[string]bool, 5)
	if params.IsOnlineOnly {
		m.SetOptions(options, false)
	}
	if params.NotOfflinePush {
		utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
	}
	if params.ContentType == constant.CustomOnlineOnly {
		m.SetOptions(options, false)
	} else if params.ContentType == constant.CustomNotTriggerConversation {
		utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, false)
	}
	pbData := msg.SendMsgReq{
		MsgData: &sdkws.MsgData{
			SendID:           params.SendID,
			GroupID:          params.GroupID,
			ClientMsgID:      utils.GetMsgID(params.SendID),
			SenderPlatformID: params.SenderPlatformID,
			SenderNickname:   params.SenderNickname,
			SenderFaceURL:    params.SenderFaceURL,
			SessionType:      params.SessionType,
			MsgFrom:          constant.SysMsgType,
			ContentType:      params.ContentType,
			Content:          []byte(newContent),
			RecvID:           params.RecvID,
			CreateTime:       utils.GetCurrentTimestampByMill(),
			Options:          options,
			OfflinePushInfo:  params.OfflinePushInfo,
		},
	}
	if params.ContentType == constant.OANotification {
		var tips sdkws.TipsComm
		tips.JsonDetail = utils.StructToJsonString(params.Content)
		pbData.MsgData.Content, err = proto.Marshal(&tips)
		if err != nil {
			log.Error(tracelog.GetOperationID(c), "Marshal failed ", err.Error(), tips.String())
		}
	}
	return &pbData
}

func (m *Msg) client() (msg.MsgClient, error) {
	conn, err := m.c.GetConn(config.Config.RpcRegisterName.OpenImMsgName)
	if err != nil {
		return nil, err
	}
	return msg.NewMsgClient(conn), nil
}

func (m *Msg) GetSeq(c *gin.Context) {
	a2r.Call(msg.MsgClient.GetMaxAndMinSeq, m.client, c)
}

func (m *Msg) PullMsgBySeqs(c *gin.Context) {
	a2r.Call(msg.MsgClient.PullMessageBySeqs, m.client, c)
}

func (m *Msg) DelMsg(c *gin.Context) {
	a2r.Call(msg.MsgClient.DelMsgs, m.client, c)
}

func (m *Msg) DelSuperGroupMsg(c *gin.Context) {
	a2r.Call(msg.MsgClient.DelSuperGroupMsg, m.client, c)
}

func (m *Msg) ClearMsg(c *gin.Context) {
	a2r.Call(msg.MsgClient.ClearMsg, m.client, c)
}

func (m *Msg) SetMessageReactionExtensions(c *gin.Context) {
	a2r.Call(msg.MsgClient.SetMessageReactionExtensions, m.client, c)
}

func (m *Msg) GetMessageListReactionExtensions(c *gin.Context) {
	a2r.Call(msg.MsgClient.GetMessagesReactionExtensions, m.client, c)
}

func (m *Msg) AddMessageReactionExtensions(c *gin.Context) {
	a2r.Call(msg.MsgClient.AddMessageReactionExtensions, m.client, c)
}

func (m *Msg) DeleteMessageReactionExtensions(c *gin.Context) {
	a2r.Call(msg.MsgClient.DeleteMessageReactionExtensions, m.client, c)
}

func (m *Msg) SendMsg(c *gin.Context) {
	params := apistruct.ManagementSendMsgReq{}
	if err := c.BindJSON(&params); err != nil {
		apiresp.GinError(c, err)
		return
	}
	var data interface{}
	switch params.ContentType {
	case constant.Text:
		data = apistruct.TextElem{}
	case constant.Picture:
		data = apistruct.PictureElem{}
	case constant.Voice:
		data = apistruct.SoundElem{}
	case constant.Video:
		data = apistruct.VideoElem{}
	case constant.File:
		data = apistruct.FileElem{}
	case constant.Custom:
		data = apistruct.CustomElem{}
	case constant.Revoke:
		data = apistruct.RevokeElem{}
	case constant.AdvancedRevoke:
		data = apistruct.MessageRevoked{}
	case constant.OANotification:
		data = apistruct.OANotificationElem{}
		params.SessionType = constant.NotificationChatType
	case constant.CustomNotTriggerConversation:
		data = apistruct.CustomElem{}
	case constant.CustomOnlineOnly:
		data = apistruct.CustomElem{}
	//case constant.HasReadReceipt:
	//case constant.Typing:
	//case constant.Quote:
	default:
		apiresp.GinError(c, errors.New("wrong contentType"))
		return
	}
	if err := mapstructure.WeakDecode(params.Content, &data); err != nil {
		apiresp.GinError(c, errs.ErrData)
		return
	} else if err := m.validate.Struct(data); err != nil {
		apiresp.GinError(c, errs.ErrData)
		return
	}
	switch params.SessionType {
	case constant.SingleChatType:
		if len(params.RecvID) == 0 {
			apiresp.GinError(c, errs.ErrData)
			return
		}
	case constant.GroupChatType, constant.SuperGroupChatType:
		if len(params.GroupID) == 0 {
			apiresp.GinError(c, errs.ErrData)
			return
		}
	}
	pbReq := m.newUserSendMsgReq(c, &params)
	conn, err := m.c.GetConn(config.Config.RpcRegisterName.OpenImMsgName)
	if err != nil {
		apiresp.GinError(c, errs.ErrInternalServer)
		return
	}
	client := msg.NewMsgClient(conn)
	var status int
	respPb, err := client.SendMsg(c, pbReq)
	if err != nil {
		status = constant.MsgSendFailed
		apiresp.GinError(c, err)
		return
	}
	status = constant.MsgSendSuccessed
	_, err = client.SetSendMsgStatus(c, &msg.SetSendMsgStatusReq{
		Status: int32(status),
	})
	if err != nil {
		log.NewError(tracelog.GetOperationID(c), "SetSendMsgStatus failed")
	}
	resp := apistruct.ManagementSendMsgResp{ResultList: sdkws.UserSendMsgResp{ServerMsgID: respPb.ServerMsgID, ClientMsgID: respPb.ClientMsgID, SendTime: respPb.SendTime}}
	apiresp.GinSuccess(c, resp)
}

func (m *Msg) ManagementBatchSendMsg(c *gin.Context) {
	a2r.Call(msg.MsgClient.SendMsg, m.client, c)
}

func (m *Msg) CheckMsgIsSendSuccess(c *gin.Context) {
	a2r.Call(msg.MsgClient.GetSendMsgStatus, m.client, c)
}

func (m *Msg) GetUsersOnlineStatus(c *gin.Context) {
	a2r.Call(msg.MsgClient.GetSendMsgStatus, m.client, c)
}

func (m *Msg) AccountCheck(c *gin.Context) {
	a2r.Call(msg.MsgClient.GetSendMsgStatus, m.client, c)
}
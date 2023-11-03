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

package msg

import (
	"context"
	"encoding/json"

	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"

	"github.com/OpenIMSDK/protocol/constant"
	msgv3 "github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"

	"github.com/OpenIMSDK/tools/utils"
)

func (m *msgServer) SetRedPacketMsgStatus(ctx context.Context, req *msgv3.SetRedPacketMsgStatusReq) (*msgv3.SetRedPacketMsgStatusResp, error) {
	defer log.ZDebug(ctx, "SetRedPacketMsgStatus return line")
	if req.UserID == "" {
		return nil, errs.ErrArgs.Wrap("user_id is empty")
	}
	if req.ConversationID == "" {
		return nil, errs.ErrArgs.Wrap("conversation_id is empty")
	}
	if req.Seq < 0 {
		return nil, errs.ErrArgs.Wrap("seq is invalid")
	}

	if !authverify.IsAppManagerUid(ctx) {
		return nil, errs.ErrNoPermission.Wrap(utils.GetSelfFuncName())
	}

	_, _, msgs, err := m.MsgDatabase.GetMsgBySeqs(ctx, req.UserID, req.ConversationID, []int64{req.Seq})
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 || msgs[0] == nil {
		return nil, errs.ErrRecordNotFound.Wrap("msg not found")
	}

	data, _ := json.Marshal(msgs[0])
	log.ZInfo(ctx, "GetMsgBySeqs", "conversationID", req.ConversationID, "seq", req.Seq, "msg", string(data))
	msg := msgs[0]

	if req.ContentType == constant.RedPacketExpiredNotification || req.ContentType == constant.RedPacketClaimedNotification {
		elem := sdkws.RedPacketElem{}
		utils.JsonStringToStruct(string(msg.Content), &elem)
		elem.Status = req.Status

		err = m.MsgDatabase.RenewRedPacketMsg(ctx, req.ConversationID, req.Seq, utils.StructToJsonString(&elem))
		if err != nil {
			return nil, err
		}
	}

	tips := sdkws.RedPacketTips{
		ClientMsgID:    msg.ClientMsgID,
		Seq:            req.Seq,
		ConversationID: req.ConversationID,
		RedPacketID:    req.RedPacketID,
		Status:         req.Status,
		ContentType:    req.ContentType,
	}

	if req.ContentType == constant.RedPacketClaimedByUserNotification {
		user, err := m.User.GetPublicUserInfo(ctx, req.ClaimUserID)
		if err != nil {
			return nil, err
		}
		tips.ClaimUser = user
	}

	var recvID string
	if msg.SessionType == constant.SuperGroupChatType {
		recvID = msg.GroupID
	} else {
		recvID = msg.RecvID
	}
	if err := m.notificationSender.NotificationWithSesstionType(ctx, req.UserID, recvID, req.ContentType, msg.SessionType, &tips, rpcclient.WithRpcGetUserName()); err != nil {
		return nil, err
	}
	return &msgv3.SetRedPacketMsgStatusResp{}, nil
}

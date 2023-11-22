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
	"time"

	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"

	"github.com/OpenIMSDK/protocol/constant"
	msgv3 "github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
)

func (m *msgServer) ModifyMsg(ctx context.Context, req *msgv3.ModifyMsgReq) (*msgv3.ModifyMsgResp, error) {
	defer log.ZDebug(ctx, "ModifyMsg return line")
	if req.UserID == "" {
		return nil, errs.ErrArgs.Wrap("user_id is empty")
	}
	if req.ConversationID == "" {
		return nil, errs.ErrArgs.Wrap("conversation_id is empty")
	}
	if req.Seq < 0 {
		return nil, errs.ErrArgs.Wrap("seq is invalid")
	}
	if err := authverify.CheckAccessV3(ctx, req.UserID); err != nil {
		return nil, err
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

	err = m.MsgDatabase.ModifyMsgBySeq(ctx, req.ConversationID, req.Seq, req.Content)
	if err != nil {
		return nil, err
	}

	tips := &sdkws.ModifyMessageTips{
		ClientMsgID:    msg.ClientMsgID,
		Seq:            msg.Seq,
		ConversationID: req.ConversationID,
		OpUser:         mcontext.GetOpUserID(ctx),
		ModifyTime:     time.Now().UnixMilli(),
		Content:        req.Content,
		ModifyType:     req.ModifyType,
	}

	var recvID string
	if msg.SessionType == constant.SuperGroupChatType {
		recvID = msg.GroupID
	} else {
		recvID = msg.RecvID
	}
	if err := m.notificationSender.NotificationWithSesstionType(
		ctx,
		req.UserID,
		recvID,
		constant.ModifyMessageNotification,
		msg.SessionType,
		tips,
		rpcclient.WithRpcGetUserName(),
	); err != nil {
		return nil, err
	}
	return &msgv3.ModifyMsgResp{}, nil
}

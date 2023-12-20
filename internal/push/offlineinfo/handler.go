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

package offlineinfo

import (
	"context"

	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
)

type SingleChatMsgHandler struct{}
type GroupMsgHandler struct{ rpc }
type ServerGroupMsgHandler struct{ rpc }

func (h SingleChatMsgHandler) Msg(ctx context.Context, msg *sdkws.MsgData) (*OfflineMsg, error) {
	info := &OfflineMsg{Title: msg.SenderNickname}
	switch msg.ContentType {
	case constant.RedPacket:
		info.Content = constant.ContentType2PushContentI18n[constant.RedPacket]
	case constant.FriendApplicationApprovedNotification:
		info.Content = config.Config.Notification.FriendApplicationApproved.OfflinePush.Desc
	default:
		info.Content = constant.ContentType2PushContentI18n[constant.Common]
	}
	return info, nil
}

func (h GroupMsgHandler) Msg(ctx context.Context, msg *sdkws.MsgData) (*OfflineMsg, error) {
	info := &OfflineMsg{}
	groupInfo, err := h.groupRpcClient.GetGroupInfo(ctx, msg.GroupID)
	if err != nil {
		log.ZError(ctx, "offline info GetGroupInfo failed", err)
		return nil, err
	}
	info.Title = groupInfo.GroupName
	switch msg.ContentType {
	case constant.RedPacket:
		info.Content = constant.ContentType2PushContentI18n[constant.RedPacket]
	case constant.AtText:
		info.Content = constant.ContentType2PushContentI18n[constant.AtText]
	default:
		info.Content = constant.ContentType2PushContentI18n[constant.Common]
	}
	return info, nil
}

func (h ServerGroupMsgHandler) Msg(ctx context.Context, msg *sdkws.MsgData) (*OfflineMsg, error) {
	info := &OfflineMsg{}
	groupInfo, err := h.groupRpcClient.GetGroupInfo(ctx, msg.GroupID)
	if err != nil {
		log.ZError(ctx, "offline info GetGroupInfo failed", err)
		return nil, err
	}
	switch msg.ContentType {
	case constant.RedPacket:
		info.Title = groupInfo.GroupName
		info.Content = constant.ContentType2PushContentI18n[constant.RedPacket]
	case constant.AtText:
		info.Title = groupInfo.GroupName + "[部落]"
		loginUserID := mcontext.GetOpUserID(ctx)
		if utils.IsContain(loginUserID, msg.AtUserIDList) {
			info.Content = constant.ContentType2PushContentI18n[constant.AtText]
		}
	default:
		info.Title = groupInfo.GroupName + "[部落]"
		info.Content = constant.ContentType2PushContentI18n[constant.Common]
	}
	return info, nil
}

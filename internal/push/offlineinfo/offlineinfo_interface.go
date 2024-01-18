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

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/openimsdk/open-im-server/v3/pkg/common/i18n"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

type OfflineMsg struct {
	Title   string
	Content string
}

// type OfflineMsg struct {
// 	Title   string
// 	Content string
// }

type rpc struct {
	groupRpcClient *rpcclient.GroupRpcClient
	clubRpcClient  *rpcclient.ClubRpcClient
}

type OfflineInfo interface {
	Msg(ctx context.Context, msg *sdkws.MsgData, lang i18n.Language, pushContentMode int32) (*OfflineMsg, error)
}

type OfflineInfoParse struct {
	groupRpcClient *rpcclient.GroupRpcClient
	clubRpcClient  *rpcclient.ClubRpcClient
}

func NewOfflineInfoParse(groupRpcClient *rpcclient.GroupRpcClient, clubRpcClient *rpcclient.ClubRpcClient) *OfflineInfoParse {
	return &OfflineInfoParse{
		groupRpcClient: groupRpcClient,
		clubRpcClient:  clubRpcClient,
	}
}

func (o *OfflineInfoParse) GetOfflineInfo(ctx context.Context, msg *sdkws.MsgData, pushContentMode int32, lang i18n.Language) (*OfflineMsg, error) {
	var info OfflineInfo
	rpc := rpc{
		groupRpcClient: o.groupRpcClient,
		clubRpcClient:  o.clubRpcClient,
	}
	switch msg.SessionType {
	case constant.SingleChatType:
		info = SingleChatMsgHandler{rpc}
	case constant.SuperGroupChatType:
		info = GroupMsgHandler{rpc}
	case constant.ServerGroupChatType:
		info = ServerGroupMsgHandler{rpc}
	default:
		info = SingleChatMsgHandler{}
	}
	return info.Msg(ctx, msg, lang, pushContentMode)
}

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
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

type OfflineMsg struct {
	Title   string
	Content string
}

type rpc struct {
	groupRpcClient *rpcclient.GroupRpcClient
}

type OfflineInfo interface {
	Msg(ctx context.Context, msg *sdkws.MsgData) (*OfflineMsg, error)
}

func GetOfflineInfo(ctx context.Context, msg *sdkws.MsgData, groupRpcClient *rpcclient.GroupRpcClient) (*OfflineMsg, error) {
	var info OfflineInfo
	rpc := rpc{
		groupRpcClient: groupRpcClient,
	}
	switch msg.SessionType {
	case constant.SingleChatType:
		info = SingleChatMsgHandler{}
	case constant.SuperGroupChatType:
		info = GroupMsgHandler{rpc}
	case constant.ServerGroupChatType:
		info = ServerGroupMsgHandler{rpc}
	default:
		info = SingleChatMsgHandler{}
	}
	return info.Msg(ctx, msg)
}

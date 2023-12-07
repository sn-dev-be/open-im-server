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
)

type OfflineMsg struct {
	Title   string
	Content string
}

type OfflineInfo interface {
	Msg(ctx context.Context, conversationID string, msg *sdkws.MsgData) (*OfflineMsg, error)
}

func GetOfflineInfo(ctx context.Context, conversationID string, msg *sdkws.MsgData) (*OfflineMsg, error) {
	var info OfflineInfo
	switch msg.ContentType {
	case constant.Text:
		info = TextMsgHandler{}
	case constant.Voice:
		info = VoiceMsgHandler{}
	default:
		info = CommonMsgHandler{}
	}
	return info.Msg(ctx, conversationID, msg)
}

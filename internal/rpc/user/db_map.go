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

package user

import (
	"context"

	"github.com/OpenIMSDK/protocol/user"
)

func UpdateUserInfoMap(ctx context.Context, user *user.SetUserSettingReq) map[string]any {
	m := make(map[string]any)
	if user.GlobalRecvMsgOpt != nil {
		m["global_recv_msg_opt"] = user.GlobalRecvMsgOpt.Value
	}
	if user.AllowBeep != nil {
		m["allow_beep"] = user.AllowBeep.Value
	}
	if user.AllowVibration != nil {
		m["allow_vibration"] = user.AllowVibration.Value
	}
	if user.AllowPushContent != nil {
		m["allow_push_content"] = user.AllowPushContent.Value
	}
	if user.AllowOnlinePush != nil {
		m["allow_online_push"] = user.AllowOnlinePush.Value
	}
	if user.Language != nil {
		m["language"] = user.Language.Value
	}
	if user.AllowStrangerMsg != nil {
		m["allow_stranger_msg"] = user.AllowStrangerMsg.Value
	}

	return m
}

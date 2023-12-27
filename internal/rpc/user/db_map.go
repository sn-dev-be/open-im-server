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

	"github.com/OpenIMSDK/protocol/sdkws"
)

func UserSettingInfoMap(ctx context.Context, userSetting *sdkws.UserSettingForSet) map[string]any {
	m := make(map[string]any)
	if userSetting.NewMsgPushMode != nil {
		m["new_msg_push_mode"] = userSetting.NewMsgPushMode.Value
	}
	if userSetting.NewMsgPushDetailMode != nil {
		m["new_msg_push_detail_mode"] = userSetting.NewMsgPushDetailMode.Value
	}
	if userSetting.NewMsgVoiceMode != nil {
		m["new_msg_voice_mode"] = userSetting.NewMsgVoiceMode.Value
	}
	if userSetting.NewMsgShakeMode != nil {
		m["new_msg_shake_mode"] = userSetting.NewMsgShakeMode.Value
	}
	if userSetting.Ex != nil {
		m["ex"] = userSetting.Ex.Value
	}
	return m
}

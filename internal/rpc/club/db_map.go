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

package club

import (
	"context"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/sdkws"
)

func UpdateServerInfoMap(ctx context.Context, server *sdkws.ServerInfoForSet) map[string]any {
	m := make(map[string]any)
	if server.ServerName != "" {
		m["name"] = server.ServerName
	}
	if server.Icon != "" {
		m["icon"] = server.Icon
	}
	if server.Description != "" {
		m["description"] = server.Description
	}
	if server.Banner != "" {
		m["banner"] = server.Banner
	}
	if server.ApplyMode != nil {
		m["apply_mode"] = server.ApplyMode.Value
	}
	if server.Searchable != nil {
		m["searchable"] = server.Searchable.Value
	}
	if server.UserMutualAccessible != nil {
		m["user_mutual_accessible"] = server.UserMutualAccessible.Value
	}
	if server.Ex != nil {
		m["ex"] = server.Ex.Value
	}
	if server.DappID != nil {
		m["dapp_id"] = server.DappID.Value
	}
	return m
}

func UpdateServerStatusMap(status int) map[string]any {
	return map[string]any{
		"status": status,
	}
}

func UpdateServerMemberMutedTimeMap(t time.Time) map[string]any {
	return map[string]any{
		"mute_end_time": t,
	}
}

func UpdateServerMemberMap(req *pbclub.SetServerMemberInfo) map[string]any {
	m := make(map[string]any)
	if req.Nickname != nil {
		m["nickname"] = req.Nickname.Value
	}
	if req.FaceURL != nil {
		m["user_server_face_url"] = req.FaceURL.Value
	}
	if req.RoleLevel != nil {
		m["role_level"] = req.RoleLevel.Value
	}
	if req.Ex != nil {
		m["ex"] = req.Ex.Value
	}
	return m
}

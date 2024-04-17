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
	"github.com/OpenIMSDK/tools/mcontext"
)

func UpdateServerInfoMap(ctx context.Context, server *sdkws.ServerInfoForSet) map[string]any {
	m := make(map[string]any)
	if server.ServerName != nil {
		m["name"] = server.ServerName.Value
	}
	if server.Icon != nil {
		m["icon"] = server.Icon.Value
	}
	if server.Description != nil {
		m["description"] = server.Description.Value
	}
	if server.Banner != nil {
		m["banner"] = server.Banner.Value
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
	if server.CommunityName != nil {
		m["community_name"] = server.CommunityName.Value
	}
	if server.CommunityBanner != nil {
		m["community_banner"] = server.CommunityBanner.Value
	}
	if server.CommunityViewMode != nil {
		m["community_view_mode"] = server.CommunityViewMode.Value
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

func UpdateGroupCategoryInfoMap(ctx context.Context, categoryName string, reorder_weight int32) map[string]any {
	m := make(map[string]any)
	if categoryName != "" {
		m["name"] = categoryName
	}
	if reorder_weight != 0 {
		m["reorder_weight"] = reorder_weight
	}
	return m
}

func UpdateGroupInfoMap(ctx context.Context, req *pbclub.SetServerGroupInfoReq) map[string]any {
	m := make(map[string]any)
	group := req.GroupInfo
	if group.GroupName != "" {
		m["name"] = group.GroupName
	}
	if group.Notification != "" {
		m["notification"] = group.Notification
		m["notification_update_time"] = time.Now()
		m["notification_user_id"] = mcontext.GetOpUserID(ctx)
	}
	if group.Introduction != "" {
		m["introduction"] = group.Introduction
	}
	if group.FaceURL != "" {
		m["face_url"] = group.FaceURL
	}
	if group.ConditionType != 0 {
		m["condition_type"] = group.ConditionType
	}
	if group.Condition != "" {
		m["condition"] = group.Condition
	}
	if group.GroupMode != 0 {
		m["group_mode"] = group.GroupMode
	}
	if req.DappID != "" {
		m["dapp_id"] = req.DappID
	}
	return m
}

func UpdateGroupStatusMap(status int) map[string]any {
	return map[string]any{
		"status": status,
	}
}

func UpdateGroupTreasuryMap(req *pbclub.SetGroupTreasuryReq) map[string]any {
	m := make(map[string]any)
	if req.Info.TreasuryID != "" {
		m["treasury_id"] = req.Info.TreasuryID
	}
	if req.Info.Icon != "" {
		m["icon"] = req.Info.Icon
	}
	if req.Info.Name != "" {
		m["name"] = req.Info.Name
	}
	if req.Info.WalletType != 0 {
		m["wallet_type"] = req.Info.WalletType
	}
	if req.Info.ContractAddress != "" {
		m["contract_address"] = req.Info.ContractAddress
	}
	if req.Info.AdministratorAddress != "" {
		m["administrator_address"] = req.Info.AdministratorAddress
	}
	return m
}

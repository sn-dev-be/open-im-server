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
	"github.com/OpenIMSDK/protocol/sdkws"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) serverDB2PB(server *relation.ServerModel, ownerUserID string, memberCount uint32) *sdkws.ServerInfo {
	return &sdkws.ServerInfo{
		ServerID:             server.ServerID,
		ServerName:           server.ServerName,
		Icon:                 server.Icon,
		Description:          server.Description,
		ApplyMode:            server.ApplyMode,
		InviteMode:           server.InviteMode,
		Searchable:           server.Searchable,
		Status:               server.Status,
		Banner:               server.Banner,
		MemberNumber:         server.MemberNumber,
		UserMutualAccessible: server.UserMutualAccessible,
		CategoryNumber:       server.CategoryNumber,
		ChannelNumber:        server.ChannelNumber,
		OwnerUserID:          ownerUserID,
		CreateTime:           server.CreateTime.UnixMilli(),
	}
}

func (s *clubServer) serverMemberDB2PB(
	member *relation.ServerMemberModel,
	appMangerLevel int32,
) *sdkws.ServerMemberFullInfo {
	return &sdkws.ServerMemberFullInfo{
		ServerID:       member.ServerID,
		UserID:         member.UserID,
		RoleLevel:      member.RoleLevel,
		JoinTime:       member.JoinTime.UnixMilli(),
		Nickname:       member.Nickname,
		FaceURL:        member.FaceURL,
		AppMangerLevel: appMangerLevel,
		JoinSource:     member.JoinSource,
		OperatorUserID: member.OperatorUserID,
		Ex:             member.Ex,
		MuteEndTime:    member.MuteEndTime.UnixMilli(),
		InviterUserID:  member.InviterUserID,
	}
}

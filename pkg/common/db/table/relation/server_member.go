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

package relation

import (
	"context"
	"time"
)

const (
	ServerMemberModelTableName = "server_members"
)

type ServerMemberModel struct {
	ID             uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;UNSIGNED"         json:"id"`
	ServerID       string    `gorm:"column:server_id;primary_key;size:64"                  json:"serverID"`
	UserID         string    `gorm:"column:user_id;primary_key;size:64"                    json:"userID"`
	Nickname       string    `gorm:"column:nickname;index;size:64"                         json:"nickname"`
	FaceURL        string    `gorm:"column:user_server_face_url;size:255"`
	ServerRoleID   string    `gorm:"column:server_role_id;primary_key;size:64"             json:"serverRoleID"`
	RoleLevel      int32     `gorm:"column:role_level"                                     json:"roleLevel"`
	JoinSource     int32     `gorm:"column:join_source"                                    json:"joinSource"`
	InviterUserID  string    `gorm:"column:inviter_user_id;size:64"                        json:"inviterUserID"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	ReorderWeight  int32     `gorm:"column:reorder_weight;default:0"                       json:"reorder_weight"`
	MuteEndTime    time.Time `gorm:"column:mute_end_time"                                  json:"muteEndTime"`
	Ex             string    `gorm:"column:ex;size:255"                                    json:"ex"`
	JoinTime       time.Time `gorm:"column:join_time;index:join_time;autoCreateTime"       json:"joinTime"`
}

func (ServerMemberModel) TableName() string {
	return ServerMemberModelTableName
}

type ServerMemberModelInterface interface {
	NewTx(tx any) ServerMemberModelInterface

	Create(ctx context.Context, serverMembers []*ServerMemberModel) (err error)
	Delete(ctx context.Context, serverID string, userIDs []string) (err error)
	DeleteServer(ctx context.Context, serverIDs []string) (err error)
	Update(ctx context.Context, serverID string, userID string, data map[string]any) (err error)
	UpdateRoleLevel(ctx context.Context, serverID string, userID string, roleLevel int32) (rowsAffected int64, err error)
	Find(
		ctx context.Context,
		serverIDs []string,
		userIDs []string,
		roleLevels []int32,
	) (serverMembers []*ServerMemberModel, err error)
	FindMemberUserID(ctx context.Context, serverID string) (userIDs []string, err error)
	Take(ctx context.Context, serverID string, userID string) (serverMember *ServerMemberModel, err error)
	TakeOwner(ctx context.Context, serverID string) (serverMember *ServerMemberModel, err error)
	SearchMember(
		ctx context.Context,
		keyword string,
		serverIDs []string,
		userIDs []string,
		roleLevels []int32,
		pageNumber, showNumber int32,
	) (total uint32, serverList []*ServerMemberModel, err error)
	MapServerMemberNum(ctx context.Context, serverIDs []string) (count map[string]uint32, err error)
	FindJoinUserID(ctx context.Context, serverIDs []string) (serverUsers map[string][]string, err error)
	FindUserJoinedServerID(ctx context.Context, userID string) (serverIDs []string, err error)
	TakeServerMemberNum(ctx context.Context, serverID string) (count int64, err error)
	FindUsersJoinedServerID(ctx context.Context, userIDs []string) (map[string][]string, error)
	FindUserManagedServerID(ctx context.Context, userID string) (serverIDs []string, err error)
	FindManageRoleUser(ctx context.Context, serverID string, roleIDs []string) (serverMembers []*ServerMemberModel, err error)
	FindLastestJoinedServerMember(ctx context.Context, serverID string, showNumber int32) (serverMembers []*ServerMemberModel, err error)
}

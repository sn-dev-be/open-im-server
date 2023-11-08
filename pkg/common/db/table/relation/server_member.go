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
	ID            uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;UNSIGNED" json:"id"`
	ServerID      string    `gorm:"column:server_id;size:64" json:"serverID"`
	UserID        string    `gorm:"column:user_id;size:64" json:"userID"`
	Nickname      string    `gorm:"column:nickname;size:64" json:"nickname"`
	ServerRoleID  string    `gorm:"column:server_role_id;size:64" json:"serverRoleID"`
	RoleLevel     int32     `gorm:"column:role_level" json:"roleLevel`
	JoinSource    int32     `gorm:"column:join_source" json:"joinSource`
	InviterUserID string    `gorm:"column:inviter_user_id;size:64" json:"inviterUserID"`
	muteEndTime   time.Time `gorm:"column:mute_end_time" json:"muteEndTime"`
	Ex            string    `gorm:"column:ex;size:255" json:"ex"`
	JoinTime      time.Time `gorm:"column:create_time;index:create_time;autoCreateTime" json:"joinTime"`
}

func (ServerMemberModel) TableName() string {
	return ServerMemberModelTableName
}

type ServerMemberModelInterface interface {
	NewTx(tx any) ServerMemberModelInterface
	Create(ctx context.Context, groups []*ServerMemberModel) (err error)
	PageServerMembers(ctx context.Context, showNumber, pageNumber int32, serverID string) (members []*ServerMemberModel, total int64, err error)
	GetServerMembers(ctx context.Context, ids []uint64, serverID string) (members []*ServerMemberModel, err error)
}

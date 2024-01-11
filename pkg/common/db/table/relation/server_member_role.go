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
)

const (
	ServerMemberRoleModelTableName = "server_member_roles"
)

type ServerMemberRoleModel struct {
	ID       uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;UNSIGNED"         json:"id"`
	RoleID   string `gorm:"column:role_id;index:role_id;size:64"                        		  json:"roleID"`
	MemberID uint64 `gorm:"column:member_id;index:member_id;UNSIGNED"         					  json:"memberID"`
}

func (ServerMemberRoleModel) TableName() string {
	return ServerMemberRoleModelTableName
}

type ServerMemberRoleModelInterface interface {
	NewTx(tx any) ServerMemberRoleModelInterface
	Create(ctx context.Context, memberRoles []*ServerMemberRoleModel) (err error)

	DeleteByRole(ctx context.Context, roleIDs []string) error
	DeleteByMember(ctx context.Context, memberIDs []uint64) error

	FindByRoleIDs(ctx context.Context, roleIDs []string) (memberRoles []*ServerMemberRoleModel, err error)
	FindByMemberIDS(ctx context.Context, memberIDs []uint64) (memberRoles []*ServerMemberRoleModel, err error)
}

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

	"gorm.io/datatypes"
)

const (
	ServerRoleModelTableName = "server_roles"
)

type ServerRoleModel struct {
	RoleID       string         `gorm:"column:id;primary_key;size:64"                        json:"roleID"`
	RoleName     string         `gorm:"column:name;size:64"                                  json:"roleName"`
	Icon         string         `gorm:"column:icon;size:64"                                  json:"icon"`
	Type         int32          `gorm:"column:type;default:0"                                json:"type"`
	Priority     int32          `gorm:"column:priority"                                      json:"priority"`
	ServerID     string         `gorm:"column:server_id;size:64"                             json:"serverID"`
	Permissions  datatypes.JSON `gorm:"column:permissions;"                                  json:"permissions"`
	ColorLevel   int32          `gorm:"column:color_level"                                   json:"colorLevel"`
	MemberNumber int32          `gorm:"column:member_number"                                 json:"memberNumber"`
	Ex           string         `gorm:"column:ex;size:255"                                   json:"ex"`
	CreateTime   time.Time      `gorm:"column:create_time;index:create_time;autoCreateTime"  json:"createTime"`
}

func (ServerRoleModel) TableName() string {
	return ServerRoleModelTableName
}

type ServerRoleModelInterface interface {
	NewTx(tx any) ServerRoleModelInterface
	Create(ctx context.Context, serverRoles []*ServerRoleModel) (err error)
	Take(ctx context.Context, serverRoleID string) (serverRole *ServerRoleModel, err error)
	TakeServerRoleByPriority(ctx context.Context, serverID string, priority int32) (serverRole *ServerRoleModel, err error)
	DeleteServer(ctx context.Context, serverIDs []string) error

	FindRoleID(ctx context.Context, serverID string) (serverRoleIDs []string, err error)
	FindDesignationRoleID(ctx context.Context, serverID, roleKey string) (serverRoleIDs []string, err error)
}

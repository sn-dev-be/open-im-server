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
	ServerModelTableName = "servers"
)

type ServerModel struct {
	ServerID             string    `gorm:"column:server_id;primary_key;size:64"                json:"serverID"           binding:"required"`
	ServerName           string    `gorm:"column:name;size:255"                                json:"serverName"`
	Icon                 string    `gorm:"column:icon;size:255"                                json:"icon"`
	Description          string    `gorm:"column:description;size:255"                         json:"description"`
	Banner               string    `gorm:"column:banner;size:255"                              json:"banner"`
	OwnerUserID          string    `gorm:"column:owner_user_id;size:255"                       json:"ownerUserID"`
	MemberNumber         int32     `gorm:"column:memberNumber"                                 json:"memberNumber"`
	ApplyMode            int32     `gorm:"column:applyMode"                                    json:"applyMode"`
	InviteMode           int32     `gorm:"column:inviteMode"                                   json:"inviteMode"`
	Searchable           int32     `gorm:"column:searchable"                                   json:"searchable"`
	UserMutualAccessible int32     `gorm:"column:userMutualAccessible"                         json:"userMutualAccessible"`
	Status               int32     `gorm:"column:status"                                       json:"status"`
	CategoryNumber       int32     `gorm:"column:categoryNumber"                               json:"categoryNumber"`
	ChannelNumber        int32     `gorm:"column:channelNumber"                                json:"channelNumber"`
	Ex                   string    `gorm:"column:ex;size:255"                                  json:"ex"`
	CreateTime           time.Time `gorm:"column:create_time;index:create_time;autoCreateTime" json:"createTime"`
}

func (ServerModel) TableName() string {
	return ServerModelTableName
}

type ServerModelInterface interface {
	NewTx(tx any) ServerModelInterface
	Create(ctx context.Context, servers []*ServerModel) (err error)
	Take(ctx context.Context, serverID string) (server *ServerModel, err error)
	FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*ServerModel, err error)
	FindServersSplit(ctx context.Context, pageNumber, showNumber int32) (servers []*ServerModel, total int64, err error)
	GetServers(ctx context.Context, serverIDs []string) (servers []*ServerModel, err error)
}

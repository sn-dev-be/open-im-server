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
	ServerName           string    `gorm:"column:name;size:255;index"                          json:"serverName"`
	Icon                 string    `gorm:"column:icon;size:255"                                json:"icon"`
	Description          string    `gorm:"column:description;size:255"                         json:"description"`
	Banner               string    `gorm:"column:banner;size:255"                              json:"banner"`
	CreatorUserID        string    `gorm:"column:creator_user_id;size:64"`
	OwnerUserID          string    `gorm:"column:owner_user_id;size:255"                       json:"ownerUserID"`
	MemberNumber         uint32    `gorm:"column:member_number"                                json:"memberNumber"`
	ApplyMode            int32     `gorm:"column:apply_mode"                                   json:"applyMode"`
	InviteMode           int32     `gorm:"column:invite_mode"                                  json:"inviteMode"`
	Searchable           int32     `gorm:"column:searchable"                                   json:"searchable"`
	UserMutualAccessible int32     `gorm:"column:user_mutual_accessible"                       json:"userMutualAccessible"`
	Status               int32     `gorm:"column:status"                                       json:"status"`
	CategoryNumber       uint32    `gorm:"column:category_number"                              json:"categoryNumber"`
	GroupNumber          uint32    `gorm:"column:group_number"                                 json:"groupNumber"`
	DappID               string    `gorm:"column:dapp_id;size:64"                              json:"dappID"`
	Ex                   string    `gorm:"column:ex;size:255"                                  json:"ex"`
	CreateTime           time.Time `gorm:"column:create_time;index:create_time;autoCreateTime" json:"createTime"`
	CommunityName        string    `gorm:"column:community_name;size:64"					   json:"communityName"`
	CommunityBanner      string    `gorm:"column:community_banner;size:255"			 		   json:"communityBanner"`
	CommunityViewMode    int32     `gorm:"column:community_view_mode;"					       json:"communityViewMode"`
}

func (ServerModel) TableName() string {
	return ServerModelTableName
}

type ServerModelInterface interface {
	NewTx(tx any) ServerModelInterface
	Create(ctx context.Context, servers []*ServerModel) (err error)
	Delete(ctx context.Context, serverID string) (err error)
	UpdateMap(ctx context.Context, serverID string, args map[string]interface{}) (err error)
	Take(ctx context.Context, serverID string) (server *ServerModel, err error)
	Search(
		ctx context.Context,
		keyword string,
		pageNumber, showNumber int32,
	) (total uint32, servers []*ServerModel, err error)
	FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*ServerModel, err error)
	GetServers(ctx context.Context, serverIDs []string) (servers []*ServerModel, err error)
}

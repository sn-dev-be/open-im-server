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
	ChannelModelTableName = "channels"
)

type ChannelModel struct {
	ChannelID     string    `gorm:"column:channel_id;primary_key;size:64" json:"channelID"`
	ServerID      string    `gorm:"column:server_id;size:64" json:"serverID"`
	CategoryID    string    `gorm:"column:category_id;size:64" json:"categoryID"`
	ChannelName   string    `gorm:"column:name;size:255"                                json:"channelName"`
	Icon          string    `gorm:"column:icon;size:255"                                json:"icon"`
	Description   string    `gorm:"column:description;size:255" json:"description"`
	OwnerUserID   string    `gorm:"column:owner;size:255" json:"ownerUserID"`
	JoinCondition int32     `gorm:"column:join_condition;default:0" json:"joinCondition"`
	ConditionType int32     `gorm:"column:condition_type;default:0" json:"conditionType"`
	ChannelType   int32     `gorm:"column:channel_type;default:0"                       json:"channelType"`
	ReorderWeight int32     `gorm:"column:reorder_weight;default:0"                     json:"reorderWeight"`
	VisitorMode   int32     `gorm:"column:visitor_mode;default:0"                       json:"visitorMode"`
	ViewMode      int32     `gorm:"column:view_mode;default:0"                          json:"viewMode"`
	Ex            string    `gorm:"column:ex;size:255"                                  json:"ex"`
	CreateTime    time.Time `gorm:"column:create_time;index:create_time;autoCreateTime" json:"createTime"`
}

func (ChannelModel) TableName() string {
	return ChannelModelTableName
}

type ChannelModelInterface interface {
	NewTx(tx any) ChannelModelInterface
	Create(ctx context.Context, groups []*ChannelModel) (err error)
}

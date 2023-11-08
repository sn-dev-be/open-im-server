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
	ChannelMemberModelTableName = "channel_members"
)

type ChannelMemberModel struct {
	ID             uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT;UNSIGNED" json:"id"`
	ChannelID      string    `gorm:"column:channel_id;size:64" json:"channelID"`
	ServerID       string    `gorm:"column:server_id;size:64" json:"serverID"`
	ServerMemberID uint64    `gorm:"column:server_member_id" json:"serverMemberId"`
	Ex             string    `gorm:"column:ex;size:255" json:"ex"`
	CreateTime     time.Time `gorm:"column:create_time;index:create_time;autoCreateTime" json:"createTime"`
}

func (ChannelMemberModel) TableName() string {
	return ChannelMemberModelTableName
}

type ChannelMemberModelInterface interface {
	NewTx(tx any) ChannelMemberModelInterface
	Create(ctx context.Context, channelMembers []*ChannelMemberModel) (err error)
	PageChannelMembers(ctx context.Context, showNumber, pageNumber int32, serverID string) (members []*ChannelMemberModel, total int64, err error)
}

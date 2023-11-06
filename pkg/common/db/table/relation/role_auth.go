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
	RoleAuthModelTableName = "role_auths"
)

type RoleAuthModel struct {
	RoleAuthID            string    `gorm:"column:role_auth_id;primary_key;size:64" json:"roleAuthID"`
	ManageServer          int8      `gorm:"column:manage_server" json:"manageServer"`
	ShareServer           int8      `gorm:"column:share_server" json:"shareServer"`
	ManageMember          int8      `gorm:"column:manage_member" json:"manageMember"`
	SendMsg               int8      `gorm:"column:send_msg" json:"sendMsg"`
	ManageMsg             int8      `gorm:"column:manage_msg" json:"manageMsg"`
	ManageCommunity       int8      `gorm:"column:manage_community" json:"manageCommunity"`
	PostTweet             int8      `gorm:"column:post_tweet" json:"postTweet"`
	TweetReply            int8      `gorm:"column:tweet_reply" json:"tweetReply"`
	ManageChannelCategory int8      `gorm:"column:manage_channel_category" json:"manageChannelCategory"`
	ManageChannel         int8      `gorm:"column:manage_channel" json:"manageChannel"`
	CreateTime            time.Time `gorm:"column:create_time;index:create_time;autoCreateTime" json:"createTime"`
}

func (RoleAuthModel) TableName() string {
	return RoleAuthModelTableName
}

type RoleAuthModelInterface interface {
	NewTx(tx any) RoleAuthModelInterface
	Create(ctx context.Context, groups []*RoleAuthModel) (err error)
}

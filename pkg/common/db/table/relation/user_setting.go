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
	UserSettingTableName = "user_settings"
)

type UserSettingModel struct {
	UserID               string    `gorm:"column:user_id;primary_key;size:64"`
	NewMsgPushMode       int32     `gorm:"column:new_msg_push_mode;"`
	NewMsgPushDetailMode int32     `gorm:"column:new_msg_push_detail_mode;"`
	NewMsgVoiceMode      int32     `gorm:"column:new_msg_voice_mode;"`
	NewMsgShakeMode      int32     `gorm:"column:new_msg_shake_mode;"`
	Ex                   string    `gorm:"column:ex;size:1024"`
	CreateTime           time.Time `gorm:"column:create_time;index:create_time;autoCreateTime"`
}

type UserSettingModelInterface interface {
	Create(ctx context.Context, settings []*UserSettingModel) (err error)
	UpdateByMap(ctx context.Context, userID string, args map[string]interface{}) (err error)
	Update(ctx context.Context, user *UserSettingModel) (err error)
	// 获取指定用户信息  不存在，也不返回错误
	Find(ctx context.Context, userIDs []string) (users []*UserSettingModel, err error)
	// 获取某个用户信息  不存在，则返回错误
	Take(ctx context.Context, userID string) (user *UserSettingModel, err error)
}

func (UserSettingModel) TableName() string {
	return UserSettingTableName
}

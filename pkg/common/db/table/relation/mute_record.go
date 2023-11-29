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
	MuteRecordModelTableName = "mute_records"
)

type MuteRecordModel struct {
	ServerID       string    `gorm:"column:server_id;primary_key;size:64"`
	BlockUserID    string    `gorm:"column:block_user_id;primary_key;size:64"`
	CreateTime     time.Time `gorm:"column:create_time"`
	MuteEndTime    time.Time `gorm:"column:mute_end_time"`
	AddSource      int32     `gorm:"column:add_source"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	Ex             string    `gorm:"column:ex;size:1024"`
}

func (MuteRecordModel) TableName() string {
	return MuteRecordModelTableName
}

type MuteRecordModelInterface interface {
	NewTx(tx any) MuteRecordModelInterface
	Create(ctx context.Context, mute_records []*MuteRecordModel) (err error)
	Delete(ctx context.Context, mute_records []*MuteRecordModel) (err error)
	DeleteByUserIDs(ctx context.Context, serverID string, userIDs []string) (err error)
	Take(ctx context.Context, blockUserID string, serverID string) (muteRecord *MuteRecordModel, err error)
	FindServerMuteRecords(ctx context.Context, serverID string, pageNumber, showNumber int32) (mute_records []*MuteRecordModel, total int64, err error)
}

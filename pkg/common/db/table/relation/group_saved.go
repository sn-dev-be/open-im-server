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
	GroupSavedModelTableName = "group_saved"
)

type GroupSavedModel struct {
	GroupID    string    `gorm:"column:group_id;primary_key;size:64"`
	UserID     string    `gorm:"column:user_id;primary_key;size:64"`
	CreateTime time.Time `gorm:"column:create_time;index:create_time;autoCreateTime"`
}

func (GroupSavedModel) TableName() string {
	return GroupSavedModelTableName
}

type GroupSavedModelInterface interface {
	NewTx(tx any) GroupSavedModelInterface
	Create(ctx context.Context, groups_saved *GroupSavedModel) (err error)
	FindByUser(ctx context.Context, userID string) (groups_saved []*GroupSavedModel, total int64, err error)
	Take(ctx context.Context, groupID string, userID string) (groupSaved *GroupSavedModel, err error)
	Delete(ctx context.Context, groupID string, userID string) (err error)
}

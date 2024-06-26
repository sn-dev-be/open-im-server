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
	GroupDappModelTableName = "group_dapps"
)

type GroupDappModel struct {
	ID         int64     `gorm:"column:id;primary_key;AUTO_INCREMENT;UNSIGNED"        json:"id"`
	GroupID    string    `gorm:"column:group_id;primary_key;size:64"                  json:"groupID"`
	DappID     string    `gorm:"column:dapp_id;primary_key;size:64"                   json:"dappID"`
	CreateTime time.Time `gorm:"column:create_time;index:create_time;autoCreateTime"  json:"createTime"`
}

func (GroupDappModel) TableName() string {
	return GroupDappModelTableName
}

type GroupDappModellInterface interface {
	NewTx(tx any) GroupDappModellInterface
	Create(ctx context.Context, groups []*GroupDappModel) (err error)
	TakeGroupDapp(ctx context.Context, groupID string) (groupDapp *GroupDappModel, err error)
	UpdateByMap(ctx context.Context, groupID string, data map[string]interface{}) (err error)
	DeleteServer(ctx context.Context, serverIDs []string) error
	DeleteByGroup(ctx context.Context, groupIDs []string) error
}

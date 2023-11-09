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
	ServerRecommendedTableName = "server_recommendeds"
)

type ServerRecommendedModel struct {
	ID             int32     `gorm:"column:id;primary_key;AUTO_INCREMENT"                	   json:"id"`
	ServerID       string    `gorm:"column:server_id;size:255"                                 json:"serverID"`
	ReorderWeight  string    `gorm:"column:reorder_weight;size:255"                            json:"reorderWeight"`
	OperatedUserID string    `gorm:"column:operated_user_id;size:255" 						   json:"operatedUserID"`
	CreateTime     time.Time `gorm:"column:create_time;index:create_time;autoCreateTime"       json:"createTime"`
}

func (ServerRecommendedModel) TableName() string {
	return ServerRecommendedTableName
}

type ServerRecommendedModelInterface interface {
	NewTx(tx any) ServerRecommendedModelInterface
	GetServerRecommendedList(ctx context.Context) (servers []*ServerRecommendedModel, err error)
}

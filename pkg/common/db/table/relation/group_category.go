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
	GroupCategoryModelTableName = "group_categories"
)

type GroupCategoryModel struct {
	CategoryID    string    `gorm:"column:category_id;primary_key;size:64"      json:"categoryID"           binding:"required"`
	CategoryName  string    `gorm:"column:name;size:255"                                    json:"categoryName"`
	ReorderWeight int32     `gorm:"column:reorder_weight"                                   json:"reorderWeight"`
	ViewMode      int32     `gorm:"column:view_mode"                                        json:"viewMode"`
	CategoryType  int32     `gorm:"column:category_type;default:1"                          json:"categoryType"`
	ServerID      string    `gorm:"column:server_id;primary_key;size:255"                   json:"serverID" binding:"required"`
	Ex            string    `gorm:"column:ex;size:255"                                      json:"ex"`
	CreateTime    time.Time `gorm:"column:create_time;index:create_time;autoCreateTime"     json:"createTime"`
}

func (GroupCategoryModel) TableName() string {
	return GroupCategoryModelTableName
}

type GroupCategoryModelInterface interface {
	NewTx(tx any) GroupCategoryModelInterface
	Create(ctx context.Context, groupCategories []*GroupCategoryModel) (err error)
	Take(ctx context.Context, groupCategoryID string) (groupCategory *GroupCategoryModel, err error)
	Delete(ctx context.Context, categoryIDs []string) error
	Find(ctx context.Context, groupCategoryIDs []string) (categories []*GroupCategoryModel, err error)
	FindGroupCategoryByType(ctx context.Context, serverID string, categoryType int32) (categories []*GroupCategoryModel, err error)
	GetGroupCategoriesByServerID(ctx context.Context, serverID string) (categories []*GroupCategoryModel, err error)
	DeleteServer(ctx context.Context, serverIDs []string) error
	UpdateMap(ctx context.Context, serverID string, categoryID string, args map[string]interface{}) (err error)
}

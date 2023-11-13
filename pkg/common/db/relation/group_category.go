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

	"gorm.io/gorm"

	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

var _ relation.GroupCategoryModelInterface = (*GroupCategoryGorm)(nil)

type GroupCategoryGorm struct {
	*MetaDB
}

func (s *GroupCategoryGorm) GetGroupCategoriesByServerID(ctx context.Context, serverID string) (categories []*relation.GroupCategoryModel, err error) {
	return categories, utils.Wrap(s.DB.Where("server_id = ?", serverID).Find(&categories).Error, "")
}

func NewGroupCategoryDB(db *gorm.DB) relation.GroupCategoryModelInterface {
	return &GroupCategoryGorm{NewMetaDB(db, &relation.GroupCategoryModel{})}
}

func (s *GroupCategoryGorm) NewTx(tx any) relation.GroupCategoryModelInterface {
	return &GroupCategoryGorm{NewMetaDB(tx.(*gorm.DB), &relation.GroupCategoryModel{})}
}

func (s *GroupCategoryGorm) Create(ctx context.Context, servers []*relation.GroupCategoryModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *GroupCategoryGorm) Take(ctx context.Context, groupCategoryID string) (groupCategory *relation.GroupCategoryModel, err error) {
	groupCategory = &relation.GroupCategoryModel{}
	return groupCategory, utils.Wrap(s.DB.Where("category_id = ?", groupCategoryID).Take(groupCategory).Error, "")
}

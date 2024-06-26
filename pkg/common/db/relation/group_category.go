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

func (s *GroupCategoryGorm) Delete(ctx context.Context, categoryIDs []string) error {
	return utils.Wrap(s.db(ctx).Where("category_id in (?)", categoryIDs).Delete(&relation.GroupCategoryModel{}).Error, "")
}

func (s *GroupCategoryGorm) DeleteServer(ctx context.Context, serverIDs []string) (err error) {
	return utils.Wrap(s.db(ctx).Where("server_id in (?)", serverIDs).Delete(&relation.GroupCategoryModel{}).Error, "")
}

func (s *GroupCategoryGorm) UpdateMap(ctx context.Context, serverID string, categoryID string, args map[string]interface{}) (err error) {
	return utils.Wrap(s.DB.Where("server_id = ? and category_id = ?", serverID, categoryID).Model(&relation.GroupCategoryModel{}).Updates(args).Error, "")
}

func (s *GroupCategoryGorm) Find(ctx context.Context, groupCategoryIDs []string) (categories []*relation.GroupCategoryModel, err error) {
	return categories, utils.Wrap(s.DB.Where("category_id in ?", groupCategoryIDs).Find(&categories).Error, "")
}

func (s *GroupCategoryGorm) FindGroupCategoryIDsByType(ctx context.Context, serverID string, categoryType int32) (categoryID []string, err error) {
	return categoryID, utils.Wrap(s.DB.Model(&relation.GroupCategoryModel{}).Where("server_id = ? and category_type = ?", serverID, categoryType).Order("reorder_weight asc").Pluck("category_id", &categoryID).Error, "")
}

func (s *GroupCategoryGorm) FindGroupCategoryIDsByServerID(ctx context.Context, serverID string) (categoryID []string, err error) {
	return categoryID, utils.Wrap(s.DB.Model(&relation.GroupCategoryModel{}).Where("server_id = ?", serverID).Order("reorder_weight asc").Pluck("category_id", &categoryID).Error, "")
}

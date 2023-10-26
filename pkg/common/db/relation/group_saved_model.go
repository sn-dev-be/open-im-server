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
	"errors"

	"gorm.io/gorm"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

type GroupSavedGorm struct {
	*MetaDB
}

func NewGroupSavedDB(db *gorm.DB) relation.GroupSavedModelInterface {
	return &GroupSavedGorm{NewMetaDB(db, &relation.GroupSavedModel{})}
}

func (g *GroupSavedGorm) NewTx(tx any) relation.GroupSavedModelInterface {
	return &GroupSavedGorm{NewMetaDB(tx.(*gorm.DB), &relation.GroupSavedModel{})}
}

func (g *GroupSavedGorm) Create(ctx context.Context, group_saved *relation.GroupSavedModel) (err error) {
	return utils.Wrap(g.DB.Create(&group_saved).Error, "")
}

func (g *GroupSavedGorm) FindByUser(ctx context.Context, user_id string) (group_saved []*relation.GroupSavedModel, total int64, err error) {
	err = g.DB.Model(&relation.GroupSavedModel{}).Where("user_id = ? ", user_id).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	err = utils.Wrap(
		g.db(ctx).
			Where("user_id = ? ", user_id).
			Find(&group_saved).
			Error,
		"",
	)
	return

}

func (g *GroupSavedGorm) Take(ctx context.Context, groupID string, userID string) (group_saved *relation.GroupSavedModel, err error) {
	group_saved = &relation.GroupSavedModel{}

	if err := g.DB.Where("group_id = ? and user_id = ?", groupID, userID).Take(group_saved).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 未找到记录
			return nil, errs.ErrRecordNotFound
		}
		return nil, utils.Wrap(err, "error retrieving record")
	}
	return group_saved, nil
}

func (g *GroupSavedGorm) Delete(ctx context.Context, groupID string, userID string) (err error) {
	return utils.Wrap(
		g.DB.WithContext(ctx).
			Where("group_id = ? and user_id = ? ", groupID, userID).
			Delete(&relation.GroupSavedModel{}).
			Error,
		utils.GetSelfFuncName(),
	)
}

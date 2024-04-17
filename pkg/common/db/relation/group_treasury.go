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

type GroupTreasuryGorm struct {
	*MetaDB
}

func NewGroupTreasuryDB(db *gorm.DB) relation.GroupTreasuryModelInterface {
	return &GroupTreasuryGorm{NewMetaDB(db, &relation.GroupTreasuryModel{})}
}

func (s *GroupTreasuryGorm) NewTx(tx any) relation.GroupTreasuryModelInterface {
	return &GroupTreasuryGorm{NewMetaDB(tx.(*gorm.DB), &relation.GroupTreasuryModel{})}
}

func (b *GroupTreasuryGorm) Create(ctx context.Context, treasuries []*relation.GroupTreasuryModel) (err error) {
	return utils.Wrap(b.db(ctx).Create(&treasuries).Error, "")
}

func (b *GroupTreasuryGorm) Delete(ctx context.Context, groupID string) (err error) {
	return utils.Wrap(b.db(ctx).Where("group_id = ?", groupID).Delete(&relation.GroupTreasuryModel{}).Error, "")
}

func (b *GroupTreasuryGorm) UpdateByMap(
	ctx context.Context,
	groupID string,
	args map[string]interface{},
) (err error) {
	return utils.Wrap(
		b.db(ctx).Where("group_id = ?", groupID).Updates(args).Error,
		"",
	)
}

func (b *GroupTreasuryGorm) Update(ctx context.Context, treasuries []*relation.GroupTreasuryModel) (err error) {
	return utils.Wrap(b.db(ctx).Updates(&treasuries).Error, "")
}

func (b *GroupTreasuryGorm) Find(
	ctx context.Context,
	groupID string,
) (treasury *relation.GroupTreasuryModel, err error) {
	return treasury, utils.Wrap(
		b.db(ctx).Where("group_id = ?", groupID).Find(&treasury).Error,
		"",
	)
}

func (b *GroupTreasuryGorm) Take(ctx context.Context, groupID string) (treasury *relation.GroupTreasuryModel, err error) {
	treasury = &relation.GroupTreasuryModel{}
	return treasury, utils.Wrap(
		b.db(ctx).Where("group_id = ?", groupID).Take(treasury).Error,
		"",
	)
}

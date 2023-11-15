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

var _ relation.GroupDappModellInterface = (*GroupDappGorm)(nil)

type GroupDappGorm struct {
	*MetaDB
}

func NewGroupDappDB(db *gorm.DB) relation.GroupDappModellInterface {
	return &GroupDappGorm{NewMetaDB(db, &relation.GroupDappModel{})}
}

func (s *GroupDappGorm) NewTx(tx any) relation.GroupDappModellInterface {
	return &GroupDappGorm{NewMetaDB(tx.(*gorm.DB), &relation.GroupDappModel{})}
}

func (s *GroupDappGorm) Create(ctx context.Context, servers []*relation.GroupDappModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *GroupDappGorm) TakeGroupDapp(ctx context.Context, groupID string) (groupDapp *relation.GroupDappModel, err error) {
	return groupDapp, utils.Wrap(s.DB.Where("group_id = ?", groupID).Take(&groupDapp).Error, "")
}

func (s *GroupDappGorm) DeleteServer(ctx context.Context, serverIDs []string) (err error) {
	return utils.Wrap(s.db(ctx).Where("server_id in (?)", serverIDs).Delete(&relation.GroupDappModel{}).Error, "")
}

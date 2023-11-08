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

var _ relation.ChannelCategoryModelInterface = (*ChannelCategoryGorm)(nil)

type ChannelCategoryGorm struct {
	*MetaDB
}

func NewChannelCategoryDB(db *gorm.DB) relation.ChannelCategoryModelInterface {
	return &ChannelCategoryGorm{NewMetaDB(db, &relation.ChannelCategoryModel{})}
}

func (s *ChannelCategoryGorm) NewTx(tx any) relation.ChannelCategoryModelInterface {
	return &ChannelCategoryGorm{NewMetaDB(tx.(*gorm.DB), &relation.ChannelCategoryModel{})}
}

func (s *ChannelCategoryGorm) Create(ctx context.Context, servers []*relation.ChannelCategoryModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *ChannelCategoryGorm) Take(ctx context.Context, channelCategoryID string) (channelCategory *relation.ChannelCategoryModel, err error) {
	channelCategory = &relation.ChannelCategoryModel{}
	return channelCategory, utils.Wrap(s.DB.Where("category_id = ?", channelCategoryID).Take(channelCategory).Error, "")
}

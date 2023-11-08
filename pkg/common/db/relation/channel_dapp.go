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

var _ relation.ChannelDappModellInterface = (*ChannelDappGorm)(nil)

type ChannelDappGorm struct {
	*MetaDB
}

func NewChannelDappDB(db *gorm.DB) relation.ChannelDappModellInterface {
	return &ChannelDappGorm{NewMetaDB(db, &relation.ChannelDappModel{})}
}

func (s *ChannelDappGorm) NewTx(tx any) relation.ChannelDappModellInterface {
	return &ChannelDappGorm{NewMetaDB(tx.(*gorm.DB), &relation.ChannelDappModel{})}
}

func (s *ChannelDappGorm) Create(ctx context.Context, servers []*relation.ChannelDappModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

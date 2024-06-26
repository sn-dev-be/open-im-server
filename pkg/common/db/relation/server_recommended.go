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

var _ relation.ServerRecommendedModelInterface = (*ServerRecommendedGorm)(nil)

type ServerRecommendedGorm struct {
	*MetaDB
}

func NewServerRecommendedDB(db *gorm.DB) relation.ServerRecommendedModelInterface {
	return &ServerRecommendedGorm{NewMetaDB(db, &relation.ServerRecommendedModel{})}
}

func (s *ServerRecommendedGorm) NewTx(tx any) relation.ServerRecommendedModelInterface {
	return &ServerRecommendedGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerRecommendedModel{})}
}

func (s *ServerRecommendedGorm) GetServerRecommendedList(ctx context.Context) (servers []*relation.ServerRecommendedModel, err error) {
	return servers, utils.Wrap(s.db(ctx).Order("reorder_weight asc").Find(&servers).Error, "")
}

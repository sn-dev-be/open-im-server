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

var _ relation.ServerBlackModelInterface = (*ServerBlackGorm)(nil)

type ServerBlackGorm struct {
	*MetaDB
}

func NewServerBlackDB(db *gorm.DB) relation.ServerBlackModelInterface {
	return &ServerBlackGorm{NewMetaDB(db, &relation.ServerBlackModel{})}
}

func (s *ServerBlackGorm) NewTx(tx any) relation.ServerBlackModelInterface {
	return &ServerBlackGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerBlackModel{})}
}

func (s *ServerBlackGorm) Create(ctx context.Context, servers []*relation.ServerBlackModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (b *ServerBlackGorm) Delete(ctx context.Context, blacks []*relation.ServerBlackModel) (err error) {
	return utils.Wrap(b.db(ctx).Delete(blacks).Error, "")
}

func (s *ServerBlackGorm) UpdateByMap(
	ctx context.Context,
	serverID, blockUserID string,
	args map[string]interface{},
) (err error) {
	return utils.Wrap(
		s.db(ctx).Where("server_id = ? and block_user_id = ?", serverID, blockUserID).Updates(args).Error,
		"",
	)
}

func (s *ServerBlackGorm) Update(ctx context.Context, blacks []*relation.ServerBlackModel) (err error) {
	return utils.Wrap(s.db(ctx).Updates(&blacks).Error, "")
}

func (s *ServerBlackGorm) Find(
	ctx context.Context,
	blacks []*relation.ServerBlackModel,
) (blackList []*relation.ServerBlackModel, err error) {
	var where [][]interface{}
	for _, black := range blacks {
		where = append(where, []interface{}{black.ServerID, black.BlockUserID})
	}
	return blackList, utils.Wrap(
		s.db(ctx).Where("(server_id, block_user_id) in ?", where).Find(&blackList).Error,
		"",
	)
}

func (s *ServerBlackGorm) Take(ctx context.Context, serverID, blockUserID string) (black *relation.ServerBlackModel, err error) {
	black = &relation.ServerBlackModel{}
	return black, utils.Wrap(
		s.db(ctx).Where("server_id = ? and block_user_id = ?", serverID, blockUserID).Take(black).Error,
		"",
	)
}

func (s *ServerBlackGorm) FindBlackUserIDs(ctx context.Context, serverID string) (blackUserIDs []string, err error) {
	return blackUserIDs, utils.Wrap(
		s.db(ctx).Where("server_id = ?", serverID).Pluck("block_user_id", &blackUserIDs).Error,
		"",
	)
}

func (s *ServerBlackGorm) FindServerBlackInfos(ctx context.Context, serverID string, showNumber, pageNumber int32) (blacks []*relation.ServerBlackModel, total int64, err error) {
	err = s.db(ctx).Where("server_id = ?", serverID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = utils.Wrap(
		s.db(ctx).
			Where("server_id = ? ", serverID).
			Order("create_time desc").
			Limit(int(showNumber)).
			Offset(int((pageNumber-1)*showNumber)).
			Find(&blacks).
			Error,
		"",
	)
	if err != nil {
		return nil, 0, err
	}
	return
}

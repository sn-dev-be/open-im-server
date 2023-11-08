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

var _ relation.ServerMemberModelInterface = (*ServerMemberGorm)(nil)

type ServerMemberGorm struct {
	*MetaDB
}

func NewServerMemberDB(db *gorm.DB) relation.ServerMemberModelInterface {
	return &ServerMemberGorm{NewMetaDB(db, &relation.ServerMemberModel{})}
}

func (s *ServerMemberGorm) NewTx(tx any) relation.ServerMemberModelInterface {
	return &ServerMemberGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerMemberModel{})}
}

func (s *ServerMemberGorm) Create(ctx context.Context, servers []*relation.ServerMemberModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *ServerMemberGorm) PageServerMembers(ctx context.Context, showNumber, pageNumber int32, serverID string) (members []*relation.ServerMemberModel, total int64, err error) {
	err = s.DB.Model(&relation.ServerMemberModel{}).Where("serverID = ? ", serverID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	err = utils.Wrap(
		s.db(ctx).
			Where("serverID = ? ", serverID).
			Order("join_time asc").
			Limit(int(showNumber)).
			Offset(int((pageNumber-1)*showNumber)).
			Find(&members).
			Error,
		"",
	)
	return
}

func (s *ServerMemberGorm) GetServerMembers(ctx context.Context, ids []uint64, serverID string) (members []*relation.ServerMemberModel, err error) {
	query := s.db(ctx).Order("join_time asc")
	if len(ids) > 0 {
		query.Where("server_id = ? and id in ?", serverID, ids)
	}
	err = utils.Wrap(query.Find(&members).Error, "")
	return
}

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

var _ relation.ChannelMemberModelInterface = (*ChannelMemberGorm)(nil)

type ChannelMemberGorm struct {
	*MetaDB
}

func NewChannelMemberDB(db *gorm.DB) relation.ChannelMemberModelInterface {
	return &ChannelMemberGorm{NewMetaDB(db, &relation.ChannelMemberModel{})}
}

func (s *ChannelMemberGorm) NewTx(tx any) relation.ChannelMemberModelInterface {
	return &ChannelMemberGorm{NewMetaDB(tx.(*gorm.DB), &relation.ChannelMemberModel{})}
}

func (s *ChannelMemberGorm) Create(ctx context.Context, servers []*relation.ChannelMemberModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *ChannelMemberGorm) PageChannelMembers(ctx context.Context, showNumber, pageNumber int32, serverID string) (members []*relation.ChannelMemberModel, total int64, err error) {
	err = s.DB.Model(&relation.ChannelMemberModel{}).Where("serverID = ? ", serverID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	err = utils.Wrap(
		s.db(ctx).
			Where("serverID = ? ", serverID).
			Order("create_time asc").
			Limit(int(showNumber)).
			Offset(int((pageNumber-1)*showNumber)).
			Find(&members).
			Error,
		"",
	)
	return
}

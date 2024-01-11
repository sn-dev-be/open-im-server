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

var _ relation.ServerMemberRoleModelInterface = (*ServerMemberRoleGorm)(nil)

type ServerMemberRoleGorm struct {
	*MetaDB
}

func NewServerMemberRoleDB(db *gorm.DB) relation.ServerMemberRoleModelInterface {
	return &ServerMemberRoleGorm{NewMetaDB(db, &relation.ServerMemberRoleModel{})}
}

func (s *ServerMemberRoleGorm) NewTx(tx any) relation.ServerMemberRoleModelInterface {
	return &ServerMemberRoleGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerMemberRoleModel{})}
}

func (s *ServerMemberRoleGorm) Create(ctx context.Context, servers []*relation.ServerMemberRoleModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *ServerMemberRoleGorm) DeleteByRole(ctx context.Context, roleIDs []string) error {
	return utils.Wrap(s.db(ctx).Where("role_id in (?)", roleIDs).Delete(&relation.ServerMemberRoleModel{}).Error, "")
}

func (s *ServerMemberRoleGorm) DeleteByMember(ctx context.Context, memberIDs []uint64) error {
	return utils.Wrap(s.db(ctx).Where("member_id in (?)", memberIDs).Delete(&relation.ServerMemberRoleModel{}).Error, "")
}

func (s *ServerMemberRoleGorm) FindByRoleIDs(ctx context.Context, roleIDs []string) (memberRoles []*relation.ServerMemberRoleModel, err error) {
	return memberRoles, utils.Wrap(s.DB.Where("role_id in (?)", roleIDs).Take(memberRoles).Error, "")
}

func (s *ServerMemberRoleGorm) FindByMemberIDS(ctx context.Context, memberIDs []uint64) (memberRoles []*relation.ServerMemberRoleModel, err error) {
	return memberRoles, utils.Wrap(s.DB.Where("member_id in (?)", memberIDs).Take(memberRoles).Error, "")
}

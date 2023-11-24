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

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

var _ relation.ServerRoleModelInterface = (*ServerRoleGorm)(nil)

type ServerRoleGorm struct {
	*MetaDB
}

func NewServerRoleDB(db *gorm.DB) relation.ServerRoleModelInterface {
	return &ServerRoleGorm{NewMetaDB(db, &relation.ServerRoleModel{})}
}

func (s *ServerRoleGorm) NewTx(tx any) relation.ServerRoleModelInterface {
	return &ServerRoleGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerRoleModel{})}
}

func (s *ServerRoleGorm) Create(ctx context.Context, servers []*relation.ServerRoleModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *ServerRoleGorm) Take(ctx context.Context, serverRoleID string) (serverRole *relation.ServerRoleModel, err error) {
	serverRole = &relation.ServerRoleModel{}
	return serverRole, utils.Wrap(s.DB.Where("id = ?", serverRoleID).Take(serverRole).Error, "")
}

func (s *ServerRoleGorm) TakeServerRoleByPriority(ctx context.Context, serverID string, priority int32) (serverRole *relation.ServerRoleModel, err error) {
	return serverRole, utils.Wrap(s.DB.Where("server_id = ? and priority = ?", serverID, priority).Take(&serverRole).Error, "")
}

func (s *ServerRoleGorm) DeleteServer(ctx context.Context, serverIDs []string) (err error) {
	return utils.Wrap(s.db(ctx).Where("server_id in (?)", serverIDs).Delete(&relation.ServerRoleModel{}).Error, "")
}

func (s *ServerRoleGorm) FindDesignationRoleID(ctx context.Context, serverID, roleKey string) (roleIDs []string, err error) {
	return roleIDs, utils.Wrap(
		s.db(ctx).
			Where("server_id = ?", serverID).
			Where(datatypes.JSONQuery("permissions").Equals(true, roleKey)).
			Pluck("id", &roleIDs).Error, "",
	)
}

func (s *ServerRoleGorm) FindRoleID(ctx context.Context, serverID string) (roleIDs []string, err error) {
	return roleIDs, utils.Wrap(s.db(ctx).Where("server_id = ?", serverID).Pluck("id", &roleIDs).Error, "")
}

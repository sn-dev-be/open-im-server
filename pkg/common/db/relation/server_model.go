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

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/ormutil"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

var _ relation.ServerModelInterface = (*ServerGorm)(nil)

type ServerGorm struct {
	*MetaDB
}

func (s *ServerGorm) Delete(ctx context.Context, serverID string) (err error) {
	return utils.Wrap(s.DB.Where("server_id = ?", serverID).Delete(&relation.ServerModel{}).Error, "")
}

func NewServerDB(db *gorm.DB) relation.ServerModelInterface {
	return &ServerGorm{NewMetaDB(db, &relation.ServerModel{})}
}

func (s *ServerGorm) NewTx(tx any) relation.ServerModelInterface {
	return &ServerGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerModel{})}
}

func (s *ServerGorm) Create(ctx context.Context, servers []*relation.ServerModel) (err error) {
	return utils.Wrap(s.DB.Create(&servers).Error, "")
}

func (s *ServerGorm) UpdateMap(ctx context.Context, serverID string, args map[string]interface{}) (err error) {
	return utils.Wrap(s.DB.Where("server_id = ?", serverID).Model(&relation.ServerModel{}).Updates(args).Error, "")
}

func (s *ServerGorm) Take(ctx context.Context, serverID string) (server *relation.ServerModel, err error) {
	server = &relation.ServerModel{}
	return server, utils.Wrap(s.DB.Where("server_id = ?", serverID).Take(server).Error, "")
}

func (s *ServerGorm) Search(ctx context.Context, keyword string, pageNumber, showNumber int32) (total uint32, groups []*relation.ServerModel, err error) {
	db := s.DB
	db = db.WithContext(ctx).Where("status!=? and searchable = 1", constant.ServerStatusDismissed)
	return ormutil.GormSearch[relation.ServerModel](db, []string{"name"}, keyword, pageNumber, showNumber)
}

func (s *ServerGorm) FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*relation.ServerModel, err error) {
	return servers, utils.Wrap(s.DB.Where("server_id in (?) and status != ?", serverIDs, constant.ServerStatusDismissed).Find(&servers).Error, "")
}

func (s *ServerGorm) GetServers(ctx context.Context, serverIDs []string) (servers []*relation.ServerModel, err error) {
	return servers, utils.Wrap(s.db(ctx).Where("server_id in ?", serverIDs).Find(&servers).Error, "")
}

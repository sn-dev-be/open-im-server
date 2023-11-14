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
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

var _ relation.ServerModelInterface = (*ServerGorm)(nil)

type ServerGorm struct {
	*MetaDB
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

func (s *ServerGorm) Take(ctx context.Context, serverID string) (server *relation.ServerModel, err error) {
	server = &relation.ServerModel{}
	return server, utils.Wrap(s.DB.Where("server_id = ?", serverID).Take(server).Error, "")
}

func (s *ServerGorm) FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*relation.ServerModel, err error) {
	return servers, utils.Wrap(s.DB.Where("server_id in (?) and status != ?", serverIDs, constant.ServerStatusDismissed).Find(&servers).Error, "")
}

func (s *ServerGorm) FindServersSplit(ctx context.Context, pageNumber, showNumber int32) (servers []*relation.ServerModel, total int64, err error) {
	err = s.DB.Model(&relation.ServerModel{}).Where("searchable = 1 ").Count(&total).Error
	if err != nil {
		return nil, 0, utils.Wrap(err, "")
	}
	err = utils.Wrap(
		s.db(ctx).
			Where("searchable = 1 ").
			Order("memberNumber desc").
			Limit(int(showNumber)).
			Offset(int((pageNumber-1)*showNumber)).
			Find(&servers).
			Error,
		"",
	)
	return servers, total, nil
}

func (s *ServerGorm) GetServers(ctx context.Context, serverIDs []string) (servers []*relation.ServerModel, err error) {
	return servers, utils.Wrap(s.db(ctx).Where("server_id in ?", serverIDs).Find(&servers).Error, "")
}

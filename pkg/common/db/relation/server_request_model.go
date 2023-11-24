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

	"github.com/OpenIMSDK/tools/ormutil"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

var _ relation.ServerRequestModelInterface = (*ServerRequestGorm)(nil)

type ServerRequestGorm struct {
	*MetaDB
}

func NewServerRequestDB(db *gorm.DB) relation.ServerRequestModelInterface {
	return &ServerRequestGorm{
		NewMetaDB(db, &relation.ServerRequestModel{}),
	}
}

func (s *ServerRequestGorm) NewTx(tx any) relation.ServerRequestModelInterface {
	return &ServerRequestGorm{NewMetaDB(tx.(*gorm.DB), &relation.ServerRequestModel{})}
}

func (s *ServerRequestGorm) Create(ctx context.Context, serverRequests []*relation.ServerRequestModel) (err error) {
	return utils.Wrap(s.DB.Create(&serverRequests).Error, utils.GetSelfFuncName())
}

func (s *ServerRequestGorm) Delete(ctx context.Context, serverID string, userID string) (err error) {
	return utils.Wrap(
		s.DB.WithContext(ctx).
			Where("server_id = ? and user_id = ? ", serverID, userID).
			Delete(&relation.ServerRequestModel{}).
			Error,
		utils.GetSelfFuncName(),
	)
}

func (s *ServerRequestGorm) UpdateHandler(
	ctx context.Context,
	serverID string,
	userID string,
	handledMsg string,
	handleResult int32,
) (err error) {
	return utils.Wrap(
		s.DB.WithContext(ctx).
			Model(&relation.ServerRequestModel{}).
			Where("server_id = ? and user_id = ? ", serverID, userID).
			Updates(map[string]any{
				"handle_msg":    handledMsg,
				"handle_result": handleResult,
			}).
			Error,
		utils.GetSelfFuncName(),
	)
}

func (s *ServerRequestGorm) Take(
	ctx context.Context,
	serverID string,
	userID string,
) (serverRequest *relation.ServerRequestModel, err error) {
	serverRequest = &relation.ServerRequestModel{}
	return serverRequest, utils.Wrap(
		s.DB.WithContext(ctx).Where("server_id = ? and user_id = ? ", serverID, userID).Take(serverRequest).Error,
		utils.GetSelfFuncName(),
	)
}

func (s *ServerRequestGorm) Page(
	ctx context.Context,
	userID string,
	pageNumber, showNumber int32,
) (total uint32, servers []*relation.ServerRequestModel, err error) {
	return ormutil.GormSearch[relation.ServerRequestModel](
		s.DB.WithContext(ctx).Where("user_id = ?", userID),
		nil,
		"",
		pageNumber,
		showNumber,
	)
}

func (s *ServerRequestGorm) PageServer(
	ctx context.Context,
	serverIDs []string,
	pageNumber, showNumber int32,
) (total uint32, servers []*relation.ServerRequestModel, err error) {
	return ormutil.GormPage[relation.ServerRequestModel](
		s.DB.WithContext(ctx).Where("server_id in ?", serverIDs),
		pageNumber,
		showNumber,
	)
}

func (s *ServerRequestGorm) FindServerRequests(
	ctx context.Context,
	serverID string,
	userIDs []string,
) (total int64, serverRequests []*relation.ServerRequestModel, err error) {
	err = s.DB.WithContext(ctx).Where("server_id = ? and user_id in ?", serverID, userIDs).Find(&serverRequests).Error
	return int64(len(serverRequests)), serverRequests, utils.Wrap(err, utils.GetSelfFuncName())
}

func (s *ServerRequestGorm) DeleteServer(ctx context.Context, serverIDs []string) (err error) {
	return utils.Wrap(s.db(ctx).Where("server_id in (?)", serverIDs).Delete(&relation.ServerRequestModel{}).Error, "")
}

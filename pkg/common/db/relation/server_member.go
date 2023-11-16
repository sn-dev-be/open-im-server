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

func (g *ServerMemberGorm) Delete(ctx context.Context, serverID string, userIDs []string) (err error) {
	return utils.Wrap(
		g.db(ctx).Where("server_id = ? and user_id in (?)", serverID, userIDs).Delete(&relation.ServerMemberModel{}).Error,
		"",
	)
}

func (g *ServerMemberGorm) DeleteServer(ctx context.Context, serverIDs []string) (err error) {
	return utils.Wrap(g.db(ctx).Where("server_id in (?)", serverIDs).Delete(&relation.ServerMemberModel{}).Error, "")
}

func (g *ServerMemberGorm) Update(ctx context.Context, serverID string, userID string, data map[string]any) (err error) {
	return utils.Wrap(g.db(ctx).Where("server_id = ? and user_id = ?", serverID, userID).Updates(data).Error, "")
}

func (g *ServerMemberGorm) UpdateRoleLevel(
	ctx context.Context,
	serverID string,
	userID string,
	roleLevel int32,
) (rowsAffected int64, err error) {
	db := g.db(ctx).Where("server_id = ? and user_id = ?", serverID, userID).Updates(map[string]any{
		"role_level": roleLevel,
	})
	return db.RowsAffected, utils.Wrap(db.Error, "")
}

func (g *ServerMemberGorm) Find(
	ctx context.Context,
	serverIDs []string,
	userIDs []string,
	roleLevels []int32,
) (serverMembers []*relation.ServerMemberModel, err error) {
	db := g.db(ctx)
	if len(serverIDs) > 0 {
		db = db.Where("server_id in (?)", serverIDs)
	}
	if len(userIDs) > 0 {
		db = db.Where("user_id in (?)", userIDs)
	}
	if len(roleLevels) > 0 {
		db = db.Where("role_level in (?)", roleLevels)
	}
	return serverMembers, utils.Wrap(db.Find(&serverMembers).Error, "")
}

func (g *ServerMemberGorm) Take(
	ctx context.Context,
	serverID string,
	userID string,
) (serverMember *relation.ServerMemberModel, err error) {
	serverMember = &relation.ServerMemberModel{}
	return serverMember, utils.Wrap(
		g.db(ctx).Where("server_id = ? and user_id = ?", serverID, userID).Take(serverMember).Error,
		"",
	)
}

func (g *ServerMemberGorm) TakeOwner(
	ctx context.Context,
	serverID string,
) (serverMember *relation.ServerMemberModel, err error) {
	serverMember = &relation.ServerMemberModel{}
	return serverMember, utils.Wrap(
		g.db(ctx).Where("server_id = ? and role_level = ?", serverID, constant.ServerOwner).Take(serverMember).Error,
		"",
	)
}

func (g *ServerMemberGorm) SearchMember(
	ctx context.Context,
	keyword string,
	serverIDs []string,
	userIDs []string,
	roleLevels []int32,
	pageNumber, showNumber int32,
) (total uint32, serverList []*relation.ServerMemberModel, err error) {
	db := g.db(ctx)
	ormutil.GormIn(&db, "server_id", serverIDs)
	ormutil.GormIn(&db, "user_id", userIDs)
	ormutil.GormIn(&db, "role_level", roleLevels)
	return ormutil.GormSearch[relation.ServerMemberModel](db, []string{"nickname"}, keyword, pageNumber, showNumber)
}

func (g *ServerMemberGorm) MapServerMemberNum(
	ctx context.Context,
	serverIDs []string,
) (count map[string]uint32, err error) {
	return ormutil.MapCount(g.db(ctx).Where("server_id in (?)", serverIDs), "server_id")
}

func (g *ServerMemberGorm) FindJoinUserID(
	ctx context.Context,
	serverIDs []string,
) (serverUsers map[string][]string, err error) {
	var serverMembers []*relation.ServerMemberModel
	if err := g.db(ctx).Select("server_id, user_id").Where("server_id in (?)", serverIDs).Find(&serverMembers).Error; err != nil {
		return nil, utils.Wrap(err, "")
	}
	serverUsers = make(map[string][]string)
	for _, item := range serverMembers {
		v, ok := serverUsers[item.ServerID]
		if !ok {
			serverUsers[item.ServerID] = []string{item.UserID}
		} else {
			serverUsers[item.ServerID] = append(v, item.UserID)
		}
	}
	return serverUsers, nil
}

func (g *ServerMemberGorm) FindMemberUserID(ctx context.Context, serverID string) (userIDs []string, err error) {
	return userIDs, utils.Wrap(g.db(ctx).Where("server_id = ?", serverID).Pluck("user_id", &userIDs).Error, "")
}

func (g *ServerMemberGorm) FindUserJoinedServerID(ctx context.Context, userID string) (serverIDs []string, err error) {
	return serverIDs, utils.Wrap(g.db(ctx).Where("user_id = ?", userID).Pluck("server_id", &serverIDs).Error, "")
}

func (g *ServerMemberGorm) TakeServerMemberNum(ctx context.Context, serverID string) (count int64, err error) {
	return count, utils.Wrap(g.db(ctx).Where("server_id = ?", serverID).Count(&count).Error, "")
}

func (g *ServerMemberGorm) FindUsersJoinedServerID(ctx context.Context, userIDs []string) (map[string][]string, error) {
	var serverMembers []*relation.ServerMemberModel
	err := g.db(ctx).Select("server_id, user_id").Where("user_id IN (?)", userIDs).Find(&serverMembers).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string][]string)
	for _, serverMember := range serverMembers {
		v, ok := result[serverMember.UserID]
		if !ok {
			result[serverMember.UserID] = []string{serverMember.ServerID}
		} else {
			result[serverMember.UserID] = append(v, serverMember.ServerID)
		}
	}
	return result, nil
}

func (g *ServerMemberGorm) FindUserManagedServerID(ctx context.Context, userID string) (serverIDs []string, err error) {
	return serverIDs, utils.Wrap(
		g.db(ctx).
			Model(&relation.ServerMemberModel{}).
			Where("user_id = ? and (role_level = ? or role_level = ?)", userID, constant.ServerOwner, constant.ServerAdmin).
			Pluck("server_id", &serverIDs).
			Error,
		"",
	)
}

func (g *ServerMemberGorm) FindManageRoleUser(ctx context.Context, serverID string, roleIDs []string) (serverMembers []*relation.ServerMemberModel, err error) {
	return serverMembers, utils.Wrap(g.db(ctx).Where("server_id = ? and server_role_id in (?)", serverID, roleIDs).Find(&serverMembers).Error, "")
}

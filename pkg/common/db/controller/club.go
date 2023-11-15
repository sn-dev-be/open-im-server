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

package controller

import (
	"context"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/tx"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/relation"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

type ClubDatabase interface {
	// server
	CreateServer(ctx context.Context, servers []*relationtb.ServerModel) error
	TakeServer(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error)
	DismissServer(ctx context.Context, serverID string) error // 解散部落，并删除群成员

	////
	FindServer(ctx context.Context, serverIDs []string) (groups []*relationtb.ServerModel, err error)
	FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error)
	SearchServer(ctx context.Context, keyword string, pageNumber, showNumber int32) (uint32, []*relationtb.ServerModel, error)
	UpdateServer(ctx context.Context, groupID string, data map[string]any) error
	GetServerRecommendedList(ctx context.Context) (servers []*relationtb.ServerModel, err error)

	// serverRole
	TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error)
	TakeServerRoleByPriority(ctx context.Context, serverID string, priority int32) (serverRole *relationtb.ServerRoleModel, err error)
	CreateServerRole(ctx context.Context, serverRoles []*relationtb.ServerRoleModel) error

	// serverRequest
	CreateServerRequest(ctx context.Context, requests []*relationtb.ServerRequestModel) error
	TakeServerRequest(ctx context.Context, serverID string, userID string) (*relationtb.ServerRequestModel, error)
	FindServerRequests(ctx context.Context, serverID string, userIDs []string) (int64, []*relationtb.ServerRequestModel, error)
	PageServerRequestUser(ctx context.Context, userID string, pageNumber, showNumber int32) (uint32, []*relationtb.ServerRequestModel, error)

	// serverBlack

	//groupCategory
	TakeGroupCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error)
	CreateGroupCategory(ctx context.Context, categories []*relationtb.GroupCategoryModel) error
	GetAllGroupCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.GroupCategoryModel, error)

	// group
	FindGroup(ctx context.Context, serverIDs []string) (groups []*relationtb.GroupModel, err error)
	TakeGroup(ctx context.Context, groupID string) (group *relationtb.GroupModel, err error)
	CreateServerGroup(ctx context.Context, groups []*relationtb.GroupModel, group_dapps []*relationtb.GroupDappModel) error

	//groupDapp
	TakeGroupDapp(ctx context.Context, groupID string) (groupDapp *relationtb.GroupDappModel, err error)

	// serverMember
	CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error

	TakeServerMember(ctx context.Context, serverID string, userID string) (groupMember *relationtb.ServerMemberModel, err error)
	TakeServerOwner(ctx context.Context, serverID string) (*relationtb.ServerMemberModel, error)
	FindServerMember(ctx context.Context, serverIDs []string, userIDs []string, roleLevels []int32) ([]*relationtb.ServerMemberModel, error)
	FindServerMemberUserID(ctx context.Context, serverID string) ([]string, error)
	FindServerMemberNum(ctx context.Context, serverID string) (uint32, error)
	FindUserManagedServerID(ctx context.Context, userID string) (serverIDs []string, err error)
	PageServerRequest(ctx context.Context, serverIDs []string, pageNumber, showNumber int32) (uint32, []*relationtb.ServerRequestModel, error)

	PageGetJoinServer(ctx context.Context, userID string, pageNumber, showNumber int32) (total uint32, totalServerMembers []*relationtb.ServerMemberModel, err error)
	PageGetServerMember(ctx context.Context, serverID string, pageNumber, showNumber int32) (total uint32, totalServerMembers []*relationtb.ServerMemberModel, err error)
	SearchServerMember(ctx context.Context, keyword string, serverIDs []string, userIDs []string, roleLevels []int32, pageNumber, showNumber int32) (uint32, []*relationtb.ServerMemberModel, error)
	HandlerServerRequest(ctx context.Context, serverID string, userID string, handledMsg string, handleResult int32, member *relationtb.ServerMemberModel) error
	DeleteServerMember(ctx context.Context, serverID string, userIDs []string) error
	MapServerMemberUserID(ctx context.Context, serverIDs []string) (map[string]*relationtb.GroupSimpleUserID, error)
	MapServerMemberNum(ctx context.Context, serverIDs []string) (map[string]uint32, error)
	TransferServerOwner(ctx context.Context, serverID string, oldOwner, newOwner *relationtb.ServerMemberModel, roleLevel int32) error // 转让群
	UpdateServerMember(ctx context.Context, serverID string, userID string, data map[string]any) error
	UpdateServerMembers(ctx context.Context, data []*relationtb.BatchUpdateGroupMember) error
}

func NewClubDatabase(
	server relationtb.ServerModelInterface,
	ServerRecommended relationtb.ServerRecommendedModelInterface,
	serverMember relationtb.ServerMemberModelInterface,
	groupCategory relationtb.GroupCategoryModelInterface,
	group relationtb.GroupModelInterface,
	serverRole relationtb.ServerRoleModelInterface,
	serverRequest relationtb.ServerRequestModelInterface,
	serverBlack relationtb.ServerBlackModelInterface,
	groupDapp relationtb.GroupDappModellInterface,
	tx tx.Tx,
	ctxTx tx.CtxTx,
	cache cache.ClubCache,
) ClubDatabase {
	database := &clubDatabase{
		serverDB:            server,
		serverRecommendedDB: ServerRecommended,
		serverMemberDB:      serverMember,
		serverRoleDB:        serverRole,
		serverRequestDB:     serverRequest,
		serverBlackDB:       serverBlack,
		groupCategoryDB:     groupCategory,
		groupDB:             group,
		groupDappDB:         groupDapp,

		tx: tx,

		ctxTx: ctxTx,
		cache: cache,
	}
	return database
}

func InitClubDatabase(db *gorm.DB, rdb redis.UniversalClient, database *mongo.Database, hashCode func(ctx context.Context, serverID string) (uint64, error)) ClubDatabase {
	rcOptions := rockscache.NewDefaultOptions()
	rcOptions.StrongConsistency = true
	rcOptions.RandomExpireAdjustment = 0.2
	return NewClubDatabase(
		relation.NewServerDB(db),
		relation.NewServerRecommendedDB(db),
		relation.NewServerMemberDB(db),
		relation.NewGroupCategoryDB(db),
		relation.NewGroupDB(db),
		relation.NewServerRoleDB(db),
		relation.NewServerRequestDB(db),
		relation.NewServerBlackDB(db),
		relation.NewGroupDappDB(db),
		tx.NewGorm(db),
		tx.NewMongo(database.Client()),
		cache.NewClubCacheRedis(
			rdb,
			relation.NewServerDB(db),
			relation.NewGroupDappDB(db),
			relation.NewGroupDB(db),
			relation.NewServerMemberDB(db),
			relation.NewServerRequestDB(db),
			hashCode,
			rcOptions,
		),
	)
}

type clubDatabase struct {
	serverDB            relationtb.ServerModelInterface
	serverRecommendedDB relationtb.ServerRecommendedModelInterface
	serverMemberDB      relationtb.ServerMemberModelInterface
	groupCategoryDB     relationtb.GroupCategoryModelInterface
	groupDB             relationtb.GroupModelInterface
	serverRoleDB        relationtb.ServerRoleModelInterface
	serverRequestDB     relationtb.ServerRequestModelInterface
	serverBlackDB       relationtb.ServerBlackModelInterface
	groupDappDB         relationtb.GroupDappModellInterface

	tx    tx.Tx
	ctxTx tx.CtxTx

	cache cache.ClubCache
}

// /server
func (c *clubDatabase) CreateServer(
	ctx context.Context,
	servers []*relationtb.ServerModel,
) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverDB.NewTx(tx).Create(ctx, servers); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (c *clubDatabase) TakeServer(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error) {
	return c.cache.GetServerInfo(ctx, serverID)
}

func (c *clubDatabase) FindServer(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error) {
	return c.cache.GetServersInfo(ctx, serverIDs)
}

func (c *clubDatabase) FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error) {
	return c.serverDB.FindNotDismissedServer(ctx, serverIDs)
}

func (c *clubDatabase) SearchServer(
	ctx context.Context,
	keyword string,
	pageNumber, showNumber int32,
) (uint32, []*relationtb.ServerModel, error) {
	return c.serverDB.Search(ctx, keyword, pageNumber, showNumber)
}

func (c *clubDatabase) UpdateServer(ctx context.Context, groupID string, data map[string]any) error {
	if err := c.serverDB.UpdateMap(ctx, groupID, data); err != nil {
		return err
	}
	return c.cache.DelServersInfo(groupID).ExecDel(ctx)
}

func (c *clubDatabase) GetServerRecommendedList(ctx context.Context) (servers []*relationtb.ServerModel, err error) {
	if recommends, err := c.serverRecommendedDB.GetServerRecommendedList(ctx); err == nil {
		server_recommendeds := []*relationtb.ServerModel{}
		for _, recommend := range recommends {
			server, err := c.serverDB.Take(ctx, recommend.ServerID)
			if err == nil {
				server_recommendeds = append(server_recommendeds, server)
			}
		}
		return server_recommendeds, nil
	} else {
		return nil, err
	}
}

// /serverRole
func (c *clubDatabase) DismissServer(ctx context.Context, serverID string) error {
	cache := c.cache.NewCache()
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverDB.NewTx(tx).Delete(ctx, serverID); err != nil {
			return err
		}
		if err := c.serverMemberDB.NewTx(tx).DeleteServer(ctx, []string{serverID}); err != nil {
			return err
		}
		if err := c.groupCategoryDB.NewTx(tx).DeleteServer(ctx, []string{serverID}); err != nil {
			return err
		}
		if err := c.groupDB.NewTx(tx).DeleteServer(ctx, []string{serverID}); err != nil {
			return err
		}
		if err := c.serverRoleDB.NewTx(tx).DeleteServer(ctx, []string{serverID}); err != nil {
			return err
		}
		// if err := c.groupDappDB.NewTx(tx).DeleteServer(ctx, []string{serverID}); err != nil {
		// 	return err
		// }
		userIDs, err := c.cache.GetServerMemberIDs(ctx, serverID)
		if err != nil {
			return err
		}
		cache = cache.DelJoinedServerID(userIDs...).DelServerMemberIDs(serverID).DelServersMemberNum(serverID).DelServerMembersHash(serverID)
		cache = cache.DelServersInfo(serverID)
		return nil
	}); err != nil {
		return err
	}
	return cache.ExecDel(ctx)
}

func (c *clubDatabase) TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error) {
	return c.serverRoleDB.Take(ctx, serverRoleID)
}

func (c *clubDatabase) TakeServerRoleByPriority(ctx context.Context, serverID string, priority int32) (serverRole *relationtb.ServerRoleModel, err error) {
	return c.serverRoleDB.TakeServerRoleByPriority(ctx, serverID, priority)
}

func (c *clubDatabase) CreateServerRole(ctx context.Context, serverRoles []*relationtb.ServerRoleModel) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverRoleDB.NewTx(tx).Create(ctx, serverRoles); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// groupCategory
func (c *clubDatabase) TakeGroupCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error) {
	return c.groupCategoryDB.Take(ctx, groupCategoryID)

}

func (c *clubDatabase) GetAllGroupCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.GroupCategoryModel, error) {
	return c.groupCategoryDB.GetGroupCategoriesByServerID(ctx, serverID)
}

func (c *clubDatabase) CreateGroupCategory(ctx context.Context, categories []*relationtb.GroupCategoryModel) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.groupCategoryDB.NewTx(tx).Create(ctx, categories); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// / group
func (c *clubDatabase) FindGroup(ctx context.Context, serverIDs []string) (groups []*relationtb.GroupModel, err error) {
	groupIDs, err := c.groupDB.GetGroupIDsByServerIDs(ctx, serverIDs)
	return c.cache.GetGroupsInfo(ctx, groupIDs)
}

func (c *clubDatabase) TakeGroup(ctx context.Context, groupID string) (group *relationtb.GroupModel, err error) {
	return c.cache.GetGroupInfo(ctx, groupID)
}

func (c *clubDatabase) CreateServerGroup(ctx context.Context, groups []*relationtb.GroupModel, group_dapps []*relationtb.GroupDappModel) error {
	cache := c.cache.NewCache()
	if err := c.tx.Transaction(func(tx any) error {
		if len(groups) > 0 {
			if err := c.groupDB.NewTx(tx).Create(ctx, groups); err != nil {
				return err
			}
		}
		if len(group_dapps) > 0 {
			if err := c.groupDappDB.NewTx(tx).Create(ctx, group_dapps); err != nil {
				return err
			}
		}
		createGroupIDs := utils.DistinctAnyGetComparable(groups, func(group *relationtb.GroupModel) string {
			return group.GroupID
		})

		cache = cache.DelGroupsInfo(createGroupIDs...)
		return nil
	}); err != nil {
		return err
	}
	return cache.ExecDel(ctx)
}

func (c *clubDatabase) TakeGroupDapp(ctx context.Context, groupID string) (groupDapp *relationtb.GroupDappModel, err error) {
	return c.cache.GetGroupDappInfo(ctx, groupID)
}

// //serverMember
func (c *clubDatabase) CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error {
	if err := c.serverMemberDB.Create(ctx, serverMembers); err != nil {
		return err
	}
	for _, serverMember := range serverMembers {
		c.cache.DelServerMembersHash(serverMember.ServerID).
			DelServerMemberIDs(serverMember.ServerID).
			DelServersMemberNum(serverMember.ServerID).
			DelJoinedServerID(serverMember.UserID).
			DelServerMembersInfo(serverMember.ServerID, serverMember.UserID).ExecDel(ctx)
	}
	return nil
}

func (c *clubDatabase) TakeServerMember(
	ctx context.Context,
	serverID string,
	userID string,
) (groupMember *relationtb.ServerMemberModel, err error) {
	return c.cache.GetServerMemberInfo(ctx, serverID, userID)
}

func (c *clubDatabase) TakeServerOwner(ctx context.Context, serverID string) (*relationtb.ServerMemberModel, error) {
	return c.serverMemberDB.TakeOwner(ctx, serverID)
}

func (c *clubDatabase) FindServerMember(ctx context.Context, serverIDs []string, userIDs []string, roleLevels []int32) (totalServerMembers []*relationtb.ServerMemberModel, err error) {
	if len(serverIDs) == 0 && len(roleLevels) == 0 && len(userIDs) == 1 {
		gIDs, err := c.cache.GetJoinedServerIDs(ctx, userIDs[0])
		if err != nil {
			return nil, err
		}
		var res []*relationtb.ServerMemberModel
		for _, serverID := range gIDs {
			v, err := c.cache.GetServerMemberInfo(ctx, serverID, userIDs[0])
			if err != nil {
				return nil, err
			}
			res = append(res, v)
		}
		return res, nil
	}
	if len(roleLevels) == 0 {
		for _, serverID := range serverIDs {
			groupMembers, err := c.cache.GetServerMembersInfo(ctx, serverID, userIDs)
			if err != nil {
				return nil, err
			}
			totalServerMembers = append(totalServerMembers, groupMembers...)
		}
		return totalServerMembers, nil
	}
	return c.serverMemberDB.Find(ctx, serverIDs, userIDs, roleLevels)
}

func (c *clubDatabase) FindServerMemberUserID(ctx context.Context, serverID string) ([]string, error) {
	return c.cache.GetServerMemberIDs(ctx, serverID)
}

func (c *clubDatabase) FindServerMemberNum(ctx context.Context, serverID string) (uint32, error) {
	num, err := c.cache.GetServerMemberNum(ctx, serverID)
	if err != nil {
		return 0, err
	}
	return uint32(num), nil
}

func (c *clubDatabase) FindUserManagedServerID(ctx context.Context, userID string) (serverIDs []string, err error) {
	return c.serverMemberDB.FindUserManagedServerID(ctx, userID)
}

func (c *clubDatabase) PageServerRequest(
	ctx context.Context,
	serverIDs []string,
	pageNumber, showNumber int32,
) (uint32, []*relationtb.ServerRequestModel, error) {
	return c.serverRequestDB.PageServer(ctx, serverIDs, pageNumber, showNumber)
}

func (c *clubDatabase) PageGetJoinServer(
	ctx context.Context,
	userID string,
	pageNumber, showNumber int32,
) (total uint32, totalServerMembers []*relationtb.ServerMemberModel, err error) {
	serverIDs, err := c.cache.GetJoinedServerIDs(ctx, userID)
	if err != nil {
		return 0, nil, err
	}
	for _, serverID := range utils.Paginate(serverIDs, int(pageNumber), int(showNumber)) {
		groupMembers, err := c.cache.GetServerMembersInfo(ctx, serverID, []string{userID})
		if err != nil {
			return 0, nil, err
		}
		totalServerMembers = append(totalServerMembers, groupMembers...)
	}
	return uint32(len(serverIDs)), totalServerMembers, nil
}

func (c *clubDatabase) PageGetServerMember(
	ctx context.Context,
	serverID string,
	pageNumber, showNumber int32,
) (total uint32, totalServerMembers []*relationtb.ServerMemberModel, err error) {
	groupMemberIDs, err := c.cache.GetServerMemberIDs(ctx, serverID)
	if err != nil {
		return 0, nil, err
	}
	pageIDs := utils.Paginate(groupMemberIDs, int(pageNumber), int(showNumber))
	if len(pageIDs) == 0 {
		return uint32(len(groupMemberIDs)), nil, nil
	}
	members, err := c.cache.GetServerMembersInfo(ctx, serverID, pageIDs)
	if err != nil {
		return 0, nil, err
	}
	return uint32(len(groupMemberIDs)), members, nil
}

func (c *clubDatabase) SearchServerMember(
	ctx context.Context,
	keyword string,
	serverIDs []string,
	userIDs []string,
	roleLevels []int32,
	pageNumber, showNumber int32,
) (uint32, []*relationtb.ServerMemberModel, error) {
	return c.serverMemberDB.SearchMember(ctx, keyword, serverIDs, userIDs, roleLevels, pageNumber, showNumber)
}

func (c *clubDatabase) HandlerServerRequest(
	ctx context.Context,
	serverID string,
	userID string,
	handledMsg string,
	handleResult int32,
	member *relationtb.ServerMemberModel,
) error {
	return c.tx.Transaction(func(tx any) error {
		if err := c.serverRequestDB.NewTx(tx).UpdateHandler(ctx, serverID, userID, handledMsg, handleResult); err != nil {
			return err
		}
		if member != nil {
			if err := c.serverMemberDB.NewTx(tx).Create(ctx, []*relationtb.ServerMemberModel{member}); err != nil {
				return err
			}
			if err := c.cache.NewCache().DelServerMembersHash(serverID).DelServerMembersInfo(serverID, member.UserID).DelServerMemberIDs(serverID).DelServersMemberNum(serverID).DelJoinedServerID(member.UserID).ExecDel(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *clubDatabase) DeleteServerMember(ctx context.Context, serverID string, userIDs []string) error {
	if err := c.serverMemberDB.Delete(ctx, serverID, userIDs); err != nil {
		return err
	}
	return c.cache.DelServerMembersHash(serverID).
		DelServerMemberIDs(serverID).
		DelServersMemberNum(serverID).
		DelJoinedServerID(userIDs...).
		DelServerMembersInfo(serverID, userIDs...).
		ExecDel(ctx)
}

func (c *clubDatabase) MapServerMemberUserID(
	ctx context.Context,
	serverIDs []string,
) (map[string]*relationtb.GroupSimpleUserID, error) {
	return c.cache.GetServerMemberHashMap(ctx, serverIDs)
}

func (c *clubDatabase) MapServerMemberNum(ctx context.Context, serverIDs []string) (m map[string]uint32, err error) {
	m = make(map[string]uint32)
	for _, serverID := range serverIDs {
		num, err := c.cache.GetServerMemberNum(ctx, serverID)
		if err != nil {
			return nil, err
		}
		m[serverID] = uint32(num)
	}
	return m, nil
}

func (c *clubDatabase) TransferServerOwner(ctx context.Context, serverID string, oldOwner, newOwner *relationtb.ServerMemberModel, roleLevel int32) error {
	return c.tx.Transaction(func(tx any) error {
		ordinaryRole, err := c.serverRoleDB.NewTx(tx).TakeServerRoleByPriority(ctx, serverID, constant.ServerOrdinaryUsers)
		if err != nil {
			return err
		}

		m := make(map[string]any, 2)
		m["server_role_id"] = oldOwner.ServerRoleID
		m["role_level"] = oldOwner.RoleLevel
		err = c.serverMemberDB.NewTx(tx).Update(ctx, serverID, newOwner.UserID, m)
		if err != nil {
			return err
		}

		m["server_role_id"] = ordinaryRole.RoleID
		m["role_level"] = ordinaryRole.Priority
		err = c.serverMemberDB.NewTx(tx).Update(ctx, serverID, oldOwner.UserID, m)
		if err != nil {
			return err
		}

		return c.cache.DelServerMembersInfo(serverID, oldOwner.UserID, newOwner.UserID).DelServerMembersHash(serverID).ExecDel(ctx)
	})
}

func (c *clubDatabase) UpdateServerMember(
	ctx context.Context,
	serverID string,
	userID string,
	data map[string]any,
) error {
	if err := c.serverMemberDB.Update(ctx, serverID, userID, data); err != nil {
		return err
	}
	return c.cache.DelServerMembersInfo(serverID, userID).ExecDel(ctx)
}

func (c *clubDatabase) UpdateServerMembers(ctx context.Context, data []*relationtb.BatchUpdateGroupMember) error {
	cache := c.cache.NewCache()
	if err := c.tx.Transaction(func(tx any) error {
		for _, item := range data {
			if err := c.serverMemberDB.NewTx(tx).Update(ctx, item.GroupID, item.UserID, item.Map); err != nil {
				return err
			}
			cache = cache.DelServerMembersInfo(item.GroupID, item.UserID)
		}
		return nil
	}); err != nil {
		return err
	}
	return cache.ExecDel(ctx)
}

// /serverRequest
func (c *clubDatabase) CreateServerRequest(ctx context.Context, requests []*relationtb.ServerRequestModel) error {
	return c.tx.Transaction(func(tx any) error {
		db := c.serverRequestDB.NewTx(tx)
		for _, request := range requests {
			if err := db.Delete(ctx, request.ServerID, request.UserID); err != nil {
				return err
			}
		}
		return db.Create(ctx, requests)
	})
}

func (c *clubDatabase) TakeServerRequest(
	ctx context.Context,
	serverID string,
	userID string,
) (*relationtb.ServerRequestModel, error) {
	return c.serverRequestDB.Take(ctx, serverID, userID)
}

func (c *clubDatabase) FindServerRequests(ctx context.Context, serverID string, userIDs []string) (int64, []*relationtb.ServerRequestModel, error) {
	return c.serverRequestDB.FindServerRequests(ctx, serverID, userIDs)
}

func (c *clubDatabase) PageServerRequestUser(
	ctx context.Context,
	userID string,
	pageNumber, showNumber int32,
) (uint32, []*relationtb.ServerRequestModel, error) {
	return c.serverRequestDB.Page(ctx, userID, pageNumber, showNumber)
}

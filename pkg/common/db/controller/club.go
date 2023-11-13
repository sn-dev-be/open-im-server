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
	"fmt"
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/tx"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/relation"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

type ClubDatabase interface {
	// Server
	CreateServer(ctx context.Context, servers []*relationtb.ServerModel) error
	TakeServer(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error)
	PageServers(ctx context.Context, pageNumber, showNumber int32) (servers []*relationtb.ServerModel, total int64, err error)
	////
	FindServer(ctx context.Context, serverIDs []string) (groups []*relationtb.ServerModel, err error)
	GetServerRecommendedList(ctx context.Context) (servers []*relationtb.ServerModel, err error)
	GetJoinedServerList(ctx context.Context, userID string) (servers []*relationtb.ServerModel, err error)

	//server_role
	TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error)
	TakeServerRoleByType(ctx context.Context, serverID string, roleType int32) (serverRole *relationtb.ServerRoleModel, err error)
	CreateServerRole(ctx context.Context, serverRoles []*relationtb.ServerRoleModel) error
	GetServerRoleByUserIDAndServerID(ctx context.Context, userID string, serverID string) (server *relationtb.ServerRoleModel, err error)

	//server_request
	CreateServerRequest(ctx context.Context, serverID, userID, invitedUserID string, reqMsg string, ex string, joinSource int32) error

	//server_black

	//group_category
	TakeGroupCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error)
	CreateGroupCategory(ctx context.Context, categories []*relationtb.GroupCategoryModel) error
	GetAllGroupCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.GroupCategoryModel, error)

	//server_member
	PageServerMembers(ctx context.Context, pageNumber, showNumber int32, serverID string) (members []*relationtb.ServerMemberModel, total int64, err error)
	GetServerMembers(ctx context.Context, ids []uint64, serverID string) (members []*relationtb.ServerMemberModel, err error)
	CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error

	//
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
	TransferServerOwner(ctx context.Context, serverID string, oldOwnerUserID, newOwnerUserID string, roleLevel int32) error // 转让群
	UpdateServerMember(ctx context.Context, serverID string, userID string, data map[string]any) error
	UpdateServerMembers(ctx context.Context, data []*relationtb.BatchUpdateGroupMember) error
	GetServerMemberByUserID(ctx context.Context, serverID, userID string) (serverMember *relationtb.ServerMemberModel, err error)
}

func NewClubDatabase(
	server relationtb.ServerModelInterface,
	ServerRecommended relationtb.ServerRecommendedModelInterface,
	serverMember relationtb.ServerMemberModelInterface,
	groupCategory relationtb.GroupCategoryModelInterface,
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
		groupDappDB:         groupDapp,
		tx:                  tx,
		ctxTx:               ctxTx,
		cache:               cache,
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
		relation.NewServerRoleDB(db),
		relation.NewServerRequestDB(db),
		relation.NewServerBlackDB(db),
		relation.NewGroupDappDB(db),
		tx.NewGorm(db),
		tx.NewMongo(database.Client()),
		cache.NewClubCacheRedis(
			rdb,
			relation.NewServerDB(db),
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
	serverRoleDB        relationtb.ServerRoleModelInterface
	serverRequestDB     relationtb.ServerRequestModelInterface
	serverBlackDB       relationtb.ServerBlackModelInterface
	groupDappDB         relationtb.GroupDappModellInterface
	tx                  tx.Tx
	ctxTx               tx.CtxTx
	cache               cache.ClubCache
}

func (c *clubDatabase) CreateServerRequest(ctx context.Context, serverID, userID, invitedUserID string, reqMsg string, ex string, joinSource int32) error {
	return c.tx.Transaction(func(tx any) error {
		_, err := c.serverRequestDB.Take(ctx, serverID, userID)
		// 有db错误
		if err != nil && errs.Unwrap(err) != gorm.ErrRecordNotFound {
			return err
		}
		// 无错误 则更新
		if err == nil {
			if err := c.serverRequestDB.NewTx(tx).UpdateHandler(ctx, serverID, userID, "", constant.ServerResponseNotHandle); err != nil {
				return err
			}
		} else {
			if err := c.serverRequestDB.NewTx(tx).Create(ctx, []*relationtb.ServerRequestModel{{
				FromUserID:    userID,
				ServerID:      serverID,
				InviterUserID: invitedUserID,
				HandleResult:  constant.ServerResponseNotHandle,
				ReqMsg:        reqMsg,
				Ex:            ex,
				JoinSource:    joinSource,
				CreateTime:    time.Now(),
				HandleTime:    time.Unix(0, 0),
			}}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *clubDatabase) TakeServerRoleByType(ctx context.Context, serverID string, roleType int32) (serverRole *relationtb.ServerRoleModel, err error) {
	return c.serverRoleDB.TakeServerRoleByType(ctx, serverID, roleType)
}

func (c *clubDatabase) GetJoinedServerList(ctx context.Context, userID string) (servers []*relationtb.ServerModel, err error) {
	if joinedServers, err := c.serverMemberDB.GetJoinedServerByUserID(ctx, userID); err != nil {
		return nil, err
	} else {
		serverIDs := []string{}
		for _, serverMember := range joinedServers {
			serverIDs = append(serverIDs, serverMember.ServerID)
		}
		return c.serverDB.GetServers(ctx, serverIDs)
	}
}

func (c *clubDatabase) GetServerMemberByUserID(ctx context.Context, serverID string, userID string) (serverMember *relationtb.ServerMemberModel, err error) {
	return c.serverMemberDB.GetServerMemberByUserID(ctx, userID, serverID)
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

func (c *clubDatabase) TakeGroupCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error) {
	return c.groupCategoryDB.Take(ctx, groupCategoryID)
}

func (c *clubDatabase) GetServerRoleByUserIDAndServerID(ctx context.Context, userID string, serverID string) (server *relationtb.ServerRoleModel, err error) {
	if member, err := c.serverMemberDB.GetServerMemberByUserID(ctx, userID, serverID); err != nil {
		return nil, err
	} else {
		roleID := member.ServerRoleID
		if role, err := c.serverRoleDB.Take(ctx, roleID); err != nil {
			return nil, err
		} else {
			return role, nil
		}
	}
}

func (c *clubDatabase) CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error {
	return c.serverMemberDB.Create(ctx, serverMembers)
}

func (c *clubDatabase) GetServerMembers(ctx context.Context, ids []uint64, serverID string) (members []*relationtb.ServerMemberModel, err error) {
	return c.serverMemberDB.GetServerMembers(ctx, ids, serverID)
}

func (c *clubDatabase) PageServerMembers(ctx context.Context, pageNumber int32, showNumber int32, serverID string) (members []*relationtb.ServerMemberModel, total int64, err error) {
	return c.serverMemberDB.PageServerMembers(ctx, showNumber, pageNumber, serverID)
}

func (c *clubDatabase) GetAllGroupCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.GroupCategoryModel, error) {
	return c.groupCategoryDB.GetGroupCategoriesByServerID(ctx, serverID)
}

func (c *clubDatabase) PageServers(ctx context.Context, pageNumber int32, showNumber int32) (servers []*relationtb.ServerModel, total int64, err error) {
	return c.serverDB.FindServersSplit(ctx, pageNumber, showNumber)
}

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
	return c.serverDB.Take(ctx, serverID)
}

func (c *clubDatabase) FindServer(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error) {
	return c.cache.GetServersInfo(ctx, serverIDs)
}

func (c *clubDatabase) TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error) {
	return c.serverRoleDB.Take(ctx, serverRoleID)
}

func (c *clubDatabase) TakeChannelCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error) {
	return c.groupCategoryDB.Take(ctx, groupCategoryID)
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

// ///////////////////////////////////////serverMember////////////////////
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

func (c *clubDatabase) TransferServerOwner(ctx context.Context, serverID string, oldOwnerUserID, newOwnerUserID string, roleLevel int32) error {
	return c.tx.Transaction(func(tx any) error {
		rowsAffected, err := c.serverMemberDB.NewTx(tx).UpdateRoleLevel(ctx, serverID, oldOwnerUserID, roleLevel)
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return utils.Wrap(fmt.Errorf("oldOwnerUserID %s rowsAffected = %d", oldOwnerUserID, rowsAffected), "")
		}
		rowsAffected, err = c.serverMemberDB.NewTx(tx).UpdateRoleLevel(ctx, serverID, newOwnerUserID, constant.ServerOwner)
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return utils.Wrap(fmt.Errorf("newOwnerUserID %s rowsAffected = %d", newOwnerUserID, rowsAffected), "")
		}
		return c.cache.DelServerMembersInfo(serverID, oldOwnerUserID, newOwnerUserID).DelServerMembersHash(serverID).ExecDel(ctx)
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

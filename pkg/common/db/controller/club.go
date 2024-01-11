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
	"time"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/tx"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/relation"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

type ClubDatabase interface {
	// server
	CreateServer(ctx context.Context, servers []*relationtb.ServerModel, roles []*relationtb.ServerRoleModel, categories []*relationtb.GroupCategoryModel, groups []*relationtb.GroupModel, members []*relationtb.ServerMemberModel) error
	TakeServer(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error)
	DismissServer(ctx context.Context, serverID string) error // 解散部落，并删除群成员

	FindServer(ctx context.Context, serverIDs []string) (groups []*relationtb.ServerModel, err error)
	FindNotDismissedServer(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error)
	SearchServer(ctx context.Context, keyword string, pageNumber, showNumber int32) (uint32, []*relationtb.ServerModel, error)
	UpdateServer(ctx context.Context, groupID string, data map[string]any) error
	GetServerRecommendedList(ctx context.Context) (servers []*relationtb.ServerModel, err error)

	// serverRole
	TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error)
	TakeServerRoleByPriority(ctx context.Context, serverID string, priority int32) (serverRole *relationtb.ServerRoleModel, err error)
	CreateServerRole(ctx context.Context, serverRoles []*relationtb.ServerRoleModel) error
	PageGetServerRole(ctx context.Context, serverID string, pageNumber, showNumber int32) (total uint32, totalServerRoles []*relationtb.ServerRoleModel, err error)
	FindServerRole(ctx context.Context, roleIDs []string) (serverRoles []*relationtb.ServerRoleModel, err error)

	// serverRequest
	CreateServerRequest(ctx context.Context, requests []*relationtb.ServerRequestModel) error
	TakeServerRequest(ctx context.Context, serverID string, userID string) (*relationtb.ServerRequestModel, error)
	FindServerRequests(ctx context.Context, serverID string, userIDs []string) (int64, []*relationtb.ServerRequestModel, error)
	PageServerRequestUser(ctx context.Context, userID string, pageNumber, showNumber int32) (uint32, []*relationtb.ServerRequestModel, error)

	// serverBlack
	CreateServerBlack(ctx context.Context, blacks []*relationtb.ServerBlackModel, kickMembers []string, serverID string) (err error)
	DeleteServerBlack(ctx context.Context, blacks []*relationtb.ServerBlackModel) (err error)
	FindServerBlacks(ctx context.Context, serverID string, showNumber, pageNumber int32) (blacks []*relationtb.ServerBlackModel, total int64, err error)
	FindBlackIDs(ctx context.Context, serverID string) (blackIDs []string, err error)

	//groupCategory
	TakeGroupCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error)
	FindGroupCategory(ctx context.Context, groupCategoryIDs []string) (groupCategorys []*relationtb.GroupCategoryModel, err error)
	CreateGroupCategory(ctx context.Context, categories []*relationtb.GroupCategoryModel) error
	GetAllGroupCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.GroupCategoryModel, error)
	UpdateGroupCategory(ctx context.Context, serverID, categoryID string, data map[string]any) error
	DeleteGroupCategorys(ctx context.Context, serverID string, categoryIDs []string) error

	// group
	FindGroup(ctx context.Context, serverIDs []string) (groups []*relationtb.GroupModel, err error)
	TakeGroup(ctx context.Context, groupID string) (group *relationtb.GroupModel, err error)
	CreateServerGroup(ctx context.Context, groups []*relationtb.GroupModel, group_dapps []*relationtb.GroupDappModel) error
	DeleteServerGroup(ctx context.Context, serverID string, groupIDs []string) error
	UpdateServerGroup(ctx context.Context, groupID string, data map[string]any) error
	UpdateServerGroupOrder(ctx context.Context, groupID string, data map[string]any) error

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
	FindServerMemberByRole(ctx context.Context, serverID, role string) ([]*relationtb.ServerMemberModel, error)

	PageGetJoinServer(ctx context.Context, userID string, pageNumber, showNumber int32) (total uint32, totalServerMembers []*relationtb.ServerMemberModel, err error)
	SetJoinServersOrder(ctx context.Context, userID string, serverIDs []string) (err error)
	PageGetServerMember(ctx context.Context, serverID string, pageNumber, showNumber int32) (total uint32, totalServerMembers []*relationtb.ServerMemberModel, err error)
	SearchServerMember(ctx context.Context, keyword string, serverIDs []string, userIDs []string, roleLevels []int32, pageNumber, showNumber int32) (uint32, []*relationtb.ServerMemberModel, error)
	HandlerServerRequest(ctx context.Context, serverID string, userID string, handledMsg string, handleResult int32, member *relationtb.ServerMemberModel) error
	DeleteServerMember(ctx context.Context, serverID string, userIDs []string) error
	MapServerMemberUserID(ctx context.Context, serverIDs []string) (map[string]*relationtb.GroupSimpleUserID, error)
	MapServerMemberNum(ctx context.Context, serverIDs []string) (map[string]uint32, error)
	TransferServerOwner(ctx context.Context, serverID string, oldOwner, newOwner *relationtb.ServerMemberModel, roleLevel int32) error // 转让群
	UpdateServerMember(ctx context.Context, serverID string, userID string, data map[string]any) error
	UpdateServerMembers(ctx context.Context, data []*relationtb.BatchUpdateGroupMember) error
	GetLastestJoinedServerMember(ctx context.Context, serverIDs []string) (members map[string][]*relationtb.ServerMemberModel, err error)

	//mute_record
	FindServerMuteRecords(ctx context.Context, serverID string, pageNumber, showNumber int32) (mute_records []*relationtb.MuteRecordModel, total int64, err error)

	// to mq
	// MsgToMQ(ctx context.Context, key string, msg2mq *sdkws.MsgData) error
	// MsgToModifyMQ(ctx context.Context, key, conversarionID string, msgs []*sdkws.MsgData) error
}

func NewClubDatabase(
	server relationtb.ServerModelInterface,
	ServerRecommended relationtb.ServerRecommendedModelInterface,
	serverMember relationtb.ServerMemberModelInterface,
	groupCategory relationtb.GroupCategoryModelInterface,
	group relationtb.GroupModelInterface,
	serverRole relationtb.ServerRoleModelInterface,
	serverMemberRole relationtb.ServerMemberRoleModelInterface,
	serverRequest relationtb.ServerRequestModelInterface,
	serverBlack relationtb.ServerBlackModelInterface,
	groupDapp relationtb.GroupDappModellInterface,
	muteRecord relationtb.MuteRecordModelInterface,
	tx tx.Tx,
	ctxTx tx.CtxTx,
	cache cache.ClubCache,
) ClubDatabase {
	database := &clubDatabase{
		serverDB:            server,
		serverRecommendedDB: ServerRecommended,
		serverMemberDB:      serverMember,
		serverRoleDB:        serverRole,
		serverMemberRoleDB:  serverMemberRole,
		serverRequestDB:     serverRequest,
		serverBlackDB:       serverBlack,
		groupCategoryDB:     groupCategory,
		groupDB:             group,
		groupDappDB:         groupDapp,
		muteRecordDB:        muteRecord,

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
		relation.NewServerMemberRoleDB(db),
		relation.NewServerRequestDB(db),
		relation.NewServerBlackDB(db),
		relation.NewGroupDappDB(db),
		relation.NewMuteRecordDB(db),
		tx.NewGorm(db),
		tx.NewMongo(database.Client()),
		cache.NewClubCacheRedis(
			rdb,
			relation.NewServerDB(db),
			relation.NewGroupDappDB(db),
			relation.NewGroupDB(db),
			relation.NewGroupCategoryDB(db),
			relation.NewServerMemberDB(db),
			relation.NewServerRequestDB(db),
			relation.NewServerBlackDB(db),
			relation.NewServerRoleDB(db),
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
	serverMemberRoleDB  relationtb.ServerMemberRoleModelInterface
	serverRequestDB     relationtb.ServerRequestModelInterface
	serverBlackDB       relationtb.ServerBlackModelInterface
	groupDappDB         relationtb.GroupDappModellInterface
	muteRecordDB        relationtb.MuteRecordModelInterface

	tx    tx.Tx
	ctxTx tx.CtxTx

	cache cache.ClubCache
}

// /server
func (c *clubDatabase) CreateServer(
	ctx context.Context,
	servers []*relationtb.ServerModel,
	roles []*relationtb.ServerRoleModel,
	categories []*relationtb.GroupCategoryModel,
	groups []*relationtb.GroupModel,
	members []*relationtb.ServerMemberModel,
) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverDB.NewTx(tx).Create(ctx, servers); err != nil {
			return err
		}

		if err := c.groupCategoryDB.NewTx(tx).Create(ctx, categories); err != nil {
			return err
		}

		if err := c.groupDB.NewTx(tx).Create(ctx, groups); err != nil {
			return err
		}

		if err := c.serverRoleDB.NewTx(tx).Create(ctx, roles); err != nil {
			return err
		}

		if err := c.serverMemberDB.NewTx(tx).Create(ctx, members); err != nil {
			return err
		}

		memberRoles := []*relationtb.ServerMemberRoleModel{}
		for _, member := range members {
			memberRoles = append(memberRoles, &relationtb.ServerMemberRoleModel{RoleID: member.ServerRoleID, MemberID: member.ID})
		}
		if err := c.serverMemberRoleDB.NewTx(tx).Create(ctx, memberRoles); err != nil {
			return err
		}

		serverID := servers[0].ServerID
		userID := members[0].UserID
		groupIDs := utils.Slice(groups, func(g *relationtb.GroupModel) string { return g.GroupID })
		categoryIDs := utils.Slice(categories, func(c *relationtb.GroupCategoryModel) string { return c.CategoryID })
		c.cache.DelServerMemberIDs(serverID).DelJoinedServerID(userID).DelGroupCategoriesInfo(categoryIDs...).DelServersInfo(serverID).DelGroupsInfo(groupIDs...).ExecDel(ctx)
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

func (c *clubDatabase) GetLastestJoinedServerMember(ctx context.Context, serverIDs []string) (members map[string][]*relationtb.ServerMemberModel, err error) {
	return c.cache.GetLastestJoinedServerMember(ctx, serverIDs)
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
	return c.cache.GetServerRoleInfo(ctx, serverRoleID)
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

func (c *clubDatabase) PageGetServerRole(
	ctx context.Context,
	serverID string,
	pageNumber, showNumber int32,
) (total uint32, totalServerRoles []*relationtb.ServerRoleModel, err error) {
	serverRoleIDs, err := c.cache.GetServerRoleIDs(ctx, serverID)
	if err != nil {
		return 0, nil, err
	}
	pageIDs := utils.Paginate(serverRoleIDs, int(pageNumber), int(showNumber))
	if len(pageIDs) == 0 {
		return uint32(len(serverRoleIDs)), nil, nil
	}
	roles, err := c.cache.GetServerRolesInfo(ctx, pageIDs)
	if err != nil {
		return 0, nil, err
	}
	return uint32(len(serverRoleIDs)), roles, nil
}

func (c *clubDatabase) FindServerRole(ctx context.Context, roleIDs []string) (serverRoles []*relationtb.ServerRoleModel, err error) {
	serverRoles, err = c.cache.GetServerRolesInfo(ctx, roleIDs)
	return serverRoles, err
}

// groupCategory
func (c *clubDatabase) TakeGroupCategory(ctx context.Context, groupCategoryID string) (groupCategory *relationtb.GroupCategoryModel, err error) {
	return c.groupCategoryDB.Take(ctx, groupCategoryID)
}

func (c *clubDatabase) FindGroupCategory(ctx context.Context, groupCategoryIDs []string) (groupCategorys []*relationtb.GroupCategoryModel, err error) {
	return c.cache.GetGroupCategoriesInfo(ctx, groupCategoryIDs)
}

func (c *clubDatabase) GetAllGroupCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.GroupCategoryModel, error) {
	categoryIDs, err := c.groupCategoryDB.FindGroupCategoryIDsByServerID(ctx, serverID)
	if err != nil {
		return nil, err
	}
	return c.cache.GetGroupCategoriesInfo(ctx, categoryIDs)
}

func (c *clubDatabase) CreateGroupCategory(ctx context.Context, categories []*relationtb.GroupCategoryModel) error {
	if err := c.tx.Transaction(func(tx any) error {

		serverID := categories[0].ServerID
		server, err := c.cache.GetServerInfo(ctx, serverID)
		if err != nil {
			return err
		}

		//order
		for i, category := range categories {
			category.ReorderWeight = int32(server.CategoryNumber + uint32(i))
		}
		if err := c.groupCategoryDB.NewTx(tx).Create(ctx, categories); err != nil {
			return err
		}

		server.CategoryNumber += uint32(len(categories))
		data := make(map[string]any)
		data["category_number"] = server.CategoryNumber
		err = c.serverDB.NewTx(tx).UpdateMap(ctx, serverID, data)
		if err != nil {
			return err
		}

		categoryIDs := utils.Slice(categories, func(e *relationtb.GroupCategoryModel) string { return e.CategoryID })
		return c.cache.DelGroupCategoriesInfo(categoryIDs...).DelServersInfo(serverID).ExecDel(ctx)
	}); err != nil {
		return err
	}
	return nil
}

func (c *clubDatabase) UpdateGroupCategory(ctx context.Context, serverID, categoryID string, data map[string]any) error {
	err := c.groupCategoryDB.UpdateMap(ctx, serverID, categoryID, data)
	if err != nil {
		return err
	}
	return c.cache.DelGroupCategoriesInfo(categoryID).ExecDel(ctx)
}

func (c *clubDatabase) DeleteGroupCategorys(ctx context.Context, serverID string, categoryIDs []string) error {
	categories, err := c.groupCategoryDB.FindGroupCategoryIDsByType(ctx, serverID, constant.DefaultCategoryType)
	if err != nil {
		return err
	}
	defaultCategoryID := categories[0]
	data := make(map[string]any)
	data["group_category_id"] = defaultCategoryID

	if err := c.tx.Transaction(func(tx any) error {
		//categoryIDs := utils.Slice(categories, func(e *relationtb.GroupCategoryModel) string { return e.CategoryID })
		for _, categoryID := range categoryIDs {

			//将分组下房间移入默认分组
			groupIDs, err := c.groupDB.GetGroupIDsByCategoryID(ctx, categoryID)
			if err != nil && !errs.ErrRecordNotFound.Is(err) {
				return err
			}
			for _, groupID := range groupIDs {
				err := c.groupDB.NewTx(tx).UpdateMap(ctx, groupID, data)
				if err != nil {
					return err
				}
			}
			c.cache.DelGroupsInfo(groupIDs...).ExecDel(ctx)
		}
		//批量删除分组
		err := c.groupCategoryDB.NewTx(tx).Delete(ctx, categoryIDs)
		if err != nil {
			return err
		}

		//更新server中的category_number
		sm, err := c.cache.GetServerInfo(ctx, serverID)
		if err != nil {
			return err
		}
		sm.CategoryNumber -= uint32(len(categoryIDs))
		data := make(map[string]any)
		data["category_number"] = sm.CategoryNumber
		err = c.serverDB.NewTx(tx).UpdateMap(ctx, serverID, data)
		if err != nil {
			return err
		}

		return c.cache.DelServersInfo(serverID).DelGroupCategoriesInfo(categoryIDs...).ExecDel(ctx)
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

		serverIDs := utils.Slice(groups, func(e *relationtb.GroupModel) string { return e.ServerID })
		sm, err := c.cache.GetServerInfo(ctx, serverIDs[0])
		if err != nil {
			return err
		}

		if len(groups) > 0 {
			for i, group := range groups {
				group.ReorderWeight = int32(sm.GroupNumber + uint32(i))
			}
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

		//维护servers group_number
		data := make(map[string]any)
		data["group_number"] = sm.GroupNumber + uint32(len(groups))
		err = c.serverDB.NewTx(tx).UpdateMap(ctx, serverIDs[0], data)
		if err != nil {
			return err
		}

		cache = cache.DelServersInfo(serverIDs[0]).DelGroupsInfo(createGroupIDs...)
		return nil
	}); err != nil {
		return err
	}
	return cache.ExecDel(ctx)
}

func (c *clubDatabase) TakeGroupDapp(ctx context.Context, groupID string) (groupDapp *relationtb.GroupDappModel, err error) {
	groupDapps, err := c.cache.GetGroupDappInfo(ctx, []string{groupID})
	if err != nil {
		return nil, err
	}
	if len(groupDapps) == 0 {
		return nil, errs.ErrRecordNotFound
	}
	return groupDapps[0], nil
}

func (c *clubDatabase) DeleteServerGroup(ctx context.Context, serverID string, groupIDs []string) error {
	cache := c.cache.NewCache()
	if err := c.tx.Transaction(func(tx any) error {
		if len(groupIDs) > 0 {

			groups, err := c.FindGroup(ctx, []string{serverID})
			if err != nil {
				return err
			}
			dbGroupIDs := utils.Slice(groups, func(e *relationtb.GroupModel) string { return e.GroupID })
			groupMap := utils.SliceToMapAny(groups, func(g *relationtb.GroupModel) (string, *relationtb.GroupModel) {
				return g.GroupID, g
			})

			deleteGroupApps := []string{}

			deleteGroupNum := 0
			for _, groupID := range groupIDs {
				if utils.Contain(groupID, dbGroupIDs...) {
					if err := c.groupDB.NewTx(tx).UpdateStatus(ctx, groupID, constant.GroupStatusDismissed); err != nil {
						return err
					}

					if group, ok := groupMap[groupID]; ok {
						if group.GroupMode == constant.AppGroupMode {
							deleteGroupApps = append(deleteGroupApps, groupID)
						}
					}

					deleteGroupNum++
					cache = cache.DelGroupsInfo(groupID)
				}
			}

			if len(deleteGroupApps) > 0 {
				c.groupDappDB.DeleteByGroup(ctx, deleteGroupApps)
			}

			//维护servers group_number
			sm, err := c.cache.GetServerInfo(ctx, serverID)
			if err != nil {
				return err
			}

			data := make(map[string]any)
			data["group_number"] = sm.GroupNumber - uint32(deleteGroupNum)
			err = c.serverDB.NewTx(tx).UpdateMap(ctx, serverID, data)
			if err != nil {
				return err
			}
			cache = cache.DelServersInfo(serverID)
		}
		return nil
	}); err != nil {
		return err
	}
	return cache.ExecDel(ctx)
}

func (c *clubDatabase) UpdateServerGroup(ctx context.Context, groupID string, data map[string]any) error {
	return c.tx.Transaction(func(tx any) error {
		dappID := ""
		groupMode := int32(0)
		if data["dapp_id"] != nil && data["group_mode"] != nil {
			dappID, _ = data["dapp_id"].(string)
			v, ok := data["group_mode"].(int32)
			if ok {
				groupMode = int32(v)
			}
		}
		if dappID != "" && groupMode == constant.AppGroupMode {
			_, err := c.groupDappDB.TakeGroupDapp(ctx, groupID)
			if err != nil && errs.Unwrap(err) != gorm.ErrRecordNotFound {
				return err
			}
			if err == nil {
				m := make(map[string]any)
				m["dapp_id"] = data["dapp_id"]
				err = c.groupDappDB.UpdateByMap(ctx, groupID, m)
			} else {
				m := &relationtb.GroupDappModel{
					GroupID:    groupID,
					DappID:     dappID,
					CreateTime: time.Now(),
				}
				err = c.groupDappDB.Create(ctx, []*relationtb.GroupDappModel{m})
			}
			if err != nil {
				return err
			}
		} else {
			err := c.groupDappDB.NewTx(tx).DeleteByGroup(ctx, []string{groupID})
			if err != nil {
				return err
			}
		}
		delete(data, "dapp_id")
		if err := c.groupDB.NewTx(tx).UpdateMap(ctx, groupID, data); err != nil {
			return err
		}
		return c.cache.DelGroupsInfo(groupID).DelGroupDappInfo(ctx, groupID).ExecDel(ctx)
	})
}

func (c *clubDatabase) UpdateServerGroupOrder(ctx context.Context, groupID string, data map[string]any) error {
	return c.tx.Transaction(func(tx any) error {
		if err := c.groupDB.NewTx(tx).UpdateMap(ctx, groupID, data); err != nil {
			return err
		}
		return c.cache.DelGroupsInfo(groupID).DelGroupDappInfo(ctx, groupID).ExecDel(ctx)
	})
}

// //serverMember
func (c *clubDatabase) CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverMemberDB.Create(ctx, serverMembers); err != nil {
			return err
		}

		memberRoles := []*relationtb.ServerMemberRoleModel{}
		for _, member := range serverMembers {
			memberRoles = append(memberRoles, &relationtb.ServerMemberRoleModel{RoleID: member.ServerRoleID, MemberID: member.ID})
		}
		if err := c.serverMemberRoleDB.NewTx(tx).Create(ctx, memberRoles); err != nil {
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
	}); err != nil {
		return err
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
	// for i := 0; i < len(serverIDs); i++ {
	// 	groupMembers, err := c.cache.GetServerMembersInfo(ctx, serverIDs[i], []string{userID})
	// 	if err != nil {
	// 		return 0, nil, err
	// 	}
	// 	totalServerMembers = append(totalServerMembers, groupMembers...)
	// }
	return uint32(len(serverIDs)), totalServerMembers, nil
}

func (c *clubDatabase) SetJoinServersOrder(ctx context.Context, userID string, serverIDs []string) (err error) {
	return c.tx.Transaction(func(tx any) error {
		for i, serverID := range serverIDs {
			data := make(map[string]any)
			data["reorder_weight"] = i
			err := c.serverMemberDB.NewTx(tx).Update(ctx, serverID, userID, data)
			if err != nil {
				return err
			}
		}
		c.cache.DelJoinedServerID(userID).ExecDel(ctx)
		return nil
	})
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

			if err := c.cache.NewCache().DelServersInfo(serverID).DelServerMembersHash(serverID).DelServerMembersInfo(serverID, member.UserID).DelServerMemberIDs(serverID).DelServersMemberNum(serverID).DelJoinedServerID(member.UserID).ExecDel(ctx); err != nil {
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
	c.muteRecordDB.DeleteByUserIDs(ctx, serverID, userIDs)

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

		sm := make(map[string]any, 1)
		sm["owner_user_id"] = newOwner.UserID
		err = c.serverDB.NewTx(tx).UpdateMap(ctx, serverID, sm)
		if err != nil {
			return err
		}

		return c.cache.DelServersInfo(serverID).DelJoinedServerID(oldOwner.UserID, newOwner.UserID).DelServerMemberIDs(serverID).DelServerMembersInfo(serverID, oldOwner.UserID, newOwner.UserID).DelServerMembersHash(serverID).ExecDel(ctx)
	})
}

func (c *clubDatabase) UpdateServerMember(
	ctx context.Context,
	serverID string,
	userID string,
	data map[string]any,
) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverMemberDB.NewTx(tx).Update(ctx, serverID, userID, data); err != nil {
			return err
		}

		mute_end_time, ok := data["mute_end_time"].(time.Time)
		if ok {
			mrm, err := c.muteRecordDB.Take(ctx, userID, serverID)
			if err != nil && errs.Unwrap(err) != gorm.ErrRecordNotFound {
				return err
			}
			if mute_end_time != time.Unix(0, 0) {
				if err != nil {
					//muteszzz
					mrm = &relationtb.MuteRecordModel{
						ServerID:       serverID,
						BlockUserID:    userID,
						OperatorUserID: mcontext.GetOpUserID(ctx),
						CreateTime:     time.Now(),
						MuteEndTime:    mute_end_time,
					}
					err = c.muteRecordDB.NewTx(tx).Create(ctx, []*relationtb.MuteRecordModel{mrm})
					if err != nil {
						return err
					}
				}
			} else {
				//cancel_mute
				err = c.muteRecordDB.NewTx(tx).Delete(ctx, []*relationtb.MuteRecordModel{mrm})
				if err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
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

func (c *clubDatabase) FindServerMemberByRole(ctx context.Context, serverID, role string) ([]*relationtb.ServerMemberModel, error) {
	roleIDs, err := c.serverRoleDB.FindDesignationRoleID(ctx, serverID, role)
	if err != nil {
		return nil, err
	}
	serverMembers, err := c.serverMemberDB.FindManageRoleUser(ctx, serverID, roleIDs)
	if err != nil {
		return nil, err
	}
	return serverMembers, nil
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

//////////////server_black/////////////////////

func (c *clubDatabase) CreateServerBlack(ctx context.Context, blacks []*relationtb.ServerBlackModel, kickMembers []string, serverID string) (err error) {
	return c.tx.Transaction(func(tx any) error {
		if len(kickMembers) > 0 {
			if err := c.serverMemberDB.NewTx(tx).Delete(ctx, serverID, kickMembers); err != nil {
				return err
			}

			c.muteRecordDB.DeleteByUserIDs(ctx, serverID, kickMembers)

			c.cache.DelServerMemberIDs(serverID).DelServerMembersHash(serverID).DelServerMembersInfo(serverID, kickMembers...).DelServersMemberNum(serverID).DelServersInfo(serverID).ExecDel(ctx)
		}

		if err := c.serverBlackDB.NewTx(tx).Create(ctx, blacks); err != nil {
			return err
		}
		return c.deleteBlackIDsCache(ctx, blacks)
	})
}

func (c *clubDatabase) DeleteServerBlack(ctx context.Context, blacks []*relationtb.ServerBlackModel) (err error) {
	if err := c.serverBlackDB.Delete(ctx, blacks); err != nil {
		return err
	}

	return c.deleteBlackIDsCache(ctx, blacks)
}

func (c *clubDatabase) FindServerBlacks(
	ctx context.Context,
	serverID string,
	showNumber int32,
	pageNumber int32,
) (blacks []*relationtb.ServerBlackModel, total int64, err error) {
	return c.serverBlackDB.FindServerBlackInfos(ctx, serverID, showNumber, pageNumber)
}

func (c *clubDatabase) FindBlackIDs(ctx context.Context, serverID string) (blackIDs []string, err error) {
	return c.cache.GetServerBlacksCache(ctx, serverID)
}

func (c *clubDatabase) deleteBlackIDsCache(ctx context.Context, blacks []*relationtb.ServerBlackModel) (err error) {
	cache := c.cache.NewCache()
	for _, black := range blacks {
		cache = cache.DeleteBlackIDsCache(black.ServerID)
	}
	return cache.ExecDel(ctx)
}

// ///mute_record
func (c *clubDatabase) FindServerMuteRecords(ctx context.Context, serverID string, pageNumber, showNumber int32) (mute_records []*relationtb.MuteRecordModel, total int64, err error) {
	return c.muteRecordDB.FindServerMuteRecords(ctx, serverID, pageNumber, showNumber)
}

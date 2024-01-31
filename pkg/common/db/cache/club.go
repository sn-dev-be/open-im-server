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

package cache

import (
	"context"
	"time"

	"github.com/OpenIMSDK/tools/log"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"

	"github.com/OpenIMSDK/tools/utils"

	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

const (
	serverExpireTime     = time.Second * 60 * 60 * 12
	halfHourExpireTime   = time.Second * 60 * 30
	serverInfoKey        = "SERVER_INFO:"
	serverMemberIDsKey   = "SERVER_MEMBER_IDS:"
	serverMembersHashKey = "SERVER_MEMBERS_HASH2:"
	serverMemberInfoKey  = "SERVER_MEMBER_INFO:"
	joinedServersKey     = "JOIN_SERVERS_KEY:"
	serverMemberNumKey   = "SERVER_MEMBER_NUM_CACHE:"
	groupDappInfoKey     = "GROUP_DAPP_INFO:"
	serverBlackInfoKey   = "SERVER_BLACK_INFO:"
	serverRoleIDsKey     = "SERVER_ROLE_IDS:"
	serverRoleInfoKey    = "SERVER_ROLE_INFO:"
)

type ClubCache interface {
	metaCache
	NewCache() ClubCache
	GetServersInfo(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error)
	GetServerInfo(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error)
	DelServersInfo(serverIDs ...string) ClubCache

	GetGroupsInfo(ctx context.Context, groupIDs []string) (groups []*relationtb.GroupModel, err error)
	GetGroupInfo(ctx context.Context, groupID string) (group *relationtb.GroupModel, err error)
	DelGroupsInfo(groupIDs ...string) ClubCache

	GetGroupCategoriesInfo(ctx context.Context, categoryIDs []string) (categories []*relationtb.GroupCategoryModel, err error)
	GetGroupCategoryInfo(ctx context.Context, categoryID string) (category *relationtb.GroupCategoryModel, err error)
	DelGroupCategoriesInfo(categoryIDs ...string) ClubCache

	GetGroupDappInfo(ctx context.Context, groupIDs []string) (groupDapps []*relationtb.GroupDappModel, err error)
	DelGroupDappInfo(ctx context.Context, groupIDs ...string) ClubCache

	GetServerMembersHash(ctx context.Context, serverID string) (hashCode uint64, err error)
	GetServerMemberHashMap(ctx context.Context, serverIDs []string) (map[string]*relationtb.GroupSimpleUserID, error)
	DelServerMembersHash(serverID string) ClubCache

	GetServerMemberIDs(ctx context.Context, serverID string) (serverMemberIDs []string, err error)
	GetServersMemberIDs(ctx context.Context, serverIDs []string) (serverMemberIDs map[string][]string, err error)

	DelServerMemberIDs(serverID string) ClubCache

	GetJoinedServerIDs(ctx context.Context, userID string) (joinedServerIDs []string, err error)
	DelJoinedServerID(userID ...string) ClubCache

	GetServerMemberInfo(ctx context.Context, serverID, userID string) (serverMember *relationtb.ServerMemberModel, err error)
	GetServerMembersInfo(ctx context.Context, serverID string, userID []string) (serverMembers []*relationtb.ServerMemberModel, err error)
	GetAllServerMembersInfo(ctx context.Context, serverID string) (serverMembers []*relationtb.ServerMemberModel, err error)
	GetServerMembersPage(ctx context.Context, serverID string, userID []string, showNumber, pageNumber int32) (total uint32, serverMembers []*relationtb.ServerMemberModel, err error)
	GetLastestJoinedServerMember(ctx context.Context, serverIDs []string) (serverMember map[string][]*relationtb.ServerMemberModel, err error)

	DelServerMembersInfo(serverID string, userID ...string) ClubCache

	GetServerMemberNum(ctx context.Context, serverID string) (memberNum int64, err error)
	DelServersMemberNum(serverID ...string) ClubCache

	DeleteBlackIDsCache(serverID ...string) ClubCache
	GetServerBlacksCache(ctx context.Context, serverID string) (blackIDs []string, err error)

	GetServerRoleIDs(ctx context.Context, serverID string) (serverRoleIDs []string, err error)
	DelServerRoleIDs(serverID string) ClubCache

	GetServerRoleInfo(ctx context.Context, roleID string) (serverRole *relationtb.ServerRoleModel, err error)
	GetServerRolesInfo(ctx context.Context, roleIDs []string) (serverRoles []*relationtb.ServerRoleModel, err error)
	DelServerRolesInfo(roleID ...string) ClubCache
}

type ClubCacheRedis struct {
	metaCache
	serverDB        relationtb.ServerModelInterface
	groupDappDB     relationtb.GroupDappModellInterface
	groupDB         relationtb.GroupModelInterface
	groupCategoryDB relationtb.GroupCategoryModelInterface
	serverMemberDB  relationtb.ServerMemberModelInterface
	serverRequestDB relationtb.ServerRequestModelInterface
	serverBlackDB   relationtb.ServerBlackModelInterface
	serverRoleDB    relationtb.ServerRoleModelInterface
	expireTime      time.Duration
	rcClient        *rockscache.Client
	hashCode        func(ctx context.Context, serverID string) (uint64, error)
}

func NewClubCacheRedis(
	rdb redis.UniversalClient,
	serverDB relationtb.ServerModelInterface,
	groupDappDB relationtb.GroupDappModellInterface,
	groupDB relationtb.GroupModelInterface,
	groupCategoryDB relationtb.GroupCategoryModelInterface,
	serverMemberDB relationtb.ServerMemberModelInterface,
	serverRequestDB relationtb.ServerRequestModelInterface,
	serverBlackDB relationtb.ServerBlackModelInterface,
	serverRoleDB relationtb.ServerRoleModelInterface,
	hashCode func(ctx context.Context, serverID string) (uint64, error),
	opts rockscache.Options,
) ClubCache {
	rcClient := rockscache.NewClient(rdb, opts)

	return &ClubCacheRedis{
		rcClient:        rcClient,
		expireTime:      serverExpireTime,
		serverDB:        serverDB,
		groupDappDB:     groupDappDB,
		groupDB:         groupDB,
		groupCategoryDB: groupCategoryDB,
		serverMemberDB:  serverMemberDB,
		serverRequestDB: serverRequestDB,
		serverBlackDB:   serverBlackDB,
		serverRoleDB:    serverRoleDB,
		hashCode:        hashCode,
		metaCache:       NewMetaCacheRedis(rcClient),
	}
}

func (c *ClubCacheRedis) NewCache() ClubCache {
	return &ClubCacheRedis{
		rcClient:        c.rcClient,
		expireTime:      c.expireTime,
		serverDB:        c.serverDB,
		serverMemberDB:  c.serverMemberDB,
		serverRequestDB: c.serverRequestDB,
		serverBlackDB:   c.serverBlackDB,
		metaCache:       NewMetaCacheRedis(c.rcClient, c.metaCache.GetPreDelKeys()...),
	}
}

func (c *ClubCacheRedis) getServerInfoKey(serverID string) string {
	return serverInfoKey + serverID
}

func (c *ClubCacheRedis) getGroupCategoryInfoKey(categoryID string) string {
	return groupCategoryKey + categoryID
}

func (c *ClubCacheRedis) getGroupInfoKey(groupID string) string {
	return groupInfoKey + groupID
}

func (c *ClubCacheRedis) getJoinedServersKey(userID string) string {
	return joinedServersKey + userID
}

func (c *ClubCacheRedis) getServerMembersHashKey(serverID string) string {
	return serverMembersHashKey + serverID
}

func (c *ClubCacheRedis) getServerMemberIDsKey(serverID string) string {
	return serverMemberIDsKey + serverID
}

func (c *ClubCacheRedis) getServerMemberInfoKey(serverID, userID string) string {
	return serverMemberInfoKey + serverID + "-" + userID
}

func (c *ClubCacheRedis) getServerMemberNumKey(serverID string) string {
	return serverMemberNumKey + serverID
}

func (c *ClubCacheRedis) GetLastestJoinedServerMemberKey(serverID string) string {
	return lastestJoinedServerMemberKey + serverID
}

func (c *ClubCacheRedis) getServerBlackInfoKey(serverID string) string {
	return serverBlackInfoKey + serverID
}

func (c *ClubCacheRedis) getGroupDappInfoKey(groupID string) string {
	return groupDappInfoKey + groupID
}

func (c *ClubCacheRedis) getServerRoleIDsKey(serverID string) string {
	return serverRoleIDsKey + serverID
}

func (c *ClubCacheRedis) getServerRoleInfoKey(roleID string) string {
	return serverRoleInfoKey + "-" + roleID
}

func (c *ClubCacheRedis) GetServerIndex(server *relationtb.ServerModel, keys []string) (int, error) {
	key := c.getServerInfoKey(server.ServerID)
	for i, _key := range keys {
		if _key == key {
			return i, nil
		}
	}

	return 0, errIndex
}

func (c *ClubCacheRedis) GetServerMemberIndex(serverMember *relationtb.ServerMemberModel, keys []string) (int, error) {
	key := c.getServerMemberInfoKey(serverMember.ServerID, serverMember.UserID)
	for i, _key := range keys {
		if _key == key {
			return i, nil
		}
	}

	return 0, errIndex
}

func (c *ClubCacheRedis) GetServersInfo(ctx context.Context, serverIDs []string) (servers []*relationtb.ServerModel, err error) {
	return batchGetCache2(ctx, c.rcClient, c.expireTime, serverIDs, func(serverID string) string {
		return c.getServerInfoKey(serverID)
	}, func(ctx context.Context, serverID string) (*relationtb.ServerModel, error) {
		return c.serverDB.Take(ctx, serverID)
	})
}

func (c *ClubCacheRedis) GetServerInfo(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error) {
	return getCache(ctx, c.rcClient, c.getServerInfoKey(serverID), c.expireTime, func(ctx context.Context) (*relationtb.ServerModel, error) {
		return c.serverDB.Take(ctx, serverID)
	})
}

func (c *ClubCacheRedis) DelServersInfo(serverIDs ...string) ClubCache {
	newClubCache := c.NewCache()
	keys := make([]string, 0, len(serverIDs))
	for _, serverID := range serverIDs {
		keys = append(keys, c.getServerInfoKey(serverID))
	}
	newClubCache.AddKeys(keys...)

	return newClubCache
}

// /group_category
// / group_category
func (c *ClubCacheRedis) GetGroupCategoriesInfo(ctx context.Context, categoryIDs []string) (categories []*relationtb.GroupCategoryModel, err error) {
	return batchGetCache2(ctx, c.rcClient, c.expireTime, categoryIDs, func(categoryID string) string {
		return c.getGroupCategoryInfoKey(categoryID)
	}, func(ctx context.Context, categoryID string) (*relationtb.GroupCategoryModel, error) {
		return c.groupCategoryDB.Take(ctx, categoryID)
	})
}
func (c *ClubCacheRedis) GetGroupCategoryInfo(ctx context.Context, categoryID string) (category *relationtb.GroupCategoryModel, err error) {
	return getCache(ctx, c.rcClient, c.getGroupCategoryInfoKey(categoryID), c.expireTime, func(ctx context.Context) (*relationtb.GroupCategoryModel, error) {
		return c.groupCategoryDB.Take(ctx, categoryID)
	})
}
func (c *ClubCacheRedis) DelGroupCategoriesInfo(categoryIDs ...string) ClubCache {
	newClubCache := c.NewCache()
	keys := make([]string, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		keys = append(keys, c.getGroupCategoryInfoKey(categoryID))
	}
	newClubCache.AddKeys(keys...)

	return newClubCache
}

// / groupInfo.
func (c *ClubCacheRedis) GetGroupsInfo(ctx context.Context, groupIDs []string) (groups []*relationtb.GroupModel, err error) {
	return batchGetCache2(ctx, c.rcClient, c.expireTime, groupIDs, func(groupID string) string {
		return c.getGroupInfoKey(groupID)
	}, func(ctx context.Context, groupID string) (*relationtb.GroupModel, error) {
		return c.groupDB.Take(ctx, groupID)
	})
}

func (c *ClubCacheRedis) GetGroupInfo(ctx context.Context, groupID string) (group *relationtb.GroupModel, err error) {
	return getCache(ctx, c.rcClient, c.getGroupInfoKey(groupID), c.expireTime, func(ctx context.Context) (*relationtb.GroupModel, error) {
		return c.groupDB.Take(ctx, groupID)
	})
}

func (c *ClubCacheRedis) DelGroupsInfo(groupIDs ...string) ClubCache {
	newClubCache := c.NewCache()
	keys := make([]string, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		keys = append(keys, c.getGroupInfoKey(groupID))
	}
	newClubCache.AddKeys(keys...)

	return newClubCache
}

// groupDapp
func (c *ClubCacheRedis) GetGroupDappInfo(ctx context.Context, groupIDs []string) (groupDapps []*relationtb.GroupDappModel, err error) {
	// return getCache(ctx, c.rcClient, c.getGroupDappInfoKey(groupID), c.expireTime, func(ctx context.Context) (*relationtb.GroupDappModel, error) {
	// 	return c.groupDappDB.TakeGroupDapp(ctx, groupID)
	// })

	return batchGetCache2(ctx, c.rcClient, c.expireTime, groupIDs, func(groupID string) string {
		return c.getGroupDappInfoKey(groupID)
	}, func(ctx context.Context, groupID string) (*relationtb.GroupDappModel, error) {
		return c.groupDappDB.TakeGroupDapp(ctx, groupID)
	})
}

func (c *ClubCacheRedis) DelGroupDappInfo(ctx context.Context, groupIDs ...string) ClubCache {
	newClubCache := c.NewCache()
	keys := make([]string, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		keys = append(keys, c.getGroupDappInfoKey(groupID))
	}
	newClubCache.AddKeys(keys...)
	return newClubCache
}

// /serverMemberInfo
func (c *ClubCacheRedis) GetServerMembersHash(ctx context.Context, serverID string) (hashCode uint64, err error) {
	return getCache(ctx, c.rcClient, c.getServerMembersHashKey(serverID), c.expireTime, func(ctx context.Context) (uint64, error) {
		return c.hashCode(ctx, serverID)
	})
}

func (c *ClubCacheRedis) GetServerMemberHashMap(ctx context.Context, serverIDs []string) (map[string]*relationtb.GroupSimpleUserID, error) {
	res := make(map[string]*relationtb.GroupSimpleUserID)
	for _, serverID := range serverIDs {
		hash, err := c.GetServerMembersHash(ctx, serverID)
		if err != nil {
			return nil, err
		}
		log.ZInfo(ctx, "GetServerMemberHashMap", "serverID", serverID, "hash", hash)
		num, err := c.GetServerMemberNum(ctx, serverID)
		if err != nil {
			return nil, err
		}
		res[serverID] = &relationtb.GroupSimpleUserID{Hash: hash, MemberNum: uint32(num)}
	}

	return res, nil
}

func (c *ClubCacheRedis) DelServerMembersHash(serverID string) ClubCache {
	cache := c.NewCache()
	cache.AddKeys(c.getServerMembersHashKey(serverID))

	return cache
}

func (c *ClubCacheRedis) GetServerMemberIDs(ctx context.Context, serverID string) (serverMemberIDs []string, err error) {
	return getCache(ctx, c.rcClient, c.getServerMemberIDsKey(serverID), c.expireTime, func(ctx context.Context) ([]string, error) {
		return c.serverMemberDB.FindMemberUserID(ctx, serverID)
	})
}

func (c *ClubCacheRedis) GetServersMemberIDs(ctx context.Context, serverIDs []string) (map[string][]string, error) {
	m := make(map[string][]string)
	for _, serverID := range serverIDs {
		userIDs, err := c.GetServerMemberIDs(ctx, serverID)
		if err != nil {
			return nil, err
		}
		m[serverID] = userIDs
	}

	return m, nil
}

func (c *ClubCacheRedis) DelServerMemberIDs(serverID string) ClubCache {
	cache := c.NewCache()
	cache.AddKeys(c.getServerMemberIDsKey(serverID))

	return cache
}

func (c *ClubCacheRedis) GetJoinedServerIDs(ctx context.Context, userID string) (joinedServerIDs []string, err error) {
	return getCache(ctx, c.rcClient, c.getJoinedServersKey(userID), c.expireTime, func(ctx context.Context) ([]string, error) {
		return c.serverMemberDB.FindUserJoinedServerID(ctx, userID)
	})
}

func (c *ClubCacheRedis) DelJoinedServerID(userIDs ...string) ClubCache {
	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, c.getJoinedServersKey(userID))
	}
	cache := c.NewCache()
	cache.AddKeys(keys...)

	return cache
}

func (c *ClubCacheRedis) GetServerMemberInfo(ctx context.Context, serverID, userID string) (serverMember *relationtb.ServerMemberModel, err error) {
	return getCache(ctx, c.rcClient, c.getServerMemberInfoKey(serverID, userID), c.expireTime, func(ctx context.Context) (*relationtb.ServerMemberModel, error) {
		return c.serverMemberDB.Take(ctx, serverID, userID)
	})
}

func (c *ClubCacheRedis) GetServerMembersInfo(ctx context.Context, serverID string, userIDs []string) ([]*relationtb.ServerMemberModel, error) {
	return batchGetCache2(ctx, c.rcClient, c.expireTime, userIDs, func(userID string) string {
		return c.getServerMemberInfoKey(serverID, userID)
	}, func(ctx context.Context, userID string) (*relationtb.ServerMemberModel, error) {
		return c.serverMemberDB.Take(ctx, serverID, userID)
	})
}

func (c *ClubCacheRedis) GetServerMembersPage(
	ctx context.Context,
	serverID string,
	userIDs []string,
	showNumber, pageNumber int32,
) (total uint32, serverMembers []*relationtb.ServerMemberModel, err error) {
	serverMemberIDs, err := c.GetServerMemberIDs(ctx, serverID)
	if err != nil {
		return 0, nil, err
	}
	if userIDs != nil {
		userIDs = utils.BothExist(userIDs, serverMemberIDs)
	} else {
		userIDs = serverMemberIDs
	}
	serverMembers, err = c.GetServerMembersInfo(ctx, serverID, utils.Paginate(userIDs, int(showNumber), int(showNumber)))

	return uint32(len(userIDs)), serverMembers, err
}

func (c *ClubCacheRedis) GetAllServerMembersInfo(ctx context.Context, serverID string) (serverMembers []*relationtb.ServerMemberModel, err error) {
	serverMemberIDs, err := c.GetServerMemberIDs(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return c.GetServerMembersInfo(ctx, serverID, serverMemberIDs)
}

func (c *ClubCacheRedis) DelServerMembersInfo(serverID string, userIDs ...string) ClubCache {
	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, c.getServerMemberInfoKey(serverID, userID))
	}
	cache := c.NewCache()
	cache.AddKeys(keys...)

	return cache
}

func (c *ClubCacheRedis) GetServerMemberNum(ctx context.Context, serverID string) (memberNum int64, err error) {
	return getCache(ctx, c.rcClient, c.getServerMemberNumKey(serverID), c.expireTime, func(ctx context.Context) (int64, error) {
		return c.serverMemberDB.TakeServerMemberNum(ctx, serverID)
	})
}

func (c *ClubCacheRedis) DelServersMemberNum(serverID ...string) ClubCache {
	keys := make([]string, 0, len(serverID))
	for _, serverID := range serverID {
		keys = append(keys, c.getServerMemberNumKey(serverID))
	}
	cache := c.NewCache()
	cache.AddKeys(keys...)

	return cache
}

func (c *ClubCacheRedis) GetLastestJoinedServerMember(ctx context.Context, serverIDs []string) (members map[string][]*relationtb.ServerMemberModel, err error) {
	res, err := batchGetCache2(ctx, c.rcClient, halfHourExpireTime, serverIDs, func(serverID string) string {
		return c.GetLastestJoinedServerMemberKey(serverID)
	}, func(ctx context.Context, serverID string) ([]*relationtb.ServerMemberModel, error) {
		return c.serverMemberDB.FindLastestJoinedServerMember(ctx, serverID, 3)
	})

	serverMembersMap := make(map[string][]*relationtb.ServerMemberModel)

	for _, result := range res {
		if len(result) == 0 {
			continue
		}
		serverID := result[0].ServerID
		// 将结果添加到 map 中
		serverMembersMap[serverID] = result
	}

	return serverMembersMap, err
}

// //////server_blacks////////////
func (c *ClubCacheRedis) DeleteBlackIDsCache(serverID ...string) ClubCache {
	keys := make([]string, 0, len(serverID))
	for _, serverID := range serverID {
		keys = append(keys, c.getServerBlackInfoKey(serverID))
	}
	cache := c.NewCache()
	cache.AddKeys(keys...)

	return cache
}

func (c *ClubCacheRedis) GetServerBlacksCache(ctx context.Context, serverID string) (blackIDs []string, err error) {
	return getCache(ctx, c.rcClient, c.getServerBlackInfoKey(serverID), c.expireTime, func(ctx context.Context) ([]string, error) {
		return c.serverBlackDB.FindBlackUserIDs(ctx, serverID)
	})
}

// / serverRole
func (c *ClubCacheRedis) GetServerRoleInfo(ctx context.Context, roleID string) (serverRole *relationtb.ServerRoleModel, err error) {
	return getCache(ctx, c.rcClient, c.getServerRoleInfoKey(roleID), c.expireTime, func(ctx context.Context) (*relationtb.ServerRoleModel, error) {
		return c.serverRoleDB.Take(ctx, roleID)
	})
}

func (c *ClubCacheRedis) DelServerRolesInfo(roleIDs ...string) ClubCache {
	keys := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		keys = append(keys, c.getServerRoleInfoKey(roleID))
	}
	cache := c.NewCache()
	cache.AddKeys(keys...)

	return cache
}

func (c *ClubCacheRedis) GetServerRoleIDs(ctx context.Context, serverID string) (serverRoleIDs []string, err error) {
	return getCache(ctx, c.rcClient, c.getServerRoleIDsKey(serverID), c.expireTime, func(ctx context.Context) ([]string, error) {
		return c.serverRoleDB.FindRoleID(ctx, serverID)
	})
}

func (c *ClubCacheRedis) DelServerRoleIDs(serverID string) ClubCache {
	cache := c.NewCache()
	cache.AddKeys(c.getServerRoleIDsKey(serverID))

	return cache
}

func (c *ClubCacheRedis) GetServerRolesInfo(ctx context.Context, roleIDs []string) ([]*relationtb.ServerRoleModel, error) {
	return batchGetCache2(ctx, c.rcClient, c.expireTime, roleIDs, func(roleID string) string {
		return c.getServerRoleInfoKey(roleID)
	}, func(ctx context.Context, roleID string) (*relationtb.ServerRoleModel, error) {
		return c.serverRoleDB.Take(ctx, roleID)
	})
}

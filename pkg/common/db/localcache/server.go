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

package localcache

import (
	"context"
	"sync"

	"github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/tools/errs"

	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

type ServerLocalCache struct {
	lock        sync.Mutex
	cache       map[string]ServerMemberIDsHash
	clubClient  *rpcclient.ClubRpcClient
	groupClient *rpcclient.GroupRpcClient
}

type ServerMemberIDsHash struct {
	memberListHash uint64
	userIDs        []string
}

func NewServerLocalCache(clubClient *rpcclient.ClubRpcClient, groupClient *rpcclient.GroupRpcClient) *ServerLocalCache {
	return &ServerLocalCache{
		cache:       make(map[string]ServerMemberIDsHash, 0),
		clubClient:  clubClient,
		groupClient: groupClient,
	}
}

func (g *ServerLocalCache) GetServerMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	group, err := g.groupClient.GetGroupInfo(ctx, groupID)
	if err != nil {
		return nil, err
	}
	serverID := group.ServerID
	resp, err := g.clubClient.Client.GetServerAbstractInfo(ctx, &club.GetServerAbstractInfoReq{
		ServerIDs: []string{serverID},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.ServerAbstractInfos) < 1 {
		return nil, errs.ErrGroupIDNotFound
	}

	g.lock.Lock()
	localHashInfo, ok := g.cache[serverID]
	if ok && localHashInfo.memberListHash == resp.ServerAbstractInfos[0].ServerMemberListHash {
		g.lock.Unlock()
		return localHashInfo.userIDs, nil
	}
	g.lock.Unlock()

	serverMembersResp, err := g.clubClient.Client.GetServerMemberUserIDs(ctx, &club.GetServerMemberUserIDsReq{
		ServerID: serverID,
	})
	if err != nil {
		return nil, err
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	g.cache[serverID] = ServerMemberIDsHash{
		memberListHash: resp.ServerAbstractInfos[0].ServerMemberListHash,
		userIDs:        serverMembersResp.UserIDs,
	}
	return g.cache[serverID].userIDs, nil
}

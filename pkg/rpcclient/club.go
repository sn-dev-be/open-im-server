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

package rpcclient

import (
	"context"
	"strings"

	"google.golang.org/grpc"

	"github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	sdkws "github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
)

type Club struct {
	conn   grpc.ClientConnInterface
	Client club.ClubClient
	discov discoveryregistry.SvcDiscoveryRegistry
}

func NewClub(discov discoveryregistry.SvcDiscoveryRegistry) *Club {
	conn, err := discov.GetConn(context.Background(), config.Config.RpcRegisterName.OpenImClubName)
	if err != nil {
		panic(err)
	}
	client := club.NewClubClient(conn)
	return &Club{discov: discov, conn: conn, Client: client}
}

type ClubRpcClient Club

func NewClubRpcClient(discov discoveryregistry.SvcDiscoveryRegistry) ClubRpcClient {
	return ClubRpcClient(*NewClub(discov))
}

func (c *ClubRpcClient) GetServerMemberInfos(
	ctx context.Context,
	serverID string,
	userIDs []string,
	complete bool,
) ([]*sdkws.ServerMemberFullInfo, error) {
	resp, err := c.Client.GetServerMembersInfo(ctx, &club.GetServerMembersInfoReq{
		ServerID: serverID,
		UserIDs:  userIDs,
	})
	if err != nil {
		return nil, err
	}
	if complete {
		if ids := utils.Single(userIDs, utils.Slice(resp.Members, func(e *sdkws.ServerMemberFullInfo) string {
			return e.UserID
		})); len(ids) > 0 {
			return nil, errs.ErrNotInGroupYet.Wrap(strings.Join(ids, ","))
		}
	}
	return resp.Members, nil
}

func (c *ClubRpcClient) GetServerMemberInfo(
	ctx context.Context,
	serverID string,
	userID string,
) (*sdkws.ServerMemberFullInfo, error) {
	members, err := c.GetServerMemberInfos(ctx, serverID, []string{userID}, true)
	if err != nil {
		return nil, err
	}
	return members[0], nil
}

func (c *ClubRpcClient) GetServerMemberInfoMap(
	ctx context.Context,
	serverID string,
	userIDs []string,
	complete bool,
) (map[string]*sdkws.ServerMemberFullInfo, error) {
	members, err := c.GetServerMemberInfos(ctx, serverID, userIDs, true)
	if err != nil {
		return nil, err
	}
	return utils.SliceToMap(members, func(e *sdkws.ServerMemberFullInfo) string {
		return e.UserID
	}), nil
}

func (c *ClubRpcClient) GetOwnerAndAdminInfos(
	ctx context.Context,
	serverID string,
) ([]*sdkws.ServerMemberFullInfo, error) {
	resp, err := c.Client.GetServerMemberRoleLevel(ctx, &club.GetServerMemberRoleLevelReq{
		ServerID:   serverID,
		RoleLevels: []int32{constant.ServerOwner, constant.ServerAdmin},
	})
	if err != nil {
		return nil, err
	}
	return resp.Members, nil
}

func (c *ClubRpcClient) GetOwnerInfo(ctx context.Context, serverID string) (*sdkws.ServerMemberFullInfo, error) {
	resp, err := c.Client.GetServerMemberRoleLevel(ctx, &club.GetServerMemberRoleLevelReq{
		ServerID:   serverID,
		RoleLevels: []int32{constant.ServerOwner},
	})
	return resp.Members[0], err
}

func (c *ClubRpcClient) GetServerMemberIDs(ctx context.Context, serverID string) ([]string, error) {
	resp, err := c.Client.GetServerMemberUserIDs(ctx, &club.GetServerMemberUserIDsReq{
		ServerID: serverID,
	})
	if err != nil {
		return nil, err
	}
	return resp.UserIDs, nil
}

func (c *ClubRpcClient) GetServerMemberCache(
	ctx context.Context,
	serverID string,
	serverMemberID string,
) (*sdkws.ServerMemberFullInfo, error) {
	resp, err := c.Client.GetServerMemberCache(ctx, &club.GetServerMemberCacheReq{
		ServerID:       serverID,
		ServerMemberID: serverMemberID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Member, nil
}

func (c *ClubRpcClient) GetServerGroupMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	resp, err := c.Client.GetServerGroupMemberUserIDs(ctx, &club.GetServerGroupMemberUserIDsReq{
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}
	return resp.UserIDs, nil
}

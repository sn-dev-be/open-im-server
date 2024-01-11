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

package club

import (
	"context"

	"github.com/openimsdk/open-im-server/v3/pkg/permissions"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient/notification"

	"google.golang.org/grpc"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/relation"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/unrelation"
)

func Start(client discoveryregistry.SvcDiscoveryRegistry, server *grpc.Server) error {
	db, err := relation.NewGormDB()
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&relationtb.ServerModel{},
		&relationtb.GroupCategoryModel{},
		&relationtb.ServerRoleModel{},
		&relationtb.ServerMemberRoleModel{},
		&relationtb.GroupDappModel{},
		&relationtb.ServerBlackModel{},
		&relationtb.ServerRequestModel{},
		&relationtb.ServerMemberModel{},
		&relationtb.ServerRecommendedModel{},
		&relationtb.MuteRecordModel{},
	); err != nil {
		return err
	}

	mongo, err := unrelation.NewMongo()
	if err != nil {
		return err
	}

	rdb, err := cache.NewRedis()
	if err != nil {
		return err
	}

	userRpcClient := rpcclient.NewUserRpcClient(client)
	msgRpcClient := rpcclient.NewMessageRpcClient(client)
	conversationRpcClient := rpcclient.NewConversationRpcClient(client)
	groupRpcClient := rpcclient.NewGroupRpcClient(client)

	var cs clubServer
	database := controller.InitClubDatabase(db, rdb, mongo.GetDatabase(), cs.serverMemberHashCode)
	cs.ClubDatabase = database
	cs.GroupDatabase = controller.InitGroupDatabase(db, rdb, mongo.GetDatabase(), nil)

	cs.User = userRpcClient
	cs.Notification = notification.NewClubNotificationSender(database, &msgRpcClient, &userRpcClient, func(ctx context.Context, userIDs []string) ([]notification.CommonUser, error) {
		users, err := userRpcClient.GetUsersInfo(ctx, userIDs)
		if err != nil {
			return nil, err
		}
		return utils.Slice(users, func(e *sdkws.UserInfo) notification.CommonUser { return e }), nil
	})
	cs.conversationRpcClient = conversationRpcClient
	cs.msgRpcClient = msgRpcClient
	cs.Group = groupRpcClient
	pbclub.RegisterClubServer(server, &cs)
	return nil
}

type clubServer struct {
	ClubDatabase          controller.ClubDatabase
	GroupDatabase         controller.GroupDatabase
	User                  rpcclient.UserRpcClient
	Group                 rpcclient.GroupRpcClient
	Notification          *notification.ClubNotificationSender
	conversationRpcClient rpcclient.ConversationRpcClient
	msgRpcClient          rpcclient.MessageRpcClient
}

func (c *clubServer) GetPublicUserInfoMap(ctx context.Context, userIDs []string, complete bool) (map[string]*sdkws.PublicUserInfo, error) {
	if len(userIDs) == 0 {
		return map[string]*sdkws.PublicUserInfo{}, nil
	}
	users, err := c.User.GetPublicUserInfos(ctx, userIDs, complete)
	if err != nil {
		return nil, err
	}
	return utils.SliceToMapAny(users, func(e *sdkws.PublicUserInfo) (string, *sdkws.PublicUserInfo) {
		return e.UserID, e
	}), nil
}

func (c *clubServer) getOpUserServerPermission(ctx context.Context, serverID string) (*permissions.Permissions, error) {
	opUserID := mcontext.GetOpUserID(ctx)
	ServerMember, err := c.ClubDatabase.TakeServerMember(ctx, serverID, opUserID)
	if err != nil {
		return nil, err
	}
	serverRole, err := c.ClubDatabase.TakeServerRole(ctx, ServerMember.ServerRoleID)
	if err != nil {
		return nil, err
	}
	permissions, err := permissions.PermissionsFromJSON(string(serverRole.Permissions))
	if err != nil {
		return nil, err
	}
	return &permissions, nil
}

func (c *clubServer) checkPermissions(ctx context.Context, serverID string, role string) bool {
	permission, err := c.getOpUserServerPermission(ctx, serverID)
	if err != nil {
		return false
	}
	return permission.HasPermission(role)
}

func (c *clubServer) checkManageServer(ctx context.Context, serverID string) bool {
	return c.checkPermissions(ctx, serverID, permissions.ManageServer)
}

func (c *clubServer) checkManageMember(ctx context.Context, serverID string) bool {
	return c.checkPermissions(ctx, serverID, permissions.ManageMember)
}

func (c *clubServer) checkManageGroup(ctx context.Context, serverID string) bool {
	return c.checkPermissions(ctx, serverID, permissions.ManageGroup)
}

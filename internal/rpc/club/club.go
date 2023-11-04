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

	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient/notification"

	"google.golang.org/grpc"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/discoveryregistry"
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

	if err := db.AutoMigrate(&relationtb.ServerModel{}); err != nil {
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

	var cs clubServer
	database := controller.InitClubDatabase(db, rdb, mongo.GetDatabase())
	cs.ClubDatabase = database
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
	pbclub.RegisterClubServer(server, &cs)
	return nil
}

type clubServer struct {
	ClubDatabase          controller.ClubDatabase
	User                  rpcclient.UserRpcClient
	Notification          *notification.ClubNotificationSender
	conversationRpcClient rpcclient.ConversationRpcClient
	msgRpcClient          rpcclient.MessageRpcClient
}

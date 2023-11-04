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

package notification

import (
	"context"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

func NewClubNotificationSender(
	db controller.ClubDatabase,
	msgRpcClient *rpcclient.MessageRpcClient,
	userRpcClient *rpcclient.UserRpcClient,
	fn func(ctx context.Context, userIDs []string) ([]CommonUser, error),
) *ClubNotificationSender {
	return &ClubNotificationSender{
		NotificationSender: rpcclient.NewNotificationSender(rpcclient.WithRpcClient(msgRpcClient), rpcclient.WithUserRpcClient(userRpcClient)),
		getUsersInfo:       fn,
		db:                 db,
	}
}

type ClubNotificationSender struct {
	*rpcclient.NotificationSender
	getUsersInfo func(ctx context.Context, userIDs []string) ([]CommonUser, error)
	db           controller.ClubDatabase
}

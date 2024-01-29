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

	"google.golang.org/grpc"

	pbcron "github.com/OpenIMSDK/protocol/cron"
	"github.com/OpenIMSDK/tools/discoveryregistry"

	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
)

type Cron struct {
	conn   grpc.ClientConnInterface
	Client pbcron.CronClient
	discov discoveryregistry.SvcDiscoveryRegistry
}

func NewCron(discov discoveryregistry.SvcDiscoveryRegistry) *Cron {
	conn, err := discov.GetConn(context.Background(), config.Config.RpcRegisterName.OpenImCronName)
	if err != nil {
		panic(err)
	}
	return &Cron{
		discov: discov,
		conn:   conn,
		Client: pbcron.NewCronClient(conn),
	}
}

type CronRpcClient Cron

func NewCronRpcClient(discov discoveryregistry.SvcDiscoveryRegistry) CronRpcClient {
	return CronRpcClient(*NewCron(discov))
}

func (c *CronRpcClient) SetCloseVoiceChannelJob(ctx context.Context, userID, channelID, groupID string, sessionType int32) error {
	var req pbcron.SetCloseVoiceChannelJobReq
	req.UserID = userID
	req.ChannelID = channelID
	req.GroupID = groupID
	req.SessionType = sessionType
	_, err := c.Client.SetCloseVoiceChannelJob(ctx, &req)
	if err != nil {
		return err
	}
	return nil
}

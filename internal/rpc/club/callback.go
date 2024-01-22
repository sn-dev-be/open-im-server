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

	"github.com/OpenIMSDK/protocol/common"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/tools/utils"
)

func (s *clubServer) SendBusinessEventToMQ(ctx context.Context, req *msg.SendBusinessEventToMQReq) {
	s.msgRpcClient.Client.SendBusinessEventToMQ(ctx, req)
}

func (s *clubServer) SendClubServerEvent(ctx context.Context, serverID, name, banner, icon string, isPublic bool) {
	event := &common.BusinessMQEvent{
		Event: utils.StructToJsonString(&common.CommonBusinessMQEvent{
			ClubServer: &common.ClubServer{
				ClubServerId: serverID,
				Name:         name,
				Icon:         icon,
				Banner:       banner,
				IsPublic:     isPublic,
			},
			EventType: constant.ClubServerMQEventType,
		}),
	}
	s.SendBusinessEventToMQ(ctx, &msg.SendBusinessEventToMQReq{
		Events: []*common.BusinessMQEvent{event},
	})
}

func (s *clubServer) SendDeleteClubServerEvent(ctx context.Context, serverID string) {
	event := &common.BusinessMQEvent{
		Event: utils.StructToJsonString(&common.CommonBusinessMQEvent{
			ClubServer: &common.ClubServer{
				ClubServerId: serverID,
			},
			EventType: constant.DeleteServerMQEventType,
		}),
	}
	s.SendBusinessEventToMQ(ctx, &msg.SendBusinessEventToMQReq{
		Events: []*common.BusinessMQEvent{event},
	})
}

func (s *clubServer) SendClubServerUserEvent(ctx context.Context, serverID, userID, nickname string) {
	event := &common.BusinessMQEvent{
		Event: utils.StructToJsonString(&common.CommonBusinessMQEvent{
			ClubServerUser: &common.ClubServerUser{
				ServerId: serverID,
				UserId:   userID,
				Nickname: nickname,
			},
			EventType: constant.ClubServerUserMQEventType,
		}),
	}
	s.SendBusinessEventToMQ(ctx, &msg.SendBusinessEventToMQReq{
		Events: []*common.BusinessMQEvent{event},
	})
}

func (s *clubServer) SendDeleteClubServerUserEvent(ctx context.Context, serverID string, userIDs []string) {
	events := []*common.BusinessMQEvent{}
	for _, userID := range userIDs {
		event := &common.BusinessMQEvent{
			Event: utils.StructToJsonString(&common.CommonBusinessMQEvent{
				ClubServerUser: &common.ClubServerUser{
					ServerId: serverID,
					UserId:   userID,
				},
				EventType: constant.DeleteClubServerUserMQEventType,
			}),
		}
		events = append(events, event)
	}

	s.SendBusinessEventToMQ(ctx, &msg.SendBusinessEventToMQReq{
		Events: events,
	})
}

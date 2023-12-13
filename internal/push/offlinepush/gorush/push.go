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

package gorush

import (
	"context"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/log"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	http2 "github.com/openimsdk/open-im-server/v3/pkg/common/http"
	"github.com/redis/go-redis/v9"
)

var (
	Terminal = []int{constant.IOSPlatformID, constant.AndroidPlatformID}
)

const (
	pushURL = "/api/push"

	pushSuccess = "ok"

	chunkSize = 300
)

type Gorush struct {
	cache cache.MsgModel
}

func NewClient(cache cache.MsgModel) *Gorush {
	return &Gorush{cache: cache}
}

func (g *Gorush) Push(ctx context.Context, userIDs []string, title, content string, opts *offlinepush.Opts) error {
	var notifications []*Notification
	for _, userID := range userIDs {
		for _, v := range Terminal {
			token, err := g.cache.GetFcmToken(ctx, userID, v)
			if err != nil {
				continue
			}
			badge := 0
			if v == constant.IOSPlatformID {
				unreadCountSum, err := g.cache.IncrUserBadgeUnreadCountSum(ctx, userID)
				if err == nil {
					badge = unreadCountSum
				}
			}
			unreadCountSum, err := g.cache.GetUserBadgeUnreadCountSum(ctx, userID)
			if err == nil && unreadCountSum != 0 {
				badge = unreadCountSum
			} else if err == redis.Nil || unreadCountSum == 0 {
				badge = 1
			}
			notification := NewNotification([]string{token}, v, title, content, opts, badge)
			notifications = append(notifications, notification)
		}
	}
	for i := 0; i < len(notifications); i += chunkSize {
		end := i + chunkSize
		if end > len(notifications) {
			end = len(notifications)
		}
		chunk := notifications[i:end]
		if err := g.request(ctx, Notifications{Notifications: chunk}); err != nil {
			log.ZError(ctx, "gorush push notifications failed", err, "notifications length", len(chunk), "title", title)
			continue
		}
	}
	return nil
}

func (g *Gorush) request(ctx context.Context, req interface{}) error {
	resp := &Resp{}
	err := http2.PostReturn(ctx, config.Config.Push.Gorush.PushUrl+pushURL, nil, req, resp, 5)
	if err != nil {
		return err
	}
	return resp.parseError()
}

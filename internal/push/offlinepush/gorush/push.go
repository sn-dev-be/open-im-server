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
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	http2 "github.com/openimsdk/open-im-server/v3/pkg/common/http"
)

var (
	Terminal = []int{constant.IOSPlatformID, constant.AndroidPlatformID}
)

const (
	pushURL = "/api/push"

	SinglePushCountLimit = 1000

	pushSuccess = "ok"
)

type Gorush struct {
	cache cache.MsgModel
}

func NewClient(cache cache.MsgModel) *Gorush {
	return &Gorush{cache: cache}
}

func (g *Gorush) Push(ctx context.Context, userIDs []string, title, content string, opts *offlinepush.Opts) error {
	var n []*Notification
	for _, platform := range Terminal {
		var platformTokens []string
		for _, account := range userIDs {
			token, err := g.cache.GetFcmToken(ctx, account, platform)
			if err == nil {
				platformTokens = append(platformTokens, token)
			}
		}
		n = append(n, NewNotifications(platformTokens, platform, title, content)...)
	}
	if len(n) > 0 {
		return g.request(ctx, Notifications{Notifications: n})
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

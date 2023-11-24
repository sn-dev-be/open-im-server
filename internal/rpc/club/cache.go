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

	pbclub "github.com/OpenIMSDK/protocol/club"

	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
)

func (c *clubServer) GetServerMemberCache(
	ctx context.Context,
	req *pbclub.GetServerMemberCacheReq,
) (resp *pbclub.GetServerMemberCacheResp, err error) {
	members, err := c.ClubDatabase.TakeServerMember(ctx, req.ServerID, req.ServerMemberID)
	if err != nil {
		return nil, err
	}
	resp = &pbclub.GetServerMemberCacheResp{Member: convert.Db2PbServerMember(members)}
	return resp, nil
}

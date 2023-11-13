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

package api

import (
	"github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/tools/a2r"
	"github.com/gin-gonic/gin"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

type ClubApi rpcclient.Club

func NewClubApi(client rpcclient.Club) ClubApi {
	return ClubApi(client)
}

func (o *ClubApi) CreateServer(c *gin.Context) {
	a2r.Call(club.ClubClient.CreateServer, o.Client, c)
}

func (o *ClubApi) GetServerRecommendedList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerRecommendedList, o.Client, c)
}

func (o *ClubApi) GetJoinedServerList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetJoinedServerList, o.Client, c)
}

func (o *ClubApi) GetServerDetails(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerDetails, o.Client, c)
}

func (o *ClubApi) CreateGroupCategory(c *gin.Context) {
	a2r.Call(club.ClubClient.CreateGroupCategory, o.Client, c)
}

func (o *ClubApi) GetJoinedServerGroupList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetJoinedServerGroupList, o.Client, c)
}

func (o *ClubApi) JoinServer(c *gin.Context) {
	a2r.Call(club.ClubClient.JoinServer, o.Client, c)
}

func (o *ClubApi) QuitServer(c *gin.Context) {
	a2r.Call(club.ClubClient.QuitServer, o.Client, c)
}

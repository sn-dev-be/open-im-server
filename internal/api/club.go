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

// /server
func (o *ClubApi) CreateServer(c *gin.Context) {
	a2r.Call(club.ClubClient.CreateServer, o.Client, c)
}

func (o *ClubApi) SetServerInfo(c *gin.Context) {
	a2r.Call(club.ClubClient.SetServerInfo, o.Client, c)
}

func (o *ClubApi) GetServerRecommendedList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerRecommendedList, o.Client, c)
}

func (o *ClubApi) GetJoinedServerList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetJoinedServerList, o.Client, c)
}

func (o *ClubApi) GetServersInfo(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServersInfo, o.Client, c)
}

func (o *ClubApi) JoinServer(c *gin.Context) {
	a2r.Call(club.ClubClient.JoinServer, o.Client, c)
}

func (o *ClubApi) QuitServer(c *gin.Context) {
	a2r.Call(club.ClubClient.QuitServer, o.Client, c)
}

func (o *ClubApi) TransferServerOwner(c *gin.Context) {
	a2r.Call(club.ClubClient.TransferServerOwner, o.Client, c)
}

func (o *ClubApi) DismissServer(c *gin.Context) {
	a2r.Call(club.ClubClient.DismissServer, o.Client, c)
}

func (o *ClubApi) SearchServer(c *gin.Context) {
	a2r.Call(club.ClubClient.SearchServer, o.Client, c)
}

// /groupCategory
func (o *ClubApi) CreateGroupCategory(c *gin.Context) {
	a2r.Call(club.ClubClient.CreateGroupCategory, o.Client, c)
}

// /group
func (o *ClubApi) GetJoinedServerGroupList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetJoinedServerGroupList, o.Client, c)
}

func (o *ClubApi) CreateServerGroup(c *gin.Context) {
	a2r.Call(club.ClubClient.CreateServerGroup, o.Client, c)
}

// /serverRequest
func (o *ClubApi) ApplicationServerResponse(c *gin.Context) {
	a2r.Call(club.ClubClient.ServerApplicationResponse, o.Client, c)
}

func (o *ClubApi) GetRecvServerApplicationList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerApplicationList, o.Client, c)
}

func (o *ClubApi) GetUserReqServerApplicationList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetUserReqApplicationList, o.Client, c)
}

func (o *ClubApi) GetServerUsersReqApplicationList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerUsersReqApplicationList, o.Client, c)
}

// /serverMember
func (o *ClubApi) GetServerMembersInfo(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerMemberList, o.Client, c)
}

func (o *ClubApi) GetServerMemberList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerMemberList, o.Client, c)
}

func (o *ClubApi) KickServerMember(c *gin.Context) {
	a2r.Call(club.ClubClient.KickServerMember, o.Client, c)
}

func (o *ClubApi) MuteServerMember(c *gin.Context) {
	a2r.Call(club.ClubClient.MuteServerMember, o.Client, c)
}

func (o *ClubApi) CancelMuteServerMember(c *gin.Context) {
	a2r.Call(club.ClubClient.CancelMuteServerMember, o.Client, c)
}

func (o *ClubApi) SetServerMemberInfo(c *gin.Context) {
	a2r.Call(club.ClubClient.SetServerMemberInfo, o.Client, c)
}

// /groupCategory
func (o *ClubApi) GetServerBlackList(c *gin.Context) {
	a2r.Call(club.ClubClient.GetServerBlackList, o.Client, c)
}

func (o *ClubApi) BanServerMember(c *gin.Context) {
	a2r.Call(club.ClubClient.BanServerMember, o.Client, c)
}

func (o *ClubApi) CancelBanServerMember(c *gin.Context) {
	a2r.Call(club.ClubClient.CancelBanServerMember, o.Client, c)
}

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
	"context"
	"net/http"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/apiresp"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/tokenverify"

	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/cache"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/OpenIMSDK/tools/discoveryregistry"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mw"

	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/rpcclient"
)

func NewGinRouter(discov discoveryregistry.SvcDiscoveryRegistry, rdb redis.UniversalClient) *gin.Engine {
	discov.AddOption(mw.GrpcClient(), grpc.WithTransportCredentials(insecure.NewCredentials())) // 默认RPC中间件
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("required_if", RequiredIf)
	}
	log.ZInfo(context.Background(), "load config", "config", config.Config)
	r.Use(gin.Recovery(), mw.CorsHandler(), mw.GinParseOperationID())
	// init rpc client here
	userRpc := rpcclient.NewUser(discov)
	groupRpc := rpcclient.NewGroup(discov)
	friendRpc := rpcclient.NewFriend(discov)
	messageRpc := rpcclient.NewMessage(discov)
	conversationRpc := rpcclient.NewConversation(discov)
	authRpc := rpcclient.NewAuth(discov)
	thirdRpc := rpcclient.NewThird(discov)
	clubRpc := rpcclient.NewClub(discov)
	cronRpc := rpcclient.NewCron(discov)

	u := NewUserApi(*userRpc)
	m := NewMessageApi(messageRpc, userRpc)
	ParseToken := GinParseToken(rdb)
	userRouterGroup := r.Group("/user")
	{
		userRouterGroup.POST("/user_register", u.UserRegister)
		userRouterGroup.POST("/update_user_info", ParseToken, u.UpdateUserInfo)
		userRouterGroup.POST("/set_user_setting", ParseToken, u.SetUserSetting)
		userRouterGroup.POST("/get_users_info", ParseToken, u.GetUsersPublicInfo)
		userRouterGroup.POST("/get_all_users_uid", ParseToken, u.GetAllUsersID)
		userRouterGroup.POST("/account_check", ParseToken, u.AccountCheck)
		userRouterGroup.POST("/get_users", ParseToken, u.GetUsers)
		userRouterGroup.POST("/get_users_online_status", ParseToken, u.GetUsersOnlineStatus)
		userRouterGroup.POST("/get_users_online_token_detail", ParseToken, u.GetUsersOnlineTokenDetail)
		userRouterGroup.POST("/subscribe_users_status", ParseToken, u.SubscriberStatus)
		userRouterGroup.POST("/get_users_status", ParseToken, u.GetUserStatus)
		userRouterGroup.POST("/get_subscribe_users_status", ParseToken, u.GetSubscribeUsersStatus)
	}
	// friend routing group
	friendRouterGroup := r.Group("/friend", ParseToken)
	{
		f := NewFriendApi(*friendRpc)
		friendRouterGroup.POST("/delete_friend", f.DeleteFriend)
		friendRouterGroup.POST("/get_friend_apply_list", f.GetFriendApplyList)
		friendRouterGroup.POST("/get_designated_friend_apply", f.GetDesignatedFriendsApply)
		friendRouterGroup.POST("/get_self_friend_apply_list", f.GetSelfApplyList)
		friendRouterGroup.POST("/get_friend_list", f.GetFriendList)
		friendRouterGroup.POST("/get_designated_friends", f.GetDesignatedFriends)
		friendRouterGroup.POST("/add_friend", f.ApplyToAddFriend)
		friendRouterGroup.POST("/add_friend_response", f.RespondFriendApply)
		friendRouterGroup.POST("/set_friend_remark", f.SetFriendRemark)
		friendRouterGroup.POST("/add_black", f.AddBlack)
		friendRouterGroup.POST("/get_black_list", f.GetPaginationBlacks)
		friendRouterGroup.POST("/remove_black", f.RemoveBlack)
		friendRouterGroup.POST("/import_friend", f.ImportFriends)
		friendRouterGroup.POST("/is_friend", f.IsFriend)
		friendRouterGroup.POST("/get_friend_id", f.GetFriendIDs)
		friendRouterGroup.POST("/get_specified_friends_info", f.GetSpecifiedFriendsInfo)
	}
	g := NewGroupApi(*groupRpc)
	groupRouterGroup := r.Group("/group", ParseToken)
	{
		groupRouterGroup.POST("/create_group", g.CreateGroup)
		groupRouterGroup.POST("/set_group_info", g.SetGroupInfo)
		groupRouterGroup.POST("/join_group", g.JoinGroup)
		groupRouterGroup.POST("/quit_group", g.QuitGroup)
		groupRouterGroup.POST("/group_application_response", g.ApplicationGroupResponse)
		groupRouterGroup.POST("/transfer_group", g.TransferGroupOwner)
		groupRouterGroup.POST("/get_recv_group_applicationList", g.GetRecvGroupApplicationList)
		groupRouterGroup.POST("/get_user_req_group_applicationList", g.GetUserReqGroupApplicationList)
		groupRouterGroup.POST("/get_group_users_req_application_list", g.GetGroupUsersReqApplicationList)
		groupRouterGroup.POST("/get_groups_info", g.GetGroupsInfo)
		groupRouterGroup.POST("/kick_group", g.KickGroupMember)
		groupRouterGroup.POST("/get_group_members_info", g.GetGroupMembersInfo)
		groupRouterGroup.POST("/get_group_member_list", g.GetGroupMemberList)
		groupRouterGroup.POST("/invite_user_to_group", g.InviteUserToGroup)
		groupRouterGroup.POST("/get_joined_group_list", g.GetJoinedGroupList)
		groupRouterGroup.POST("/dismiss_group", g.DismissGroup) //
		groupRouterGroup.POST("/mute_group_member", g.MuteGroupMember)
		groupRouterGroup.POST("/cancel_mute_group_member", g.CancelMuteGroupMember)
		groupRouterGroup.POST("/mute_group", g.MuteGroup)
		groupRouterGroup.POST("/cancel_mute_group", g.CancelMuteGroup)
		groupRouterGroup.POST("/set_group_member_info", g.SetGroupMemberInfo)
		groupRouterGroup.POST("/get_group_abstract_info", g.GetGroupAbstractInfo)
		groupRouterGroup.POST("/get_groups", g.GetGroups)
		groupRouterGroup.POST("/get_group_member_user_id", g.GetGroupMemberUserIDs)
		groupRouterGroup.POST("/save_group", g.SaveGroup)
		groupRouterGroup.POST("/unsave_group", g.UnsaveGroup)
		groupRouterGroup.POST("/get_saved_group_list", g.GetSavedGroupList)
	}
	superGroupRouterGroup := r.Group("/super_group", ParseToken)
	{
		superGroupRouterGroup.POST("/get_joined_group_list", g.GetJoinedSuperGroupList)
		superGroupRouterGroup.POST("/get_groups_info", g.GetSuperGroupsInfo)
	}
	// certificate
	authRouterGroup := r.Group("/auth")
	{
		a := NewAuthApi(*authRpc)
		authRouterGroup.POST("/user_token", a.UserToken)
		authRouterGroup.POST("/parse_token", a.ParseToken)
		authRouterGroup.POST("/force_logout", ParseToken, a.ForceLogout)
	}
	// Third service
	thirdGroup := r.Group("/third", ParseToken)
	{
		thirdGroup.GET("/prometheus", GetPrometheus)
		t := NewThirdApi(*thirdRpc)
		thirdGroup.POST("/fcm_update_token", t.FcmUpdateToken)
		thirdGroup.POST("/set_app_badge", t.SetAppBadge)

		logs := thirdGroup.Group("/logs")
		logs.POST("/upload", t.UploadLogs)
		logs.POST("/delete", t.DeleteLogs)
		logs.POST("/search", t.SearchLogs)

		objectGroup := r.Group("/object", ParseToken)

		objectGroup.POST("/part_limit", t.PartLimit)
		objectGroup.POST("/part_size", t.PartSize)
		objectGroup.POST("/initiate_multipart_upload", t.InitiateMultipartUpload)
		objectGroup.POST("/auth_sign", t.AuthSign)
		objectGroup.POST("/complete_multipart_upload", t.CompleteMultipartUpload)
		objectGroup.POST("/access_url", t.AccessURL)
		objectGroup.GET("/*name", t.ObjectRedirect)
	}
	// Message
	msgGroup := r.Group("/msg", ParseToken)
	{
		msgGroup.POST("/newest_seq", m.GetSeq)
		msgGroup.POST("/search_msg", m.SearchMsg)
		msgGroup.POST("/send_msg", m.SendMessage)
		msgGroup.POST("/send_business_notification", m.SendBusinessNotification)
		msgGroup.POST("/pull_msg_by_seq", m.PullMsgBySeqs)
		msgGroup.POST("/revoke_msg", m.RevokeMsg)
		msgGroup.POST("/mark_msgs_as_read", m.MarkMsgsAsRead)
		msgGroup.POST("/mark_conversation_as_read", m.MarkConversationAsRead)
		msgGroup.POST("/get_conversations_has_read_and_max_seq", m.GetConversationsHasReadAndMaxSeq)
		msgGroup.POST("/set_conversation_has_read_seq", m.SetConversationHasReadSeq)

		msgGroup.POST("/clear_conversation_msg", m.ClearConversationsMsg)
		msgGroup.POST("/user_clear_all_msg", m.UserClearAllMsg)
		msgGroup.POST("/delete_msgs", m.DeleteMsgs)
		msgGroup.POST("/delete_msg_phsical_by_seq", m.DeleteMsgPhysicalBySeq)
		msgGroup.POST("/delete_msg_physical", m.DeleteMsgPhysical)

		msgGroup.POST("/batch_send_msg", m.BatchSendMsg)
		msgGroup.POST("/check_msg_is_send_success", m.CheckMsgIsSendSuccess)
		msgGroup.POST("/get_server_time", m.GetServerTime)

		msgGroup.POST("/set_red_packet_msg_status", m.SetRedPacketMsgStatus)
	}
	// Conversation
	conversationGroup := r.Group("/conversation", ParseToken)
	{
		c := NewConversationApi(*conversationRpc)
		conversationGroup.POST("/get_all_conversations", c.GetAllConversations)
		conversationGroup.POST("/get_conversation", c.GetConversation)
		conversationGroup.POST("/get_conversations", c.GetConversations)
		conversationGroup.POST("/set_conversations", c.SetConversations)
		conversationGroup.POST("/get_conversation_offline_push_user_ids", c.GetConversationOfflinePushUserIDs)
	}

	statisticsGroup := r.Group("/statistics", ParseToken)
	{
		statisticsGroup.POST("/user/register", u.UserRegisterCount)
		statisticsGroup.POST("/user/active", m.GetActiveUser)
		statisticsGroup.POST("/group/create", g.GroupCreateCount)
		statisticsGroup.POST("/group/active", m.GetActiveGroup)
	}

	//club
	clubGroup := r.Group("/club", ParseToken)
	{
		c := NewClubApi(*clubRpc)
		clubGroup.POST("/create_server", c.CreateServer)
		clubGroup.POST("/set_server_info", c.SetServerInfo)
		clubGroup.POST("/join_server", c.JoinServer)
		clubGroup.POST("/quit_server", c.QuitServer)
		clubGroup.POST("/transfer_server", c.TransferServerOwner)
		clubGroup.POST("/get_server_recommended_list", c.GetServerRecommendedList)
		clubGroup.POST("/get_servers_info", c.GetServersInfo)
		clubGroup.POST("/dismiss_server", c.DismissServer)
		clubGroup.POST("/search_server", c.SearchServer)
		clubGroup.POST("/mute_server", c.MuteServer)
		clubGroup.POST("/cancel_mute_server", c.CancelMuteServer)

		clubGroup.POST("/create_category", c.CreateGroupCategory)
		clubGroup.POST("/set_category_info", c.SetGroupCategoryInfo)
		clubGroup.POST("/delete_category", c.DeleteGroupCategory)
		clubGroup.POST("/set_category_order", c.SetGroupCategoryOrder)
		clubGroup.POST("/get_categories", c.GetGroupCategories)

		clubGroup.POST("/get_server_role_list", c.GetServerRoleList)
		clubGroup.POST("/get_server_roles_info", c.GetServerRolesInfo)
		clubGroup.POST("/create_server_role", nil)
		clubGroup.POST("/set_server_role_info", nil)
		clubGroup.POST("/delete_server_role", nil)

		clubGroup.POST("/get_server_black_list", c.GetServerBlackList)
		clubGroup.POST("/ban_server_member", c.BanServerMember)
		clubGroup.POST("/cancel_ban_server_member", c.CancelBanServerMember)

		clubGroup.POST("/get_joined_server_group_list", c.GetJoinedServerGroupList)
		clubGroup.POST("/get_server_groups_info", c.GetServerGroupsInfo)
		clubGroup.POST("/create_server_group", c.CreateServerGroup)
		clubGroup.POST("/set_server_group_info", c.SetServerGroupInfo)
		clubGroup.POST("/set_server_group_order", c.SetServerGroupOrder)
		clubGroup.POST("/delete_server_group", c.DeleteServerGroup)
		clubGroup.POST("/mute_server_group", c.MuteServerGroup)
		clubGroup.POST("/cancel_mute_server_group", c.CancelMuteServerGroup)
		clubGroup.POST("/get_server_group_members_info", c.GetServerGroupMembersInfo)
		clubGroup.POST("/get_server_group_base_infos", c.GetServerGroupBaseInfos)

		clubGroup.POST("/get_server_members_info", c.GetServerMembersInfo)
		clubGroup.POST("/get_server_member_list", c.GetServerMemberList)
		clubGroup.POST("/kick_server_member", c.KickServerMember)
		clubGroup.POST("/mute_server_member", c.MuteServerMember)
		clubGroup.POST("/get_server_mute_records", c.GetServerMuteRecords)
		clubGroup.POST("/cancel_mute_server_member", c.CancelMuteServerMember)
		clubGroup.POST("/set_server_member_info", c.SetServerMemberInfo)
		clubGroup.POST("/get_server_mute_list", nil)
		clubGroup.POST("/get_joined_server_list", c.GetJoinedServerList)
		clubGroup.POST("/set_joined_servers_order", c.SetJoinedServersOrder)

		clubGroup.POST("/server_application_response", c.ApplicationServerResponse)
		clubGroup.POST("/get_recv_server_application_list", c.GetRecvServerApplicationList)
		clubGroup.POST("/get_user_req_server_application_list", c.GetUserReqServerApplicationList)
		clubGroup.POST("/get_server_users_req_application_list", c.GetServerUsersReqApplicationList)
	}

	// cron
	cronGroup := r.Group("/cron", ParseToken)
	{
		c := NewCronApi(*cronRpc)
		cronGroup.POST("/set_clear_msg_job", c.SetClearMsgJob)
		cronGroup.POST("/get_clear_msg_job", c.GetClearMsgJob)
	}

	return r
}

func GinParseToken(rdb redis.UniversalClient) gin.HandlerFunc {
	dataBase := controller.NewAuthDatabase(
		cache.NewMsgCacheModel(rdb),
		config.Config.Secret,
		config.Config.TokenPolicy.Expire,
	)
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost:
			token := c.Request.Header.Get(constant.Token)
			if token == "" {
				log.ZWarn(c, "header get token error", errs.ErrArgs.Wrap("header must have token"))
				apiresp.GinError(c, errs.ErrArgs.Wrap("header must have token"))
				c.Abort()
				return
			}
			if token == config.Config.Secret {
				c.Set(constant.OpUserID, config.Config.Manager.UserID[0])
				c.Next()
				return
			}
			claims, err := tokenverify.GetClaimFromToken(token, authverify.Secret())
			if err != nil {
				log.ZWarn(c, "jwt get token error", errs.ErrTokenUnknown.Wrap())
				apiresp.GinError(c, errs.ErrTokenUnknown.Wrap())
				c.Abort()
				return
			}
			m, err := dataBase.GetTokensWithoutError(c, claims.UserID, claims.PlatformID)
			if err != nil {
				log.ZWarn(c, "cache get token error", errs.ErrTokenNotExist.Wrap())
				apiresp.GinError(c, errs.ErrTokenNotExist.Wrap())
				c.Abort()
				return
			}
			if len(m) == 0 {
				log.ZWarn(c, "cache do not exist token error", errs.ErrTokenNotExist.Wrap())
				apiresp.GinError(c, errs.ErrTokenNotExist.Wrap())
				c.Abort()
				return
			}
			if v, ok := m[token]; ok {
				switch v {
				case constant.NormalToken:
				case constant.KickedToken:
					log.ZWarn(c, "cache kicked token error", errs.ErrTokenKicked.Wrap())
					apiresp.GinError(c, errs.ErrTokenKicked.Wrap())
					c.Abort()
					return
				default:
					log.ZWarn(c, "cache unknown token error", errs.ErrTokenUnknown.Wrap())
					apiresp.GinError(c, errs.ErrTokenUnknown.Wrap())
					c.Abort()
					return
				}
			} else {
				apiresp.GinError(c, errs.ErrTokenNotExist.Wrap())
				c.Abort()
				return
			}
			c.Set(constant.OpUserPlatform, constant.PlatformIDToName(claims.PlatformID))
			c.Set(constant.OpUserID, claims.UserID)
			c.Next()
		}
	}
}

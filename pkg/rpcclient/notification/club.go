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
	"errors"
	"fmt"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	"github.com/openimsdk/open-im-server/v3/pkg/permissions"
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

func (c *ClubNotificationSender) getUser(ctx context.Context, userID string) (*sdkws.PublicUserInfo, error) {
	users, err := c.getUsersInfo(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errs.ErrUserIDNotFound.Wrap(fmt.Sprintf("user %s not found", userID))
	}
	return &sdkws.PublicUserInfo{
		UserID:   users[0].GetUserID(),
		Nickname: users[0].GetNickname(),
		FaceURL:  users[0].GetFaceURL(),
		Ex:       users[0].GetEx(),
	}, nil
}

func (c *ClubNotificationSender) getServerInfo(ctx context.Context, serverID string) (*sdkws.ServerInfo, error) {
	svr, err := c.db.TakeServer(ctx, serverID)
	if err != nil {
		return nil, err
	}
	num, err := c.db.FindServerMemberNum(ctx, serverID)
	if err != nil {
		return nil, err
	}
	owner, err := c.db.TakeServerOwner(ctx, serverID)
	if err != nil {
		return nil, err
	}
	return &sdkws.ServerInfo{
		ServerID:             svr.ServerID,
		ServerName:           svr.ServerName,
		Icon:                 svr.Icon,
		Banner:               svr.Banner,
		Description:          svr.Description,
		OwnerUserID:          owner.UserID,
		CreateTime:           svr.CreateTime.UnixMilli(),
		MemberNumber:         num,
		Ex:                   svr.Ex,
		Status:               svr.Status,
		ApplyMode:            svr.ApplyMode,
		InviteMode:           svr.InviteMode,
		Searchable:           svr.Searchable,
		UserMutualAccessible: svr.UserMutualAccessible,
		CategoryNumber:       svr.CategoryNumber,
		GroupNumber:          svr.GroupNumber,
		DappID:               svr.DappID,
	}, nil
}

func (c *ClubNotificationSender) serverMemberDB2PB(member *relation.ServerMemberModel, appMangerLevel int32) *sdkws.ServerMemberFullInfo {
	return &sdkws.ServerMemberFullInfo{
		ServerID:       member.ServerID,
		UserID:         member.UserID,
		RoleLevel:      member.RoleLevel,
		JoinTime:       member.JoinTime.UnixMilli(),
		Nickname:       member.Nickname,
		FaceURL:        member.FaceURL,
		AppMangerLevel: appMangerLevel,
		JoinSource:     member.JoinSource,
		OperatorUserID: member.OperatorUserID,
		Ex:             member.Ex,
		MuteEndTime:    member.MuteEndTime.UnixMilli(),
		InviterUserID:  member.InviterUserID,
	}
}

func (c *ClubNotificationSender) fillOpUser(ctx context.Context, opUser **sdkws.ServerMemberFullInfo, serverID string) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	if opUser == nil {
		return errs.ErrInternalServer.Wrap("**sdkws.ServerMemberFullInfo is nil")
	}
	if *opUser != nil {
		return nil
	}
	userID := mcontext.GetOpUserID(ctx)
	if serverID != "" {
		if authverify.IsManagerUserID(userID) {
			*opUser = &sdkws.ServerMemberFullInfo{
				ServerID:       serverID,
				UserID:         userID,
				RoleLevel:      constant.ServerAdmin,
				AppMangerLevel: constant.AppAdmin,
			}
		} else {
			member, err := c.db.TakeServerMember(ctx, serverID, userID)
			if err == nil {
				*opUser = c.serverMemberDB2PB(member, 0)
			} else if !errs.ErrRecordNotFound.Is(err) {
				return err
			}
		}
	}
	user, err := c.getUser(ctx, userID)
	if err != nil {
		return err
	}
	if *opUser == nil {
		*opUser = &sdkws.ServerMemberFullInfo{
			ServerID:       serverID,
			UserID:         userID,
			Nickname:       user.Nickname,
			FaceURL:        user.FaceURL,
			OperatorUserID: userID,
		}
	} else {
		if (*opUser).Nickname == "" {
			(*opUser).Nickname = user.Nickname
		}
		if (*opUser).FaceURL == "" {
			(*opUser).FaceURL = user.FaceURL
		}
	}
	return nil
}

func (c *ClubNotificationSender) getNotificationAdminUserID() string {
	if len(config.Config.Manager.UserID) < 6 {
		panic("wrong manage user id config")
	}
	return config.Config.Manager.UserID[5]
}

func (c *ClubNotificationSender) getServerManageRoleUserID(ctx context.Context, serverID string) ([]string, error) {
	members, err := c.db.FindServerMemberByRole(ctx, serverID, permissions.ManageMember)
	if err != nil {
		return nil, err
	}
	fn := func(e *relation.ServerMemberModel) string { return e.UserID }
	return utils.Slice(members, fn), nil
}

func (c *ClubNotificationSender) JoinServerApplicationNotification(ctx context.Context, req *pbclub.JoinServerReq) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	server, err := c.getServerInfo(ctx, req.ServerID)
	if err != nil {
		return err
	}
	user, err := c.getUser(ctx, req.InviterUserID)
	if err != nil {
		return err
	}
	userIDs, err := c.getServerManageRoleUserID(ctx, req.ServerID)
	if err != nil {
		return err
	}
	userIDs = append(userIDs)
	log.ZInfo(ctx, "JoinServerApplicationNotification", "server", req.ServerID, "ManageMember userIDs", userIDs)
	tips := &sdkws.JoinServerApplicationTips{Server: server, Applicant: user, ReqMsg: req.ReqMessage, HandleResult: constant.ServerResponseNotHandle}
	for _, userID := range utils.Distinct(userIDs) {
		err = c.Notification(ctx, c.getNotificationAdminUserID(), userID, constant.JoinServerApplicationNotification, tips)
		if err != nil {
			log.ZError(ctx, "JoinServerApplicationNotification failed", err, "server", req.ServerID, "userID", userID)
		}
	}
	return nil
}

func (c *ClubNotificationSender) ServerApplicationAcceptedNotification(ctx context.Context, req *pbclub.ServerApplicationResponseReq) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	server, err := c.getServerInfo(ctx, req.ServerID)
	if err != nil {
		return err
	}
	tips := &sdkws.ServerApplicationAcceptedTips{Server: server, HandleMsg: req.HandledMsg, ReceiverAs: 1}
	if err := c.fillOpUser(ctx, &tips.OpUser, tips.Server.ServerID); err != nil {
		return err
	}
	err = c.Notification(ctx, c.getNotificationAdminUserID(), req.FromUserID, constant.ServerApplicationAcceptedNotification, tips)
	if err != nil {
		log.ZError(ctx, "failed", err)
	}
	return nil
}

func (c *ClubNotificationSender) ServerApplicationRejectedNotification(ctx context.Context, req *pbclub.ServerApplicationResponseReq) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	server, err := c.getServerInfo(ctx, req.ServerID)
	if err != nil {
		return err
	}
	tips := &sdkws.ServerApplicationRejectedTips{Server: server, HandleMsg: req.HandledMsg}
	if err := c.fillOpUser(ctx, &tips.OpUser, tips.Server.ServerID); err != nil {
		return err
	}
	err = c.Notification(ctx, c.getNotificationAdminUserID(), req.FromUserID, constant.ServerApplicationRejectedNotification, tips)
	if err != nil {
		log.ZError(ctx, "failed", err)
	}
	return nil
}

func (c *ClubNotificationSender) ServerCreatedNotification(ctx context.Context, tips *sdkws.ServerCreatedTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	if err := c.fillOpUser(ctx, &tips.OpUser, tips.Server.ServerID); err != nil {
		return err
	}

	if tips.ServerGroupList != nil && len(tips.ServerGroupList) > 0 {
		groupIDs := utils.Slice(tips.ServerGroupList, func(g *sdkws.GroupInfo) string { return g.GroupID })
		c.sendNotificationToServerGroups(ctx, mcontext.GetOpUserID(ctx), groupIDs, constant.ServerCreatedNotification, tips)

	}
	// for _, group := range tips.ServerGroupList {
	// 	c.Notification(ctx, mcontext.GetOpUserID(ctx), group.GroupID, constant.ServerCreatedNotification, tips)
	// }

	return nil
}

func (c *ClubNotificationSender) ServerDismissNotification(ctx context.Context, tips *sdkws.ServerDissmissedTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	// if err := c.fillOpUser(ctx, &tips.OpUser, tips.ServerID); err != nil {
	// 	return err
	// }

	// if tips.ServerGroupList != nil && len(tips.ServerGroupList) > 0 {
	// 	groupIDs := utils.Slice(tips.ServerGroupList, func(g *sdkws.GroupInfo) string { return g.GroupID })
	// 	c.sendNotificationToServerGroups(ctx, mcontext.GetOpUserID(ctx), groupIDs, constant.ServerCreatedNotification, tips)

	// }
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerDismissedNotification, tips)
	}

	return nil
}

func (c *ClubNotificationSender) ServerInfoSetNotification(ctx context.Context, tips *sdkws.ServerInfoSetTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	// if err := c.fillOpUser(ctx, &tips.OpUser, tips.Server.ServerID); err != nil {
	// 	return err
	// }

	// if tips.ServerGroupList != nil && len(tips.ServerGroupList) > 0 {
	// 	groupIDs := utils.Slice(tips.ServerGroupList, func(g *sdkws.GroupInfo) string { return g.GroupID })
	// 	c.sendNotificationToServerGroups(ctx, mcontext.GetOpUserID(ctx), groupIDs, constant.ServerCreatedNotification, tips)

	// }
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerInfoSetNotification, tips)
	}

	return nil
}

func (c *ClubNotificationSender) ServerMemberMutedNotification(ctx context.Context, tips *sdkws.ServerMemberMutedTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerMemberMutedNotification, tips)
	}
	return nil
}

func (c *ClubNotificationSender) ServerMemberCancelMutedNotification(ctx context.Context, tips *sdkws.ServerMemberCancelMutedTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerMemberCancelMutedNotification, tips)
	}
	return nil
}

func (c *ClubNotificationSender) ServerMemberEnterNotification(ctx context.Context, serverID, userID string) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	tips := &sdkws.ServerMemberEnterTips{ServerID: serverID, MemberUserIDList: []string{userID}}

	if err := c.fillOpUser(ctx, &tips.User, serverID); err != nil {
		return err
	}
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerMemberEnterNotification, tips)
	}
	return nil
}

func (c *ClubNotificationSender) ServerMemberQuitNotification(ctx context.Context, tips *sdkws.ServerMemberQuitTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerMemberQuitNotification, tips)
	}
	return nil
}

func (c *ClubNotificationSender) ServerMemberKickedNotification(ctx context.Context, tips *sdkws.ServerMemberKickedTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerMemberKickedNotification, tips)
	}
	return nil
}

func (c *ClubNotificationSender) ServerMemberInfoSetNotification(ctx context.Context, tips *sdkws.ServerMemberInfoSetTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerMemberInfoSetNotification, tips)
	}
	return nil
}

func (c *ClubNotificationSender) sendNotificationToServerGroups(ctx context.Context, opUserID string, groupIDs []string, notificationType int32, tips interface{}) error {
	for _, groupID := range groupIDs {
		switch v := tips.(type) {
		case *sdkws.ServerCreatedTips:
			c.Notification(ctx, mcontext.GetOpUserID(ctx), groupID, notificationType, v)
		default:
			return utils.Wrap(errors.New("unsupported tips type"), "")
		}
	}
	return nil
}

func (c *ClubNotificationSender) ServerGroupCreatedNotification(ctx context.Context, tips *sdkws.ServerGroupCreatedTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	if err := c.fillOpUser(ctx, &tips.OpUser, tips.ServerID); err != nil {
		return err
	}
	return c.Notification(ctx, mcontext.GetOpUserID(ctx), tips.Group.GroupID, constant.ServerGroupCreatedNotification, tips)
}

func (c *ClubNotificationSender) ServerGroupDismissNotification(ctx context.Context, tips *sdkws.ServerGroupDismissTips) (err error) {
	defer log.ZDebug(ctx, "return")
	defer func() {
		if err != nil {
			log.ZError(ctx, utils.GetFuncName(1)+" failed", err)
		}
	}()
	// if err := c.fillOpUser(ctx, &tips.OpUser, tips.ServerID); err != nil {
	// 	return err
	// }
	for _, userID := range tips.MemberUserIDList {
		c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.ServerGroupDismissNotification, tips)
	}
	return nil
}

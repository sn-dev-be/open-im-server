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
	"fmt"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/controller"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
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

func (c *ClubNotificationSender) getServerManageRoleUserID(ctx context.Context, serverID string) ([]string, error) {
	members, err := c.db.FindServerMemberByRole(ctx, serverID, "managerMember")
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
	userIDs = append(userIDs, req.InviterUserID, mcontext.GetOpUserID(ctx))
	tips := &sdkws.JoinServerApplicationTips{Server: server, Applicant: user, ReqMsg: req.ReqMessage}
	for _, userID := range utils.Distinct(userIDs) {
		err = c.Notification(ctx, mcontext.GetOpUserID(ctx), userID, constant.JoinServerApplicationNotification, tips)
		if err != nil {
			log.ZError(ctx, "JoinServerApplicationNotification failed", err, "server", req.ServerID, "userID", userID)
		}
	}
	return nil
}

package club

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	"github.com/openimsdk/open-im-server/v3/pkg/msgprocessor"
)

func (c *clubServer) JoinServer(ctx context.Context, req *pbclub.JoinServerReq) (resp *pbclub.JoinServerResp, err error) {
	defer log.ZInfo(ctx, "JoinServer.Return")
	user, err := c.User.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	server, err := c.ClubDatabase.TakeServer(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	if server.Status != constant.ServerOk {
		return nil, errs.ErrDismissedAlready.Wrap()
	}

	_, err = c.ClubDatabase.TakeServerMember(ctx, req.ServerID, req.UserID)
	if err == nil {
		return nil, errs.ErrArgs.Wrap("already in server")
	} else if !c.IsNotFound(err) && utils.Unwrap(err) != errs.ErrRecordNotFound {
		return nil, err
	}

	serverRole, err := c.getServerRoleByPriority(ctx, req.ServerID, constant.ServerOrdinaryUsers)
	if err != nil {
		return nil, errs.ErrRecordNotFound.Wrap("server role is not exists")
	}

	if server.ApplyMode == constant.JoinServerDirectly {
		serverMember := &relationtb.ServerMemberModel{
			ServerID:      req.ServerID,
			UserID:        req.UserID,
			Nickname:      user.Nickname,
			ServerRoleID:  serverRole.RoleID,
			JoinSource:    req.JoinSource,
			InviterUserID: req.InviterUserID,
			JoinTime:      time.Now(),
			MuteEndTime:   time.Unix(0, 0),
		}
		err = c.ClubDatabase.CreateServerMember(ctx, []*relationtb.ServerMemberModel{serverMember})
		if err != nil {
			return nil, err
		}
		//todo 是否需要发送notification
		return &pbclub.JoinServerResp{}, nil
	} else {
		serverRequest := relationtb.ServerRequestModel{
			UserID:      req.InviterUserID,
			ReqMsg:      req.ReqMessage,
			ServerID:    req.ServerID,
			JoinSource:  req.JoinSource,
			ReqTime:     time.Now(),
			HandledTime: time.Unix(0, 0),
		}
		if err := c.ClubDatabase.CreateServerRequest(ctx, []*relationtb.ServerRequestModel{&serverRequest}); err != nil {
			return nil, err
		}
		c.Notification.JoinServerApplicationNotification(ctx, req)
	}
	return resp, nil
}

func (c *clubServer) QuitServer(ctx context.Context, req *pbclub.QuitServerReq) (*pbclub.QuitServerResp, error) {
	resp := &pbclub.QuitServerResp{}
	if req.UserID == "" {
		req.UserID = mcontext.GetOpUserID(ctx)
	} else {
		if err := authverify.CheckAccessV3(ctx, req.UserID); err != nil {
			return nil, err
		}
	}

	info, err := c.TakeServerMember(ctx, req.ServerID, req.UserID)
	if err != nil {
		return nil, err
	}
	if info.RoleLevel == constant.ServerOwner {
		return nil, errs.ErrNoPermission.Wrap("server owner can't quit")
	}
	err = c.ClubDatabase.DeleteServerMember(ctx, req.ServerID, []string{req.UserID})
	if err != nil {
		return nil, err
	}

	//todo 发送notification
	//_ = c.Notification.MemberQuitNotification(ctx, c.groupMemberDB2PB(info, 0))

	if err := c.deleteMemberAndSetConversationSeq(ctx, req.ServerID, []string{req.UserID}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *clubServer) createServerMember(ctx context.Context, serverID, user_id, nickname, serverRoleID, invitedUserID, ex string, roleLevel, joinSource int32) error {
	server_member := &relationtb.ServerMemberModel{
		ServerID:      serverID,
		UserID:        user_id,
		Nickname:      nickname,
		ServerRoleID:  serverRoleID,
		RoleLevel:     roleLevel,
		JoinSource:    joinSource,
		InviterUserID: invitedUserID,
		Ex:            ex,
		MuteEndTime:   time.UnixMilli(0),
		JoinTime:      time.Now(),
	}
	if err := c.ClubDatabase.CreateServerMember(ctx, []*relationtb.ServerMemberModel{server_member}); err != nil {
		return err
	}
	return nil
}

func (c *clubServer) serverMemberHashCode(ctx context.Context, serverID string) (uint64, error) {
	userIDs, err := c.ClubDatabase.FindServerMemberUserID(ctx, serverID)
	if err != nil {
		return 0, err
	}
	var members []*sdkws.ServerMemberFullInfo
	if len(userIDs) > 0 {
		resp, err := c.GetServerMembersInfo(ctx, &pbclub.GetServerMembersInfoReq{ServerID: serverID, UserIDs: userIDs})
		if err != nil {
			return 0, err
		}
		members = resp.Members
		utils.Sort(userIDs, true)
	}
	memberMap := utils.SliceToMap(members, func(e *sdkws.ServerMemberFullInfo) string {
		return e.UserID
	})
	res := make([]*sdkws.ServerMemberFullInfo, 0, len(members))
	for _, userID := range userIDs {
		member, ok := memberMap[userID]
		if !ok {
			continue
		}
		member.AppMangerLevel = 0
		res = append(res, member)
	}
	data, err := json.Marshal(res)
	if err != nil {
		return 0, err
	}
	sum := md5.Sum(data)
	return binary.BigEndian.Uint64(sum[:]), nil
}

func (c *clubServer) GetServerMemberList(ctx context.Context, req *pbclub.GetServerMemberListReq) (*pbclub.GetServerMemberListResp, error) {
	resp := &pbclub.GetServerMemberListResp{}
	total, members, err := c.PageGetServerMember(ctx, req.ServerID, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	log.ZDebug(ctx, "GetServerMemberList", "total", total, "members", members, "length", len(members))
	if err != nil {
		return nil, err
	}
	resp.Total = total
	resp.Members = utils.Batch(convert.Db2PbServerMember, members)
	log.ZDebug(ctx, "GetServerMemberList", "resp", resp, "length", len(resp.Members))
	return resp, nil
}

func (c *clubServer) GetServerMembersInfo(ctx context.Context, req *pbclub.GetServerMembersInfoReq) (*pbclub.GetServerMembersInfoResp, error) {
	resp := &pbclub.GetServerMembersInfoResp{}
	if len(req.UserIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("userIDs empty")
	}
	if req.ServerID == "" {
		return nil, errs.ErrArgs.Wrap("serverID empty")
	}
	members, err := c.FindServerMember(ctx, []string{req.ServerID}, req.UserIDs, nil)
	if err != nil {
		return nil, err
	}
	publicUserInfoMap, err := c.GetPublicUserInfoMap(ctx, utils.Filter(members, func(e *relationtb.ServerMemberModel) (string, bool) {
		return e.UserID, e.Nickname == "" || e.FaceURL == ""
	}), true)
	if err != nil {
		return nil, err
	}
	resp.Members = utils.Slice(members, func(e *relationtb.ServerMemberModel) *sdkws.ServerMemberFullInfo {
		if userInfo, ok := publicUserInfoMap[e.UserID]; ok {
			if e.Nickname == "" {
				e.Nickname = userInfo.Nickname
			}
			if e.FaceURL == "" {
				e.FaceURL = userInfo.FaceURL
			}
		}
		return convert.Db2PbServerMember(e)
	})
	return resp, nil
}

func (c *clubServer) KickServerMember(ctx context.Context, req *pbclub.KickServerMemberReq) (*pbclub.KickServerMemberResp, error) {
	resp := &pbclub.KickServerMemberResp{}
	server, err := c.ClubDatabase.TakeServer(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	if len(req.KickedUserIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("KickedUserIDs empty")
	}
	if utils.IsDuplicateStringSlice(req.KickedUserIDs) {
		return nil, errs.ErrArgs.Wrap("KickedUserIDs duplicate")
	}
	opUserID := mcontext.GetOpUserID(ctx)
	if utils.IsContain(opUserID, req.KickedUserIDs) {
		return nil, errs.ErrArgs.Wrap("opUserID in KickedUserIDs")
	}

	members, err := c.FindServerMember(ctx, []string{req.ServerID}, append(req.KickedUserIDs, opUserID), nil)
	if err != nil {
		return nil, err
	}
	memberMap := make(map[string]*relationtb.ServerMemberModel)
	for i, member := range members {
		memberMap[member.UserID] = members[i]
	}
	isAppManagerUid := authverify.IsAppManagerUid(ctx)
	opMember := memberMap[opUserID]
	for _, userID := range req.KickedUserIDs {
		member, ok := memberMap[userID]
		if !ok {
			return nil, errs.ErrUserIDNotFound.Wrap(userID)
		}
		if !isAppManagerUid {
			if opMember == nil {
				return nil, errs.ErrNoPermission.Wrap("opUserID no in server")
			}
			switch opMember.RoleLevel {
			case constant.ServerOwner:
			case constant.ServerAdmin:
				if member.RoleLevel == constant.ServerOwner || member.RoleLevel == constant.ServerAdmin {
					return nil, errs.ErrNoPermission.Wrap("server admins cannot remove the server owner and other admins")
				}
			case constant.ServerOrdinaryUsers:
				return nil, errs.ErrNoPermission.Wrap("opUserID no permission")
			default:
				return nil, errs.ErrNoPermission.Wrap("opUserID roleLevel unknown")
			}
		}
	}
	// num, err := c.ClubDatabase.FindServerMemberNum(ctx, req.ServerID)
	// if err != nil {
	// 	return nil, err
	// }
	// owner, err := c.FindServerMember(ctx, []string{req.ServerID}, nil, []int32{constant.ServerOwner})
	// if err != nil {
	// 	return nil, err
	// }
	if err := c.ClubDatabase.DeleteServerMember(ctx, server.ServerID, req.KickedUserIDs); err != nil {
		return nil, err
	}
	// tips := &sdkws.MemberKickedTips{
	// 	Server: &sdkws.ServerInfo{
	// 		ServerID:     server.ServerID,
	// 		ServerName:   server.ServerName,
	// 		Notification: server.Notification,
	// 		Introduction: server.Introduction,
	// 		FaceURL:      server.FaceURL,
	// 		// OwnerUserID:            owner[0].UserID,
	// 		CreateTime:             server.CreateTime.UnixMilli(),
	// 		MemberCount:            num,
	// 		Ex:                     server.Ex,
	// 		Status:                 server.Status,
	// 		CreatorUserID:          server.CreatorUserID,
	// 		ServerType:             server.ServerType,
	// 		NeedVerification:       server.NeedVerification,
	// 		LookMemberInfo:         server.LookMemberInfo,
	// 		ApplyMemberFriend:      server.ApplyMemberFriend,
	// 		NotificationUpdateTime: server.NotificationUpdateTime.UnixMilli(),
	// 		NotificationUserID:     server.NotificationUserID,
	// 	},
	// 	KickedUserList: []*sdkws.ServerMemberFullInfo{},
	// }
	// if len(owner) > 0 {
	// 	tips.Server.OwnerUserID = owner[0].UserID
	// }
	// if opMember, ok := memberMap[opUserID]; ok {
	// 	tips.OpUser = convert.Db2PbServerMember(opMember)
	// }
	// for _, userID := range req.KickedUserIDs {
	// 	tips.KickedUserList = append(tips.KickedUserList, convert.Db2PbServerMember(memberMap[userID]))
	// }
	// c.Notification.MemberKickedNotification(ctx, tips)
	// if err := c.deleteMemberAndSetConversationSeq(ctx, req.ServerID, req.KickedUserIDs); err != nil {
	// 	return nil, err
	// }
	return resp, nil
}

func (c *clubServer) GetJoinedServerList(ctx context.Context, req *pbclub.GetJoinedServerListReq) (*pbclub.GetJoinedServerListResp, error) {
	resp := &pbclub.GetJoinedServerListResp{}
	if err := authverify.CheckAccessV3(ctx, req.FromUserID); err != nil {
		return nil, err
	}
	var pageNumber, showNumber int32
	if req.Pagination != nil {
		pageNumber = req.Pagination.PageNumber
		showNumber = req.Pagination.ShowNumber
	}
	total, members, err := c.ClubDatabase.PageGetJoinServer(ctx, req.FromUserID, pageNumber, showNumber)
	if err != nil {
		return nil, err
	}
	resp.Total = total
	if len(members) == 0 {
		return resp, nil
	}
	serverIDs := utils.Slice(members, func(e *relationtb.ServerMemberModel) string {
		return e.ServerID
	})
	servers, err := c.ClubDatabase.FindServer(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	serverMemberNum, err := c.ClubDatabase.MapServerMemberNum(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	owners, err := c.FindServerMember(ctx, serverIDs, nil, []int32{constant.ServerOwner})
	if err != nil {
		return nil, err
	}
	ownerMap := utils.SliceToMap(owners, func(e *relationtb.ServerMemberModel) string {
		return e.ServerID
	})
	resp.Servers = utils.Slice(utils.Order(serverIDs, servers, func(server *relationtb.ServerModel) string {
		return server.ServerID
	}), func(server *relationtb.ServerModel) *sdkws.ServerInfo {
		var userID string
		if user := ownerMap[server.ServerID]; user != nil {
			userID = user.UserID
		}
		return convert.Db2PbServerInfo(server, userID, serverMemberNum[server.ServerID])
	})
	return resp, nil
}

func (c *clubServer) GetServerMembersCMS(ctx context.Context, req *pbclub.GetServerMembersCMSReq) (*pbclub.GetServerMembersCMSResp, error) {
	resp := &pbclub.GetServerMembersCMSResp{}
	total, members, err := c.ClubDatabase.SearchServerMember(ctx, req.UserName, []string{req.ServerID}, nil, nil, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	resp.Total = total
	publicUserInfoMap, err := c.GetPublicUserInfoMap(ctx, utils.Filter(members, func(e *relationtb.ServerMemberModel) (string, bool) {
		return e.UserID, e.Nickname == "" || e.FaceURL == ""
	}), true)
	if err != nil {
		return nil, err
	}
	resp.Members = utils.Slice(members, func(e *relationtb.ServerMemberModel) *sdkws.ServerMemberFullInfo {
		if userInfo, ok := publicUserInfoMap[e.UserID]; ok {
			if e.Nickname == "" {
				e.Nickname = userInfo.Nickname
			}
			if e.FaceURL == "" {
				e.FaceURL = userInfo.FaceURL
			}
		}
		return convert.Db2PbServerMember(e)
	})
	return resp, nil
}

func (c *clubServer) MuteServerMember(ctx context.Context, req *pbclub.MuteServerMemberReq) (*pbclub.MuteServerMemberResp, error) {
	resp := &pbclub.MuteServerMemberResp{}
	//if err := tokenverify.CheckAccessV3(ctx, req.UserID); err != nil {
	//	return nil, err
	//}
	member, err := c.TakeServerMember(ctx, req.ServerID, req.UserID)
	if err != nil {
		return nil, err
	}
	if !authverify.IsAppManagerUid(ctx) {
		opMember, err := c.TakeServerMember(ctx, req.ServerID, mcontext.GetOpUserID(ctx))
		if err != nil {
			return nil, err
		}
		switch member.RoleLevel {
		case constant.ServerOwner:
			return nil, errs.ErrNoPermission.Wrap("set server owner mute")
		case constant.ServerAdmin:
			if opMember.RoleLevel != constant.ServerOwner {
				return nil, errs.ErrNoPermission.Wrap("set server admin mute")
			}
		case constant.ServerOrdinaryUsers:
			if !(opMember.RoleLevel == constant.ServerAdmin || opMember.RoleLevel == constant.ServerOwner) {
				return nil, errs.ErrNoPermission.Wrap("set server ordinary users mute")
			}
		}
	}
	data := UpdateServerMemberMutedTimeMap(time.Now().Add(time.Second * time.Duration(req.MutedSeconds)))
	if err := c.ClubDatabase.UpdateServerMember(ctx, member.ServerID, member.UserID, data); err != nil {
		return nil, err
	}
	// c.Notification.ServerMemberMutedNotification(ctx, req.ServerID, req.UserID, req.MutedSeconds)
	return resp, nil
}

func (c *clubServer) CancelMuteServerMember(ctx context.Context, req *pbclub.CancelMuteServerMemberReq) (*pbclub.CancelMuteServerMemberResp, error) {
	resp := &pbclub.CancelMuteServerMemberResp{}
	//member, err := c.ClubDatabase.TakeServerMember(ctx, req.ServerID, req.UserID)
	//if err != nil {
	//	return nil, err
	//}
	//if !(mcontext.GetOpUserID(ctx) == req.UserID || tokenverify.IsAppManagerUid(ctx)) {
	//	opMember, err := c.ClubDatabase.TakeServerMember(ctx, req.ServerID, mcontext.GetOpUserID(ctx))
	//	if err != nil {
	//		return nil, err
	//	}
	//	if opMember.RoleLevel <= member.RoleLevel {
	//		return nil, errs.ErrNoPermission.Wrap(fmt.Sprintf("self RoleLevel %d target %d", opMember.RoleLevel, member.RoleLevel))
	//	}
	//}
	//if err := tokenverify.CheckAccessV3(ctx, req.UserID); err != nil {
	//	return nil, err
	//}
	member, err := c.TakeServerMember(ctx, req.ServerID, req.UserID)
	if err != nil {
		return nil, err
	}
	if !authverify.IsAppManagerUid(ctx) {
		opMember, err := c.TakeServerMember(ctx, req.ServerID, mcontext.GetOpUserID(ctx))
		if err != nil {
			return nil, err
		}
		switch member.RoleLevel {
		case constant.ServerOwner:
			return nil, errs.ErrNoPermission.Wrap("set server owner mute")
		case constant.ServerAdmin:
			if opMember.RoleLevel != constant.ServerOwner {
				return nil, errs.ErrNoPermission.Wrap("set server admin mute")
			}
		case constant.ServerOrdinaryUsers:
			if !(opMember.RoleLevel == constant.ServerAdmin || opMember.RoleLevel == constant.ServerOwner) {
				return nil, errs.ErrNoPermission.Wrap("set server ordinary users mute")
			}
		}
	}
	data := UpdateServerMemberMutedTimeMap(time.Unix(0, 0))
	if err := c.ClubDatabase.UpdateServerMember(ctx, member.ServerID, member.UserID, data); err != nil {
		return nil, err
	}
	// c.Notification.ServerMemberCancelMutedNotification(ctx, req.ServerID, req.UserID)
	return resp, nil
}

func (c *clubServer) SetServerMemberInfo(ctx context.Context, req *pbclub.SetServerMemberInfoReq) (*pbclub.SetServerMemberInfoResp, error) {
	resp := &pbclub.SetServerMemberInfoResp{}
	if len(req.Members) == 0 {
		return nil, errs.ErrArgs.Wrap("members empty")
	}
	for i := range req.Members {
		req.Members[i].FaceURL = nil
	}
	duplicateMap := make(map[[2]string]struct{})
	userIDMap := make(map[string]struct{})
	serverIDMap := make(map[string]struct{})
	for _, member := range req.Members {
		key := [...]string{member.ServerID, member.UserID}
		if _, ok := duplicateMap[key]; ok {
			return nil, errs.ErrArgs.Wrap("server user duplicate")
		}
		duplicateMap[key] = struct{}{}
		userIDMap[member.UserID] = struct{}{}
		serverIDMap[member.ServerID] = struct{}{}
	}
	serverIDs := utils.Keys(serverIDMap)
	userIDs := utils.Keys(userIDMap)
	members, err := c.FindServerMember(ctx, serverIDs, append(userIDs, mcontext.GetOpUserID(ctx)), nil)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		delete(duplicateMap, [...]string{member.ServerID, member.UserID})
	}
	if len(duplicateMap) > 0 {
		return nil, errs.ErrArgs.Wrap("user not found" + strings.Join(utils.Slice(utils.Keys(duplicateMap), func(e [2]string) string {
			return fmt.Sprintf("[server: %s user: %s]", e[0], e[1])
		}), ","))
	}
	memberMap := utils.SliceToMap(members, func(e *relationtb.ServerMemberModel) [2]string {
		return [...]string{e.ServerID, e.UserID}
	})
	if !authverify.IsAppManagerUid(ctx) {
		opUserID := mcontext.GetOpUserID(ctx)
		for _, member := range req.Members {
			if member.RoleLevel != nil {
				switch member.RoleLevel.Value {
				case constant.ServerOrdinaryUsers, constant.ServerAdmin:
				default:
					return nil, errs.ErrArgs.Wrap("invalid role level")
				}
			}
			opMember, ok := memberMap[[...]string{member.ServerID, opUserID}]
			if !ok {
				return nil, errs.ErrArgs.Wrap(fmt.Sprintf("user %s not in server %s", opUserID, member.ServerID))
			}
			if member.UserID == opUserID {
				if member.RoleLevel != nil {
					return nil, errs.ErrNoPermission.Wrap("can not change self role level")
				}
				continue
			}
			if opMember.RoleLevel == constant.ServerOrdinaryUsers {
				return nil, errs.ErrNoPermission.Wrap("ordinary users can not change other role level")
			}
			dbMember, ok := memberMap[[...]string{member.ServerID, member.UserID}]
			if !ok {
				return nil, errs.ErrRecordNotFound.Wrap(fmt.Sprintf("user %s not in server %s", member.UserID, member.ServerID))
			}
			//if opMember.RoleLevel == constant.ServerOwner {
			//	continue
			//}
			//if dbMember.RoleLevel == constant.ServerOwner {
			//	return nil, errs.ErrNoPermission.Wrap("change server owner")
			//}
			//if opMember.RoleLevel == constant.ServerAdmin && dbMember.RoleLevel == constant.ServerAdmin {
			//	return nil, errs.ErrNoPermission.Wrap("admin can not change other admin role info")
			//}
			switch opMember.RoleLevel {
			case constant.ServerOrdinaryUsers:
				return nil, errs.ErrNoPermission.Wrap("ordinary users can not change other role level")
			case constant.ServerAdmin:
				if dbMember.RoleLevel != constant.ServerOrdinaryUsers {
					return nil, errs.ErrNoPermission.Wrap("admin can not change other role level")
				}
				if member.RoleLevel != nil {
					return nil, errs.ErrNoPermission.Wrap("admin can not change other role level")
				}
			case constant.ServerOwner:
				//if member.RoleLevel != nil && member.RoleLevel.Value == constant.ServerOwner {
				//	return nil, errs.ErrNoPermission.Wrap("owner only one")
				//}
			}
		}
	}
	for _, member := range req.Members {
		if member.RoleLevel == nil {
			continue
		}
		if memberMap[[...]string{member.ServerID, member.UserID}].RoleLevel == constant.ServerOwner {
			return nil, errs.ErrArgs.Wrap(fmt.Sprintf("server %s user %s is owner", member.ServerID, member.UserID))
		}
	}
	// for i := 0; i < len(req.Members); i++ {
	// 	if err := CallbackBeforeSetServerMemberInfo(ctx, req.Members[i]); err != nil {
	// 		return nil, err
	// 	}
	// }
	if err = c.ClubDatabase.UpdateServerMembers(ctx, utils.Slice(req.Members, func(e *pbclub.SetServerMemberInfo) *relationtb.BatchUpdateGroupMember {
		return &relationtb.BatchUpdateGroupMember{
			GroupID: e.ServerID,
			UserID:  e.UserID,
			Map:     UpdateServerMemberMap(e),
		}
	})); err != nil {
		return nil, err
	}
	for _, member := range req.Members {
		if member.RoleLevel != nil {
			switch member.RoleLevel.Value {
			case constant.ServerAdmin:
				// c.Notification.ServerMemberSetToAdminNotification(ctx, member.ServerID, member.UserID)
			case constant.ServerOrdinaryUsers:
				// c.Notification.ServerMemberSetToOrdinaryUserNotification(ctx, member.ServerID, member.UserID)
			}
		}
		// if member.Nickname != nil || member.FaceURL != nil || member.Ex != nil {
		// 	log.ZDebug(ctx, "setServerMemberInfo notification", "member", member.UserID)
		// 	if err := c.Notification.ServerMemberInfoSetNotification(ctx, member.ServerID, member.UserID); err != nil {
		// 		log.ZError(ctx, "setServerMemberInfo notification failed", err, "member", member.UserID, "serverID", member.ServerID)
		// 	}
		// }
	}
	return resp, nil
}

func (c *clubServer) GetUserInServerMembers(ctx context.Context, req *pbclub.GetUserInServerMembersReq) (*pbclub.GetUserInServerMembersResp, error) {
	resp := &pbclub.GetUserInServerMembersResp{}
	if len(req.ServerIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("serverIDs empty")
	}
	members, err := c.FindServerMember(ctx, []string{req.UserID}, req.ServerIDs, nil)
	if err != nil {
		return nil, err
	}
	publicUserInfoMap, err := c.GetPublicUserInfoMap(ctx, utils.Filter(members, func(e *relationtb.ServerMemberModel) (string, bool) {
		return e.UserID, e.Nickname == "" || e.FaceURL == ""
	}), true)
	if err != nil {
		return nil, err
	}
	resp.Members = utils.Slice(members, func(e *relationtb.ServerMemberModel) *sdkws.ServerMemberFullInfo {
		if userInfo, ok := publicUserInfoMap[e.UserID]; ok {
			if e.Nickname == "" {
				e.Nickname = userInfo.Nickname
			}
			if e.FaceURL == "" {
				e.FaceURL = userInfo.FaceURL
			}
		}
		return convert.Db2PbServerMember(e)
	})
	return resp, nil
}

func (c *clubServer) GetServerMemberUserIDs(ctx context.Context, req *pbclub.GetServerMemberUserIDsReq) (resp *pbclub.GetServerMemberUserIDsResp, err error) {
	resp = &pbclub.GetServerMemberUserIDsResp{}
	resp.UserIDs, err = c.ClubDatabase.FindServerMemberUserID(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clubServer) GetServerMemberRoleLevel(ctx context.Context, req *pbclub.GetServerMemberRoleLevelReq) (*pbclub.GetServerMemberRoleLevelResp, error) {
	resp := &pbclub.GetServerMemberRoleLevelResp{}
	if len(req.RoleLevels) == 0 {
		return nil, errs.ErrArgs.Wrap("RoleLevels empty")
	}
	members, err := c.FindServerMember(ctx, []string{req.ServerID}, nil, req.RoleLevels)
	if err != nil {
		return nil, err
	}
	publicUserInfoMap, err := c.GetPublicUserInfoMap(ctx, utils.Filter(members, func(e *relationtb.ServerMemberModel) (string, bool) {
		return e.UserID, e.Nickname == "" || e.FaceURL == ""
	}), true)
	if err != nil {
		return nil, err
	}
	resp.Members = utils.Slice(members, func(e *relationtb.ServerMemberModel) *sdkws.ServerMemberFullInfo {
		if userInfo, ok := publicUserInfoMap[e.UserID]; ok {
			if e.Nickname == "" {
				e.Nickname = userInfo.Nickname
			}
			if e.FaceURL == "" {
				e.FaceURL = userInfo.FaceURL
			}
		}
		return convert.Db2PbServerMember(e)
	})
	return resp, nil
}

func (c *clubServer) deleteMemberAndSetConversationSeq(ctx context.Context, serverID string, userIDs []string) error {
	groups, err := c.ClubDatabase.FindGroup(ctx, []string{serverID})
	if err != nil {
		return err
	}
	for _, group := range groups {
		conevrsationID := msgprocessor.GetConversationIDBySessionType(constant.ServerGroupChatType, group.GroupID)
		maxSeq, err := c.msgRpcClient.GetConversationMaxSeq(ctx, conevrsationID)
		if err != nil {
			return err
		}
		c.conversationRpcClient.SetConversationMaxSeq(ctx, userIDs, conevrsationID, maxSeq)
	}
	return nil
}

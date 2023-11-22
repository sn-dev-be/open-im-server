package club

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
	pbuser "github.com/OpenIMSDK/protocol/user"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/mw/specialerror"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) CreateServer(ctx context.Context, req *pbclub.CreateServerReq) (*pbclub.CreateServerResp, error) {
	if req.OwnerUserID == "" {
		return nil, errs.ErrArgs.Wrap("no club owner")
	}
	if req.ServerName == "" {
		return nil, errs.ErrArgs.Wrap("no club name")
	}
	if req.Icon == "" {
		return nil, errs.ErrArgs.Wrap("no club icon")
	}

	//todo 后期加上创建部落数量限制

	if err := authverify.CheckAccessV3(ctx, req.OwnerUserID); err != nil {
		return nil, err
	}

	opUserID := mcontext.GetOpUserID(ctx)
	serverDB := &relation.ServerModel{
		ServerName:           req.ServerName,
		Icon:                 req.Icon,
		Description:          req.Description,
		ApplyMode:            constant.JoinServerNeedVerification, //开发测试阶段直接进，生产环境记得改成审核
		InviteMode:           constant.ServerInvitedDenied,
		Searchable:           constant.ServerSearchableDenied,
		Status:               constant.ServerOk,
		Banner:               req.Banner,
		UserMutualAccessible: req.UserMutualAccessible,
		OwnerUserID:          req.OwnerUserID,
		CreateTime:           time.Now(),
		Ex:                   req.Ex,
	}
	serverDB.OwnerUserID = opUserID
	//这几个配置是默认写死的，后期根据需求调整
	serverDB.CategoryNumber = 3
	serverDB.GroupNumber = 4
	serverDB.MemberNumber = 1

	if err := s.GenServerID(ctx, &serverDB.ServerID); err != nil {
		return nil, err
	}
	if err := s.ClubDatabase.CreateServer(ctx, []*relationtb.ServerModel{serverDB}); err != nil {
		return nil, err
	}

	//创建默认身份组
	go func() {
		s.CreateServerRoleForEveryone(ctx, serverDB.ServerID)
	}()

	roleID, err := s.CreateServerRoleForOwner(ctx, serverDB.ServerID)
	if err != nil {
		return nil, err
	}

	//部落主进部落
	err = s.createServerMember(ctx, serverDB.ServerID, opUserID, "", roleID, opUserID, "", constant.ServerOwner, 0)
	if err != nil {
		return nil, err
	}

	//创建默认分组与房间
	if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "", constant.DefaultCategoryType, 0); err == nil {
		createServerReq := s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "公告栏", opUserID)
		s.CreateServerGroup(ctx, createServerReq)
	}
	if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "文字房间", constant.SysCategoryType, 1); err == nil {
		createServerReq := s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "日常聊天", opUserID)
		s.CreateServerGroup(ctx, createServerReq)
		createServerReq = s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "资讯互动", opUserID)
		s.CreateServerGroup(ctx, createServerReq)
	}
	if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "部落管理", constant.SysCategoryType, 2); err == nil {
		createServerReq := s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "部落事务讨论", opUserID)
		s.CreateServerGroup(ctx, createServerReq)
	}

	return &pbclub.CreateServerResp{ServerID: serverDB.ServerID}, nil
}

// 获取所有热门部落
func (s *clubServer) GetServerRecommendedList(ctx context.Context, req *pbclub.GetServerRecommendedListReq) (*pbclub.GetServerRecommendecListResp, error) {
	resp := &pbclub.GetServerRecommendecListResp{}

	servers, err := s.ClubDatabase.GetServerRecommendedList(ctx)
	if err != nil {
		return nil, err
	}
	respServers := utils.Batch(convert.Db2PbServerFullInfo, servers)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, server := range respServers {
		wg.Add(1)
		go func(m *sdkws.ServerFullInfo) {
			defer wg.Done()
			s.genClubMembersAvatar(ctx, m)
		}(server)
	}
	wg.Wait()

	resp.Servers = respServers
	return resp, nil
}

func (s *clubServer) GetServersInfo(ctx context.Context, req *pbclub.GetServersInfoReq) (*pbclub.GetServersInfoResp, error) {
	resp := &pbclub.GetServersInfoResp{}
	respServerList := []*pbclub.GetServerInfoResp{}
	for _, serverID := range req.ServerIDs {
		respServer := &pbclub.GetServerInfoResp{}
		loginUserID := mcontext.GetOpUserID(ctx)
		isJoined := false

		if _, err := s.ClubDatabase.TakeServerMember(ctx, serverID, loginUserID); err == nil {
			isJoined = true
		}
		respServer.Joined = isJoined

		server, err := s.ClubDatabase.TakeServer(ctx, serverID)
		if err != nil {
			return nil, err
		}
		serverPb := convert.DB2PbServerInfo(server)
		if err != nil {
			return nil, err
		}
		respServer.Server = serverPb

		//查询分组与房间信息
		categories, _ := s.ClubDatabase.GetAllGroupCategoriesByServer(ctx, server.ServerID)
		if len(categories) > 0 {
			serverGroups, err := s.ClubDatabase.FindGroup(ctx, []string{server.ServerID})
			if err == nil {
				for _, category := range categories {
					temp := []*sdkws.ServerGroupListInfo{}

					for _, server := range serverGroups {
						if category.CategoryID == server.GroupCategoryID {
							pbGroupInfo := convert.Db2PbServerGroupInfo(server)
							if server.GroupMode == constant.AppGroupMode {
								if serverDapp, err := s.ClubDatabase.TakeGroupDapp(ctx, server.GroupID); err != nil {
									return nil, err
								} else {
									pbGroupDapp := convert.Db2PbGroupDapp(serverDapp)
									pbGroupInfo.Dapp = pbGroupDapp
								}

							}
							temp = append(temp, pbGroupInfo)
						}
					}
					groupCategory := convert.Db2PbGroupCategory(category)
					list := sdkws.GroupCategoryListInfo{}
					list.CategoryInfo = groupCategory
					list.GroupList = temp
					respServer.CategoryList = append(respServer.CategoryList, &list)
				}
			}
		}
		respServerList = append(respServerList, respServer)
	}
	resp.Servers = respServerList
	return resp, nil
}

func (s *clubServer) DismissServer(ctx context.Context, req *pbclub.DismissServerReq) (*pbclub.DismissServerResp, error) {
	defer log.ZInfo(ctx, "DismissServer.return")
	resp := &pbclub.DismissServerResp{}
	owner, err := s.ClubDatabase.TakeServerOwner(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	if !authverify.IsAppManagerUid(ctx) {
		if owner.UserID != mcontext.GetOpUserID(ctx) {
			return nil, errs.ErrNoPermission.Wrap("not group owner")
		}
	}
	if err := s.ClubDatabase.DismissServer(ctx, req.ServerID); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *clubServer) IsNotFound(err error) bool {
	return errs.ErrRecordNotFound.Is(specialerror.ErrCode(errs.Unwrap(err)))
}

func (s *clubServer) GenServerID(ctx context.Context, serverID *string) error {
	if *serverID != "" {
		_, err := s.ClubDatabase.TakeServer(ctx, *serverID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("server id existed " + *serverID)
		} else if s.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}
	for i := 0; i < 10; i++ {
		id := utils.Md5(strings.Join([]string{mcontext.GetOperationID(ctx), strconv.FormatInt(time.Now().UnixNano(), 10), strconv.Itoa(rand.Int())}, ",;,"))
		bi := big.NewInt(0)
		bi.SetString(id[0:8], 16)
		id = bi.String()
		_, err := s.ClubDatabase.TakeServer(ctx, id)
		if err == nil {
			continue
		} else if s.IsNotFound(err) {
			*serverID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("server id gen error")
}

func (s *clubServer) genClubMembersAvatar(ctx context.Context, server *sdkws.ServerFullInfo) error {
	_, members, err := s.ClubDatabase.PageGetServerMember(ctx, server.ServerInfo.ServerID, 1, 3)
	if err == nil {
		userIDs := []string{}
		for _, member := range members {
			userIDs = append(userIDs, member.UserID)
		}
		getDesignateUsersReq := &pbuser.GetDesignateUsersReq{
			UserIDs: userIDs,
		}
		getDesignateUsersResp, err := s.User.Client.GetDesignateUsers(ctx, getDesignateUsersReq)
		if err != nil {
			return err
		}

		userAvatarList := []string{}
		for _, user := range getDesignateUsersResp.UsersInfo {
			userAvatarList = append(userAvatarList, user.FaceURL)
		}
		server.MemberAvatarList = userAvatarList
	}
	return nil
}

func (s *clubServer) genCreateServerGroupReq(serverID, categoryID, groupName, ownerUserID string) *pbclub.CreateServerGroupReq {
	req := &pbclub.CreateServerGroupReq{}

	groupInfo := &sdkws.GroupInfo{
		GroupName:       groupName,
		OwnerUserID:     ownerUserID,
		Status:          constant.GroupOk,
		CreatorUserID:   ownerUserID,
		GroupType:       constant.ServerGroup,
		ConditionType:   1,
		Condition:       "",
		GroupCategoryID: categoryID,
		ServerID:        serverID,
	}
	req.GroupInfo = groupInfo
	return req
}

func (s *clubServer) SetServerInfo(ctx context.Context, req *pbclub.SetServerInfoReq) (*pbclub.SetServerInfoResp, error) {
	var opMember *relationtb.ServerMemberModel
	if !authverify.IsAppManagerUid(ctx) {
		var err error
		opMember, err = s.TakeServerMember(ctx, req.ServerInfoForSet.ServerID, mcontext.GetOpUserID(ctx))
		if err != nil {
			return nil, err
		}
		if !(opMember.RoleLevel == constant.ServerOwner || opMember.RoleLevel == constant.ServerAdmin) {
			return nil, errs.ErrNoPermission.Wrap("no server owner or admin")
		}
	}
	server, err := s.ClubDatabase.TakeServer(ctx, req.ServerInfoForSet.ServerID)
	if err != nil {
		return nil, err
	}
	if server.Status == constant.ServerStatusDismissed {
		return nil, utils.Wrap(errs.ErrDismissedAlready, "")
	}
	resp := &pbclub.SetServerInfoResp{}
	// count, err := s.ClubDatabase.FindServerMemberNum(ctx, server.ServerID)
	// if err != nil {
	// 	return nil, err
	// }
	// owner, err := s.TakeServerOwner(ctx, server.ServerID)
	// if err != nil {
	// 	return nil, err
	// }
	data := UpdateServerInfoMap(ctx, req.ServerInfoForSet)
	if len(data) == 0 {
		return resp, nil
	}
	if err := s.ClubDatabase.UpdateServer(ctx, server.ServerID, data); err != nil {
		return nil, err
	}
	server, err = s.ClubDatabase.TakeServer(ctx, req.ServerInfoForSet.ServerID)
	if err != nil {
		return nil, err
	}
	// tips := &sdkws.ServerInfoSetTips{
	// 	Server:   s.serverDB2PB(server, owner.UserID, count),
	// 	MuteTime: 0,
	// 	OpUser:   &sdkws.ServerMemberFullInfo{},
	// }
	// if opMember != nil {
	// 	tips.OpUser = s.serverMemberDB2PB(opMember, 0)
	// }
	// switch len(data) - num {
	// case 0:
	// case 1:
	// 	if req.ServerInfoForSet.ServerName == "" {
	// 		s.Notification.ServerInfoSetNotification(ctx, tips)
	// 	} else {
	// 		s.Notification.ServerInfoSetNameNotification(ctx, &sdkws.ServerInfoSetNameTips{Server: tips.Server, OpUser: tips.OpUser})
	// 	}
	// default:
	// 	s.Notification.ServerInfoSetNotification(ctx, tips)
	// }
	return resp, nil
}

func (s *clubServer) SearchServer(ctx context.Context, req *pbclub.SearchServerReq) (*pbclub.SearchServerResp, error) {
	resp := &pbclub.SearchServerResp{}
	total, servers, err := s.ClubDatabase.SearchServer(ctx, req.Keyword, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	resp.Total = total
	resp.ServerInfos = utils.Batch(convert.DB2PbServerInfo, servers)
	return resp, nil
}

func (s *clubServer) GetServerAbstractInfo(ctx context.Context, req *pbclub.GetServerAbstractInfoReq) (*pbclub.GetServerAbstractInfoResp, error) {
	resp := &pbclub.GetServerAbstractInfoResp{}
	if len(req.ServerIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("serverIDs empty")
	}
	if utils.Duplicate(req.ServerIDs) {
		return nil, errs.ErrArgs.Wrap("serverIDs duplicate")
	}
	servers, err := s.ClubDatabase.FindServer(ctx, req.ServerIDs)
	if err != nil {
		return nil, err
	}
	if ids := utils.Single(req.ServerIDs, utils.Slice(servers, func(server *relationtb.ServerModel) string {
		return server.ServerID
	})); len(ids) > 0 {
		return nil, errs.ErrGroupIDNotFound.Wrap("not found server " + strings.Join(ids, ","))
	}
	serverUserMap, err := s.ClubDatabase.MapServerMemberUserID(ctx, req.ServerIDs)
	if err != nil {
		return nil, err
	}
	if ids := utils.Single(req.ServerIDs, utils.Keys(serverUserMap)); len(ids) > 0 {
		return nil, errs.ErrGroupIDNotFound.Wrap(fmt.Sprintf("server %s not found member", strings.Join(ids, ",")))
	}
	resp.ServerAbstractInfos = utils.Slice(servers, func(server *relationtb.ServerModel) *pbclub.ServerAbstractInfo {
		users := serverUserMap[server.ServerID]
		return convert.Db2PbServerAbstractInfo(server.ServerID, users.MemberNum, users.Hash)
	})
	return resp, nil
}

func (s *clubServer) MuteServer(ctx context.Context, req *pbclub.MuteServerReq) (*pbclub.MuteServerResp, error) {
	resp := &pbclub.MuteServerResp{}
	if !s.checkManageServer(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}
	if err := s.ClubDatabase.UpdateServer(ctx, req.ServerID, UpdateServerStatusMap(constant.ServerStatusMuted)); err != nil {
		return nil, err
	}
	// s.Notification.ServerMutedNotification(ctx, req.ServerID)
	return resp, nil
}

func (s *clubServer) CancelMuteServer(ctx context.Context, req *pbclub.CancelMuteServerReq) (*pbclub.CancelMuteServerResp, error) {
	resp := &pbclub.CancelMuteServerResp{}
	if !s.checkManageServer(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}
	if err := s.ClubDatabase.UpdateServer(ctx, req.ServerID, UpdateServerStatusMap(constant.ServerOk)); err != nil {
		return nil, err
	}
	// s.Notification.ServerCancelMutedNotification(ctx, req.ServerID)
	return resp, nil
}

package club

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	pbgroup "github.com/OpenIMSDK/protocol/group"
	"github.com/OpenIMSDK/protocol/sdkws"
	pbuser "github.com/OpenIMSDK/protocol/user"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/mw/specialerror"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
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
	serverDB := convert.Pb2DBServerInfo(req)
	serverDB.OwnerUserID = opUserID
	//这几个配置是默认写死的，后期根据需求调整
	serverDB.CategoryNumber = 3
	serverDB.ChannelNumber = 4
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
	err = s.createServerMember(ctx, serverDB.ServerID, opUserID, "", roleID, opUserID, "", 0, 0)
	if err != nil {
		return nil, err
	}

	//创建默认分组与房间
	if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "", constant.DefaultCategoryType, 0); err == nil {
		createServerReq := s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "公告栏", opUserID)
		s.Group.Client.CreateServerGroup(ctx, createServerReq)
	}
	if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "文字房间", constant.SysCategoryType, 1); err == nil {
		createServerReq := s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "日常聊天", opUserID)
		s.Group.Client.CreateServerGroup(ctx, createServerReq)
		createServerReq = s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "资讯互动", opUserID)
		s.Group.Client.CreateServerGroup(ctx, createServerReq)
	}
	if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "部落管理", constant.SysCategoryType, 2); err == nil {
		createServerReq := s.genCreateServerGroupReq(serverDB.ServerID, categoryID, "部落事务讨论", opUserID)
		s.Group.Client.CreateServerGroup(ctx, createServerReq)

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
	resp_servers, err := convert.DB2PbServerFullInfoList(servers)
	if err != nil {
		return nil, err
	}

	//get member avatar limit 3
	var wg sync.WaitGroup
	for _, server := range resp_servers {
		wg.Add(1) // 增加 WaitGroup 计数
		go func(m *sdkws.ServerFullInfo) {
			defer wg.Done()
			s.genClubMembersAvatar(ctx, m)
		}(server)
	}
	// 等待所有协程完成
	wg.Wait()

	resp.Servers = resp_servers
	return resp, nil
}

func (s *clubServer) GetServerDetails(ctx context.Context, req *pbclub.GetServerDetailsReq) (*pbclub.GetServerDetailsResp, error) {
	resp := &pbclub.GetServerDetailsResp{}
	loginUserID := mcontext.GetOpUserID(ctx)
	isJoined := false

	if _, err := s.ClubDatabase.GetServerMemberByUserID(ctx, req.ServerID, loginUserID); err == nil {
		isJoined = true
	}
	resp.Joined = isJoined

	server, err := s.ClubDatabase.TakeServer(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	resp_server, err := convert.DB2PbServerInfo(server)
	if err != nil {
		return nil, err
	}
	resp.Server = resp_server

	//查询分组与房间信息
	categories, _ := s.ClubDatabase.GetAllGroupCategoriesByServer(ctx, server.ServerID)
	if len(categories) > 0 {
		serverGroups, err := s.ClubDatabase.FindGroup(ctx, []string{server.ServerID})
		if err == nil {
			for _, category := range categories {
				temp := []*sdkws.ServerGroupListInfo{}

				for _, group := range serverGroups {
					if category.CategoryID == group.GroupCategoryID {
						temp = append(temp, convert.Db2PbServerGroupInfo(group))
					}
				}
				resp_category, _ := convert.DB2PbCategory(category, temp)
				resp.CategoryList = append(resp.CategoryList, resp_category)
			}
		}
	}
	return resp, nil
}

func (s *clubServer) BatchDeleteServers(ctx context.Context, req *pbclub.DeleteServerReq) (*pbclub.DeleteServerResp, error) {
	return nil, nil
}

func (s *clubServer) GetJoinedServerList(ctx context.Context, req *pbclub.GetJoinedServerListReq) (*pbclub.GetJoinedServerListResp, error) {
	return nil, nil
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
	members, _, err := s.ClubDatabase.PageServerMembers(ctx, 1, 3, server.ServerID)
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

func (s *clubServer) genCreateServerGroupReq(serverID, categoryID, groupName, ownerUserID string) *pbgroup.CreateServerGroupReq {
	req := &pbgroup.CreateServerGroupReq{
		OwnerUserID: ownerUserID,
	}

	groupInfo := &sdkws.GroupInfo{
		GroupName:       groupName,
		OwnerUserID:     ownerUserID,
		Status:          constant.GroupOk,
		CreatorUserID:   ownerUserID,
		GroupType:       constant.ServerGroup,
		ConditionType:   0,
		Condition:       "",
		GroupCategoryID: categoryID,
		ServerID:        serverID,
	}
	req.GroupInfo = groupInfo
	return req
}

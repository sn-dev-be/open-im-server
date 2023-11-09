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
	//if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "", constant.DefaultCategoryType, 0); err == nil {
	//todo 创建分组

	// if channelID, err := s.CreateChannelByDefault(ctx, serverDB.ServerID, categoryID, "公告栏", opUserID, constant.ChatChannelType, 0); err == nil {
	// 	channelIDs = append(channelIDs, channelID)
	// }
	//}
	//i//f categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "文字房间", constant.SysCategoryType, 1); err == nil {
	// if channelID, err := s.CreateChannelByDefault(ctx, serverDB.ServerID, categoryID, "日常聊天", opUserID, constant.ChatChannelType, 0); err == nil {
	// 	channelIDs = append(channelIDs, channelID)
	// }
	// if channelID, err := s.CreateChannelByDefault(ctx, serverDB.ServerID, categoryID, "资讯互动", opUserID, constant.ChatChannelType, 1); err == nil {
	// 	channelIDs = append(channelIDs, channelID)
	// }
	//}
	//if categoryID, err := s.createGroupCategoryByDefault(ctx, serverDB.ServerID, "部落管理", constant.SysCategoryType, 2); err == nil {
	// if channelID, err := s.CreateChannelByDefault(ctx, serverDB.ServerID, categoryID, "部落事务讨论", opUserID, constant.ChatChannelType, 0); err == nil {
	// 	channelIDs = append(channelIDs, channelID)
	// }
	//}

	// //todo 部落主统一进入所有房间
	// if err := s.CreateChannelMembser(ctx, serverDB.ServerID, channelIDs, serverMemberId); err != nil {
	// 	log.ZDebug(ctx, "owner join channel failed", "user_id", opUserID, "channelIDs", channelIDs)
	// }
	return &pbclub.CreateServerResp{}, nil
}

func (s *clubServer) GetServerList(ctx context.Context, req *pbclub.GetServerListReq) (*pbclub.GetServerListResp, error) {
	resp := &pbclub.GetServerListResp{}

	servers, total, err := s.ClubDatabase.PageServers(ctx, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	resp_servers, err := convert.DB2PbServerInfo(servers)
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
	resp.Total = int32(total)
	return resp, nil
}

func (s *clubServer) GetServerDetails(ctx context.Context, req *pbclub.GetServerDetailsReq) (*pbclub.GetServerDetailsResp, error) {
	return nil, nil
}

func (s *clubServer) BatchDeleteServers(ctx context.Context, req *pbclub.DeleteServerReq) (*pbclub.DeleteServerResp, error) {
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

		userAvatarList := make([]string, len(getDesignateUsersResp.UsersInfo))
		for _, user := range getDesignateUsersResp.UsersInfo {
			userAvatarList = append(userAvatarList, user.FaceURL)
		}
		server.MemberAvatarList = userAvatarList
	}
	return nil
}

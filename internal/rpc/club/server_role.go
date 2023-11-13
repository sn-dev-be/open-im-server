package club

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/OpenIMSDK/protocol/constant"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) GenServerRoleID(ctx context.Context, serverRoleID *string) error {
	if *serverRoleID != "" {
		_, err := s.ClubDatabase.TakeServerRole(ctx, *serverRoleID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("serverRole id existed " + *serverRoleID)
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
		_, err := s.ClubDatabase.TakeServerRole(ctx, id)
		if err == nil {
			continue
		} else if s.IsNotFound(err) {
			*serverRoleID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("server_role id gen error")
}

func (s *clubServer) CreateServerRole(ctx context.Context, server_roles []*relationtb.ServerRoleModel) error {
	if err := s.ClubDatabase.CreateServerRole(ctx, server_roles); err != nil {
		return err
	}
	return nil
}

// 创建两个默认身份组
func (s *clubServer) CreateServerRoleForEveryone(ctx context.Context, serverID string) error {
	//全体成员
	everyoneAuth := &pbclub.RoleAuth{
		ManageServer:        constant.ServerRoleAuthDenied,
		ShareServer:         constant.ServerRoleAuthAllowed,
		ManageMember:        constant.ServerRoleAuthDenied,
		SendMsg:             constant.ServerRoleAuthAllowed,
		ManageMsg:           constant.ServerRoleAuthDenied,
		ManageCommunity:     constant.ServerRoleAuthDenied,
		PostTweet:           constant.ServerRoleAuthAllowed,
		TweetReply:          constant.ServerRoleAuthAllowed,
		ManageGroupCategory: constant.ServerRoleAuthDenied,
		ManageGroup:         constant.ServerRoleAuthDenied,
	}
	everyone := &relationtb.ServerRoleModel{
		RoleName:     "全体成员",
		Icon:         "",
		Type:         constant.ServerRoleTypeEveryOne,
		Priority:     0,
		ServerID:     serverID,
		RoleAuth:     utils.StructToJsonString(everyoneAuth),
		ColorLevel:   0,
		MemberNumber: 1,
		Ex:           "",
		CreateTime:   time.Now(),
	}
	if err := s.GenServerRoleID(ctx, &everyone.RoleID); err != nil {
		return err
	}
	if err := s.ClubDatabase.CreateServerRole(ctx, []*relationtb.ServerRoleModel{everyone}); err != nil {
		return err
	}

	return nil
}

func (s *clubServer) CreateServerRoleForOwner(ctx context.Context, serverID string) (string, error) {
	//部落主
	ownerAuth := &pbclub.RoleAuth{
		ManageServer:        constant.ServerRoleAuthAllowed,
		ShareServer:         constant.ServerRoleAuthAllowed,
		ManageMember:        constant.ServerRoleAuthAllowed,
		SendMsg:             constant.ServerRoleAuthAllowed,
		ManageMsg:           constant.ServerRoleAuthAllowed,
		ManageCommunity:     constant.ServerRoleAuthAllowed,
		PostTweet:           constant.ServerRoleAuthAllowed,
		TweetReply:          constant.ServerRoleAuthAllowed,
		ManageGroupCategory: constant.ServerRoleAuthAllowed,
		ManageGroup:         constant.ServerRoleAuthAllowed,
	}
	owner := &relationtb.ServerRoleModel{
		RoleName:     "部落主",
		Icon:         "",
		Type:         constant.ServerRoleTypeOwner,
		Priority:     0,
		ServerID:     serverID,
		RoleAuth:     utils.StructToJsonString(ownerAuth),
		ColorLevel:   0,
		MemberNumber: 1,
		Ex:           "",
		CreateTime:   time.Now(),
	}
	if err := s.GenServerRoleID(ctx, &owner.RoleID); err != nil {
		return "", err
	}
	if err := s.ClubDatabase.CreateServerRole(ctx, []*relationtb.ServerRoleModel{owner}); err != nil {
		return "", err
	}

	return owner.RoleID, nil
}

func (s *clubServer) CreateServerRoleByPb(ctx context.Context, server_roles []*relationtb.ServerRoleModel) error {
	if err := s.ClubDatabase.CreateServerRole(ctx, server_roles); err != nil {
		return err
	}
	return nil
}

func (s *clubServer) getServerRoleByType(ctx context.Context, serverID string, roleType int32) (*relationtb.ServerRoleModel, error) {
	return s.ClubDatabase.TakeServerRoleByType(ctx, serverID, roleType)
}

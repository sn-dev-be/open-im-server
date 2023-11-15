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
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	"github.com/openimsdk/open-im-server/v3/pkg/permissions"
)

func (c *clubServer) GenServerRoleID(ctx context.Context, serverRoleID *string) error {
	if *serverRoleID != "" {
		_, err := c.ClubDatabase.TakeServerRole(ctx, *serverRoleID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("serverRole id existed " + *serverRoleID)
		} else if c.IsNotFound(err) {
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
		_, err := c.ClubDatabase.TakeServerRole(ctx, id)
		if err == nil {
			continue
		} else if c.IsNotFound(err) {
			*serverRoleID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("server_role id gen error")
}

func (c *clubServer) CreateServerRoleForEveryone(ctx context.Context, serverID string) error {
	permissions, _ := permissions.NewDefaultEveryonePermissions().ToJSON()
	everyone := &relationtb.ServerRoleModel{
		RoleName:     "全体成员",
		Icon:         "",
		Type:         constant.ServerRoleTypeEveryOne,
		Priority:     constant.ServerOrdinaryUsers,
		ServerID:     serverID,
		RoleAuth:     permissions,
		ColorLevel:   0,
		MemberNumber: 1,
		Ex:           "",
		CreateTime:   time.Now(),
	}
	if err := c.GenServerRoleID(ctx, &everyone.RoleID); err != nil {
		return err
	}
	if err := c.ClubDatabase.CreateServerRole(ctx, []*relationtb.ServerRoleModel{everyone}); err != nil {
		return err
	}
	return nil
}

func (c *clubServer) CreateServerRoleForOwner(ctx context.Context, serverID string) (string, error) {
	permissions, _ := permissions.NewDefaultAdminPermissions().ToJSON()
	owner := &relationtb.ServerRoleModel{
		RoleName:     "部落主",
		Icon:         "",
		Type:         constant.ServerRoleTypeOwner,
		Priority:     constant.ServerOwner,
		ServerID:     serverID,
		RoleAuth:     permissions,
		ColorLevel:   0,
		MemberNumber: 1,
		Ex:           "",
		CreateTime:   time.Now(),
	}
	if err := c.GenServerRoleID(ctx, &owner.RoleID); err != nil {
		return "", err
	}
	if err := c.ClubDatabase.CreateServerRole(ctx, []*relationtb.ServerRoleModel{owner}); err != nil {
		return "", err
	}

	return owner.RoleID, nil
}

func (c *clubServer) getServerRoleByType(ctx context.Context, serverID string, roleType int32) (*relationtb.ServerRoleModel, error) {
	return c.ClubDatabase.TakeServerRoleByType(ctx, serverID, roleType)
}

func (s *clubServer) TransferServerOwner(ctx context.Context, req *pbclub.TransferServerOwnerReq) (*pbclub.TransferServerOwnerResp, error) {
	resp := &pbclub.TransferServerOwnerResp{}
	_, err := s.ClubDatabase.TakeServer(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	if req.OldOwnerUserID == req.NewOwnerUserID {
		return nil, errs.ErrArgs.Wrap("OldOwnerUserID == NewOwnerUserID")
	}
	members, err := s.ClubDatabase.FindServerMember(ctx, []string{req.ServerID}, []string{req.OldOwnerUserID, req.NewOwnerUserID}, nil)
	if err != nil {
		return nil, err
	}
	memberMap := utils.SliceToMap(members, func(e *relationtb.ServerMemberModel) string { return e.UserID })
	oldOwner := memberMap[req.OldOwnerUserID]
	if oldOwner == nil {
		return nil, errs.ErrArgs.Wrap("OldOwnerUserID not in group " + req.NewOwnerUserID)
	}
	newOwner := memberMap[req.NewOwnerUserID]
	if newOwner == nil {
		return nil, errs.ErrArgs.Wrap("NewOwnerUser not in group " + req.NewOwnerUserID)
	}
	if !authverify.IsAppManagerUid(ctx) {
		if !(mcontext.GetOpUserID(ctx) == oldOwner.UserID && oldOwner.RoleLevel == constant.ServerOwner) {
			return nil, errs.ErrNoPermission.Wrap("no permission transfer group owner")
		}
	}
	if err := s.ClubDatabase.TransferServerOwner(ctx, req.ServerID, req.OldOwnerUserID, req.NewOwnerUserID, constant.ServerOrdinaryUsers); err != nil {
		return nil, err
	}
	//s.Notification.GroupOwnerTransferredNotification(ctx, req)
	return resp, nil
}

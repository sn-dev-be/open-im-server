package club

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/OpenIMSDK/protocol/constant"
	"gorm.io/datatypes"

	pbclub "github.com/OpenIMSDK/protocol/club"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
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
		Permissions:  datatypes.JSON(permissions),
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
		Permissions:  datatypes.JSON(permissions),
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

func (c *clubServer) getServerRoleByPriority(ctx context.Context, serverID string, priority int32) (*relationtb.ServerRoleModel, error) {
	return c.ClubDatabase.TakeServerRoleByPriority(ctx, serverID, priority)
}

func (c *clubServer) TransferServerOwner(ctx context.Context, req *pbclub.TransferServerOwnerReq) (*pbclub.TransferServerOwnerResp, error) {
	resp := &pbclub.TransferServerOwnerResp{}
	server, err := c.ClubDatabase.TakeServer(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	if req.OldOwnerUserID == req.NewOwnerUserID {
		return nil, errs.ErrArgs.Wrap("OldOwnerUserID == NewOwnerUserID")
	}
	members, err := c.ClubDatabase.FindServerMember(ctx, []string{req.ServerID}, []string{req.OldOwnerUserID, req.NewOwnerUserID}, nil)
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
		if !(mcontext.GetOpUserID(ctx) == oldOwner.UserID && oldOwner.UserID == server.OwnerUserID) {
			return nil, errs.ErrNoPermission.Wrap("no permission transfer group owner")
		}
	}
	if err := c.ClubDatabase.TransferServerOwner(ctx, req.ServerID, oldOwner, newOwner, constant.ServerOrdinaryUsers); err != nil {
		return nil, err
	}
	//s.Notification.GroupOwnerTransferredNotification(ctx, req)
	return resp, nil
}

func (c *clubServer) GetServerRoleList(ctx context.Context, req *pbclub.GetServerRoleListReq) (*pbclub.GetServerRoleListResp, error) {
	resp := &pbclub.GetServerRoleListResp{}
	total, roles, err := c.ClubDatabase.PageGetServerRole(ctx, req.ServerID, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	log.ZDebug(ctx, "GetServerRoleList", "total", total, "roles", roles, "length", len(roles))
	if err != nil {
		return nil, err
	}
	resp.Total = total
	resp.Roles = utils.Batch(convert.Db2PbServerRole, roles)
	log.ZDebug(ctx, "GetServerRoleList", "resp", resp, "length", len(resp.Roles))

	return resp, nil
}

func (c *clubServer) GetServerRolesInfo(ctx context.Context, req *pbclub.GetServerRolesInfoReq) (*pbclub.GetServerRolesInfoResp, error) {
	resp := &pbclub.GetServerRolesInfoResp{}
	roles, err := c.ClubDatabase.FindServerRole(ctx, req.RoleIDs)
	if err != nil {
		return nil, err
	}
	resp.Roles = utils.Batch(convert.Db2PbServerRole, roles)
	return resp, nil
}

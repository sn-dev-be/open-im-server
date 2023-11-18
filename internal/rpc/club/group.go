package club

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (c *clubServer) GetJoinedServerGroupList(ctx context.Context, req *pbclub.GetJoinedServerGroupListReq) (*pbclub.GetJoinedServerGroupListResp, error) {
	resp := &pbclub.GetJoinedServerGroupListResp{}
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
	groups, err := c.ClubDatabase.FindGroup(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	groupIDs := utils.Slice(groups, func(e *relationtb.GroupModel) string {
		return e.GroupID
	})

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

	resp.Groups = utils.Slice(utils.Order(groupIDs, groups, func(group *relationtb.GroupModel) string {
		return group.GroupID
	}), func(group *relationtb.GroupModel) *sdkws.GroupInfo {
		var userID string
		if user := ownerMap[group.ServerID]; user != nil {
			userID = user.UserID
		}
		return convert.Db2PbGroupInfo(group, userID, serverMemberNum[group.ServerID])
	})
	return resp, nil
}

// SetGroupCategoryOrder implements club.ClubServer.
func (*clubServer) SetGroupCategoryOrder(context.Context, *pbclub.SetGroupCategoryOrderReq) (*pbclub.SetGroupCategoryOrderResp, error) {
	panic("unimplemented")
}

// SetServerGroupInfo implements club.ClubServer.
func (*clubServer) SetServerGroupInfo(context.Context, *pbclub.SetServerGroupInfoReq) (*pbclub.SetServerGroupInfoResp, error) {
	panic("unimplemented")
}

// SetServerGroupOrder implements club.ClubServer.
func (*clubServer) SetServerGroupOrder(context.Context, *pbclub.SetServerGroupOrderReq) (*pbclub.SetServerGroupOrderResp, error) {
	panic("unimplemented")
}

func (c *clubServer) CreateServerGroup(ctx context.Context, req *pbclub.CreateServerGroupReq) (*pbclub.CreateServerGroupResp, error) {
	if req.GroupInfo.OwnerUserID == "" {
		return nil, errs.ErrArgs.Wrap("no group owner")
	}
	if req.GroupInfo.GroupType != constant.ServerGroup {
		return nil, errs.ErrArgs.Wrap(fmt.Sprintf("group type %d not support", req.GroupInfo.GroupType))
	}
	if err := authverify.CheckAccessV3(ctx, req.GroupInfo.OwnerUserID); err != nil {
		return nil, err
	}
	opUserID := mcontext.GetOpUserID(ctx)
	group := convert.Pb2DBGroupInfo(req.GroupInfo)
	group.CreatorUserID = opUserID
	if err := c.GenGroupID(ctx, &group.GroupID); err != nil {
		return nil, err
	}
	if group.GroupMode == constant.AppGroupMode {
		if req.DappID == "" {
			return nil, errs.ErrArgs.Wrap("no group dapp bind")
		}
		groupDapp := &relationtb.GroupDappModel{
			GroupID:    group.GroupID,
			DappID:     req.DappID,
			CreateTime: time.Now(),
		}
		if err := c.ClubDatabase.CreateServerGroup(ctx, []*relationtb.GroupModel{group}, []*relationtb.GroupDappModel{groupDapp}); err != nil {
			return nil, err
		}
	} else {
		if err := c.ClubDatabase.CreateServerGroup(ctx, []*relationtb.GroupModel{group}, nil); err != nil {
			return nil, err
		}
	}

	resp := &pbclub.CreateServerGroupResp{GroupInfo: &sdkws.GroupInfo{}}
	resp.GroupInfo = convert.Db2PbGroupInfo(group, req.GroupInfo.OwnerUserID, 0)
	resp.GroupInfo.MemberCount = 0
	return resp, nil
}

func (c *clubServer) GenGroupID(ctx context.Context, groupID *string) error {
	if *groupID != "" {
		_, err := c.ClubDatabase.TakeGroup(ctx, *groupID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("group id existed " + *groupID)
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
		_, err := c.ClubDatabase.TakeGroup(ctx, id)
		if err == nil {
			continue
		} else if c.IsNotFound(err) {
			*groupID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("group id gen error")
}

func (c *clubServer) MuteServerGroup(ctx context.Context, req *pbclub.MuteServerGroupReq) (*pbclub.MuteServerGroupResp, error) {
	resp := &pbclub.MuteServerGroupResp{}
	if !c.checkManageServer(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}
	if err := c.GroupDatabase.UpdateGroup(ctx, req.GroupID, UpdateGroupStatusMap(constant.GroupStatusMuted)); err != nil {
		return nil, err
	}
	// c.Notification.GroupMutedNotification(ctx, req.GroupID)
	return resp, nil
}

func (c *clubServer) CancelMuteServerGroup(ctx context.Context, req *pbclub.CancelMuteServerGroupReq) (*pbclub.CancelMuteServerGroupResp, error) {
	resp := &pbclub.CancelMuteServerGroupResp{}
	if !c.checkManageServer(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}
	if err := c.GroupDatabase.UpdateGroup(ctx, req.GroupID, UpdateGroupStatusMap(constant.GroupOk)); err != nil {
		return nil, err
	}
	// c.Notification.GroupCancelMutedNotification(ctx, req.GroupID)
	return resp, nil
}

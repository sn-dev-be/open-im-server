package club

import (
	"context"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"

	"github.com/OpenIMSDK/tools/log"
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

func (c *clubServer) GetServerGroups(ctx context.Context, req *pbclub.GetServerGroupsReq) (*pbclub.GetServerGroupsResp, error) {
	defer log.ZDebug(ctx, "return")
	resp := &pbclub.GetServerGroupsResp{}
	groups, err := c.ClubDatabase.FindGroup(ctx, []string{req.ServerID})
	if err != nil {
		return nil, err
	}
	//convert
	serverGroups := []*sdkws.ServerGroupListInfo{}
	for _, group := range groups {
		serverGroups = append(serverGroups, convert.Db2PbServerGroupInfo(group))
	}

	resp.Groups = serverGroups
	return resp, nil
}

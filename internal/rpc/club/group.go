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

func (c *clubServer) SetServerGroupInfo(ctx context.Context, req *pbclub.SetServerGroupInfoReq) (*pbclub.SetServerGroupInfoResp, error) {

	if !c.checkManageGroup(ctx, req.GroupInfo.ServerID) {
		return nil, errs.ErrNoPermission
	}

	//var opMember *relationtb.ServerMemberModel
	if !authverify.IsAppManagerUid(ctx) {
		var err error
		_, err = c.TakeServerMember(ctx, req.GroupInfo.ServerID, mcontext.GetOpUserID(ctx))
		if err != nil {
			return nil, err
		}
	}
	group, err := c.ClubDatabase.TakeGroup(ctx, req.GroupInfo.GroupID)
	if err != nil {
		return nil, err
	}
	if group.Status == constant.GroupStatusDismissed {
		return nil, utils.Wrap(errs.ErrDismissedAlready, "")
	}
	resp := &pbclub.SetServerGroupInfoResp{}

	data := UpdateGroupInfoMap(ctx, req)
	if len(data) == 0 {
		return resp, nil
	}
	if err := c.ClubDatabase.UpdateServerGroup(ctx, group.GroupID, data); err != nil {
		return nil, err
	}
	group, err = c.ClubDatabase.TakeGroup(ctx, req.GroupInfo.GroupID)
	if err != nil {
		return nil, err
	}
	if req.DappID != "" && req.GroupInfo.GroupMode == constant.AppGroupMode {
		gdm, err := c.ClubDatabase.TakeGroupDapp(ctx, req.GroupInfo.GroupID)
		if err != nil {
			return nil, err
		}
		resp.Dapp = convert.Db2PbGroupDapp(gdm)
	}
	// tips := &sdkws.GroupInfoSetTips{
	// 	Group:    s.groupDB2PB(group, owner.UserID, count),
	// 	MuteTime: 0,
	// 	OpUser:   &sdkws.GroupMemberFullInfo{},
	// }
	// if opMember != nil {
	// 	tips.OpUser = s.groupMemberDB2PB(opMember, 0)
	// }
	// var num int
	// if req.GroupInfoForSet.Notification != "" {
	// 	go func() {
	// 		nctx := mcontext.NewCtx("@@@" + mcontext.GetOperationID(ctx))
	// 		conversation := &pbconversation.ConversationReq{
	// 			ConversationID:   msgprocessor.GetConversationIDBySessionType(constant.SuperGroupChatType, req.GroupInfoForSet.GroupID),
	// 			ConversationType: constant.SuperGroupChatType,
	// 			GroupID:          req.GroupInfoForSet.GroupID,
	// 		}
	// 		resp, err := s.GetGroupMemberUserIDs(nctx, &pbgroup.GetGroupMemberUserIDsReq{GroupID: req.GroupInfoForSet.GroupID})
	// 		if err != nil {
	// 			log.ZWarn(ctx, "GetGroupMemberIDs", err)
	// 			return
	// 		}
	// 		conversation.GroupAtType = &wrapperspb.Int32Value{Value: constant.GroupNotification}
	// 		if err := s.conversationRpcClient.SetConversations(nctx, resp.UserIDs, conversation); err != nil {
	// 			log.ZWarn(ctx, "SetConversations", err, resp.UserIDs, conversation)
	// 		}
	// 	}()
	// 	num++
	// 	s.Notification.GroupInfoSetAnnouncementNotification(ctx, &sdkws.GroupInfoSetAnnouncementTips{Group: tips.Group, OpUser: tips.OpUser})
	// }
	// switch len(data) - num {
	// case 0:
	// case 1:
	// 	if req.GroupInfoForSet.GroupName == "" {
	// 		s.Notification.GroupInfoSetNotification(ctx, tips)
	// 	} else {
	// 		s.Notification.GroupInfoSetNameNotification(ctx, &sdkws.GroupInfoSetNameTips{Group: tips.Group, OpUser: tips.OpUser})
	// 	}
	// default:
	// 	s.Notification.GroupInfoSetNotification(ctx, tips)
	// }
	resp.GroupInfo = convert.Db2PbGroupInfo(group, group.CreatorUserID, 0)
	return resp, nil
}

func (c *clubServer) SetServerGroupOrder(ctx context.Context, req *pbclub.SetServerGroupOrderReq) (*pbclub.SetServerGroupOrderResp, error) {
	if !c.checkManageGroup(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	resp := &pbclub.SetServerGroupOrderResp{}

	groups, err := c.ClubDatabase.FindGroup(ctx, []string{req.ServerID})
	if err != nil {
		return nil, err
	}

	DbGroupIDs := utils.Slice(groups, func(e *relationtb.GroupModel) string { return e.GroupID })
	for _, category := range req.CategoryList {
		for i, group := range category.GroupList {
			if utils.Contain(group.GroupID, DbGroupIDs...) {
				data := make(map[string]any)
				if group.GroupName != "" {
					data["name"] = group.GroupName
				}
				data["reorder_weight"] = i
				data["group_category_id"] = category.CategoryInfo.CategoryID
				err := c.ClubDatabase.UpdateServerGroupOrder(ctx, group.GroupID, data)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return resp, nil
}

func (c *clubServer) DeleteServerGroup(ctx context.Context, req *pbclub.DeleteServerGroupReq) (*pbclub.DeleteServerGroupResp, error) {
	if !c.checkManageGroup(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	resp := &pbclub.DeleteServerGroupResp{}
	if len(req.GroupIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("groupIDs is empty")
	}

	err := c.ClubDatabase.DeleteServerGroup(ctx, req.ServerID, req.GroupIDs)
	if err != nil {
		return nil, err
	}

	return resp, nil
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
	if !c.checkManageGroup(ctx, req.GroupInfo.ServerID) {
		return nil, errs.ErrNoPermission
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

func (c *clubServer) GetServerGroupMemberUserIDs(ctx context.Context, req *pbclub.GetServerGroupMemberUserIDsReq) (resp *pbclub.GetServerGroupMemberUserIDsResp, err error) {
	resp = &pbclub.GetServerGroupMemberUserIDsResp{}
	group, err := c.ClubDatabase.TakeGroup(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	resp.UserIDs, err = c.ClubDatabase.FindServerMemberUserID(ctx, group.ServerID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clubServer) GetServerGroupsInfo(ctx context.Context, req *pbclub.GetServerGroupsInfoReq) (*pbclub.GetServerGroupsInfoResp, error) {
	resp := &pbclub.GetServerGroupsInfoResp{}
	if len(req.GroupIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("groupID is empty")
	}
	groups, err := c.GroupDatabase.FindGroup(ctx, req.GroupIDs)
	if err != nil {
		return nil, err
	}
	ServerIDs := utils.Slice(groups, func(e *relationtb.GroupModel) string {
		return e.ServerID
	})
	serverMemberNumMap, err := c.ClubDatabase.MapServerMemberNum(ctx, ServerIDs)
	if err != nil {
		return nil, err
	}
	resp.GroupInfos = utils.Slice(groups, func(e *relationtb.GroupModel) *sdkws.GroupInfo {
		return convert.Db2PbGroupInfo(e, e.CreatorUserID, serverMemberNumMap[e.ServerID])
	})
	return resp, nil
}

func (c *clubServer) GetServerGroupMembersInfo(ctx context.Context, req *pbclub.GetServerGroupMembersInfoReq) (*pbclub.GetServerGroupMembersInfoResp, error) {
	resp := &pbclub.GetServerGroupMembersInfoResp{}
	if len(req.UserIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("userIDs empty")
	}
	if req.GroupID == "" {
		return nil, errs.ErrArgs.Wrap("groupID empty")
	}
	group, err := c.ClubDatabase.TakeGroup(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	members, err := c.FindServerMember(ctx, []string{group.ServerID}, req.UserIDs, nil)
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

func (c *clubServer) GetServerGroupBaseInfos(ctx context.Context, req *pbclub.GetServerGroupBaseInfosReq) (*pbclub.GetServerGroupBaseInfosResp, error) {
	if req.GroupIDs == nil || len(req.GroupIDs) == 0 {
		return nil, errs.ErrArgs
	}
	resp := &pbclub.GetServerGroupBaseInfosResp{}

	groups, err := c.GroupDatabase.FindGroup(ctx, req.GroupIDs)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		each := &sdkws.ServerGroupBaseInfo{
			GroupID:       group.GroupID,
			GroupName:     group.GroupName,
			GroupFaceUrl:  group.FaceURL,
			Condition:     group.Condition,
			ConditionType: group.ConditionType,
		}
		if group.ServerID != "" {
			if server, err := c.ClubDatabase.FindServer(ctx, []string{group.ServerID}); err != nil {
				continue
			} else {
				each.ServerID = server[0].ServerID
				each.ServerName = server[0].ServerName
				each.ServerIconUrl = server[0].Icon
				each.ServerOwnerUserID = server[0].OwnerUserID
			}
		}
		resp.ServerGroupBaseInfos = append(resp.ServerGroupBaseInfos, each)
	}
	return resp, nil
}

func (c *clubServer) GetGroupsByServer(ctx context.Context, req *pbclub.GetGroupsByServerReq) (*pbclub.GetGroupsByServerResp, error) {
	resp := &pbclub.GetGroupsByServerResp{}
	groups, err := c.ClubDatabase.FindGroup(ctx, req.ServerIDs)
	if err != nil {
		return nil, err
	}

	serverMemberNumMap, err := c.ClubDatabase.MapServerMemberNum(ctx, req.ServerIDs)
	if err != nil {
		return nil, err
	}
	resp.Groups = utils.Slice(groups, func(e *relationtb.GroupModel) *sdkws.GroupInfo {
		return convert.Db2PbGroupInfo(e, e.CreatorUserID, serverMemberNumMap[e.ServerID])
	})
	return resp, nil
}

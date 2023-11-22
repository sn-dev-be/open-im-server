package club

import (
	"context"
	"strings"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	pbmsg "github.com/OpenIMSDK/protocol/msg"
	"github.com/OpenIMSDK/protocol/sdkws"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (c *clubServer) ServerApplicationResponse(ctx context.Context, req *pbclub.ServerApplicationResponseReq) (*pbclub.ServerApplicationResponseResp, error) {
	defer log.ZInfo(ctx, utils.GetFuncName()+" Return")
	if !utils.Contain(req.HandleResult, constant.ServerResponseAgree, constant.ServerResponseRefuse) {
		return nil, errs.ErrArgs.Wrap("HandleResult unknown")
	}
	if !c.checkManageServer(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	serverRequest, err := c.ClubDatabase.TakeServerRequest(ctx, req.ServerID, req.FromUserID)
	if err != nil {
		return nil, err
	}
	if serverRequest.HandleResult != 0 {
		return nil, errs.ErrGroupRequestHandled.Wrap("server request already processed")
	}
	var inServer bool
	if _, err := c.ClubDatabase.TakeServerMember(ctx, req.ServerID, req.FromUserID); err == nil {
		inServer = true
	} else if !c.IsNotFound(err) {
		return nil, err
	}
	user, err := c.User.GetPublicUserInfo(ctx, req.FromUserID)
	if err != nil {
		return nil, err
	}
	serverRole, err := c.getServerRoleByPriority(ctx, req.ServerID, constant.ServerOrdinaryUsers)
	if err != nil {
		return nil, errs.ErrRecordNotFound.Wrap("server role is not exists")
	}
	var member *relationtb.ServerMemberModel
	if (!inServer) && req.HandleResult == constant.ServerResponseAgree {
		member = &relationtb.ServerMemberModel{
			ServerID:       req.ServerID,
			UserID:         req.FromUserID,
			ServerRoleID:   serverRole.RoleID,
			Nickname:       "",
			FaceURL:        "",
			RoleLevel:      constant.ServerOrdinaryUsers,
			JoinTime:       time.Now(),
			JoinSource:     serverRequest.JoinSource,
			MuteEndTime:    time.Unix(0, 0),
			InviterUserID:  serverRequest.InviterUserID,
			OperatorUserID: mcontext.GetOpUserID(ctx),
			Ex:             serverRequest.Ex,
		}
		// if err = CallbackBeforeMemberJoinServer(ctx, member, server.Ex); err != nil {
		// 	return nil, err
		// }
	}
	log.ZDebug(ctx, "ServerApplicationResponse", "inServer", inServer, "HandleResult", req.HandleResult, "member", member)
	if err := c.ClubDatabase.HandlerServerRequest(ctx, req.ServerID, req.FromUserID, req.HandledMsg, req.HandleResult, member); err != nil {
		return nil, err
	}
	switch req.HandleResult {
	case constant.ServerResponseAgree:
		if err := c.conversationRpcClient.ServerGroupChatFirstCreateConversation(ctx, req.ServerID, []string{req.FromUserID}); err != nil {
			return nil, err
		}
		c.Notification.ServerApplicationAcceptedNotification(ctx, req)
		if member == nil {
			log.ZDebug(ctx, "ServerApplicationResponse", "member is nil")
		} else {
			// c.Notification.MemberEnterNotification(ctx, req.ServerID, req.FromUserID)
		}
	case constant.ServerResponseRefuse:
		c.Notification.ServerApplicationRejectedNotification(ctx, req)
	}
	if err := c.modifyServerApplicationStatus(ctx, req, user, serverRequest); err != nil {
		return nil, err
	}
	return &pbclub.ServerApplicationResponseResp{}, nil
}

func (c *clubServer) GetServerApplicationList(ctx context.Context, req *pbclub.GetServerApplicationListReq) (*pbclub.GetServerApplicationListResp, error) {
	serverIDs, err := c.ClubDatabase.FindUserManagedServerID(ctx, req.FromUserID)
	if err != nil {
		return nil, err
	}
	resp := &pbclub.GetServerApplicationListResp{}
	if len(serverIDs) == 0 {
		return resp, nil
	}
	total, serverRequests, err := c.ClubDatabase.PageServerRequest(ctx, serverIDs, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		return nil, err
	}
	resp.Total = total
	if len(serverRequests) == 0 {
		return resp, nil
	}
	var userIDs []string

	for _, gr := range serverRequests {
		userIDs = append(userIDs, gr.UserID)
	}
	userIDs = utils.Distinct(userIDs)
	userMap, err := c.User.GetPublicUserInfoMap(ctx, userIDs, true)
	if err != nil {
		return nil, err
	}
	servers, err := c.ClubDatabase.FindServer(ctx, utils.Distinct(serverIDs))
	if err != nil {
		return nil, err
	}
	serverMap := utils.SliceToMap(servers, func(e *relationtb.ServerModel) string {
		return e.ServerID
	})
	if ids := utils.Single(utils.Keys(serverMap), serverIDs); len(ids) > 0 {
		return nil, errs.ErrGroupIDNotFound.Wrap(strings.Join(ids, ","))
	}
	serverMemberNumMap, err := c.ClubDatabase.MapServerMemberNum(ctx, serverIDs)
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
	resp.ServerRequests = utils.Slice(serverRequests, func(e *relationtb.ServerRequestModel) *sdkws.ServerRequest {
		return convert.Db2PbServerRequest(e, userMap[e.UserID], convert.Db2PbServerInfo(serverMap[e.ServerID], ownerMap[e.ServerID].UserID, serverMemberNumMap[e.ServerID]))
	})
	return resp, nil
}

func (c *clubServer) GetUserReqApplicationList(ctx context.Context, req *pbclub.GetUserReqApplicationListReq) (*pbclub.GetUserReqApplicationListResp, error) {
	resp := &pbclub.GetUserReqApplicationListResp{}
	user, err := c.User.GetPublicUserInfo(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	var pageNumber, showNumber int32
	if req.Pagination != nil {
		pageNumber = req.Pagination.PageNumber
		showNumber = req.Pagination.ShowNumber
	}
	total, requests, err := c.ClubDatabase.PageServerRequestUser(ctx, req.UserID, pageNumber, showNumber)
	if err != nil {
		return nil, err
	}
	resp.Total = total
	if len(requests) == 0 {
		return resp, nil
	}
	serverIDs := utils.Distinct(utils.Slice(requests, func(e *relationtb.ServerRequestModel) string {
		return e.ServerID
	}))
	servers, err := c.ClubDatabase.FindNotDismissedServer(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	serverMap := utils.SliceToMap(servers, func(e *relationtb.ServerModel) string {
		return e.ServerID
	})
	if ids := utils.Single(serverIDs, utils.Keys(serverMap)); len(ids) > 0 {
		return nil, errs.ErrGroupIDNotFound.Wrap(strings.Join(ids, ","))
	}
	owners, err := c.FindServerMember(ctx, serverIDs, nil, []int32{constant.ServerOwner})
	if err != nil {
		return nil, err
	}
	ownerMap := utils.SliceToMap(owners, func(e *relationtb.ServerMemberModel) string {
		return e.ServerID
	})
	if ids := utils.Single(serverIDs, utils.Keys(ownerMap)); len(ids) > 0 {
		return nil, errs.ErrData.Wrap("server no owner", strings.Join(ids, ","))
	}
	serverMemberNum, err := c.ClubDatabase.MapServerMemberNum(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	resp.ServerRequests = utils.Slice(requests, func(e *relationtb.ServerRequestModel) *sdkws.ServerRequest {
		return convert.Db2PbServerRequest(e, user, convert.Db2PbServerInfo(serverMap[e.ServerID], ownerMap[e.ServerID].UserID, uint32(serverMemberNum[e.ServerID])))
	})
	return resp, nil
}

func (c *clubServer) GetServerUsersReqApplicationList(ctx context.Context, req *pbclub.GetServerUsersReqApplicationListReq) (*pbclub.GetServerUsersReqApplicationListResp, error) {
	resp := &pbclub.GetServerUsersReqApplicationListResp{}
	total, requests, err := c.ClubDatabase.FindServerRequests(ctx, req.ServerID, req.UserIDs)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return resp, nil
	}
	serverIDs := utils.Distinct(utils.Slice(requests, func(e *relationtb.ServerRequestModel) string {
		return e.ServerID
	}))
	servers, err := c.ClubDatabase.FindServer(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	serverMap := utils.SliceToMap(servers, func(e *relationtb.ServerModel) string {
		return e.ServerID
	})
	if ids := utils.Single(serverIDs, utils.Keys(serverMap)); len(ids) > 0 {
		return nil, errs.ErrGroupIDNotFound.Wrap(strings.Join(ids, ","))
	}
	owners, err := c.FindServerMember(ctx, serverIDs, nil, []int32{constant.ServerOwner})
	if err != nil {
		return nil, err
	}
	ownerMap := utils.SliceToMap(owners, func(e *relationtb.ServerMemberModel) string {
		return e.ServerID
	})
	if ids := utils.Single(serverIDs, utils.Keys(ownerMap)); len(ids) > 0 {
		return nil, errs.ErrData.Wrap("server no owner", strings.Join(ids, ","))
	}
	serverMemberNum, err := c.ClubDatabase.MapServerMemberNum(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	resp.ServerRequests = utils.Slice(requests, func(e *relationtb.ServerRequestModel) *sdkws.ServerRequest {
		return convert.Db2PbServerRequest(e, nil, convert.Db2PbServerInfo(serverMap[e.ServerID], ownerMap[e.ServerID].UserID, uint32(serverMemberNum[e.ServerID])))
	})
	resp.Total = total
	return resp, nil
}

func (c *clubServer) modifyServerApplicationStatus(
	ctx context.Context,
	req *pbclub.ServerApplicationResponseReq,
	user *sdkws.PublicUserInfo,
	serverRequest *relationtb.ServerRequestModel,
) error {
	server, err := c.ClubDatabase.TakeServer(ctx, req.ServerID)
	if err != nil {
		return err
	}
	tips := &sdkws.JoinServerApplicationTips{
		Server:       convert.DB2PbServerInfo(server),
		Applicant:    user,
		ReqMsg:       serverRequest.ReqMsg,
		HandleResult: req.HandleResult,
	}
	modifyReq := pbmsg.ModifyMsgReq{
		ConversationID: req.ConversationID,
		Seq:            req.Seq,
		UserID:         user.UserID,
		ModifyType:     constant.MsgModifyServerRequestStatus,
		Content:        utils.StructToJsonString(&tips),
	}
	_, err = c.msgRpcClient.ModifyMsg(ctx, &modifyReq)
	if err != nil {
		return err
	}
	return nil
}

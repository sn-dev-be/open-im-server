package group

import (
	"Open_IM/internal/rpc/fault_tolerant"
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	rocksCache "Open_IM/pkg/common/db/rocks_cache"
	"Open_IM/pkg/common/log"
	promePkg "Open_IM/pkg/common/prometheus"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/common/trace_log"
	cp "Open_IM/pkg/common/utils"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"github.com/OpenIMSDK/getcdv3"

	pbCache "Open_IM/pkg/proto/cache"
	pbConversation "Open_IM/pkg/proto/conversation"
	pbGroup "Open_IM/pkg/proto/group"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"
)

type groupServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewGroupServer(port int) *groupServer {
	log.NewPrivateLog(constant.LogFileName)
	return &groupServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImGroupName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *groupServer) Run() {
	log.NewInfo("", "group rpc start ")
	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)
	//listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("", "listen network success, ", address, listener)
	defer listener.Close()
	//grpc server
	recvSize := 1024 * 1024 * constant.GroupRPCRecvSize
	sendSize := 1024 * 1024 * constant.GroupRPCSendSize
	var grpcOpts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(recvSize),
		grpc.MaxSendMsgSize(sendSize),
	}
	if config.Config.Prometheus.Enable {
		promePkg.NewGrpcRequestCounter()
		promePkg.NewGrpcRequestFailedCounter()
		promePkg.NewGrpcRequestSuccessCounter()
		grpcOpts = append(grpcOpts, []grpc.ServerOption{
			// grpc.UnaryInterceptor(promePkg.UnaryServerInterceptorProme),
			grpc.StreamInterceptor(grpcPrometheus.StreamServerInterceptor),
			grpc.UnaryInterceptor(grpcPrometheus.UnaryServerInterceptor),
		}...)
	}
	srv := grpc.NewServer(grpcOpts...)
	defer srv.GracefulStop()
	//Service registers with etcd
	pbGroup.RegisterGroupServer(srv, s)

	rpcRegisterIP := config.Config.RpcRegisterIP
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName, 10, "")
	if err != nil {
		log.NewError("", "RegisterEtcd failed ", err.Error())
		panic(utils.Wrap(err, "register group module  rpc to etcd err"))

	}
	log.Info("", "RegisterEtcd ", s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName)
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("", "group rpc success")
}

func (s *groupServer) CreateGroup(ctx context.Context, req *pbGroup.CreateGroupReq) (resp *pbGroup.CreateGroupResp, _ error) {
	resp = &pbGroup.CreateGroupResp{CommonResp: &open_im_sdk.CommonResp{}, GroupInfo: &open_im_sdk.GroupInfo{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetFuncName(1), nil, "req", req.String(), "resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if err := token_verify.CheckAccessV2(ctx, req.OpUserID, req.OwnerUserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	var groupOwnerNum int
	var userIDs []string
	for _, info := range req.InitMemberList {
		if info.RoleLevel == constant.GroupOwner {
			groupOwnerNum++
		}
		userIDs = append(userIDs, info.UserID)
	}
	if req.OwnerUserID != "" {
		groupOwnerNum++
		userIDs = append(userIDs, req.OwnerUserID)
	}
	if groupOwnerNum != 1 {
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}
	if utils.IsRepeatStringSlice(userIDs) {
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}
	users, err := rocksCache.GetUserInfoFromCacheBatch(userIDs)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if len(users) != len(userIDs) {
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}
	userMap := make(map[string]*imdb.User)
	for i, user := range users {
		userMap[user.UserID] = users[i]
	}
	if err := s.DelGroupAndUserCache(req.OperationID, "", userIDs); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if err := callbackBeforeCreateGroup(ctx, req); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupId := req.GroupInfo.GroupID
	if groupId == "" {
		groupId = utils.Md5(req.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
		bi := big.NewInt(0)
		bi.SetString(groupId[0:8], 16)
		groupId = bi.String()
	}
	groupInfo := imdb.Group{}
	utils.CopyStructFields(&groupInfo, req.GroupInfo)
	groupInfo.CreatorUserID = req.OpUserID
	groupInfo.GroupID = groupId
	groupInfo.CreateTime = time.Now()
	if groupInfo.NotificationUpdateTime.Unix() < 0 {
		groupInfo.NotificationUpdateTime = utils.UnixSecondToTime(0)
	}
	if req.GroupInfo.GroupType != constant.SuperGroup {
		var groupMembers []*imdb.GroupMember
		joinGroup := func(userID string, roleLevel int32) error {
			groupMember := &imdb.GroupMember{GroupID: groupId, RoleLevel: roleLevel, OperatorUserID: req.OpUserID, JoinSource: constant.JoinByInvitation, InviterUserID: req.OpUserID}
			user := userMap[userID]
			utils.CopyStructFields(&groupMember, user)
			if err := CallbackBeforeMemberJoinGroup(ctx, req.OperationID, groupMember, groupInfo.Ex); err != nil {
				return err
			}
			groupMembers = append(groupMembers, groupMember)
			return nil
		}
		if req.OwnerUserID == "" {
			if err := joinGroup(req.OwnerUserID, constant.GroupOwner); err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
		}
		for _, info := range req.InitMemberList {
			if err := joinGroup(info.UserID, info.RoleLevel); err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
		}
		if err := (*imdb.GroupMember)(nil).Create(ctx, groupMembers); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
	} else {
		if err := db.DB.CreateSuperGroup(groupId, userIDs, len(userIDs)); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
	}
	if err := (*imdb.Group)(nil).Create(ctx, []*imdb.Group{&groupInfo}); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	utils.CopyStructFields(resp.GroupInfo, groupInfo)
	resp.GroupInfo.MemberCount = uint32(len(userIDs))
	if req.GroupInfo.GroupType != constant.SuperGroup {
		chat.GroupCreatedNotification(req.OperationID, req.OpUserID, groupId, userIDs)
	} else {
		for _, userID := range userIDs {
			if err := rocksCache.DelJoinedSuperGroupIDListFromCache(userID); err != nil {
				trace_log.SetContextInfo(ctx, "DelJoinedSuperGroupIDListFromCache", err, "userID", userID)
			}
		}
		go func() {
			for _, v := range userIDs {
				chat.SuperGroupNotification(req.OperationID, v, v)
			}
		}()
	}
	return
}

func (s *groupServer) GetJoinedGroupList(ctx context.Context, req *pbGroup.GetJoinedGroupListReq) (resp *pbGroup.GetJoinedGroupListResp, _ error) {
	resp = &pbGroup.GetJoinedGroupListResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetFuncName(1), nil, "req", req.String(), "resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if err := token_verify.CheckAccessV2(ctx, req.OpUserID, req.FromUserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	joinedGroupList, err := rocksCache.GetJoinedGroupIDListFromCache(req.FromUserID)
	if err != nil {
		SetErr(ctx, "GetJoinedGroupIDListFromCache", err, &resp.CommonResp.ErrCode, &resp.CommonResp.ErrMsg, "userID", req.FromUserID)
		return
	}
	for _, groupID := range joinedGroupList {
		var groupNode open_im_sdk.GroupInfo
		num, err := rocksCache.GetGroupMemberNumFromCache(groupID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetGroupMemberNumFromCache", err, "groupID", groupID)
			continue
		}
		owner, err := (*imdb.GroupMember)(nil).TakeOwnerInfo(ctx, groupID)
		//owner, err2 := imdb.GetGroupOwnerInfoByGroupID(groupID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "TakeOwnerInfo", err, "groupID", groupID)
			continue
		}
		group, err := rocksCache.GetGroupInfoFromCache(ctx, groupID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetGroupInfoFromCache", err, "groupID", groupID)
			continue
		}
		if group.GroupType == constant.SuperGroup {
			continue
		}
		if group.Status == constant.GroupStatusDismissed {
			trace_log.SetContextInfo(ctx, "group.Status == constant.GroupStatusDismissed", nil, "groupID", groupID)
			continue
		}
		utils.CopyStructFields(&groupNode, group)
		groupNode.CreateTime = uint32(group.CreateTime.Unix())
		groupNode.NotificationUpdateTime = uint32(group.NotificationUpdateTime.Unix())
		if group.NotificationUpdateTime.Unix() < 0 {
			groupNode.NotificationUpdateTime = 0
		}

		groupNode.MemberCount = uint32(num)
		groupNode.OwnerUserID = owner.UserID
		resp.GroupList = append(resp.GroupList, &groupNode)
	}
	return
}

func (s *groupServer) InviteUserToGroup(ctx context.Context, req *pbGroup.InviteUserToGroupReq) (resp *pbGroup.InviteUserToGroupResp, _ error) {
	resp = &pbGroup.InviteUserToGroupResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetFuncName(1), nil, "req", req.String(), "resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if !imdb.IsExistGroupMember(req.GroupID, req.OpUserID) && !token_verify.IsManagerUserID(req.OpUserID) {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}
	groupInfo, err := (*imdb.Group)(nil).Take(ctx, req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		// TODO
		//errMsg := " group status is dismissed "
		//return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}, nil
		SetErrorForResp(constant.ErrStatus, resp.CommonResp)
		return
	}
	if groupInfo.NeedVerification == constant.AllNeedVerification &&
		!imdb.IsGroupOwnerAdmin(req.GroupID, req.OpUserID) && !token_verify.IsManagerUserID(req.OpUserID) {
		joinReq := pbGroup.JoinGroupReq{}
		for _, v := range req.InvitedUserIDList {
			var groupRequest imdb.GroupRequest
			groupRequest.UserID = v
			groupRequest.GroupID = req.GroupID
			groupRequest.JoinSource = constant.JoinByInvitation
			groupRequest.InviterUserID = req.OpUserID
			err = imdb.InsertIntoGroupRequest(groupRequest)
			if err != nil {
				var resultNode pbGroup.Id2Result
				resultNode.Result = -1
				resultNode.UserID = v
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
				// log.NewError(req.OperationID, "InsertIntoGroupRequest failed ", err.Error(), groupRequest)
				//	return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			} else {
				var resultNode pbGroup.Id2Result
				resultNode.Result = 0
				resultNode.UserID = v
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				joinReq.GroupID = req.GroupID
				joinReq.OperationID = req.OperationID
				joinReq.OpUserID = v
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				chat.JoinGroupApplicationNotification(&joinReq)
			}
		}
		return
	}
	if err := s.DelGroupAndUserCache(req.OperationID, req.GroupID, req.InvitedUserIDList); err != nil {
		//log.NewError(req.OperationID, "DelGroupAndUserCache failed", err.Error())
		//return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}, nil
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	//from User:  invite: applicant
	//to user:  invite: invited
	var okUserIDList []string
	if groupInfo.GroupType != constant.SuperGroup {
		for _, v := range req.InvitedUserIDList {
			var resultNode pbGroup.Id2Result
			resultNode.UserID = v
			resultNode.Result = 0
			toUserInfo, err := imdb.GetUserByUserID(v)
			if err != nil {
				trace_log.SetContextInfo(ctx, "GetUserByUserID", err, "userID", v)
				resultNode.Result = -1
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
			}

			if imdb.IsExistGroupMember(req.GroupID, v) {
				trace_log.SetContextInfo(ctx, "IsExistGroupMember", err, "groupID", req.GroupID, "userID", v)
				resultNode.Result = -1
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
			}
			var toInsertInfo imdb.GroupMember
			utils.CopyStructFields(&toInsertInfo, toUserInfo)
			toInsertInfo.GroupID = req.GroupID
			toInsertInfo.RoleLevel = constant.GroupOrdinaryUsers
			toInsertInfo.OperatorUserID = req.OpUserID
			toInsertInfo.InviterUserID = req.OpUserID
			toInsertInfo.JoinSource = constant.JoinByInvitation
			if err := CallbackBeforeMemberJoinGroup(ctx, req.OperationID, &toInsertInfo, groupInfo.Ex); err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
			err = imdb.InsertIntoGroupMember(toInsertInfo)
			if err != nil {
				trace_log.SetContextInfo(ctx, "InsertIntoGroupMember", err, "args", toInsertInfo)
				resultNode.Result = -1
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
			}
			okUserIDList = append(okUserIDList, v)
			err = db.DB.AddGroupMember(req.GroupID, toUserInfo.UserID)
			if err != nil {
				trace_log.SetContextInfo(ctx, "AddGroupMember", err, "groupID", req.GroupID, "userID", toUserInfo.UserID)
			}
			resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
		}
	} else {
		okUserIDList = req.InvitedUserIDList
		if err := db.DB.AddUserToSuperGroup(req.GroupID, req.InvitedUserIDList); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
	}

	// set conversations
	var haveConUserID []string
	var sessionType int
	if groupInfo.GroupType == constant.NormalGroup {
		sessionType = constant.GroupChatType
	} else {
		sessionType = constant.SuperGroupChatType
	}
	conversations, err := imdb.GetConversationsByConversationIDMultipleOwner(okUserIDList, utils.GetConversationIDBySessionType(req.GroupID, sessionType))
	if err != nil {
		trace_log.SetContextInfo(ctx, "GetConversationsByConversationIDMultipleOwner", err, "OwnerUserIDList", okUserIDList, "groupID", req.GroupID, "sessionType", sessionType)
	}
	for _, v := range conversations {
		haveConUserID = append(haveConUserID, v.OwnerUserID)
	}
	var reqPb pbUser.SetConversationReq
	var c pbConversation.Conversation
	for _, v := range conversations {
		reqPb.OperationID = req.OperationID
		c.OwnerUserID = v.OwnerUserID
		c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, sessionType)
		c.RecvMsgOpt = v.RecvMsgOpt
		c.ConversationType = int32(sessionType)
		c.GroupID = req.GroupID
		c.IsPinned = v.IsPinned
		c.AttachedInfo = v.AttachedInfo
		c.IsPrivateChat = v.IsPrivateChat
		c.GroupAtType = v.GroupAtType
		c.IsNotInGroup = false
		c.Ex = v.Ex
		reqPb.Conversation = &c
		etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		client := pbUser.NewUserClient(etcdConn)
		respPb, err := client.SetConversation(ctx, &reqPb)
		trace_log.SetContextInfo(ctx, "SetConversation", err, "req", &reqPb, "resp", respPb)
	}
	for _, v := range utils.DifferenceString(haveConUserID, okUserIDList) {
		reqPb.OperationID = req.OperationID
		c.OwnerUserID = v
		c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, sessionType)
		c.ConversationType = int32(sessionType)
		c.GroupID = req.GroupID
		c.IsNotInGroup = false
		c.UpdateUnreadCountTime = utils.GetCurrentTimestampByMill()
		reqPb.Conversation = &c
		etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		client := pbUser.NewUserClient(etcdConn)
		respPb, err := client.SetConversation(context.Background(), &reqPb)
		trace_log.SetContextInfo(ctx, "SetConversation", err, "req", &reqPb, "resp", respPb)
	}

	if groupInfo.GroupType != constant.SuperGroup {
		chat.MemberInvitedNotification(req.OperationID, req.GroupID, req.OpUserID, req.Reason, okUserIDList)
	} else {
		for _, v := range req.InvitedUserIDList {
			if err := rocksCache.DelJoinedSuperGroupIDListFromCache(v); err != nil {
				trace_log.SetContextInfo(ctx, "DelJoinedSuperGroupIDListFromCache", err, "userID", v)
			}
		}
		for _, v := range req.InvitedUserIDList {
			chat.SuperGroupNotification(req.OperationID, v, v)
		}
	}
	return
}

func (s *groupServer) GetGroupAllMember(ctx context.Context, req *pbGroup.GetGroupAllMemberReq) (resp *pbGroup.GetGroupAllMemberResp, err error) {
	resp = &pbGroup.GetGroupAllMemberResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetFuncName(1), nil, "req", req.String(), "resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	groupInfo, err := rocksCache.GetGroupInfoFromCache(ctx, req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if groupInfo.GroupType != constant.SuperGroup {
		memberList, err := rocksCache.GetGroupMembersInfoFromCache(req.Count, req.Offset, req.GroupID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		for _, v := range memberList {
			var node open_im_sdk.GroupMemberFullInfo
			cp.GroupMemberDBCopyOpenIM(&node, v)
			resp.MemberList = append(resp.MemberList, &node)
		}
	}
	return
}

func (s *groupServer) GetGroupMemberList(ctx context.Context, req *pbGroup.GetGroupMemberListReq) (resp *pbGroup.GetGroupMemberListResp, err error) {
	resp = &pbGroup.GetGroupMemberListResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetFuncName(1), nil, "req", req.String(), "resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	memberList, err := imdb.GetGroupMemberByGroupID(req.GroupID, req.Filter, req.NextSeq, 30)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	for _, v := range memberList {
		var node open_im_sdk.GroupMemberFullInfo
		utils.CopyStructFields(&node, &v)
		resp.MemberList = append(resp.MemberList, &node)
	}
	//db operate  get db sorted by join time
	if int32(len(memberList)) < 30 {
		resp.NextSeq = 0
	} else {
		resp.NextSeq = req.NextSeq + int32(len(memberList))
	}
	return
}

func (s *groupServer) getGroupUserLevel(groupID, userID string) (int, error) {
	opFlag := 0
	if !token_verify.IsManagerUserID(userID) {
		opInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(groupID, userID)
		if err != nil {
			return opFlag, utils.Wrap(err, "")
		}
		if opInfo.RoleLevel == constant.GroupOrdinaryUsers {
			opFlag = 0
		} else if opInfo.RoleLevel == constant.GroupOwner {
			opFlag = 2 //owner
		} else {
			opFlag = 3 //admin
		}
	} else {
		opFlag = 1 //app manager
	}
	return opFlag, nil
}

func (s *groupServer) KickGroupMember(ctx context.Context, req *pbGroup.KickGroupMemberReq) (resp *pbGroup.KickGroupMemberResp, _ error) {
	resp = &pbGroup.KickGroupMemberResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetFuncName(1), nil, "req", req.String(), "resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	groupInfo, err := rocksCache.GetGroupInfoFromCache(ctx, req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	var okUserIDList []string
	if groupInfo.GroupType != constant.SuperGroup {
		opFlag := 0
		if !token_verify.IsManagerUserID(req.OpUserID) {
			opInfo, err := rocksCache.GetGroupMemberInfoFromCache(req.GroupID, req.OpUserID)
			if err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
			if opInfo.RoleLevel == constant.GroupOrdinaryUsers {
				// TODO
				//errMsg := req.OperationID + " opInfo.RoleLevel == constant.GroupOrdinaryUsers " + opInfo.UserID + opInfo.GroupID
				//log.Error(req.OperationID, errMsg)
				//return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
				SetErrorForResp(err, resp.CommonResp)
				return
			} else if opInfo.RoleLevel == constant.GroupOwner {
				opFlag = 2 //owner
			} else {
				opFlag = 3 //admin
			}
		} else {
			opFlag = 1 //app manager
		}

		//op is group owner?
		if len(req.KickedUserIDList) == 0 {
			//log.NewError(req.OperationID, "failed, kick list 0")
			//return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}, nil
			SetErrorForResp(constant.ErrArgs, resp.CommonResp)
			return
		}
		if err := s.DelGroupAndUserCache(req.OperationID, req.GroupID, req.KickedUserIDList); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		//remove
		for _, v := range req.KickedUserIDList {
			kickedInfo, err := rocksCache.GetGroupMemberInfoFromCache(req.GroupID, v)
			if err != nil {
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
				trace_log.SetContextInfo(ctx, "GetGroupMemberInfoFromCache", err, "groupID", req.GroupID, "userID", v)
				continue
			}

			if kickedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
				trace_log.SetContextInfo(ctx, "", nil, "msg", "is constant.GroupAdmin, can't kicked", "groupID", req.GroupID, "userID", v)
				continue
			}
			if kickedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
				trace_log.SetContextInfo(ctx, "", nil, "msg", "is constant.GroupOwner, can't kicked", "groupID", req.GroupID, "userID", v)
				continue
			}

			err = imdb.DeleteGroupMemberByGroupIDAndUserID(req.GroupID, v)
			trace_log.SetContextInfo(ctx, "RemoveGroupMember", err, "groupID", req.GroupID, "userID", v)
			if err != nil {
				log.NewError(req.OperationID, "RemoveGroupMember failed ", err.Error(), req.GroupID, v)
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
			} else {
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: 0})
				okUserIDList = append(okUserIDList, v)
			}
		}
		var reqPb pbUser.SetConversationReq
		var c pbConversation.Conversation
		for _, v := range okUserIDList {
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = v
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsNotInGroup = true
			reqPb.Conversation = &c
			etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			trace_log.SetContextInfo(ctx, "SetConversation", err, "req", &reqPb, "resp", respPb)
		}
	} else {
		okUserIDList = req.KickedUserIDList
		if err := db.DB.RemoverUserFromSuperGroup(req.GroupID, okUserIDList); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
	}

	if groupInfo.GroupType != constant.SuperGroup {
		for _, userID := range okUserIDList {
			if err := rocksCache.DelGroupMemberInfoFromCache(req.GroupID, userID); err != nil {
				trace_log.SetContextInfo(ctx, "DelGroupMemberInfoFromCache", err, "groupID", req.GroupID, "userID", userID)
			}
		}
		chat.MemberKickedNotification(req, okUserIDList)
	} else {
		for _, userID := range okUserIDList {
			if err = rocksCache.DelJoinedSuperGroupIDListFromCache(userID); err != nil {
				trace_log.SetContextInfo(ctx, "DelGroupMemberInfoFromCache", err, "userID", userID)
			}
		}
		go func() {
			for _, v := range req.KickedUserIDList {
				chat.SuperGroupNotification(req.OperationID, v, v)
			}
		}()

	}
	return
}

func (s *groupServer) GetGroupMembersInfo(ctx context.Context, req *pbGroup.GetGroupMembersInfoReq) (resp *pbGroup.GetGroupMembersInfoResp, err error) {
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), err, "rpc req", req.String(), "rpc resp", resp.String())
		trace_log.ShowLog(ctx)
	}()
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)

	resp = &pbGroup.GetGroupMembersInfoResp{CommonResp: &open_im_sdk.CommonResp{}}
	resp.MemberList = []*open_im_sdk.GroupMemberFullInfo{}
	for _, userID := range req.MemberList {
		groupMember, err := rocksCache.GetGroupMemberInfoFromCache(req.GroupID, userID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetGroupMemberInfoFromCache", err, "groupID", req.GroupID, "userID", userID)
			continue
		}
		var memberNode open_im_sdk.GroupMemberFullInfo
		utils.CopyStructFields(&memberNode, groupMember)
		memberNode.JoinTime = int32(groupMember.JoinTime.Unix())
		resp.MemberList = append(resp.MemberList, &memberNode)
	}
	SetErr(ctx, "GetGroupMemberInfoFromCache", err, &resp.CommonResp.ErrCode, &resp.CommonResp.ErrMsg, "groupID", req.GroupID)
	return
}

func FillGroupInfoByGroupID(operationID, groupID string, groupInfo *open_im_sdk.GroupInfo) error {
	group, err := imdb.TakeGroupInfoByGroupID(groupID)
	if err != nil {
		log.Error(operationID, "TakeGroupInfoByGroupID failed ", err.Error(), groupID)
		return utils.Wrap(err, "")
	}
	if group.Status == constant.GroupStatusDismissed {
		log.Debug(operationID, " group constant.GroupStatusDismissed ", group.GroupID)
		return utils.Wrap(constant.ErrDismissedAlready, "")
	}
	return utils.Wrap(cp.GroupDBCopyOpenIM(groupInfo, group), "")
}

func FillPublicUserInfoByUserID(operationID, userID string, userInfo *open_im_sdk.PublicUserInfo) error {
	user, err := imdb.TakeUserByUserID(userID)
	if err != nil {
		log.Error(operationID, "TakeUserByUserID failed ", err.Error(), userID)
		return utils.Wrap(err, "")
	}
	cp.UserDBCopyOpenIMPublicUser(userInfo, user)
	return nil
}

func SetErr(ctx context.Context, funcName string, err error, errCode *int32, errMsg *string, args ...interface{}) {
	errInfo := constant.ToAPIErrWithErr(err)
	*errCode = errInfo.ErrCode
	*errMsg = errInfo.WrapErrMsg
	trace_log.SetContextInfo(ctx, funcName, err, args)
}

func SetErrCodeMsg(ctx context.Context, funcName string, errCodeFrom int32, errMsgFrom string, errCode *int32, errMsg *string, args ...interface{}) {
	*errCode = errCodeFrom
	*errMsg = errMsgFrom
	trace_log.SetContextInfo(ctx, funcName, constant.ToAPIErrWithErrCode(errCodeFrom), args)
}

func SetErrorForResp(err error, commonResp *open_im_sdk.CommonResp) {
	errInfo := constant.ToAPIErrWithErr(err)
	commonResp.ErrCode = errInfo.ErrCode
	commonResp.ErrMsg = errInfo.ErrMsg
	commonResp.DetailErrMsg = err.Error()
}

func (s *groupServer) GetGroupApplicationList(ctx context.Context, req *pbGroup.GetGroupApplicationListReq) (resp *pbGroup.GetGroupApplicationListResp, err error) {
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), err, "rpc req ", req.String(), "rpc resp ", resp.String())
	trace_log.ShowLog(ctx)

	resp = &pbGroup.GetGroupApplicationListResp{CommonResp: &open_im_sdk.CommonResp{}}
	reply, err := imdb.GetRecvGroupApplicationList(req.FromUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return resp, nil
	}
	var errResult error
	trace_log.SetContextInfo(ctx, "GetRecvGroupApplicationList", nil, " FromUserID: ", req.FromUserID, "GroupApplicationList: ", reply)
	for _, v := range reply {
		node := open_im_sdk.GroupRequest{UserInfo: &open_im_sdk.PublicUserInfo{}, GroupInfo: &open_im_sdk.GroupInfo{}}
		err := FillGroupInfoByGroupID(req.OperationID, v.GroupID, node.GroupInfo)
		if err != nil {
			if !errors.Is(errors.Unwrap(err), constant.ErrDismissedAlready) {
				errResult = err
			}
			continue
		}
		trace_log.SetContextInfo(ctx, "FillGroupInfoByGroupID ", nil, " groupID: ", v.GroupID, " groupInfo: ", node.GroupInfo)
		err = FillPublicUserInfoByUserID(req.OperationID, v.UserID, node.UserInfo)
		if err != nil {
			errResult = err
			continue
		}
		cp.GroupRequestDBCopyOpenIM(&node, &v)
		resp.GroupRequestList = append(resp.GroupRequestList, &node)
	}
	if errResult != nil && len(resp.GroupRequestList) == 0 {
		SetErr(ctx, "", errResult, &resp.CommonResp.ErrCode, &resp.CommonResp.ErrMsg)
		return resp, nil
	}
	trace_log.SetRpcRespInfo(ctx, utils.GetSelfFuncName(), resp.String())
	return resp, nil
}

func (s *groupServer) GetGroupsInfo(ctx context.Context, req *pbGroup.GetGroupsInfoReq) (resp *pbGroup.GetGroupsInfoResp, _ error) {
	resp = &pbGroup.GetGroupsInfoResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	groupsInfoList := make([]*open_im_sdk.GroupInfo, 0)
	for _, groupID := range req.GroupIDList {
		groupInfoFromRedis, err := rocksCache.GetGroupInfoFromCache(ctx, groupID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			continue
		}
		var groupInfo open_im_sdk.GroupInfo
		cp.GroupDBCopyOpenIM(&groupInfo, groupInfoFromRedis)
		groupInfo.NeedVerification = groupInfoFromRedis.NeedVerification
		groupsInfoList = append(groupsInfoList, &groupInfo)
	}
	resp.GroupInfoList = groupsInfoList
	return
}

func CheckPermission(ctx context.Context, groupID string, userID string) (err error) {
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), err, "groupID", groupID, "userID", userID)
	}()
	if !token_verify.IsManagerUserID(userID) && !imdb.IsGroupOwnerAdmin(groupID, userID) {
		return utils.Wrap(constant.ErrNoPermission, utils.GetSelfFuncName())
	}
	return nil
}

func (s *groupServer) GroupApplicationResponse(ctx context.Context, req *pbGroup.GroupApplicationResponseReq) (resp *pbGroup.GroupApplicationResponseResp, _ error) {
	resp = &pbGroup.GroupApplicationResponseResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if err := CheckPermission(ctx, req.GroupID, req.OpUserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupRequest := imdb.GroupRequest{}
	utils.CopyStructFields(&groupRequest, req)
	groupRequest.UserID = req.FromUserID
	groupRequest.HandleUserID = req.OpUserID
	groupRequest.HandledTime = time.Now()
	if err := (&imdb.GroupRequest{}).Update(ctx, []*imdb.GroupRequest{&groupRequest}); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupInfo, err := rocksCache.GetGroupInfoFromCache(ctx, req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if req.HandleResult == constant.GroupResponseAgree {
		user, err := imdb.GetUserByUserID(req.FromUserID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		request, err := (&imdb.GroupRequest{}).Take(ctx, req.GroupID, req.FromUserID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		member := imdb.GroupMember{}
		member.GroupID = req.GroupID
		member.UserID = req.FromUserID
		member.RoleLevel = constant.GroupOrdinaryUsers
		member.OperatorUserID = req.OpUserID
		member.FaceURL = user.FaceURL
		member.Nickname = user.Nickname
		member.JoinSource = request.JoinSource
		member.InviterUserID = request.InviterUserID
		member.MuteEndTime = time.Unix(int64(time.Now().Second()), 0)
		err = CallbackBeforeMemberJoinGroup(ctx, req.OperationID, &member, groupInfo.Ex)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}

		err = (&imdb.GroupMember{}).Create(ctx, []*imdb.GroupMember{&member})
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		etcdCacheConn, err := fault_tolerant.GetDefaultConn(config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		cacheClient := pbCache.NewCacheClient(etcdCacheConn)
		cacheResp, err := cacheClient.DelGroupMemberIDListFromCache(context.Background(), &pbCache.DelGroupMemberIDListFromCacheReq{OperationID: req.OperationID, GroupID: req.GroupID})
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		if err := rocksCache.DelGroupMemberListHashFromCache(req.GroupID); err != nil {
			trace_log.SetContextInfo(ctx, "DelGroupMemberListHashFromCache", err, "groupID ", req.GroupID)
		}
		if err := rocksCache.DelJoinedGroupIDListFromCache(req.FromUserID); err != nil {
			trace_log.SetContextInfo(ctx, "DelJoinedGroupIDListFromCache", err, "userID ", req.FromUserID)
		}
		if err := rocksCache.DelGroupMemberNumFromCache(req.GroupID); err != nil {
			trace_log.SetContextInfo(ctx, "DelJoinedGroupIDListFromCache", err, "groupID ", req.GroupID)
		}
		chat.GroupApplicationAcceptedNotification(req)
		chat.MemberEnterNotification(req)
	} else if req.HandleResult == constant.GroupResponseRefuse {
		chat.GroupApplicationRejectedNotification(req)
	} else {
		// TODO
		SetErr(ctx, "", constant.ErrArgs, &resp.CommonResp.ErrCode, &resp.CommonResp.ErrMsg)
		return
	}
	return
}

func (s *groupServer) JoinGroup(ctx context.Context, req *pbGroup.JoinGroupReq) (resp *pbGroup.JoinGroupResp, _ error) {
	resp = &pbGroup.JoinGroupResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if _, err := imdb.GetUserByUserID(req.OpUserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupInfo, err := rocksCache.GetGroupInfoFromCache(ctx, req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		// TODO
		//errMsg := " group status is dismissed "
		//return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}

	if groupInfo.NeedVerification == constant.Directly {
		if groupInfo.GroupType != constant.SuperGroup {
			us, err := imdb.GetUserByUserID(req.OpUserID)
			if err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
			//to group member
			groupMember := imdb.GroupMember{GroupID: req.GroupID, RoleLevel: constant.GroupOrdinaryUsers, OperatorUserID: req.OpUserID}
			utils.CopyStructFields(&groupMember, us)
			if err := CallbackBeforeMemberJoinGroup(ctx, req.OperationID, &groupMember, groupInfo.Ex); err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}

			if err := s.DelGroupAndUserCache(req.OperationID, req.GroupID, []string{req.OpUserID}); err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}

			err = imdb.InsertIntoGroupMember(groupMember)
			if err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}

			var sessionType int
			if groupInfo.GroupType == constant.NormalGroup {
				sessionType = constant.GroupChatType
			} else {
				sessionType = constant.SuperGroupChatType
			}
			var reqPb pbUser.SetConversationReq
			var c pbConversation.Conversation
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = req.OpUserID
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, sessionType)
			c.ConversationType = int32(sessionType)
			c.GroupID = req.GroupID
			c.IsNotInGroup = false
			c.UpdateUnreadCountTime = utils.GetCurrentTimestampByMill()
			reqPb.Conversation = &c
			etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			trace_log.SetContextInfo(ctx, "SetConversation", err, "req", reqPb, "resp", respPb)
			chat.MemberEnterDirectlyNotification(req.GroupID, req.OpUserID, req.OperationID)
			return
		} else {
			// todo
			SetErrorForResp(constant.ErrArgs, resp.CommonResp)
			return
			//log.Error(req.OperationID, "JoinGroup rpc failed, group type:  ", groupInfo.GroupType, "not support directly")
			//return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
		}
	}
	var groupRequest imdb.GroupRequest
	groupRequest.UserID = req.OpUserID
	groupRequest.ReqMsg = req.ReqMessage
	groupRequest.GroupID = req.GroupID
	groupRequest.JoinSource = req.JoinSource
	err = imdb.InsertIntoGroupRequest(groupRequest)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	//_, err = imdb.GetGroupMemberListByGroupIDAndRoleLevel(req.GroupID, constant.GroupOwner)
	//if err != nil {
	//	log.NewError(req.OperationID, "GetGroupMemberListByGroupIDAndRoleLevel failed ", err.Error(), req.GroupID, constant.GroupOwner)
	//	return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil

	chat.JoinGroupApplicationNotification(req)
	return
}

func (s *groupServer) QuitGroup(ctx context.Context, req *pbGroup.QuitGroupReq) (resp *pbGroup.QuitGroupResp, _ error) {
	resp = &pbGroup.QuitGroupResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if groupInfo.GroupType != constant.SuperGroup {
		_, err = imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}

		if err := s.DelGroupAndUserCache(req.OperationID, req.GroupID, []string{req.OpUserID}); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}

		err = imdb.DeleteGroupMemberByGroupIDAndUserID(req.GroupID, req.OpUserID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}

		err = db.DB.DelGroupMember(req.GroupID, req.OpUserID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		//modify quitter conversation info
		var reqPb pbUser.SetConversationReq
		var c pbConversation.Conversation
		reqPb.OperationID = req.OperationID
		c.OwnerUserID = req.OpUserID
		c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
		c.ConversationType = constant.GroupChatType
		c.GroupID = req.GroupID
		c.IsNotInGroup = true
		reqPb.Conversation = &c
		etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
		if etcdConn == nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		client := pbUser.NewUserClient(etcdConn)
		respPb, err := client.SetConversation(context.Background(), &reqPb)
		trace_log.SetContextInfo(ctx, "SetConversation", err, "req", &reqPb, "resp", respPb)
	} else {
		okUserIDList := []string{req.OpUserID}
		if err := db.DB.RemoverUserFromSuperGroup(req.GroupID, okUserIDList); err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
	}

	if groupInfo.GroupType != constant.SuperGroup {
		if err := rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.OpUserID); err != nil {
			trace_log.SetContextInfo(ctx, "DelGroupMemberInfoFromCache", err, "groupID", req.GroupID, "userID", req.OpUserID)
		}
		chat.MemberQuitNotification(req)
	} else {
		if err := rocksCache.DelJoinedSuperGroupIDListFromCache(req.OpUserID); err != nil {
			trace_log.SetContextInfo(ctx, "DelJoinedSuperGroupIDListFromCache", err, "userID", req.OpUserID)
		}
		if err := rocksCache.DelGroupMemberListHashFromCache(req.GroupID); err != nil {
			trace_log.SetContextInfo(ctx, "DelGroupMemberListHashFromCache", err, "groupID", req.GroupID)
		}
		chat.SuperGroupNotification(req.OperationID, req.OpUserID, req.OpUserID)
	}
	return
}

func hasAccess(req *pbGroup.SetGroupInfoReq) bool {
	if utils.IsContain(req.OpUserID, config.Config.Manager.AppManagerUid) {
		return true
	}
	groupUserInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupInfoForSet.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed, ", err.Error(), req.GroupInfoForSet.GroupID, req.OpUserID)
		return false

	}
	if groupUserInfo.RoleLevel == constant.GroupOwner || groupUserInfo.RoleLevel == constant.GroupAdmin {
		return true
	}
	return false
}

func (s *groupServer) SetGroupInfo(ctx context.Context, req *pbGroup.SetGroupInfoReq) (resp *pbGroup.SetGroupInfoResp, err error) {
	resp = &pbGroup.SetGroupInfoResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if !hasAccess(req) {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}
	group, err := imdb.GetGroupInfoByGroupID(req.GroupInfoForSet.GroupID)
	if err != nil {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}
	if group.Status == constant.GroupStatusDismissed {
		// TODO
		//errMsg := " group status is dismissed "
		//return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}

	////bitwise operators: 0001:groupName; 0010:Notification  0100:Introduction; 1000:FaceUrl; 10000:owner
	var changedType int32
	groupName := ""
	notification := ""
	introduction := ""
	faceURL := ""
	if group.GroupName != req.GroupInfoForSet.GroupName && req.GroupInfoForSet.GroupName != "" {
		changedType = 1
		groupName = req.GroupInfoForSet.GroupName
	}
	if group.Notification != req.GroupInfoForSet.Notification && req.GroupInfoForSet.Notification != "" {
		changedType = changedType | (1 << 1)
		notification = req.GroupInfoForSet.Notification
	}
	if group.Introduction != req.GroupInfoForSet.Introduction && req.GroupInfoForSet.Introduction != "" {
		changedType = changedType | (1 << 2)
		introduction = req.GroupInfoForSet.Introduction
	}
	if group.FaceURL != req.GroupInfoForSet.FaceURL && req.GroupInfoForSet.FaceURL != "" {
		changedType = changedType | (1 << 3)
		faceURL = req.GroupInfoForSet.FaceURL
	}

	if req.GroupInfoForSet.NeedVerification != nil {
		changedType = changedType | (1 << 4)
		m := make(map[string]interface{})
		m["need_verification"] = req.GroupInfoForSet.NeedVerification.Value
		if err := imdb.UpdateGroupInfoDefaultZero(req.GroupInfoForSet.GroupID, m); err != nil {
			SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
			return
		}
	}
	if req.GroupInfoForSet.LookMemberInfo != nil {
		changedType = changedType | (1 << 5)
		m := make(map[string]interface{})
		m["look_member_info"] = req.GroupInfoForSet.LookMemberInfo.Value
		if err := imdb.UpdateGroupInfoDefaultZero(req.GroupInfoForSet.GroupID, m); err != nil {
			SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
			return
		}
	}
	if req.GroupInfoForSet.ApplyMemberFriend != nil {
		changedType = changedType | (1 << 6)
		m := make(map[string]interface{})
		m["apply_member_friend"] = req.GroupInfoForSet.ApplyMemberFriend.Value
		if err := imdb.UpdateGroupInfoDefaultZero(req.GroupInfoForSet.GroupID, m); err != nil {
			SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
			return
		}
	}
	//
	//if req.RoleLevel != nil {
	//
	//}
	//only administrators can set group information
	var groupInfo imdb.Group
	utils.CopyStructFields(&groupInfo, req.GroupInfoForSet)
	if req.GroupInfoForSet.Notification != "" {
		groupInfo.NotificationUserID = req.OpUserID
		groupInfo.NotificationUpdateTime = time.Now()
	}
	if err := rocksCache.DelGroupInfoFromCache(req.GroupInfoForSet.GroupID); err != nil {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}
	err = imdb.SetGroupInfo(groupInfo)
	if err != nil {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}
	log.NewInfo(req.OperationID, "SetGroupInfo rpc return ", pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}})
	if changedType != 0 {
		chat.GroupInfoSetNotification(req.OperationID, req.OpUserID, req.GroupInfoForSet.GroupID, groupName, notification, introduction, faceURL, req.GroupInfoForSet.NeedVerification)
	}
	if req.GroupInfoForSet.Notification != "" {
		//get group member user id
		getGroupMemberIDListFromCacheReq := &pbCache.GetGroupMemberIDListFromCacheReq{OperationID: req.OperationID, GroupID: req.GroupInfoForSet.GroupID}
		etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if err != nil {
			SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
			return
		}
		client := pbCache.NewCacheClient(etcdConn)
		cacheResp, err := client.GetGroupMemberIDListFromCache(ctx, getGroupMemberIDListFromCacheReq)
		if err != nil {
			SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
			return
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			//log.NewError(req.OperationID, "GetGroupMemberIDListFromCache rpc logic call failed ", cacheResp.String())
			//return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}}, nil
			// TODO
			resp.CommonResp.ErrCode = cacheResp.CommonResp.ErrCode
			resp.CommonResp.ErrMsg = cacheResp.CommonResp.ErrMsg
			return
		}
		var conversationReq pbConversation.ModifyConversationFieldReq

		conversation := pbConversation.Conversation{
			OwnerUserID:      req.OpUserID,
			ConversationID:   utils.GetConversationIDBySessionType(req.GroupInfoForSet.GroupID, constant.GroupChatType),
			ConversationType: constant.GroupChatType,
			GroupID:          req.GroupInfoForSet.GroupID,
		}
		conversationReq.Conversation = &conversation
		conversationReq.OperationID = req.OperationID
		conversationReq.FieldType = constant.FieldGroupAtType
		conversation.GroupAtType = constant.GroupNotification
		conversationReq.UserIDList = cacheResp.UserIDList
		nEtcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImConversationName, req.OperationID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		nClient := pbConversation.NewConversationClient(nEtcdConn)
		conversationReply, err := nClient.ModifyConversationField(context.Background(), &conversationReq)
		trace_log.SetContextInfo(ctx, "ModifyConversationField", err, "req", &conversationReq, "resp", conversationReply)
	}
	return
}

func (s *groupServer) TransferGroupOwner(ctx context.Context, req *pbGroup.TransferGroupOwnerReq) (resp *pbGroup.TransferGroupOwnerResp, _ error) {
	resp = &pbGroup.TransferGroupOwnerResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		// TODO
		//errMsg := " group status is dismissed "
		//return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	if req.OldOwnerUserID == req.NewOwnerUserID {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	err = rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.NewOwnerUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	err = rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.OldOwnerUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	groupMemberInfo := imdb.GroupMember{GroupID: req.GroupID, UserID: req.OldOwnerUserID, RoleLevel: constant.GroupOrdinaryUsers}
	err = imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupMemberInfo = imdb.GroupMember{GroupID: req.GroupID, UserID: req.NewOwnerUserID, RoleLevel: constant.GroupOwner}
	err = imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	chat.GroupOwnerTransferredNotification(req)
	return
}

func (s *groupServer) GetGroups(ctx context.Context, req *pbGroup.GetGroupsReq) (resp *pbGroup.GetGroupsResp, err error) {
	resp = &pbGroup.GetGroupsResp{
		CommonResp: &open_im_sdk.CommonResp{},
		CMSGroups:  []*pbGroup.CMSGroup{},
		Pagination: &open_im_sdk.ResponsePagination{CurrentPage: req.Pagination.PageNumber, ShowNumber: req.Pagination.ShowNumber},
	}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if req.GroupID != "" {
		groupInfoDB, err := imdb.GetGroupInfoByGroupID(req.GroupID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return resp, nil
			}
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		resp.GroupNum = 1
		groupInfo := &open_im_sdk.GroupInfo{}
		utils.CopyStructFields(groupInfo, groupInfoDB)
		groupMember, err := imdb.GetGroupOwnerInfoByGroupID(req.GroupID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		memberNum, err := imdb.GetGroupMembersCount(req.GroupID, "")
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		groupInfo.MemberCount = uint32(memberNum)
		groupInfo.CreateTime = uint32(groupInfoDB.CreateTime.Unix())
		resp.CMSGroups = append(resp.CMSGroups, &pbGroup.CMSGroup{GroupInfo: groupInfo, GroupOwnerUserName: groupMember.Nickname, GroupOwnerUserID: groupMember.UserID})
	} else {
		groups, count, err := imdb.GetGroupsByName(req.GroupName, req.Pagination.PageNumber, req.Pagination.ShowNumber)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetGroupsByName", err, "GroupName", req.GroupName, "PageNumber", req.Pagination.PageNumber, "ShowNumber", req.Pagination.ShowNumber)
		}
		for _, v := range groups {
			group := &pbGroup.CMSGroup{GroupInfo: &open_im_sdk.GroupInfo{}}
			utils.CopyStructFields(group.GroupInfo, v)
			groupMember, err := imdb.GetGroupOwnerInfoByGroupID(v.GroupID)
			if err != nil {
				trace_log.SetContextInfo(ctx, "GetGroupOwnerInfoByGroupID", err, "GroupID", v.GroupID)
				continue
			}
			group.GroupInfo.CreateTime = uint32(v.CreateTime.Unix())
			group.GroupOwnerUserID = groupMember.UserID
			group.GroupOwnerUserName = groupMember.Nickname
			resp.CMSGroups = append(resp.CMSGroups, group)
		}
		resp.GroupNum = int32(count)
	}
	return
}

func (s *groupServer) GetGroupMembersCMS(ctx context.Context, req *pbGroup.GetGroupMembersCMSReq) (resp *pbGroup.GetGroupMembersCMSResp, _ error) {
	resp = &pbGroup.GetGroupMembersCMSResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	groupMembers, err := imdb.GetGroupMembersByGroupIdCMS(req.GroupID, req.UserName, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupMembersCount, err := imdb.GetGroupMembersCount(req.GroupID, req.UserName)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	log.NewInfo(req.OperationID, groupMembersCount)
	resp.MemberNums = int32(groupMembersCount)
	for _, groupMember := range groupMembers {
		member := open_im_sdk.GroupMemberFullInfo{}
		utils.CopyStructFields(&member, groupMember)
		member.JoinTime = int32(groupMember.JoinTime.Unix())
		member.MuteEndTime = uint32(groupMember.MuteEndTime.Unix())
		resp.Members = append(resp.Members, &member)
	}
	resp.Pagination = &open_im_sdk.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	return
}

func (s *groupServer) GetUserReqApplicationList(ctx context.Context, req *pbGroup.GetUserReqApplicationListReq) (resp *pbGroup.GetUserReqApplicationListResp, _ error) {
	resp = &pbGroup.GetUserReqApplicationListResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	groupRequests, err := imdb.GetUserReqGroupByUserID(req.UserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	for _, groupReq := range groupRequests {
		node := open_im_sdk.GroupRequest{UserInfo: &open_im_sdk.PublicUserInfo{}, GroupInfo: &open_im_sdk.GroupInfo{}}
		group, err := imdb.GetGroupInfoByGroupID(groupReq.GroupID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetGroupInfoByGroupID", err, "GroupID", groupReq.GroupID)
			continue
		}
		user, err := imdb.GetUserByUserID(groupReq.UserID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetUserByUserID", err, "UserID", groupReq.UserID)
			continue
		}
		cp.GroupRequestDBCopyOpenIM(&node, &groupReq)
		cp.UserDBCopyOpenIMPublicUser(node.UserInfo, user)
		cp.GroupDBCopyOpenIM(node.GroupInfo, group)
		resp.GroupRequestList = append(resp.GroupRequestList, &node)
	}
	return
}

func (s *groupServer) DismissGroup(ctx context.Context, req *pbGroup.DismissGroupReq) (resp *pbGroup.DismissGroupResp, _ error) {
	resp = &pbGroup.DismissGroupResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if !token_verify.IsManagerUserID(req.OpUserID) && !imdb.IsGroupOwnerAdmin(req.GroupID, req.OpUserID) {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}

	if err := rocksCache.DelGroupInfoFromCache(req.GroupID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if err := s.DelGroupAndUserCache(req.OperationID, req.GroupID, nil); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	err := imdb.OperateGroupStatus(req.GroupID, constant.GroupStatusDismissed)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if groupInfo.GroupType != constant.SuperGroup {
		memberList, err := imdb.GetGroupMemberListByGroupID(req.GroupID)
		if err != nil {
			trace_log.SetContextInfo(ctx, "GetGroupMemberListByGroupID", err, "groupID", req.GroupID)
		}
		//modify quitter conversation info
		var reqPb pbUser.SetConversationReq
		var c pbConversation.Conversation
		for _, v := range memberList {
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = v.UserID
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsNotInGroup = true
			reqPb.Conversation = &c
			etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if err != nil {
				SetErrorForResp(err, resp.CommonResp)
				return
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			trace_log.SetContextInfo(ctx, "SetConversation", err, "req", &reqPb, "resp", respPb)
		}
		err = imdb.DeleteGroupMemberByGroupID(req.GroupID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		chat.GroupDismissedNotification(req)
	} else {
		err = db.DB.DeleteSuperGroup(req.GroupID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
	}
	return
}

//  rpc MuteGroupMember(MuteGroupMemberReq) returns(MuteGroupMemberResp);
//  rpc CancelMuteGroupMember(CancelMuteGroupMemberReq) returns(CancelMuteGroupMemberResp);
//  rpc MuteGroup(MuteGroupReq) returns(MuteGroupResp);
//  rpc CancelMuteGroup(CancelMuteGroupReq) returns(CancelMuteGroupResp);

func (s *groupServer) MuteGroupMember(ctx context.Context, req *pbGroup.MuteGroupMemberReq) (resp *pbGroup.MuteGroupMemberResp, _ error) {
	resp = &pbGroup.MuteGroupMemberResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	opFlag, err := s.getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if opFlag == 0 {
		// TODO
		//errMsg := req.OperationID + "opFlag == 0  " + req.GroupID + req.OpUserID
		//log.Error(req.OperationID, errMsg)
		//return &pbGroup.MuteGroupMemberResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}

	mutedInfo, err := rocksCache.GetGroupMemberInfoFromCache(req.GroupID, req.UserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}
	if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
		SetErrorForResp(constant.ErrArgs, resp.CommonResp)
		return
	}

	if err := rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.UserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupMemberInfo := imdb.GroupMember{GroupID: req.GroupID, UserID: req.UserID}
	groupMemberInfo.MuteEndTime = time.Unix(int64(time.Now().Second())+int64(req.MutedSeconds), time.Now().UnixNano())
	err = imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	chat.GroupMemberMutedNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, req.MutedSeconds)
	return
}

func (s *groupServer) CancelMuteGroupMember(ctx context.Context, req *pbGroup.CancelMuteGroupMemberReq) (resp *pbGroup.CancelMuteGroupMemberResp, _ error) {
	resp = &pbGroup.CancelMuteGroupMemberResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()

	opFlag, err := s.getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if opFlag == 0 {
		SetErrorForResp(constant.ErrAccess, resp.CommonResp)
		return
	}

	mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.UserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if err := rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.UserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	groupMemberInfo := imdb.GroupMember{GroupID: req.GroupID, UserID: req.UserID}
	groupMemberInfo.MuteEndTime = time.Unix(0, 0)
	err = imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	chat.GroupMemberCancelMutedNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
	return
}

func (s *groupServer) MuteGroup(ctx context.Context, req *pbGroup.MuteGroupReq) (resp *pbGroup.MuteGroupResp, _ error) {
	resp = &pbGroup.MuteGroupResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	opFlag, err := s.getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if opFlag == 0 {
		//errMsg := req.OperationID + "opFlag == 0  " + req.GroupID + req.OpUserID
		//log.Error(req.OperationID, errMsg)
		//return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
		SetErrorForResp(constant.ErrAccess, resp.CommonResp)
		return
	}

	//mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.UserID)
	//if err != nil {
	//	errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID failed " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	//if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupOwner " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	//if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupAdmin " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	if err := rocksCache.DelGroupInfoFromCache(req.GroupID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	err = imdb.OperateGroupStatus(req.GroupID, constant.GroupStatusMuted)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	chat.GroupMutedNotification(req.OperationID, req.OpUserID, req.GroupID)
	return
}

func (s *groupServer) CancelMuteGroup(ctx context.Context, req *pbGroup.CancelMuteGroupReq) (resp *pbGroup.CancelMuteGroupResp, _ error) {
	resp = &pbGroup.CancelMuteGroupResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	opFlag, err := s.getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if opFlag == 0 {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	//mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.)
	//if err != nil {
	//	errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID failed " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	//if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupOwner " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	//if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupAdmin " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	log.Debug(req.OperationID, "UpdateGroupInfoDefaultZero ", req.GroupID, map[string]interface{}{"status": constant.GroupOk})
	if err := rocksCache.DelGroupInfoFromCache(req.GroupID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	err = imdb.UpdateGroupInfoDefaultZero(req.GroupID, map[string]interface{}{"status": constant.GroupOk})
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	chat.GroupCancelMutedNotification(req.OperationID, req.OpUserID, req.GroupID)
	return
}

func (s *groupServer) SetGroupMemberNickname(ctx context.Context, req *pbGroup.SetGroupMemberNicknameReq) (resp *pbGroup.SetGroupMemberNicknameResp, _ error) {
	resp = &pbGroup.SetGroupMemberNicknameResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if req.OpUserID != req.UserID && !token_verify.IsManagerUserID(req.OpUserID) {
		SetErrorForResp(constant.ErrIdentity, resp.CommonResp)
		return
	}
	cbReq := &pbGroup.SetGroupMemberInfoReq{
		GroupID:     req.GroupID,
		UserID:      req.UserID,
		OperationID: req.OperationID,
		OpUserID:    req.OpUserID,
		Nickname:    &wrapperspb.StringValue{Value: req.Nickname},
	}
	if err := CallbackBeforeSetGroupMemberInfo(ctx, cbReq); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	nickName := cbReq.Nickname.Value
	groupMemberInfo := imdb.GroupMember{}
	groupMemberInfo.UserID = req.UserID
	groupMemberInfo.GroupID = req.GroupID
	if nickName == "" {
		userNickname, err := imdb.GetUserNameByUserID(groupMemberInfo.UserID)
		if err != nil {
			SetErrorForResp(err, resp.CommonResp)
			return
		}
		groupMemberInfo.Nickname = userNickname
	} else {
		groupMemberInfo.Nickname = nickName
	}

	if err := rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.UserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}

	if err := imdb.UpdateGroupMemberInfo(groupMemberInfo); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
	return
}

func (s *groupServer) SetGroupMemberInfo(ctx context.Context, req *pbGroup.SetGroupMemberInfoReq) (resp *pbGroup.SetGroupMemberInfoResp, _ error) {
	resp = &pbGroup.SetGroupMemberInfoResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	if err := rocksCache.DelGroupMemberInfoFromCache(req.GroupID, req.UserID); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if err := CallbackBeforeSetGroupMemberInfo(ctx, req); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	groupMember := imdb.GroupMember{
		GroupID: req.GroupID,
		UserID:  req.UserID,
	}
	m := make(map[string]interface{})
	if req.RoleLevel != nil {
		m["role_level"] = req.RoleLevel.Value
	}
	if req.FaceURL != nil {
		m["user_group_face_url"] = req.FaceURL.Value
	}
	if req.Nickname != nil {
		m["nickname"] = req.Nickname.Value
	}
	if req.Ex != nil {
		m["ex"] = req.Ex.Value
	} else {
		m["ex"] = nil
	}
	if err := imdb.UpdateGroupMemberInfoByMap(groupMember, m); err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	if req.RoleLevel != nil {
		switch req.RoleLevel.Value {
		case constant.GroupOrdinaryUsers:
			//msg.GroupMemberRoleLevelChangeNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, constant.GroupMemberSetToOrdinaryUserNotification)
			chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
		case constant.GroupAdmin, constant.GroupOwner:
			//msg.GroupMemberRoleLevelChangeNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, constant.GroupMemberSetToAdminNotification)
			chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
		}
	} else {
		chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
	}
	return
}

func (s *groupServer) GetGroupAbstractInfo(ctx context.Context, req *pbGroup.GetGroupAbstractInfoReq) (resp *pbGroup.GetGroupAbstractInfoResp, _ error) {
	resp = &pbGroup.GetGroupAbstractInfoResp{CommonResp: &open_im_sdk.CommonResp{}}
	ctx = trace_log.NewRpcCtx(ctx, utils.GetSelfFuncName(), req.OperationID)
	defer func() {
		trace_log.SetContextInfo(ctx, utils.GetSelfFuncName(), nil, "rpc req ", req.String(), "rpc resp ", resp.String())
		trace_log.ShowLog(ctx)
	}()
	hashCode, err := rocksCache.GetGroupMemberListHashFromCache(req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	resp.GroupMemberListHash = hashCode
	num, err := rocksCache.GetGroupMemberNumFromCache(req.GroupID)
	if err != nil {
		SetErrorForResp(err, resp.CommonResp)
		return
	}
	resp.GroupMemberNumber = int32(num)
	return resp, nil
}

func (s *groupServer) DelGroupAndUserCache(operationID, groupID string, userIDList []string) error {
	if groupID != "" {
		etcdConn, err := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, operationID)
		if etcdConn == nil {
			errMsg := operationID + "getcdv3.GetDefaultConn == nil"
			log.NewError(operationID, errMsg)
			return errors.New("etcdConn is nil")
		}
		cacheClient := pbCache.NewCacheClient(etcdConn)
		cacheResp, err := cacheClient.DelGroupMemberIDListFromCache(context.Background(), &pbCache.DelGroupMemberIDListFromCacheReq{
			GroupID:     groupID,
			OperationID: operationID,
		})
		if err != nil {
			log.NewError(operationID, "DelGroupMemberIDListFromCache rpc call failed ", err.Error())
			return utils.Wrap(err, "")
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			log.NewError(operationID, "DelGroupMemberIDListFromCache rpc logic call failed ", cacheResp.String())
			return errors.New(fmt.Sprintf("errMsg is %s, errCode is %d", cacheResp.CommonResp.ErrMsg, cacheResp.CommonResp.ErrCode))
		}
		err = rocksCache.DelGroupMemberListHashFromCache(groupID)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), groupID, err.Error())
			return utils.Wrap(err, "")
		}
		err = rocksCache.DelGroupMemberNumFromCache(groupID)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error(), groupID)
			return utils.Wrap(err, "")
		}
	}
	if userIDList != nil {
		for _, userID := range userIDList {
			err := rocksCache.DelJoinedGroupIDListFromCache(userID)
			if err != nil {
				log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
				return utils.Wrap(err, "")
			}
		}
	}
	return nil
}

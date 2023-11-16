package club

import (
	"context"
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (c *clubServer) BanServerMember(ctx context.Context, req *pbclub.BanServerMemberReq) (*pbclub.BanServerMemberResp, error) {

	//todo 校验权限
	// if !permissions.PermissionsFromJSON().CanManageMember() {
	// 	return nil, errs.ErrNoPermission
	// }

	serverMember, err := c.ClubDatabase.FindServerMember(ctx, []string{req.ServerID}, req.BlockUserIDs, nil)
	if err != nil && !errs.ErrRecordNotFound.Is(err) {
		return nil, err
	}

	//需要踢出部落的members
	kickMembers := utils.Slice(serverMember, func(e *relationtb.ServerMemberModel) string { return e.UserID })

	blacks := []*relationtb.ServerBlackModel{}
	for _, blockUserID := range req.BlockUserIDs {
		black := &relationtb.ServerBlackModel{
			ServerID:       req.ServerID,
			BlockUserID:    blockUserID,
			Ex:             req.Ex,
			AddSource:      0,
			CreateTime:     time.Now(),
			OperatorUserID: mcontext.GetOpUserID(ctx),
		}
		blacks = append(blacks, black)
	}

	if err = c.ClubDatabase.CreateServerBlack(ctx, blacks, kickMembers, req.ServerID); err != nil {
		return nil, err
	}

	return &pbclub.BanServerMemberResp{}, nil
}

func (c *clubServer) CancelBanServerMember(ctx context.Context, req *pbclub.CancelBanServerMemberReq) (*pbclub.CancelBanServerMemberResp, error) {
	//todo 校验权限

	blacks := []*relationtb.ServerBlackModel{}
	for _, blockUserID := range req.BlockUserIDs {
		black := &relationtb.ServerBlackModel{
			ServerID:       req.ServerID,
			BlockUserID:    blockUserID,
			Ex:             req.Ex,
			AddSource:      0,
			CreateTime:     time.Now(),
			OperatorUserID: mcontext.GetOpUserID(ctx),
		}
		blacks = append(blacks, black)
	}
	if err := c.ClubDatabase.DeleteServerBlack(ctx, blacks); err != nil {
		return nil, err
	}
	return &pbclub.CancelBanServerMemberResp{}, nil
}

func (c *clubServer) GetServerBlackList(ctx context.Context, req *pbclub.GetServerBlackListReq) (*pbclub.GetServerBlackListResp, error) {
	resp := &pbclub.GetServerBlackListResp{}

	blacks, total, err := c.ClubDatabase.FindServerBlacks(ctx, req.ServerID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		return nil, err
	}
	respBlacks := utils.Batch(convert.DB2PbServerBlack, blacks)

	emptyUserIDs := make(map[string]struct{})
	for _, member := range blacks {
		emptyUserIDs[member.BlockUserID] = struct{}{}
	}
	if len(emptyUserIDs) > 0 {
		users, err := c.User.GetPublicUserInfoMap(ctx, utils.Keys(emptyUserIDs), true)
		if err != nil {
			return nil, err
		}
		for _, member := range respBlacks {
			user, ok := users[member.BlockUserID]
			if !ok {
				continue
			}
			if member.Nickname == "" {
				member.Nickname = user.Nickname
			}
			if member.FaceUrl == "" {
				member.FaceUrl = user.FaceURL
			}
		}
	}

	resp.Total = total
	resp.Blacks = respBlacks
	return resp, nil
}

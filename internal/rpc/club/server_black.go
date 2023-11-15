package club

import (
	"context"

	pbclub "github.com/OpenIMSDK/protocol/club"
)

func (c *clubServer) BanServerMember(ctx context.Context, req *pbclub.BanServerMemberReq) (*pbclub.BanServerMemberResp, error) {
	panic("unimplemented")
}

func (c *clubServer) CancelBanServerMember(ctx context.Context, req *pbclub.CancelBanServerMemberReq) (*pbclub.CancelBanServerMemberResp, error) {
	panic("unimplemented")
}

func (c *clubServer) GetServerBlackList(ctx context.Context, req *pbclub.GetServerBlackListReq) (*pbclub.GetServerBlackListResp, error) {
	panic("unimplemented")
}

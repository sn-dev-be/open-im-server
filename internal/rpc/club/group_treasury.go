package club

import (
	"context"
	"errors"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
	"gorm.io/gorm"
)

// GetGroupTreasure implements club.ClubGroup.
func (c *clubServer) GetGroupTreasure(ctx context.Context, req *pbclub.GetGroupTreasuryReq) (*pbclub.GetGroupTreasuryResp, error) {
	records, err := c.ClubDatabase.FindGroupTreasuryByGroupIDs(ctx, req.GroupIDs)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	result := utils.Batch(convert.Db2PbGroupTreasury, records)
	return &pbclub.GetGroupTreasuryResp{Records: result}, nil
}

// SetGroupTreasure implements club.ClubGroup.
func (c *clubServer) SetGroupTreasure(ctx context.Context, req *pbclub.SetGroupTreasuryReq) (*pbclub.SetGroupTreasuryResp, error) {
	if req.Info.GroupID == "" || req.Info.TreasuryID == "" {
		return nil, errs.ErrArgs.Wrap("no group id or treasury id")
	}

	record, err := c.ClubDatabase.FindGroupTreasuryByGroupIDs(ctx, []string{req.Info.GroupID})
	if err != nil {
		return nil, err
	}

	if record == nil || len(record) == 0 {
		treasury := &relationtb.GroupTreasuryModel{
			GroupID:              req.Info.GroupID,
			TreasuryID:           req.Info.TreasuryID,
			Icon:                 req.Info.Icon,
			Name:                 req.Info.Name,
			WalletType:           req.Info.WalletType,
			ContractAddress:      req.Info.ContractAddress,
			AdministratorAddress: req.Info.AdministratorAddress,
			Symbol:               req.Info.Symbol,
			Ex:                   req.Info.Ex,
		}
		err = c.ClubDatabase.CreateGroupTreasury(ctx, []*relationtb.GroupTreasuryModel{treasury})
		if err != nil {
			return nil, err
		}
		return &pbclub.SetGroupTreasuryResp{}, nil
	}

	data := UpdateGroupTreasuryMap(req)
	err = c.ClubDatabase.UpdateGroupTreasury(ctx, req.Info.GroupID, data)
	if err != nil {
		return nil, err
	}

	return &pbclub.SetGroupTreasuryResp{}, nil
}

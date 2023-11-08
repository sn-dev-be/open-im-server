package club

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) GenChannelCategoryID(ctx context.Context, categoryID *string) error {
	if *categoryID != "" {
		_, err := s.ClubDatabase.TakeChannelCategory(ctx, *categoryID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("category id existed " + *categoryID)
		} else if s.IsNotFound(err) {
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
		_, err := s.ClubDatabase.TakeChannelCategory(ctx, id)
		if err == nil {
			continue
		} else if s.IsNotFound(err) {
			*categoryID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("channel_category id gen error")
}

func (s *clubServer) CreateChannelCategory(ctx context.Context, categories []*relationtb.ChannelCategoryModel) error {
	if err := s.ClubDatabase.CreateChannelCategory(ctx, categories); err != nil {
		return err
	}
	return nil
}

func (s *clubServer) CreateChannelCategoryByDefault(ctx context.Context, serverID, categoryName string, categoryType, reorderWeight int32) (string, error) {
	category := &relationtb.ChannelCategoryModel{
		CategoryName:  categoryName,
		ReorderWeight: reorderWeight,
		ViewMode:      1,
		CategoryType:  categoryType,
		ServerID:      serverID,
		Ex:            "",
		CreateTime:    time.Now(),
	}
	if err := s.GenChannelCategoryID(ctx, &category.CategoryID); err != nil {
		return "", err
	}
	if err := s.ClubDatabase.CreateChannelCategory(ctx, []*relationtb.ChannelCategoryModel{category}); err != nil {
		return "", err
	}
	return category.CategoryID, nil
}

func (s *clubServer) CreateChannelCategoryByPb(ctx context.Context, server_roles []*relationtb.ServerRoleModel) error {
	if err := s.ClubDatabase.CreateServerRole(ctx, server_roles); err != nil {
		return err
	}
	return nil
}

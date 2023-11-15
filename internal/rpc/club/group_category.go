package club

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/OpenIMSDK/protocol/constant"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/authverify"
	"github.com/openimsdk/open-im-server/v3/pkg/common/convert"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (c *clubServer) CreateGroupCategory(ctx context.Context, req *pbclub.CreateGroupCategoryReq) (*pbclub.CreateGroupCategoryResp, error) {
	if err := authverify.CheckAccessV3(ctx, req.OwnerUserID); err != nil {
		return nil, err
	}

	category := &relationtb.GroupCategoryModel{
		CategoryName: req.CategoryName,
		CategoryType: constant.CustomCategoryType,
		ServerID:     req.ServerID,
		CreateTime:   time.Now(),
	}
	c.GenGroupCategoryID(ctx, &category.CategoryID)

	if err := c.ClubDatabase.CreateGroupCategory(ctx, []*relationtb.GroupCategoryModel{category}); err != nil {
		return nil, err
	}
	gc, err := convert.DB2PbCategory(category)
	if err != nil {
		return nil, err
	}
	return &pbclub.CreateGroupCategoryResp{GroupCategory: gc}, nil
}

func (c *clubServer) GenGroupCategoryID(ctx context.Context, categoryID *string) error {
	if *categoryID != "" {
		_, err := c.ClubDatabase.TakeGroupCategory(ctx, *categoryID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("category id existed " + *categoryID)
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
		_, err := c.ClubDatabase.TakeGroupCategory(ctx, id)
		if err == nil {
			continue
		} else if c.IsNotFound(err) {
			*categoryID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("group_category id gen error")
}

func (c *clubServer) createGroupCategoryByDefault(ctx context.Context, serverID, categoryName string, categoryType, reorderWeight int32) (string, error) {
	category := &relationtb.GroupCategoryModel{
		CategoryName:  categoryName,
		ReorderWeight: reorderWeight,
		ViewMode:      1,
		CategoryType:  categoryType,
		ServerID:      serverID,
		Ex:            "",
		CreateTime:    time.Now(),
	}
	if err := c.GenGroupCategoryID(ctx, &category.CategoryID); err != nil {
		return "", err
	}
	if err := c.ClubDatabase.CreateGroupCategory(ctx, []*relationtb.GroupCategoryModel{category}); err != nil {
		return "", err
	}
	return category.CategoryID, nil
}

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

	if !c.checkManageGroup(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	category := &relationtb.GroupCategoryModel{
		CategoryName:  req.CategoryName,
		CategoryType:  constant.CustomCategoryType,
		ServerID:      req.ServerID,
		ReorderWeight: 1,
		CreateTime:    time.Now(),
	}
	c.GenGroupCategoryID(ctx, &category.CategoryID)

	if err := c.ClubDatabase.CreateGroupCategory(ctx, []*relationtb.GroupCategoryModel{category}); err != nil {
		return nil, err
	}
	gc := convert.Db2PbGroupCategory(category)
	return &pbclub.CreateGroupCategoryResp{GroupCategory: gc}, nil
}

func (c *clubServer) DeleteGroupCategory(ctx context.Context, req *pbclub.DeleteGroupCategoryReq) (*pbclub.DeleteGroupCategoryResp, error) {
	if len(req.CategoryIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("categoryID is empty")
	}

	if !c.checkManageGroup(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	resp := &pbclub.DeleteGroupCategoryResp{}
	groupCategorys, err := c.ClubDatabase.FindGroupCategory(ctx, req.CategoryIDs)
	if err != nil {
		return nil, err
	}

	categoryIDs := utils.Slice(groupCategorys, func(e *relationtb.GroupCategoryModel) string { return e.CategoryID })

	for _, groupCategory := range groupCategorys {
		if groupCategory.ServerID != req.ServerID {
			return nil, errs.ErrArgs.Wrap("serverID and categoryID not match")
		}
		if groupCategory.CategoryType == constant.DefaultCategoryType {
			return nil, errs.ErrArgs.Wrap("default category cannot edit")
		}
	}
	err = c.ClubDatabase.DeleteGroupCategorys(ctx, req.ServerID, categoryIDs)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clubServer) SetGroupCategoryOrder(ctx context.Context, req *pbclub.SetGroupCategoryOrderReq) (*pbclub.SetGroupCategoryOrderResp, error) {
	if len(req.CategoryIDs) == 0 {
		return nil, errs.ErrArgs.Wrap("categoryID is empty")
	}

	if !c.checkManageGroup(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	resp := &pbclub.SetGroupCategoryOrderResp{}
	groupCategorys, err := c.ClubDatabase.FindGroupCategory(ctx, req.CategoryIDs)
	if err != nil {
		return nil, err
	}

	for i, groupCategory := range groupCategorys {
		if groupCategory.ServerID != req.ServerID {
			return nil, errs.ErrArgs.Wrap("serverID and categoryID not match")
		}
		if groupCategory.CategoryType != constant.DefaultCategoryType {
			data := UpdateGroupCategoryInfoMap(ctx, "", int32(i+1))
			err := c.ClubDatabase.UpdateGroupCategory(ctx, req.ServerID, groupCategory.CategoryID, data)
			if err != nil {
				return nil, err
			}
		}
	}
	return resp, nil
}

func (c *clubServer) SetGroupCategoryInfo(ctx context.Context, req *pbclub.SetGroupCategoryInfoReq) (*pbclub.SetGroupCategoryInfoResp, error) {

	if !c.checkManageGroup(ctx, req.ServerID) {
		return nil, errs.ErrNoPermission
	}

	groupCategory, err := c.ClubDatabase.TakeGroupCategory(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if groupCategory.ServerID != req.ServerID {
		return nil, errs.ErrArgs.Wrap("serverID and categoryID not match")
	}
	if groupCategory.CategoryType == constant.DefaultCategoryType {
		return nil, errs.ErrArgs.Wrap("default category cannot edit")
	}

	resp := &pbclub.SetGroupCategoryInfoResp{}

	data := UpdateGroupCategoryInfoMap(ctx, req.CategoryName, 0)
	if len(data) == 0 {
		return resp, nil
	}

	err = c.ClubDatabase.UpdateGroupCategory(ctx, req.ServerID, req.CategoryID, data)
	if err != nil {
		return nil, err
	}
	groupCategory.CategoryName = req.CategoryName
	resp.GroupCategory = convert.Db2PbGroupCategory(groupCategory)
	return resp, nil
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

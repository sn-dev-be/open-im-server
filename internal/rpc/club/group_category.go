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
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) CreateGroupCategory(ctx context.Context, req *pbclub.CreateGroupCategoryReq) (*pbclub.CreateGroupCategoryResp, error) {
	serverID := req.ServerID
	ownerUserID := req.OwnerUserID
	opUserID := mcontext.GetOpUserID(ctx)
	if opUserID != ownerUserID {
		return nil, errs.ErrNoPermission
	}

	//校验当前用户是否具有分组管理权限
	if serverRole, err := s.ClubDatabase.GetServerRoleByUserIDAndServerID(ctx, opUserID, serverID); err != nil {
		return nil, errs.ErrNoPermission
	} else {
		if !serverRole.AllowManageGroupCategory() {
			return nil, errs.ErrNoPermission
		}
	}

	category := &relationtb.GroupCategoryModel{
		CategoryName: req.CategoryName,
		CategoryType: constant.CustomCategoryType,
		ServerID:     req.ServerID,
		CreateTime:   time.Now(),
	}
	s.GenGroupCategoryID(ctx, &category.CategoryID)

	if err := s.createGroupCategory(ctx, []*relationtb.GroupCategoryModel{category}); err != nil {
		return nil, err
	}
	return &pbclub.CreateGroupCategoryResp{CategoryID: category.CategoryID}, nil
}

func (s *clubServer) GenGroupCategoryID(ctx context.Context, categoryID *string) error {
	if *categoryID != "" {
		_, err := s.ClubDatabase.TakeGroupCategory(ctx, *categoryID)
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
		_, err := s.ClubDatabase.TakeGroupCategory(ctx, id)
		if err == nil {
			continue
		} else if s.IsNotFound(err) {
			*categoryID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("group_category id gen error")
}

func (s *clubServer) createGroupCategory(ctx context.Context, categories []*relationtb.GroupCategoryModel) error {
	if err := s.ClubDatabase.CreateGroupCategory(ctx, categories); err != nil {
		return err
	}
	return nil
}

func (s *clubServer) createGroupCategoryByDefault(ctx context.Context, serverID, categoryName string, categoryType, reorderWeight int32) (string, error) {
	category := &relationtb.GroupCategoryModel{
		CategoryName:  categoryName,
		ReorderWeight: reorderWeight,
		ViewMode:      1,
		CategoryType:  categoryType,
		ServerID:      serverID,
		Ex:            "",
		CreateTime:    time.Now(),
	}
	if err := s.GenGroupCategoryID(ctx, &category.CategoryID); err != nil {
		return "", err
	}
	if err := s.ClubDatabase.CreateGroupCategory(ctx, []*relationtb.GroupCategoryModel{category}); err != nil {
		return "", err
	}
	return category.CategoryID, nil
}

func (s *clubServer) CreateGroupCategoryByPb(ctx context.Context, server_roles []*relationtb.ServerRoleModel) error {
	if err := s.ClubDatabase.CreateServerRole(ctx, server_roles); err != nil {
		return err
	}
	return nil
}

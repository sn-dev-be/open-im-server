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

func (s *clubServer) GenChannelID(ctx context.Context, channelID *string) error {
	if *channelID != "" {
		_, err := s.ClubDatabase.TakeChannel(ctx, *channelID)
		if err == nil {
			return errs.ErrGroupIDExisted.Wrap("channel id existed " + *channelID)
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
		_, err := s.ClubDatabase.TakeChannel(ctx, id)
		if err == nil {
			continue
		} else if s.IsNotFound(err) {
			*channelID = id
			return nil
		} else {
			return err
		}
	}
	return errs.ErrData.Wrap("channel id gen error")
}

func (s *clubServer) CreateChannel(ctx context.Context, channels []*relationtb.ChannelModel) error {
	if err := s.ClubDatabase.CreateChannel(ctx, channels); err != nil {
		return err
	}
	return nil
}

func (s *clubServer) CreateChannelByDefault(ctx context.Context, serverID, categoryID, channelName, ownerUserID string, channelType, reorderWeight int32) (string, error) {
	channel := &relationtb.ChannelModel{
		CategoryID:    categoryID,
		ServerID:      serverID,
		ChannelName:   channelName,
		Icon:          "",
		Description:   "",
		OwnerUserID:   ownerUserID,
		JoinCondition: 0,
		ConditionType: 0,
		ChannelType:   channelType,
		ReorderWeight: reorderWeight,
		VisitorMode:   0,
		ViewMode:      0,
		Ex:            "",
		CreateTime:    time.Now(),
	}
	if err := s.GenChannelID(ctx, &channel.ChannelID); err != nil {
		return "", err
	}
	if err := s.ClubDatabase.CreateChannel(ctx, []*relationtb.ChannelModel{channel}); err != nil {
		return "", err
	}

	return channel.ChannelID, nil
}

func (s *clubServer) CreateChannelByPb(ctx context.Context, server_roles []*relationtb.ServerRoleModel) error {
	if err := s.ClubDatabase.CreateServerRole(ctx, server_roles); err != nil {
		return err
	}
	return nil
}

package club

import (
	"context"
	"time"

	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) CreateChannelMembser(ctx context.Context, serverID string, channelIDs []*string, serverMemberID uint64) error {
	channelMembers := []*relationtb.ChannelMemberModel{}
	for _, channelID := range channelIDs {
		channel_member := &relationtb.ChannelMemberModel{
			ServerID:       serverID,
			ChannelID:      *channelID,
			ServerMemberID: serverMemberID,
			Ex:             "",
			CreateTime:     time.Now(),
		}
		channelMembers = append(channelMembers, channel_member)
	}

	if err := s.ClubDatabase.CreateChannelMember(ctx, channelMembers); err != nil {
		return err
	}

	return nil
}

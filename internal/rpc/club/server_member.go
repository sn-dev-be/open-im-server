package club

import (
	"context"
	"time"

	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (s *clubServer) createServerMember(ctx context.Context, serverID, user_id, nickname, serverRoleID, invitedUserID, ex string, roleLevel, joinSource int32) (uint64, error) {
	server_member := &relationtb.ServerMemberModel{
		ServerID:      serverID,
		UserID:        user_id,
		Nickname:      nickname,
		ServerRoleID:  serverRoleID,
		RoleLevel:     roleLevel,
		JoinSource:    joinSource,
		InviterUserID: invitedUserID,
		Ex:            ex,
		JoinTime:      time.Now(),
	}
	if err := s.ClubDatabase.CreateServerMember(ctx, []*relationtb.ServerMemberModel{server_member}); err != nil {
		return 0, err
	}
	return server_member.ID, nil
}

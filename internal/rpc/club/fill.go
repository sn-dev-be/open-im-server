package club

import (
	"context"

	"github.com/OpenIMSDK/tools/utils"

	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func (c *clubServer) FindServerMember(ctx context.Context, serverIDs []string, userIDs []string, roleLevels []int32) ([]*relationtb.ServerMemberModel, error) {
	members, err := c.ClubDatabase.FindServerMember(ctx, serverIDs, userIDs, roleLevels)
	if err != nil {
		return nil, err
	}
	emptyUserIDs := make(map[string]struct{})
	for _, member := range members {
		if member.Nickname == "" || member.FaceURL == "" {
			emptyUserIDs[member.UserID] = struct{}{}
		}
	}
	if len(emptyUserIDs) > 0 {
		users, err := c.User.GetPublicUserInfoMap(ctx, utils.Keys(emptyUserIDs), true)
		if err != nil {
			return nil, err
		}
		for i, member := range members {
			user, ok := users[member.UserID]
			if !ok {
				continue
			}
			if member.Nickname == "" {
				members[i].Nickname = user.Nickname
			}
			if member.FaceURL == "" {
				members[i].FaceURL = user.FaceURL
			}
		}
	}
	return members, nil
}

func (c *clubServer) TakeServerMember(
	ctx context.Context,
	serverID string,
	userID string,
) (*relationtb.ServerMemberModel, error) {
	member, err := c.ClubDatabase.TakeServerMember(ctx, serverID, userID)
	if err != nil {
		return nil, err
	}
	if member.Nickname == "" || member.FaceURL == "" {
		user, err := c.User.GetPublicUserInfo(ctx, userID)
		if err != nil {
			return nil, err
		}
		if member.Nickname == "" {
			member.Nickname = user.Nickname
		}
		if member.FaceURL == "" {
			member.FaceURL = user.FaceURL
		}
	}
	return member, nil
}

func (c *clubServer) TakeServerOwner(ctx context.Context, serverID string) (*relationtb.ServerMemberModel, error) {
	owner, err := c.ClubDatabase.TakeServerOwner(ctx, serverID)
	if err != nil {
		return nil, err
	}
	if owner.Nickname == "" || owner.FaceURL == "" {
		user, err := c.User.GetUserInfo(ctx, owner.UserID)
		if err != nil {
			return nil, err
		}
		if owner.Nickname == "" {
			owner.Nickname = user.Nickname
		}
		if owner.FaceURL == "" {
			owner.FaceURL = user.FaceURL
		}
	}
	return owner, nil
}

func (c *clubServer) PageGetServerMember(
	ctx context.Context,
	serverID string,
	pageNumber, showNumber int32,
) (uint32, []*relationtb.ServerMemberModel, error) {
	total, members, err := c.ClubDatabase.PageGetServerMember(ctx, serverID, pageNumber, showNumber)
	if err != nil {
		return 0, nil, err
	}
	emptyUserIDs := make(map[string]struct{})
	for _, member := range members {
		if member.Nickname == "" || member.FaceURL == "" {
			emptyUserIDs[member.UserID] = struct{}{}
		}
	}
	if len(emptyUserIDs) > 0 {
		users, err := c.User.GetPublicUserInfoMap(ctx, utils.Keys(emptyUserIDs), true)
		if err != nil {
			return 0, nil, err
		}
		for i, member := range members {
			user, ok := users[member.UserID]
			if !ok {
				continue
			}
			if member.Nickname == "" {
				members[i].Nickname = user.Nickname
			}
			if member.FaceURL == "" {
				members[i].FaceURL = user.FaceURL
			}
		}
	}
	return total, members, nil
}

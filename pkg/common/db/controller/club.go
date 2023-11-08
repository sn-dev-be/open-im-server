// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"

	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"github.com/OpenIMSDK/tools/tx"

	"github.com/openimsdk/open-im-server/v3/pkg/common/db/relation"
	relationtb "github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

type ClubDatabase interface {
	// Server
	CreateServer(ctx context.Context, servers []*relationtb.ServerModel) error
	TakeServer(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error)
	PageServers(ctx context.Context, pageNumber, showNumber int32) (servers []*relationtb.ServerModel, total int64, err error)

	//server_role
	TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error)
	CreateServerRole(ctx context.Context, serverRoles []*relationtb.ServerRoleModel) error

	//server_request

	//server_black

	//channel_category
	TakeChannelCategory(ctx context.Context, channelCategoryID string) (channelCategory *relationtb.ChannelCategoryModel, err error)
	CreateChannelCategory(ctx context.Context, categories []*relationtb.ChannelCategoryModel) error
	GetAllChannelCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.ChannelCategoryModel, error)

	//channel
	TakeChannel(ctx context.Context, channelID string) (channel *relationtb.ChannelModel, err error)
	CreateChannel(ctx context.Context, channels []*relationtb.ChannelModel) error
	GetAllChannelByCategory(ctx context.Context, categoryID string) ([]*relationtb.ChannelModel, error)

	//channel_member
	PageChannelMembers(ctx context.Context, pageNumber, showNumber int32, serverID string) (members []*relationtb.ChannelMemberModel, total int64, err error)
	CreateChannelMember(ctx context.Context, channelMembers []*relationtb.ChannelMemberModel) error

	//server_member
	PageServerMembers(ctx context.Context, pageNumber, showNumber int32, serverID string) (members []*relationtb.ServerMemberModel, total int64, err error)
	GetServerMembers(ctx context.Context, ids []uint64, serverID string) (members []*relationtb.ServerMemberModel, err error)
	CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error
}

func NewClubDatabase(
	server relationtb.ServerModelInterface,
	server_member relationtb.ServerMemberModelInterface,
	channel_category relationtb.ChannelCategoryModelInterface,
	channel relationtb.ChannelModelInterface,
	channelMember relationtb.ChannelMemberModelInterface,
	serverRole relationtb.ServerRoleModelInterface,
	serverRequest relationtb.ServerRequestModelInterface,
	serverBlack relationtb.ServerBlackModelInterface,
	channelDapp relationtb.ChannelDappModellInterface,
	tx tx.Tx,
	ctxTx tx.CtxTx,
) ClubDatabase {
	database := &clubDatabase{
		serverDB:          server,
		serverMemberDB:    server_member,
		serverRoleDB:      serverRole,
		serverRequestDB:   serverRequest,
		serverBlackDB:     serverBlack,
		channelCategoryDB: channel_category,
		channelDappDB:     channelDapp,
		channelDB:         channel,
		tx:                tx,
		ctxTx:             ctxTx,
	}
	return database
}

func InitClubDatabase(db *gorm.DB, rdb redis.UniversalClient, database *mongo.Database) ClubDatabase {
	rcOptions := rockscache.NewDefaultOptions()
	rcOptions.StrongConsistency = true
	rcOptions.RandomExpireAdjustment = 0.2
	return NewClubDatabase(
		relation.NewServerDB(db),
		relation.NewServerMemberDB(db),
		relation.NewChannelCategoryDB(db),
		relation.NewChannelDB(db),
		relation.NewChannelMemberDB(db),
		relation.NewServerRoleDB(db),
		relation.NewServerRequestDB(db),
		relation.NewServerBlackDB(db),
		relation.NewChannelDappDB(db),
		tx.NewGorm(db),
		tx.NewMongo(database.Client()),
	)
}

type clubDatabase struct {
	serverDB          relationtb.ServerModelInterface
	serverMemberDB    relationtb.ServerMemberModelInterface
	channelCategoryDB relationtb.ChannelCategoryModelInterface
	channelDB         relationtb.ChannelModelInterface
	channelMemberDB   relationtb.ChannelMemberModelInterface
	serverRoleDB      relationtb.ServerRoleModelInterface
	serverRequestDB   relationtb.ServerRequestModelInterface
	serverBlackDB     relationtb.ServerBlackModelInterface
	channelDappDB     relationtb.ChannelDappModellInterface
	tx                tx.Tx
	ctxTx             tx.CtxTx
}

// CreateServerMember implements ClubDatabase.
func (c *clubDatabase) CreateServerMember(ctx context.Context, serverMembers []*relationtb.ServerMemberModel) error {
	return c.serverMemberDB.Create(ctx, serverMembers)
}

// GetServerMembers implements ClubDatabase.
func (c *clubDatabase) GetServerMembers(ctx context.Context, ids []uint64, serverID string) (members []*relationtb.ServerMemberModel, err error) {
	return c.serverMemberDB.GetServerMembers(ctx, ids, serverID)
}

// PageServerMembers implements ClubDatabase.
func (c *clubDatabase) PageServerMembers(ctx context.Context, pageNumber int32, showNumber int32, serverID string) (members []*relationtb.ServerMemberModel, total int64, err error) {
	return c.serverMemberDB.PageServerMembers(ctx, showNumber, pageNumber, serverID)
}

// CreateChannelMember implements ClubDatabase.
func (c *clubDatabase) CreateChannelMember(ctx context.Context, channelMembers []*relationtb.ChannelMemberModel) error {
	return c.channelMemberDB.Create(ctx, channelMembers)
}

// PageChannelMembers implements ClubDatabase.
func (c *clubDatabase) PageChannelMembers(ctx context.Context, pageNumber int32, showNumber int32, serverID string) (members []*relationtb.ChannelMemberModel, total int64, err error) {
	return c.channelMemberDB.PageChannelMembers(ctx, pageNumber, showNumber, serverID)
}

// GetAllChannelByCategory implements ClubDatabase.
func (c *clubDatabase) GetAllChannelByCategory(ctx context.Context, categoryID string) ([]*relationtb.ChannelModel, error) {
	return nil, nil
}

// GetAllChannelCategoriesByServer implements ClubDatabase.
func (c *clubDatabase) GetAllChannelCategoriesByServer(ctx context.Context, serverID string) ([]*relationtb.ChannelCategoryModel, error) {
	panic("unimplemented")
}

// PageServers implements ClubDatabase.
func (c *clubDatabase) PageServers(ctx context.Context, pageNumber int32, showNumber int32) (servers []*relationtb.ServerModel, total int64, err error) {
	return c.serverDB.FindServersSplit(ctx, pageNumber, showNumber)
}

func (c *clubDatabase) CreateServer(
	ctx context.Context,
	servers []*relationtb.ServerModel,
) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverDB.NewTx(tx).Create(ctx, servers); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (c *clubDatabase) TakeServer(ctx context.Context, serverID string) (server *relationtb.ServerModel, err error) {
	return c.serverDB.Take(ctx, serverID)
}

func (c *clubDatabase) TakeServerRole(ctx context.Context, serverRoleID string) (serverRole *relationtb.ServerRoleModel, err error) {
	return c.serverRoleDB.Take(ctx, serverRoleID)
}

func (c *clubDatabase) TakeChannelCategory(ctx context.Context, channelCategoryID string) (channelCategory *relationtb.ChannelCategoryModel, err error) {
	return c.channelCategoryDB.Take(ctx, channelCategoryID)
}

func (c *clubDatabase) TakeChannel(ctx context.Context, channelID string) (channel *relationtb.ChannelModel, err error) {
	return c.channelDB.Take(ctx, channelID)
}

// CreateChannel implements ClubDatabase.
func (c *clubDatabase) CreateChannel(ctx context.Context, channels []*relationtb.ChannelModel) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.channelDB.NewTx(tx).Create(ctx, channels); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// CreateChannelCategory implements ClubDatabase.
func (c *clubDatabase) CreateChannelCategory(ctx context.Context, categories []*relationtb.ChannelCategoryModel) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.channelCategoryDB.NewTx(tx).Create(ctx, categories); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// CreateServerRole implements ClubDatabase.
func (c *clubDatabase) CreateServerRole(ctx context.Context, serverRoles []*relationtb.ServerRoleModel) error {
	if err := c.tx.Transaction(func(tx any) error {
		if err := c.serverRoleDB.NewTx(tx).Create(ctx, serverRoles); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

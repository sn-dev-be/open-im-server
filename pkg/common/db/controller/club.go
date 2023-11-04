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
}

func NewClubDatabase(
	server relationtb.ServerModelInterface,
	tx tx.Tx,
	ctxTx tx.CtxTx,
) ClubDatabase {
	database := &clubDatabase{
		serverDB: server,
		tx:       tx,
		ctxTx:    ctxTx,
	}
	return database
}

func InitClubDatabase(db *gorm.DB, rdb redis.UniversalClient, database *mongo.Database) ClubDatabase {
	rcOptions := rockscache.NewDefaultOptions()
	rcOptions.StrongConsistency = true
	rcOptions.RandomExpireAdjustment = 0.2
	return NewClubDatabase(
		relation.NewServerDB(db),
		tx.NewGorm(db),
		tx.NewMongo(database.Client()),
	)
}

type clubDatabase struct {
	serverDB relationtb.ServerModelInterface
	tx       tx.Tx
	ctxTx    tx.CtxTx
}

func (c *clubDatabase) CreateServer(
	ctx context.Context,
	servers []*relationtb.ServerModel,
) error {
	return nil
}

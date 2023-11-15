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

package relation

import (
	"context"
	"time"
)

const ServerBlackModelTableName = "server_blacks"

type ServerBlackModel struct {
	ServerID       string    `gorm:"column:server_id;primary_key;size:64"`
	BlockUserID    string    `gorm:"column:block_user_id;primary_key;size:64"`
	CreateTime     time.Time `gorm:"column:create_time"`
	AddSource      int32     `gorm:"column:add_source"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	Ex             string    `gorm:"column:ex;size:1024"`
}

func (ServerBlackModel) TableName() string {
	return ServerBlackModelTableName
}

type ServerBlackModelInterface interface {
	Create(ctx context.Context, serverBlacks []*ServerBlackModel) (err error)
	NewTx(tx any) ServerBlackModelInterface
	Delete(ctx context.Context, blacks []*ServerBlackModel) (err error)
	UpdateByMap(ctx context.Context, serverID, blockUserID string, args map[string]interface{}) (err error)
	Update(ctx context.Context, blacks []*ServerBlackModel) (err error)
	Find(ctx context.Context, blacks []*ServerBlackModel) (blackList []*ServerBlackModel, err error)
	Take(ctx context.Context, serverID, blockUserID string) (black *ServerBlackModel, err error)
	FindServerBlackInfos(ctx context.Context, serverID string) (blacks []*ServerBlackModel, err error)
	FindBlackUserIDs(ctx context.Context, serverID string) (blackUserIDs []string, err error)
}

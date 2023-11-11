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

const ServerRequestModelTableName = "server_requests"

type ServerRequestModel struct {
	FromUserID    string    `gorm:"column:from_user_id;primary_key;size:64"`
	ServerID      string    `gorm:"column:server_id;primary_key;size:64"`
	HandleResult  int32     `gorm:"column:handle_result"`
	ReqMsg        string    `gorm:"column:req_msg;size:255"`
	CreateTime    time.Time `gorm:"column:create_time; autoCreateTime"`
	HandlerUserID string    `gorm:"column:handler_user_id;size:64"`
	HandleMsg     string    `gorm:"column:handle_msg;size:255"`
	HandleTime    time.Time `gorm:"column:handle_time"`
	Ex            string    `gorm:"column:ex;size:1024"`
}

func (ServerRequestModel) TableName() string {
	return ServerRequestModelTableName
}

type ServerRequestModelInterface interface {
	// 插入多条记录
	Create(ctx context.Context, serverRequests []*ServerRequestModel) (err error)
	NewTx(tx any) ServerRequestModelInterface
	UpdateHandler(ctx context.Context, serverID string, userID string, handledMsg string, handleResult int32) (err error)

	PageServer(
		ctx context.Context,
		serverIDs []string,
		pageNumber, showNumber int32,
	) (total uint32, servers []*ServerRequestModel, err error)
}

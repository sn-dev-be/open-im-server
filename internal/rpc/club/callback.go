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

package club

import (
	"context"

	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/utils"

	cbapi "github.com/openimsdk/open-im-server/v3/pkg/callbackstruct"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/http"
)

const (
	RemarkServerMemberURI = "/openim-callback/club-server-user"
	DeleteServerMemberURI = "/openim-callback/club-server-user/delete"
)

func CallbackAfterRemarkServerMember(ctx context.Context, serverID, userID, nickname string) error {
	remarkServerMemberUri := config.Config.Callback.CallbackZapBusinessUrl + RemarkServerMemberURI

	if !config.Config.Callback.CallbackAfterRemarkServerMember.Enable {
		return nil
	}

	ClubServerUser := &cbapi.ClubServerUserStruct{
		ServerID: serverID,
		UserID:   userID,
		Nickname: nickname,
	}
	cbReq := &cbapi.CallbackAfterRemarkServerMemberReq{
		ClubServerUser: *ClubServerUser,
	}

	// 启动goroutine，异步执行http.Post
	// go func() {
	// 	defer func() {
	// 		if err := recover(); err != nil {
	// 			// 处理 panic（如果有）
	// 			log.ZError(ctx, "Recovered from panic in goroutine:", err.(error))
	// 		}
	// 	}()

	// 	if _, err := http.Post(ctx, remarkServerMemberUri, nil, cbReq, config.Config.Callback.CallbackAfterRemarkServerMember.CallbackTimeOut); err != nil {
	// 		log.ZInfo(ctx, "CallbackAfterRemarkServerMember", utils.Unwrap(err))
	// 	}
	// }()

	if _, err := http.Post(ctx, remarkServerMemberUri, nil, cbReq, config.Config.Callback.CallbackAfterRemarkServerMember.CallbackTimeOut); err != nil {
		log.ZInfo(ctx, "CallbackAfterRemarkServerMember", utils.Unwrap(err))
	}

	return nil
}

func CallbackAfterQuitServer(ctx context.Context, serverID, userID, nickname string) error {
	deleteServerMemberURI := config.Config.Callback.CallbackZapBusinessUrl + DeleteServerMemberURI

	if !config.Config.Callback.CallbackAfterQuitServer.Enable {
		return nil
	}

	ClubServerUser := &cbapi.ClubServerUserStruct{
		ServerID: serverID,
		UserID:   userID,
		Nickname: nickname,
	}
	cbReq := &cbapi.CallbackQuitServerReq{
		ClubServerUser: *ClubServerUser,
	}

	// go func() {
	// 	defer func() {
	// 		if err := recover(); err != nil {
	// 			// 处理 panic（如果有）
	// 			log.ZError(ctx, "Recovered from panic in goroutine:", err.(error))
	// 		}
	// 	}()

	// 	if _, err := http.Post(ctx, deleteServerMemberURI, nil, cbReq, config.Config.Callback.CallbackAfterQuitServer.CallbackTimeOut); err != nil {
	// 		log.ZError(ctx, "CallbackAfterQuitServer", utils.Unwrap(err))
	// 	}
	// }()

	if _, err := http.Post(ctx, deleteServerMemberURI, nil, cbReq, config.Config.Callback.CallbackAfterQuitServer.CallbackTimeOut); err != nil {
		log.ZError(ctx, "CallbackAfterQuitServer", utils.Unwrap(err))
	}

	return nil
}

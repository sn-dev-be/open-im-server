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

package friend

import (
	"context"

	"github.com/OpenIMSDK/protocol/constant"
	pbfriend "github.com/OpenIMSDK/protocol/friend"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	cbapi "github.com/openimsdk/open-im-server/v3/pkg/callbackstruct"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/http"
)

const (
	AddFriendURI    = "/openim-callback/user-relation"
	DeleteFriendURI = "/openim-callback/user-relation/delete"
)

func CallbackBeforeAddFriend(ctx context.Context, req *pbfriend.ApplyToAddFriendReq) error {
	if !config.Config.Callback.CallbackBeforeAddFriend.Enable {
		return nil
	}
	cbReq := &cbapi.CallbackBeforeAddFriendReq{
		CallbackCommand: constant.CallbackBeforeAddFriendCommand,
		FromUserID:      req.FromUserID,
		ToUserID:        req.ToUserID,
		ReqMsg:          req.ReqMsg,
		OperationID:     mcontext.GetOperationID(ctx),
	}
	resp := &cbapi.CallbackBeforeAddFriendResp{}
	if err := http.CallBackPostReturn(ctx, config.Config.Callback.CallbackUrl, cbReq, resp, config.Config.Callback.CallbackBeforeAddFriend); err != nil {
		if err == errs.ErrCallbackContinue {
			return nil
		}
		return err
	}
	return nil
}

func CallbackAfterAddFriend(ctx context.Context, fromUserID, targetUserID, remark string) error {
	addFriendUri := config.Config.Callback.CallbackZapBusinessUrl + AddFriendURI

	if !config.Config.Callback.CallbackAfterChangefriendRelation.Enable {
		return nil
	}
	userRelation := &cbapi.UserRelationStruct{
		UserID:       fromUserID,
		TargetUserID: targetUserID,
		Remark:       remark,
	}
	cbReq := &cbapi.CallbackAfterFriendRelationChangedReq{
		UserRelation: *userRelation,
	}
	// 启动goroutine，异步执行http.Post
	// go func() {
	// 	defer func() {
	// 		if err := recover(); err != nil {
	// 			// 处理 panic（如果有）
	// 			log.ZError(ctx, "Recovered from panic in goroutine:", err.(error))
	// 		}
	// 	}()

	// 	if _, err := http.Post(ctx, addFriendUri, nil, cbReq, config.Config.Callback.CallbackAfterChangefriendRelation.CallbackTimeOut); err != nil {
	// 		log.ZInfo(ctx, "CallbackAfterAddFriend", utils.Unwrap(err))
	// 	}
	// }()
	if _, err := http.Post(ctx, addFriendUri, nil, cbReq, config.Config.Callback.CallbackAfterChangefriendRelation.CallbackTimeOut); err != nil {
		log.ZInfo(ctx, "CallbackAfterAddFriend", utils.Unwrap(err))
	}
	return nil
}

func CallbackAfterDeleteFriend(ctx context.Context, fromUserID, targetUserID, remark string) error {
	deleteFriendURI := config.Config.Callback.CallbackZapBusinessUrl + DeleteFriendURI

	if !config.Config.Callback.CallbackAfterChangefriendRelation.Enable {
		return nil
	}
	userRelation := &cbapi.UserRelationStruct{
		UserID:       fromUserID,
		TargetUserID: targetUserID,
		Remark:       remark,
	}
	cbReq := &cbapi.CallbackAfterFriendRelationChangedReq{
		UserRelation: *userRelation,
	}
	// 启动goroutine，异步执行http.Post
	// go func() {
	// 	defer func() {
	// 		if err := recover(); err != nil {
	// 			// 处理 panic（如果有）
	// 			log.ZError(ctx, "Recovered from panic in goroutine:", err.(error))
	// 		}
	// 	}()

	// 	if _, err := http.Post(ctx, deleteFriendURI, nil, cbReq, config.Config.Callback.CallbackAfterChangefriendRelation.CallbackTimeOut); err != nil {
	// 		log.ZInfo(ctx, "CallbackAfterDeleteFriend", utils.Unwrap(err))
	// 	}
	// }()
	if _, err := http.Post(ctx, deleteFriendURI, nil, cbReq, config.Config.Callback.CallbackAfterChangefriendRelation.CallbackTimeOut); err != nil {
		log.ZInfo(ctx, "CallbackAfterDeleteFriend", utils.Unwrap(err))
	}
	return nil
}

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
	RemarkServerMemberURI    = "/openim-callback/club-server-user"
	DeleteServerMemberURI    = "/openim-callback/club-server-user/delete"
	ServerChangedCallbackURI = "/openim-callback/club-server"
	ServerDeleteCallbackURI  = "/openim-callback/club-server/delete"
)

func CallbackAfterRemarkServerMember(ctx context.Context, serverID, userID, nickname string) error {
	remarkServerMemberUri := config.Config.Callback.CallbackZapBusinessUrl + RemarkServerMemberURI

	if !config.Config.Callback.CallbackAfterSetServerMember.Enable {
		return nil
	}

	clubServerUser := &cbapi.ClubServerUserStruct{
		ServerID: serverID,
		UserID:   userID,
		Nickname: nickname,
	}
	cbReq := &cbapi.CallbackAfterRemarkServerMemberReq{
		ClubServerUser: *clubServerUser,
	}

	if _, err := http.Post(ctx, remarkServerMemberUri, nil, cbReq, config.Config.Callback.CallbackAfterSetServerMember.CallbackTimeOut); err != nil {
		log.ZInfo(ctx, "CallbackAfterRemarkServerMember", utils.Unwrap(err))
	}

	return nil
}

func CallbackAfterQuitServer(ctx context.Context, serverID, userID, nickname string) error {
	deleteServerMemberURI := config.Config.Callback.CallbackZapBusinessUrl + DeleteServerMemberURI

	if !config.Config.Callback.CallbackAfterSetServerMember.Enable {
		return nil
	}

	clubServerUser := &cbapi.ClubServerUserStruct{
		ServerID: serverID,
		UserID:   userID,
		Nickname: nickname,
	}
	cbReq := &cbapi.CallbackQuitServerReq{
		ClubServerUser: *clubServerUser,
	}

	if _, err := http.Post(ctx, deleteServerMemberURI, nil, cbReq, config.Config.Callback.CallbackAfterSetServerMember.CallbackTimeOut); err != nil {
		log.ZError(ctx, "CallbackAfterQuitServer", utils.Unwrap(err))
	}

	return nil
}

func CallbackAfterServerChanged(ctx context.Context, req *cbapi.CallbackAfterServerChangedReq) error {
	serverChangedCallbackURI := config.Config.Callback.CallbackZapBusinessUrl + ServerChangedCallbackURI

	if !config.Config.Callback.CallbackAfterServerChanged.Enable {
		return nil
	}

	if _, err := http.Post(ctx, serverChangedCallbackURI, nil, req, config.Config.Callback.CallbackAfterServerChanged.CallbackTimeOut); err != nil {
		log.ZError(ctx, "CallbackAfterServerChanged", utils.Unwrap(err))
	}

	return nil
}

func CallbackAfterServerDelete(ctx context.Context, req *cbapi.CallbackAfterServerChangedReq) error {
	serverChangedCallbackURI := config.Config.Callback.CallbackZapBusinessUrl + ServerDeleteCallbackURI

	if !config.Config.Callback.CallbackAfterServerChanged.Enable {
		return nil
	}
	if _, err := http.Post(ctx, serverChangedCallbackURI, nil, req, config.Config.Callback.CallbackAfterServerChanged.CallbackTimeOut); err != nil {
		log.ZError(ctx, "CallbackAfterServerDelete", utils.Unwrap(err))
	}

	return nil
}

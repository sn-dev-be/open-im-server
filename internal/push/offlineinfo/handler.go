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

package offlineinfo

import (
	"context"
	"fmt"

	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/utils"
	"github.com/openimsdk/open-im-server/v3/pkg/apistruct"
	"github.com/openimsdk/open-im-server/v3/pkg/common/i18n"

	"github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
)

type SingleChatMsgHandler struct{ rpc }
type GroupMsgHandler struct{ rpc }
type ServerGroupMsgHandler struct{ rpc }

func (h SingleChatMsgHandler) Msg(ctx context.Context, msg *sdkws.MsgData, lang i18n.Language, pushContentMode int32) (*OfflineMsg, error) {
	info := &OfflineMsg{
		Title:   i18n.Tr(lang, "msg.push.common.title"),
		Content: i18n.Tr(lang, "msg.push.common.base"),
	}

	if pushContentMode == constant.NewMsgPushSettingAllowed {
		info.Title = msg.SenderNickname
		switch msg.ContentType {
		case constant.Text:
			t := apistruct.TextElem{}
			utils.JsonStringToStruct(string(msg.Content), &t)
			info.Content = string(t.Content)
		case constant.Picture:
			info.Content = i18n.Tr(lang, "msg.push.common.picture")
		case constant.Voice:
			info.Content = i18n.Tr(lang, "msg.push.common.voice")
		case constant.Video:
			info.Content = i18n.Tr(lang, "msg.push.common.video")
		case constant.VoiceCall:
			info.Title = i18n.Tr(lang, "msg.push.common.title")
			info.Content = i18n.Tr(lang, "msg.push.common.voiceCall")
		case constant.RedPacket:
			t := apistruct.RedPacketElem{}
			err := utils.JsonStringToStruct(string(msg.Content), &t)
			if err != nil {
				return nil, err
			}
			i18n.TrWithData(lang, "msg.push.common.redPacket", map[string]interface{}{
				"greetings": t.Greetings,
			})
		case constant.FriendApplicationApprovedNotification:
			info.Content = i18n.Tr(lang, "msg.push.common.title")
			info.Content = i18n.Tr(lang, "friendApplicationApproved.desc")
		default:
			info.Content = i18n.Tr(lang, "msg.push.common.base")
		}
	}

	return info, nil
}

func (h GroupMsgHandler) Msg(ctx context.Context, msg *sdkws.MsgData, lang i18n.Language, pushContentMode int32) (*OfflineMsg, error) {
	info := &OfflineMsg{
		Title:   i18n.Tr(lang, "msg.push.common.title"),
		Content: i18n.Tr(lang, "msg.push.common.base"),
	}

	if pushContentMode == constant.NewMsgPushSettingAllowed {
		groupInfo, err := h.groupRpcClient.GetGroupInfo(ctx, msg.GroupID)
		if err != nil {
			log.ZError(ctx, "offline info GetGroupInfo failed", err)
			return nil, err
		}
		info.Title = groupInfo.GroupName
		switch msg.ContentType {
		case constant.Text:
			t := apistruct.TextElem{}
			utils.JsonStringToStruct(string(msg.Content), &t)
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, t.Content)
		case constant.Picture:
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.picture"))
		case constant.Voice:
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.voice"))
		case constant.Video:
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.video"))
		case constant.VoiceCall:
			info.Title = i18n.Tr(lang, "msg.push.common.title")
			info.Content = fmt.Sprintf("%s:%s:", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.voiceCall"))
		case constant.RedPacket:
			t := sdkws.RedPacketElem{}
			err := utils.JsonStringToStruct(string(msg.Content), &t)
			if err != nil {
				return nil, err
			}
			content := i18n.TrWithData(lang, "msg.push.common.redPacket", map[string]interface{}{
				"greetings": t.Greetings,
			})
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, content)
		default:
			info.Content = i18n.Tr(lang, "msg.push.common.base")
		}
	}
	return info, nil
}

func (h ServerGroupMsgHandler) Msg(ctx context.Context, msg *sdkws.MsgData, lang i18n.Language, pushContentMode int32) (*OfflineMsg, error) {
	info := &OfflineMsg{
		Title:   i18n.Tr(lang, "msg.push.common.title"),
		Content: i18n.Tr(lang, "msg.push.common.base"),
	}

	if pushContentMode == constant.NewMsgPushSettingAllowed {
		resp, err := h.clubRpcClient.Client.GetServerGroupBaseInfos(ctx, &club.GetServerGroupBaseInfosReq{GroupIDs: []string{msg.GroupID}})
		if err != nil {
			log.ZError(ctx, "offline info GetGroupInfo failed", err)
			return nil, err
		}
		serverGroup := resp.ServerGroupBaseInfos[0]
		groupName := serverGroup.GroupName
		serverName := serverGroup.ServerName
		info.Title = fmt.Sprintf("%s[%s]", serverName, groupName)

		switch msg.ContentType {
		case constant.Text:
			t := apistruct.TextElem{}
			utils.JsonStringToStruct(string(msg.Content), &t)
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, t.Content)
		case constant.Picture:
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.picture"))
		case constant.Voice:
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.voice"))
		case constant.Video:
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.video"))
		case constant.VoiceCall:
			info.Title = i18n.Tr(lang, "msg.push.common.title")
			info.Content = fmt.Sprintf("%s:%s:", msg.SenderNickname, i18n.Tr(lang, "msg.push.common.voiceCall"))
		case constant.RedPacket:
			t := sdkws.RedPacketElem{}
			err := utils.JsonStringToStruct(string(msg.Content), &t)
			if err != nil {
				return nil, err
			}
			content := i18n.TrWithData(lang, "msg.push.common.redPacket", map[string]interface{}{
				"greetings": t.Greetings,
			})
			info.Content = fmt.Sprintf("%s:%s", msg.SenderNickname, content)
		default:
			info.Content = i18n.Tr(lang, "msg.push.common.base")
		}
	}
	return info, nil
}

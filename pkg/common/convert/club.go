package convert

import (
	"time"

	"github.com/OpenIMSDK/protocol/constant"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func Db2PbServerInfo(m *relation.ServerModel, ownerUserID string, memberCount uint32) *sdkws.ServerInfo {
	return &sdkws.ServerInfo{
		ServerID:             m.ServerID,
		ServerName:           m.ServerName,
		Icon:                 m.Icon,
		Description:          m.Description,
		ApplyMode:            m.ApplyMode,
		InviteMode:           m.InviteMode,
		Searchable:           m.Searchable,
		Status:               m.Status,
		Banner:               m.Banner,
		MemberNumber:         m.MemberNumber,
		UserMutualAccessible: m.UserMutualAccessible,
		CategoryNumber:       m.CategoryNumber,
		ChannelNumber:        m.ChannelNumber,
		CreateTime:           m.CreateTime.UnixMilli(),
	}
}

func Pb2DBServerInfo(m *pbclub.CreateServerReq) *relation.ServerModel {
	return &relation.ServerModel{
		ServerName:           m.ServerName,
		Icon:                 m.Icon,
		Description:          m.Description,
		ApplyMode:            constant.JoinServerNeedVerification,
		InviteMode:           constant.ServerInvitedDenied,
		Searchable:           constant.ServerSearchableDenied,
		Status:               constant.ServerOk,
		Banner:               m.Banner,
		UserMutualAccessible: m.UserMutualAccessible,
		OwnerUserID:          m.OwnerUserID,
		CreateTime:           time.Now(),
		Ex:                   m.Ex,
	}
}

func DB2PbServerFullInfoList(servers []*relation.ServerModel) ([]*sdkws.ServerFullInfo, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	//res := make([]*sdkws.ServerFullInfo, 0, len(servers))
	res := []*sdkws.ServerFullInfo{}
	for _, m := range servers {
		res = append(res, &sdkws.ServerFullInfo{
			ServerInfo: &sdkws.ServerInfo{
				ServerID:             m.ServerID,
				ServerName:           m.ServerName,
				ChannelNumber:        m.ChannelNumber,
				MemberNumber:         m.MemberNumber,
				Icon:                 m.Icon,
				Description:          m.Description,
				ApplyMode:            m.ApplyMode,
				InviteMode:           m.InviteMode,
				Searchable:           m.Searchable,
				Status:               m.Status,
				Banner:               m.Banner,
				UserMutualAccessible: m.UserMutualAccessible,
				CategoryNumber:       m.CategoryNumber,
				OwnerUserID:          m.OwnerUserID,
				CreateTime:           m.CreateTime.UnixMilli(),
				Ex:                   m.Ex,
			},
		})
	}
	return res, nil
}

func Db2PbServerMember(m *relation.ServerMemberModel) *sdkws.ServerMemberFullInfo {
	return &sdkws.ServerMemberFullInfo{
		ServerID:       m.ServerID,
		UserID:         m.UserID,
		RoleLevel:      m.RoleLevel,
		JoinTime:       m.JoinTime.UnixMilli(),
		Nickname:       m.Nickname,
		FaceURL:        m.FaceURL,
		JoinSource:     m.JoinSource,
		OperatorUserID: m.OperatorUserID,
		Ex:             m.Ex,
		MuteEndTime:    m.MuteEndTime.UnixMilli(),
		InviterUserID:  m.InviterUserID,
	}
}

func Pb2DbServerMember(m *sdkws.UserInfo) *relation.ServerMemberModel {
	return &relation.ServerMemberModel{
		UserID:   m.UserID,
		Nickname: m.Nickname,
		FaceURL:  m.FaceURL,
		Ex:       m.Ex,
	}
}
func DB2PbServerBaseInfoList(servers []*relation.ServerModel) ([]*sdkws.ServersListInfo, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	res := []*sdkws.ServersListInfo{}
	for _, m := range servers {
		res = append(res, &sdkws.ServersListInfo{
			ServerID:   m.ServerID,
			ServerName: m.ServerName,
			Icon:       m.Icon,
		})
	}
	return res, nil
}

func DB2PbServerInfo(m *relation.ServerModel) (*sdkws.ServerInfo, error) {
	res := &sdkws.ServerInfo{
		ServerID:             m.ServerID,
		ServerName:           m.ServerName,
		ChannelNumber:        m.ChannelNumber,
		MemberNumber:         m.MemberNumber,
		Icon:                 m.Icon,
		Description:          m.Description,
		ApplyMode:            m.ApplyMode,
		InviteMode:           m.InviteMode,
		Searchable:           m.Searchable,
		Status:               m.Status,
		Banner:               m.Banner,
		UserMutualAccessible: m.UserMutualAccessible,
		CategoryNumber:       m.CategoryNumber,
		OwnerUserID:          m.OwnerUserID,
		DappID:               m.DappID,
		CreateTime:           m.CreateTime.UnixMilli(),
		Ex:                   m.Ex,
	}
	return res, nil
}

func DB2PbCategory(m *relation.GroupCategoryModel, g []*sdkws.ServerGroupListInfo) (*sdkws.GroupCategoryListInfo, error) {
	res := &sdkws.GroupCategoryListInfo{
		CategoryID:    m.CategoryID,
		CategoryName:  m.CategoryName,
		ReorderWeight: m.ReorderWeight,
		CategoryType:  m.CategoryType,
		GroupList:     g,
	}
	return res, nil
}

func Db2PbServerRequest(
	m *relation.ServerRequestModel,
	user *sdkws.PublicUserInfo,
	server *sdkws.ServerInfo,
) *sdkws.ServerRequest {
	return &sdkws.ServerRequest{
		UserInfo:      user,
		ServerInfo:    server,
		HandleResult:  m.HandleResult,
		ReqMsg:        m.ReqMsg,
		HandleMsg:     m.HandledMsg,
		ReqTime:       m.ReqTime.UnixMilli(),
		HandleUserID:  m.HandleUserID,
		HandleTime:    m.HandledTime.UnixMilli(),
		Ex:            m.Ex,
		JoinSource:    m.JoinSource,
		InviterUserID: m.InviterUserID,
	}
}
func DB2PbGroupDapp(m *relation.GroupDappModel) (*sdkws.GroupDappFullInfo, error) {
	res := &sdkws.GroupDappFullInfo{
		Id:         m.ID,
		GroupID:    m.GroupID,
		DappID:     m.DappID,
		CreateTime: m.CreateTime.UnixMilli(),
	}
	return res, nil
}

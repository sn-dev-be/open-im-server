package convert

import (
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
		OwnerUserID:          ownerUserID,
		MemberNumber:         memberCount,
		UserMutualAccessible: m.UserMutualAccessible,
		CategoryNumber:       m.CategoryNumber,
		GroupNumber:          m.GroupNumber,
		DappID:               m.DappID,
		CreateTime:           m.CreateTime.UnixMilli(),
		CommunityName:        m.CommunityName,
		CommunityBanner:      m.CommunityBanner,
		CommunityViewMode:    m.CommunityViewMode,
	}
}

func DB2PbServerInfo(m *relation.ServerModel) *sdkws.ServerInfo {
	return &sdkws.ServerInfo{
		ServerID:             m.ServerID,
		ServerName:           m.ServerName,
		GroupNumber:          m.GroupNumber,
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
		CommunityName:        m.CommunityName,
		CommunityBanner:      m.CommunityBanner,
		CommunityViewMode:    m.CommunityViewMode,
	}
}

// func Db2PbServerFullInfo(m *relation.ServerModel) *sdkws.ServerFullInfo {

// 	return &sdkws.ServerFullInfo{
// 		ServerInfo: &sdkws.ServerInfo{
// 			ServerID:             m.ServerID,
// 			ServerName:           m.ServerName,
// 			GroupNumber:          m.GroupNumber,
// 			MemberNumber:         m.MemberNumber,
// 			Icon:                 m.Icon,
// 			Description:          m.Description,
// 			ApplyMode:            m.ApplyMode,
// 			InviteMode:           m.InviteMode,
// 			Searchable:           m.Searchable,
// 			Status:               m.Status,
// 			Banner:               m.Banner,
// 			UserMutualAccessible: m.UserMutualAccessible,
// 			CategoryNumber:       m.CategoryNumber,
// 			OwnerUserID:          m.OwnerUserID,
// 			CreateTime:           m.CreateTime.UnixMilli(),
// 			Ex:                   m.Ex,
// 		},
// 	}
// }

func Db2PbServerAbstractInfo(
	serverID string,
	serverMemberNumber uint32,
	serverMemberListHash uint64,
) *pbclub.ServerAbstractInfo {
	return &pbclub.ServerAbstractInfo{
		ServerID:             serverID,
		ServerMemberNumber:   serverMemberNumber,
		ServerMemberListHash: serverMemberListHash,
	}
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
		ServerRoleID:   m.ServerRoleID,
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

func Db2PbGroupCategory(m *relation.GroupCategoryModel) *sdkws.GroupCategoryInfo {
	return &sdkws.GroupCategoryInfo{
		CategoryID:    m.CategoryID,
		CategoryName:  m.CategoryName,
		ReorderWeight: m.ReorderWeight,
		CategoryType:  m.CategoryType,
		ServerID:      m.ServerID,
		Ex:            m.Ex,
		CreateTime:    m.CreateTime.UnixMilli(),
	}
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

func Db2PbGroupDapp(m *relation.GroupDappModel) *sdkws.GroupDappFullInfo {
	return &sdkws.GroupDappFullInfo{
		ID:         m.ID,
		GroupID:    m.GroupID,
		DappID:     m.DappID,
		CreateTime: m.CreateTime.UnixMilli(),
	}
}

func DB2PbServerBlack(m *relation.ServerBlackModel) *sdkws.ServerBlackFullInfo {
	res := &sdkws.ServerBlackFullInfo{
		ServerID:       m.ServerID,
		BlockUserID:    m.BlockUserID,
		AddSource:      m.AddSource,
		OperatorUserID: m.OperatorUserID,
		Ex:             m.Ex,
		CreateTime:     m.CreateTime.UnixMilli(),
	}
	return res
}

func Db2PbServerRole(m *relation.ServerRoleModel) *sdkws.ServerRole {
	return &sdkws.ServerRole{
		RoleID:      m.RoleID,
		ServerID:    m.ServerID,
		RoleName:    m.RoleName,
		Icon:        m.Icon,
		Priority:    m.Priority,
		ColorLevel:  m.ColorLevel,
		Ex:          m.Ex,
		CreateTime:  m.CreateTime.UnixMilli(),
		Permissions: m.Permissions.String(),
	}
}

func Db2PbGroupTreasury(m *relation.GroupTreasuryModel) *sdkws.GroupTreasuryInfo {
	return &sdkws.GroupTreasuryInfo{
		GroupID:              m.GroupID,
		TreasuryID:           m.TreasuryID,
		Name:                 m.Name,
		Icon:                 m.Icon,
		ContractAddress:      m.ContractAddress,
		TokenAddress:         m.TokenAddress,
		Symbol:               m.Symbol,
		WalletType:           m.WalletType,
		AdministratorAddress: m.AdministratorAddress,
		Decimal:              m.Decimal,
	}
}

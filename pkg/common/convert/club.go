package convert

import (
	"time"

	pbclub "github.com/OpenIMSDK/protocol/club"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/openimsdk/open-im-server/v3/pkg/common/db/table/relation"
)

func Pb2DBServerInfo(m *pbclub.CreateServerReq) *relation.ServerModel {
	return &relation.ServerModel{
		ServerName:           m.ServerName,
		Icon:                 m.Icon,
		Description:          m.Description,
		ApplyMode:            1,
		InviteMode:           0,
		Searchable:           0,
		Status:               0,
		Banner:               m.Banner,
		UserMutualAccessible: m.UserMutualAccessible,
		OwnerUserID:          m.OwnerUserID,
		CreateTime:           time.Now(),
		Ex:                   m.Ex,
	}
}

func DB2PbServerInfo(servers []*relation.ServerModel) ([]*sdkws.ServerFullInfo, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	//res := make([]*sdkws.ServerFullInfo, 0, len(servers))
	res := []*sdkws.ServerFullInfo{}
	for _, m := range servers {
		res = append(res, &sdkws.ServerFullInfo{
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
			CreateTime:           m.CreateTime.Format("2006-01-02 15:04:05"),
			Ex:                   m.Ex,
		})
	}
	return res, nil
}

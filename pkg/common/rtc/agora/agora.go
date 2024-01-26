package agora

import (
	"context"

	agbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/OpenIMSDK/tools/log"

	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/open-im-server/v3/pkg/common/rtc"
)

type Agora struct {
	appID          string
	appCertificate string
	expire         uint32
}

func NewAgora() rtc.Rtc {
	a := &Agora{
		appID:          config.Config.Agora.AppID,
		appCertificate: config.Config.Agora.AppCertificate,
		expire:         config.Config.Agora.Expire,
	}
	return a
}

func (a *Agora) Token(
	ctx context.Context,
	userID string,
	channelID string,
	roleType uint16,
) (string, error) {

	token, err := agbuilder.BuildTokenWithAccount(
		a.appID,
		a.appCertificate,
		channelID,
		userID,
		agbuilder.Role(roleType),
		a.expire*60,
	)
	if err != nil {
		log.ZError(ctx, "agora generate token error", err)
		return "", err
	}
	return token, err
}

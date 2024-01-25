package third

import (
	"context"

	"github.com/OpenIMSDK/protocol/third"
)

func (t *thirdServer) GetRtcToken(ctx context.Context, req *third.InitiateRtcTokenReq) (*third.InitiateRtcTokenResp, error) {
	token, err := t.rtc.Token(ctx, req.UserID, req.ChannelID, uint16(req.RoleType))
	if err != nil {
		return nil, err
	}
	return &third.InitiateRtcTokenResp{Token: token}, nil
}

package errs

import "github.com/OpenIMSDK/tools/errs"

var (
	ErrVoiceChannelClosed     = errs.NewCodeError(VoiceChannelClosedErr, "VoiceChannelClosedError")
	ErrVoiceAlreadyInvitation = errs.NewCodeError(VoiceAlreadyInvitationErr, "VoiceAlreadyInviationError")
	ErrMsgBeBlocked           = errs.NewCodeError(MsgBeBlocked, "MsgBeBlocked") //陌生人消息被拦截

)

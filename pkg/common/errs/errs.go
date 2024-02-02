package errs

import "github.com/OpenIMSDK/tools/errs"

var (
	ErrVoiceChannelClosed     = errs.NewCodeError(VoiceChannelClosedErr, "VoiceChannelClosedError")
	ErrVoiceAlreadyInvitation = errs.NewCodeError(VoiceAlreadyInvitationErr, "VoiceAlreadyInviationError")
)

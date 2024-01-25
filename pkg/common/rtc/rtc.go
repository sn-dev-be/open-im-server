package rtc

import "context"

const (
	Inviter uint16 = 1
	Invitee uint16 = 2
)

type Rtc interface {
	Token(ctx context.Context, userID string, channelID string, roleType uint16) (string, error)
}

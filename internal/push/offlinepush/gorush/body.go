package gorush

import (
	"fmt"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
)

const IOSSoundName = "default"

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success string `json:"success"`
	Counts  uint32 `json:"counts"`
}

func (r *Resp) parseError() (err error) {
	switch r.Success {
	case pushSuccess:
		err = nil
	default:
		err = fmt.Errorf("code %d, msg %s", r.Code, r.Message)
	}
	return err
}

type RespI interface {
	parseError() error
}

type Payload struct {
	ConversationID string `json:"conversationID,omitempty"`
	ServerID       string `json:"serverID,omitempty"`
	ContentType    int32  `json:"contentType,omitempty"`
}

type Notification struct {
	Tokens    *[]string `json:"tokens,omitempty"`
	Platform  int       `json:"platform,omitempty"`
	Title     string    `json:"title,omitempty"`
	Message   string    `json:"message,omitempty"`
	Topic     string    `json:"topic,omitempty"`
	Retry     uint32    `json:"retry,omitempty"`
	SoundNmae string    `json:"name,omitempty"`
	Badge     int       `json:"badge,omitempty"`
	Data      *Payload  `json:"data,omitempty"`
}

type Notifications struct {
	Notifications []*Notification `json:"notifications"`
}

func NewNotification(
	tokens []string,
	platform int,
	title, message string,
	opts *offlinepush.Opts,
	badge int,
) *Notification {
	n := &Notification{
		Tokens:   &tokens,
		Platform: platform,
		Message:  message,
		Title:    title,
		Data: &Payload{
			ConversationID: opts.Msg.ConversationID,
			ContentType:    opts.Msg.ContentType,
		},
	}
	if platform == constant.IOSPlatformID {
		n.Topic = config.Config.Push.Gorush.BundleID
		n.SoundNmae = IOSSoundName
		n.Badge = badge
	}
	if opts.Server != nil {
		n.Data.ServerID = opts.Server.ServerID
	}
	return n
}

package gorush

import (
	"fmt"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/utils/splitter"
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

type Notification struct {
	Tokens    *[]string `json:"tokens,omitempty"`
	Platform  int       `json:"platform,omitempty"`
	Title     string    `json:"title,omitempty"`
	Message   string    `json:"message,omitempty"`
	Topic     string    `json:"topic,omitempty"`
	Retry     uint32    `json:"retry,omitempty"`
	SoundNmae string    `json:"name,omitempty"`
}

type Notifications struct {
	Notifications []*Notification `json:"notifications"`
}

func NewNotifications(
	tokens []string,
	platform int,
	title, message string,
) []*Notification {
	var notifications []*Notification
	maxNum := 100
	s := splitter.NewSplitter(maxNum, tokens)
	for _, v := range s.GetSplitResult() {
		n := &Notification{
			Tokens:   &v.Item,
			Platform: platform,
			Message:  message,
			Title:    title,
		}
		if platform == constant.IOSPlatformID {
			n.Topic = config.Config.Push.Gorush.BundleID
			n.SoundNmae = IOSSoundName
		}
		notifications = append(notifications, n)
	}
	return notifications
}

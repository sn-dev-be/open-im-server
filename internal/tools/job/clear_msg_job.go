package job

import (
	"encoding/json"
	"time"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/internal/tools/msg"
)

type ClearMsgJob struct {
	CommonJob
	ConversationID string       `json:"conversationID"`
	CronCycle      int32        `json:"cronCycle"`
	MsgTool        *msg.MsgTool `json:"-"`
}

func NewClearMsgJob(ConversationID, cronExpr string, cronCycle int32, msgTool *msg.MsgTool) *ClearMsgJob {
	return &ClearMsgJob{
		ConversationID: ConversationID,
		CronCycle:      cronCycle,
		CommonJob:      CommonJob{Name: ClearMsgJobNamePrefix + ConversationID, CronExpr: cronExpr},
		MsgTool:        msgTool,
	}
}

func (c *ClearMsgJob) Run() {
	ctx := mcontext.NewCtx(utils.GetSelfFuncName())
	log.ZInfo(ctx, "start clearMsg job", "jobName", c.Name)
	c.MsgTool.ClearMsgsByConversationID(c.ConversationID, c.GetSeconds())
	log.ZInfo(ctx, "clearMsg job finished", "jobName", c.Name)
}

func (c *ClearMsgJob) GetSeconds() int64 {
	var days int32
	switch c.CronCycle {
	case constant.CrontabDayOne:
		days = 1
	case constant.CrontabDayTwo:
		days = 2
	case constant.CrontabDayThree:
		days = 3
	case constant.CrontabDayFour:
		days = 4
	case constant.CrontabDayFive:
		days = 5
	case constant.CrontabDaySix:
		days = 6
	case constant.CrontabWeekOne:
		days = 7
	case constant.CrontabWeekTwo:
		days = 14
	case constant.CrontabWeekThree:
		days = 21
	case constant.CrontabMonth:
		days = 30
	}
	return int64((time.Duration(days) * 24 * time.Hour).Seconds())
}

func (c *ClearMsgJob) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func (c *ClearMsgJob) UnSerialize(b []byte) error {
	return json.Unmarshal(b, c)
}

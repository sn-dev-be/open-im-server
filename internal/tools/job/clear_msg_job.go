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

var cronCycleMap = map[int32]int32{
	constant.CrontabDayOne:    1,
	constant.CrontabDayTwo:    2,
	constant.CrontabDayThree:  3,
	constant.CrontabDayFour:   4,
	constant.CrontabDayFive:   5,
	constant.CrontabDaySix:    6,
	constant.CrontabWeekOne:   7,
	constant.CrontabWeekTwo:   14,
	constant.CrontabWeekThree: 21,
	constant.CrontabMonth:     30,
}

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
	days, ok := cronCycleMap[c.CronCycle]
	if !ok {
	}
	return int64((time.Duration(days) * 24 * time.Hour).Seconds())
}

func (c *ClearMsgJob) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func (c *ClearMsgJob) UnSerialize(b []byte) error {
	return json.Unmarshal(b, c)
}

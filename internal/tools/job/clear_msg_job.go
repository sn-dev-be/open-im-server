package job

import (
	"encoding/json"

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
	c.MsgTool.ClearMsgsByConversationID(c.ConversationID)
	log.ZInfo(ctx, "clearMsg job finished", "jobName", c.Name)
}

func (c *ClearMsgJob) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func (c *ClearMsgJob) UnSerialize(b []byte) error {
	return json.Unmarshal(b, c)
}

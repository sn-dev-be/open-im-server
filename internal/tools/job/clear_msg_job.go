package job

import (
	"encoding/json"

	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	"github.com/openimsdk/open-im-server/v3/internal/tools/msg"
)

const JobNamePrefix = "clearMsgJob_"

type ClearMsgJob struct {
	CommonJob
	ConversationID string `json:"conversationID"`
	MsgTool        *msg.MsgTool
}

func NewClearMsgJob(ConversationID, cronExpr string, msgTool *msg.MsgTool) *ClearMsgJob {
	return &ClearMsgJob{
		ConversationID: ConversationID,
		CommonJob:      CommonJob{Name: JobNamePrefix + ConversationID, CronExpr: cronExpr},
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

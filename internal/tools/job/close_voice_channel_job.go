package job

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/log"
	"github.com/OpenIMSDK/tools/mcontext"
	"github.com/OpenIMSDK/tools/utils"

	dcron "github.com/openimsdk/open-im-server/v3/internal/tools/cron"
	"github.com/openimsdk/open-im-server/v3/internal/tools/msg"
)

type CloseVocieChannelJob struct {
	CommonJob
	ChannelID   string       `json:"channelID"`
	OpUserID    string       `json:"opUserID"`
	SessionType int32        `json:"sessionType"`
	MsgTool     *msg.MsgTool `json:"-"`
	Cron        *dcron.Dcron `json:"-"`
}

func NewCloseVocieChannelJob(
	channelID string,
	userID string,
	sessionType int32,
	msgTool *msg.MsgTool,
	cron *dcron.Dcron,
) *CloseVocieChannelJob {
	now := time.Unix(time.Now().Unix(), 0)
	expr := fmt.Sprintf("%d */1  * * *", now.Minute())
	return &CloseVocieChannelJob{
		ChannelID:   channelID,
		OpUserID:    userID,
		SessionType: sessionType,
		MsgTool:     msgTool,
		Cron:        cron,
		CommonJob:   CommonJob{Name: expr + channelID, CronExpr: expr},
	}
}

func (c *CloseVocieChannelJob) Run() {
	ctx := mcontext.NewCtx(utils.GetSelfFuncName())
	log.ZInfo(ctx, "start closeVocieChannel", "jobName", c.Name)
	usersID, err := c.MsgTool.MsgDatabase.GetVoiceChannelUsersID(ctx, c.ChannelID, "")
	if err != nil {
		log.ZError(ctx, "get channel usersID err", err, "channelID", c.ChannelID)
		return
	}
	_, elapsedSec, err := c.MsgTool.MsgDatabase.GetVoiceChannelDuration(ctx, c.ChannelID)
	if err != nil {
		return
	}
	for _, userID := range usersID {
		c.MsgTool.MsgNotificationSender.Notification(ctx, c.OpUserID, userID, constant.SignalingClosedNotification, nil)
		if c.SessionType == constant.SingleChatType {
			tips := &sdkws.SignalVoiceSingleChatTips{
				ElapsedSec: int32(elapsedSec),
				OpUserID:   c.OpUserID,
			}
			c.MsgTool.MsgNotificationSender.Notification(ctx, c.OpUserID, userID, constant.SignalingSingleChatClosedNotification, tips)
		}
	}
	if err := c.MsgTool.MsgDatabase.DelVoiceChannel(ctx, c.ChannelID); err != nil {
		return
	}
	c.Cron.Remove(c.Name)
	log.ZDebug(ctx, "delete closeVocieChannel", "jobName", c.Name)
	log.ZInfo(ctx, "closeVocieChannel job finished", "jobName", c.Name)
}

func (c *CloseVocieChannelJob) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CloseVocieChannelJob) UnSerialize(b []byte) error {
	return json.Unmarshal(b, c)
}

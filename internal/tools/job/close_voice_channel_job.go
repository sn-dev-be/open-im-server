package job

import (
	"encoding/json"
	"strconv"

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
	JobType     int          `json:"jobType"`
	MsgTool     *msg.MsgTool `json:"-"`
	Cron        *dcron.Dcron `json:"-"`
}

func NewCloseVocieChannelJob(
	channelID string,
	userID string,
	sessionType int32,
	expr string,
	jobType int,
	msgTool *msg.MsgTool,
	cron *dcron.Dcron,
) *CloseVocieChannelJob {
	return &CloseVocieChannelJob{
		ChannelID:   channelID,
		OpUserID:    userID,
		SessionType: sessionType,
		JobType:     jobType,
		MsgTool:     msgTool,
		Cron:        cron,
		CommonJob:   CommonJob{Name: CloseVoiceChannelJobNamePrefix + channelID + "_" + strconv.Itoa(jobType), CronExpr: expr},
	}
}

func (c *CloseVocieChannelJob) Run() {
	ctx := mcontext.NewCtx(utils.GetSelfFuncName())
	log.ZInfo(ctx, "start close voice channel", "jobName", c.Name)

	status, err := c.MsgTool.MsgDatabase.GetVoiceChannelStatus(ctx, c.ChannelID)
	if err != nil {
		return
	}
	if c.JobType == OneMinuteCloseVoiceChannelJob && status == constant.OnTheCall {
		c.Cron.Remove(c.Name)
		return
	}

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
		c.MsgTool.MsgNotificationSender.
			Notification(ctx, c.OpUserID, userID, constant.SignalingClosedNotification, nil)

		if c.SessionType == constant.SingleChatType {
			tips := &sdkws.SignalVoiceSingleChatTips{
				ElapsedSec: int32(elapsedSec),
				OpUserID:   c.OpUserID,
			}
			contentType := constant.SignalingSingleChatClosedNotification
			if c.JobType == OneMinuteCloseVoiceChannelJob {
				contentType = constant.SignalingSingleChatNoAnswerNotification
			}
			c.MsgTool.MsgNotificationSender.
				Notification(ctx, c.OpUserID, userID, int32(contentType), tips)
		}
	}

	c.Cron.Remove(c.Name)
	log.ZDebug(ctx, "delete close voice channel", "jobName", c.Name)
	if err := c.MsgTool.MsgDatabase.DelVoiceChannel(ctx, c.ChannelID); err != nil {
	}

	log.ZInfo(ctx, "job finished", "jobName", c.Name)
}

func (c *CloseVocieChannelJob) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CloseVocieChannelJob) UnSerialize(b []byte) error {
	return json.Unmarshal(b, c)
}

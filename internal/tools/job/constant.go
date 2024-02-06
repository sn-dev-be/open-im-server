package job

import "reflect"

const (
	ClearMsgJobNamePrefix          = "clearMsgJob_"
	CloseVoiceChannelJobNamePrefix = "closeVoiceChannelJob_"
)

const (
	OneHourCloseVoiceChannelJob   = 1
	OneMinuteCloseVoiceChannelJob = 2
)

const (
	TClearMsg          = 1
	TCloseVoiceChannel = 2
)

var JobTypeMap = map[int]reflect.Type{
	TClearMsg:          reflect.TypeOf(ClearMsgJob{}),
	TCloseVoiceChannel: reflect.TypeOf(CloseVocieChannelJob{}),
}

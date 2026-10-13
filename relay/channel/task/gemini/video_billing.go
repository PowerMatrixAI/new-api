package gemini

import (
	"fmt"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"strconv"
)

func (a *TaskAdaptor) VideoBillingUsage(c *gin.Context, info *relaycommon.RelayInfo) (types.VideoUsage, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return types.VideoUsage{}, err
	}
	params := VeoParameters{}
	if err := taskcommon.UnmarshalMetadata(req.Metadata, &params); err != nil {
		return types.VideoUsage{}, err
	}
	seconds := 8.0
	if _, explicit := req.Metadata["durationSeconds"]; explicit {
		seconds = float64(params.DurationSeconds)
	} else if req.Duration != 0 {
		seconds = float64(req.Duration)
	} else if req.Seconds != "" {
		seconds, err = strconv.ParseFloat(req.Seconds, 64)
		if err != nil {
			return types.VideoUsage{}, err
		}
	}
	if !types.ValidVideoSeconds(seconds) || seconds != float64(int(seconds)) {
		return types.VideoUsage{}, fmt.Errorf("video duration must be a whole number of seconds")
	}
	audio := true
	if params.GenerateAudio != nil {
		audio = *params.GenerateAudio
	}
	return types.VideoUsage{
		Seconds:    seconds,
		Resolution: ResolveVeoResolution(req.Metadata, req.Size), GenerateAudio: &audio,
	}, nil
}

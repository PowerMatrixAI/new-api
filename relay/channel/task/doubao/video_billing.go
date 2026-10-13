package doubao

import (
	"fmt"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func (a *TaskAdaptor) VideoBillingUsage(c *gin.Context, _ *relaycommon.RelayInfo) (types.VideoUsage, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return types.VideoUsage{}, err
	}
	body, err := a.convertToRequestPayload(&req)
	if err != nil {
		return types.VideoUsage{}, err
	}
	seconds := 5.0
	if req.Duration != 0 {
		seconds = float64(req.Duration)
	}
	if body.Duration != nil {
		seconds = float64(*body.Duration)
	}
	// Frame-driven requests cannot safely use an assumed duration.
	if body.Frames != nil {
		return types.VideoUsage{}, fmt.Errorf("frame-based requests are not supported by per-second pricing; use duration")
	}
	resolution := body.Resolution
	if resolution == "" {
		resolution = "720p"
	}
	var audio *bool
	if body.GenerateAudio != nil {
		value := bool(*body.GenerateAudio)
		audio = &value
	}
	return types.VideoUsage{Seconds: seconds, Resolution: types.NormalizeVideoResolution(resolution), GenerateAudio: audio}, nil
}

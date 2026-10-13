package sora

import (
	"strconv"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func (a *TaskAdaptor) VideoBillingUsage(c *gin.Context, info *relaycommon.RelayInfo) (types.VideoUsage, error) {
	if info.TaskVideoUsage != nil {
		return *info.TaskVideoUsage, nil
	}
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return types.VideoUsage{}, err
	}
	seconds := float64(req.Duration)
	if req.Seconds != "" {
		seconds, err = strconv.ParseFloat(req.Seconds, 64)
		if err != nil {
			return types.VideoUsage{}, err
		}
	}
	if req.Seconds == "" && req.Duration == 0 {
		seconds = 4
	}
	size := req.Size
	if size == "" {
		size = "720x1280"
	}
	return types.VideoUsage{Seconds: seconds, Resolution: types.NormalizeVideoResolution(size)}, nil
}

package ali

import (
	"strings"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func (a *TaskAdaptor) VideoBillingUsage(c *gin.Context, info *relaycommon.RelayInfo) (types.VideoUsage, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return types.VideoUsage{}, err
	}
	body, err := a.convertToAliRequest(info, req)
	if err != nil {
		return types.VideoUsage{}, err
	}
	params := body.Parameters
	resolution := params.Resolution
	if params.Size != "" {
		resolution = params.Size
	}
	audio := params.Audio
	if audio == nil && (strings.HasPrefix(body.Model, "wan2.5") || strings.HasPrefix(body.Model, "wan2.6")) {
		value := true
		audio = &value
	}
	return types.VideoUsage{Seconds: float64(params.Duration), Resolution: types.NormalizeVideoResolution(resolution), GenerateAudio: audio}, nil
}

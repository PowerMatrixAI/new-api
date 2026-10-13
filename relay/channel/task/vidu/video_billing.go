package vidu

import (
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func (a *TaskAdaptor) VideoBillingUsage(c *gin.Context, info *relaycommon.RelayInfo) (types.VideoUsage, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return types.VideoUsage{}, err
	}
	body, err := a.convertToRequestPayload(&req, info)
	if err != nil {
		return types.VideoUsage{}, err
	}
	return types.VideoUsage{Seconds: float64(body.Duration), Resolution: types.NormalizeVideoResolution(body.Resolution), GenerateAudio: &body.Bgm}, nil
}

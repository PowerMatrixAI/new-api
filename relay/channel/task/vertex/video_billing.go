package vertex

import (
	geminitask "github.com/QuantumNous/new-api/relay/channel/task/gemini"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

func (a *TaskAdaptor) VideoBillingUsage(c *gin.Context, info *relaycommon.RelayInfo) (types.VideoUsage, error) {
	return (&geminitask.TaskAdaptor{}).VideoBillingUsage(c, info)
}

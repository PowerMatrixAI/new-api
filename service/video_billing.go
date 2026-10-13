package service

import (
	"context"
	"fmt"

	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
)

func addVideoBillingOther(other map[string]interface{}, snapshot *types.VideoBillingSnapshot) {
	if snapshot == nil {
		return
	}
	other["billing_mode"] = "video_per_second"
	other["video_seconds"] = snapshot.Usage.Seconds
	other["video_price"] = snapshot.Pricing.UnitPrice(snapshot.Usage)
	other["video_resolution"] = snapshot.Usage.Resolution
	if snapshot.Usage.GenerateAudio != nil {
		other["video_generate_audio"] = *snapshot.Usage.GenerateAudio
	}
}

func settleVideoTaskQuota(ctx context.Context, task *model.Task, result *relaycommon.TaskInfo) {
	if result.Status != model.TaskStatusSuccess {
		return
	}
	bc := task.PrivateData.BillingContext
	snapshot := bc.VideoBilling
	usage := snapshot.Usage
	if actual := result.VideoUsage; actual != nil {
		if actual.Seconds != 0 {
			if !types.ValidVideoSeconds(actual.Seconds) {
				logger.LogWarn(ctx, fmt.Sprintf("invalid video usage for task %s; keeping reserved charge", task.TaskID))
				return
			}
			usage.Seconds = actual.Seconds
		}
		if actual.Resolution != "" {
			usage.Resolution = types.NormalizeVideoResolution(actual.Resolution)
		}
		if actual.GenerateAudio != nil {
			usage.GenerateAudio = actual.GenerateAudio
		}
	}
	actualQuota, err := snapshot.Pricing.Quota(usage, bc.GroupRatio)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("video settlement %s failed: %v; keeping reserved charge", task.TaskID, err))
		return
	}
	previousUsage := snapshot.Usage
	snapshot.Usage = usage
	reason := fmt.Sprintf("视频按秒结算：%.4g 秒 × $%.6g/秒 × 分组倍率 %.4g", usage.Seconds, snapshot.Pricing.UnitPrice(usage), bc.GroupRatio)
	RecalculateTaskQuota(ctx, task, actualQuota, reason)
	if task.Quota != actualQuota {
		snapshot.Usage = previousUsage
		return
	}
	if err := task.UpdateVideoBilling(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("persist video settlement %s failed: %v", task.TaskID, err))
	}
}

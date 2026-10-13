package relay

import (
	"fmt"
	"math"
	"strconv"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

type videoUsageAdaptor interface {
	VideoBillingUsage(*gin.Context, *relaycommon.RelayInfo) (types.VideoUsage, error)
}

func taskPriceData(c *gin.Context, info *relaycommon.RelayInfo, adaptor any) (types.PriceData, error) {
	pricing, ok := ratio_setting.GetVideoPrice(info.OriginModelName)
	if !ok {
		return helper.ModelPriceHelperPerCall(c, info)
	}
	videoAdaptor, ok := adaptor.(videoUsageAdaptor)
	if !ok {
		return types.PriceData{}, fmt.Errorf("this video provider does not support per-second pricing")
	}
	if err := normalizeVideoDuration(c); err != nil {
		return types.PriceData{}, err
	}
	if info.Action == constant.TaskActionRemix && info.TaskVideoUsage == nil {
		return types.PriceData{}, fmt.Errorf("the original video's duration is unavailable for per-second remix pricing")
	}
	usage, err := videoAdaptor.VideoBillingUsage(c, info)
	if err != nil {
		return types.PriceData{}, err
	}
	usage.Resolution = types.NormalizeVideoResolution(usage.Resolution)
	group := helper.HandleGroupRatio(c, info)
	quota, err := pricing.Quota(usage, group.GroupRatio)
	if err != nil {
		return types.PriceData{}, err
	}
	return types.PriceData{
		ModelPrice: pricing.UnitPrice(usage), UsePrice: true, Quota: quota,
		FreeModel: quota == 0, GroupRatioInfo: group,
		VideoBilling: &types.VideoBillingSnapshot{Pricing: pricing, Usage: usage},
	}, nil
}

// Validate metadata too: adapters can overlay nested provider parameters after
// parsing the common request. Explicit zero and negative durations are invalid.
func normalizeVideoDuration(c *gin.Context) error {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return err
	}
	parse := func(value any) (int, error) {
		seconds, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil || !types.ValidVideoSeconds(seconds) || math.Trunc(seconds) != seconds {
			return 0, fmt.Errorf("video duration must be a whole number between 1 and %d seconds", types.MaxVideoBillingSeconds)
		}
		return int(seconds), nil
	}
	if req.DurationProvided || req.Duration != 0 {
		if _, err := parse(req.Duration); err != nil {
			return err
		}
	}
	if req.Seconds != "" {
		seconds, err := parse(req.Seconds)
		if err != nil {
			return err
		}
		if req.Duration == 0 {
			req.Duration = seconds
		}
	}
	var validate func(any) error
	validate = func(value any) error {
		switch data := value.(type) {
		case map[string]interface{}:
			for key, value := range data {
				if key == "duration" || key == "seconds" || key == "durationSeconds" {
					if _, err := parse(value); err != nil {
						return err
					}
				} else if err := validate(value); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, value := range data {
				if err := validate(value); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := validate(req.Metadata); err != nil {
		return err
	}
	c.Set("task_request", req)
	return nil
}

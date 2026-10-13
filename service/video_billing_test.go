package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/stretchr/testify/require"
)

type videoPollingAdaptor struct{ result relaycommon.TaskInfo }

func (a *videoPollingAdaptor) Init(*relaycommon.RelayInfo) {}
func (a *videoPollingAdaptor) FetchTask(string, string, map[string]any, string) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"vendor":true}`))}, nil
}
func (a *videoPollingAdaptor) ParseTaskResult([]byte) (*relaycommon.TaskInfo, error) {
	return &a.result, nil
}
func (a *videoPollingAdaptor) AdjustBillingOnComplete(*model.Task, *relaycommon.TaskInfo) int {
	return 1
}

func TestVideoBillingPollingSettlement(t *testing.T) {
	for _, source := range []string{BillingSourceWallet, BillingSourceSubscription} {
		for _, tc := range []struct {
			name    string
			usage   *types.VideoUsage
			seconds float64
		}{
			{"refund", &types.VideoUsage{Seconds: 6}, 6},
			{"extra-charge", &types.VideoUsage{Seconds: 10}, 10},
			{"missing-usage", nil, 8},
			{"invalid-usage", &types.VideoUsage{Seconds: 9999999}, 8},
		} {
			t.Run(source+"/"+tc.name, func(t *testing.T) {
				truncate(t)
				const userID, tokenID, channelID, subID = 110, 110, 110, 110
				pre := int(8 * .1 * common.QuotaPerUnit)
				actual := int(tc.seconds * .1 * common.QuotaPerUnit)
				seedUser(t, userID, 1000000)
				seedToken(t, tokenID, userID, "video-test-token", 1000000)
				seedChannel(t, channelID)
				seedSubscription(t, subID, userID, 2000000, int64(pre))
				require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{"used_quota": pre, "request_count": 1}).Error)
				require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", channelID).Update("used_quota", pre).Error)
				task := makeTask(userID, channelID, pre, tokenID, source, subID)
				task.PrivateData.BillingContext.VideoBilling = &types.VideoBillingSnapshot{Pricing: types.VideoPricing{PricePerSecond: .1}, Usage: types.VideoUsage{Seconds: 8}}
				require.NoError(t, model.DB.Create(task).Error)
				var stale model.Task
				require.NoError(t, model.DB.First(&stale, task.ID).Error)
				// Neither live administrator pricing nor the adaptor's token/credit
				// adjustment may replace the tariff saved for this task.
				before := ratio_setting.VideoPrice2JSONString()
				t.Cleanup(func() { _ = ratio_setting.UpdateVideoPriceByJSONString(before) })
				require.NoError(t, ratio_setting.UpdateVideoPriceByJSONString(`{"test-model":{"price_per_second":9}}`))
				adaptor := &videoPollingAdaptor{result: relaycommon.TaskInfo{Status: model.TaskStatusSuccess, VideoUsage: tc.usage, TotalTokens: 99999}}
				ch := &model.Channel{Id: channelID, Key: "test"}
				require.NoError(t, updateVideoSingleTask(context.Background(), adaptor, ch, task.TaskID, map[string]*model.Task{task.TaskID: task}))
				var reloaded model.Task
				require.NoError(t, model.DB.First(&reloaded, task.ID).Error)
				require.Equal(t, actual, reloaded.Quota)
				require.Equal(t, tc.seconds, reloaded.PrivateData.BillingContext.VideoBilling.Usage.Seconds)
				if source == BillingSourceWallet {
					require.Equal(t, 1000000+pre-actual, getUserQuota(t, userID))
				} else {
					require.Equal(t, int64(actual), getSubscriptionUsed(t, subID))
					require.Equal(t, 1000000, getUserQuota(t, userID))
				}
				require.Equal(t, 1000000+pre-actual, getTokenRemainQuota(t, tokenID))
				var user model.User
				require.NoError(t, model.DB.First(&user, userID).Error)
				require.Equal(t, actual, user.UsedQuota)
				require.Equal(t, 1, user.RequestCount)
				balance, logs := getTokenRemainQuota(t, tokenID), countLogs(t)
				require.NoError(t, updateVideoSingleTask(context.Background(), adaptor, ch, task.TaskID, map[string]*model.Task{task.TaskID: &stale}))
				require.NoError(t, updateVideoSingleTask(context.Background(), adaptor, ch, task.TaskID, map[string]*model.Task{task.TaskID: &reloaded}))
				require.Equal(t, balance, getTokenRemainQuota(t, tokenID))
				require.Equal(t, logs, countLogs(t))
				if actual != pre {
					log := getLastLog(t)
					require.Contains(t, log.Other, `"billing_mode":"video_per_second"`)
				}
			})
		}
	}
}

func TestVideoBillingFreeTierAndFailureRefund(t *testing.T) {
	for _, failed := range []bool{false, true} {
		name := "free-tier"
		if failed {
			name = "failure"
		}
		t.Run(name, func(t *testing.T) {
			truncate(t)
			seedUser(t, 111, 1000000)
			seedChannel(t, 111)
			pre := int(8 * .1 * common.QuotaPerUnit)
			require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", 111).Update("used_quota", pre).Error)
			require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", 111).Update("used_quota", pre).Error)
			task := makeTask(111, 111, pre, 0, BillingSourceWallet, 0)
			task.PrivateData.BillingContext.VideoBilling = &types.VideoBillingSnapshot{Pricing: types.VideoPricing{PricePerSecond: .1, Tiers: []types.VideoPriceTier{{Resolution: "720p", PricePerSecond: 0}}}, Usage: types.VideoUsage{Seconds: 8, Resolution: "1080p"}}
			require.NoError(t, model.DB.Create(task).Error)
			result := relaycommon.TaskInfo{Status: model.TaskStatusSuccess, VideoUsage: &types.VideoUsage{Seconds: 6, Resolution: "720p"}}
			if failed {
				result.Status = model.TaskStatusFailure
				result.Reason = "vendor failed"
			}
			adaptor := &videoPollingAdaptor{result: result}
			require.NoError(t, updateVideoSingleTask(context.Background(), adaptor, &model.Channel{Id: 111}, task.TaskID, map[string]*model.Task{task.TaskID: task}))
			require.Equal(t, 1000000+pre, getUserQuota(t, 111))
			var reloaded model.Task
			require.NoError(t, model.DB.First(&reloaded, task.ID).Error)
			require.Zero(t, reloaded.Quota)
			log := getLastLog(t)
			require.Equal(t, model.LogTypeRefund, log.Type)
		})
	}
}

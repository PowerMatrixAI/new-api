package relay

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relay/channel/task/ali"
	"github.com/QuantumNous/new-api/relay/channel/task/doubao"
	"github.com/QuantumNous/new-api/relay/channel/task/gemini"
	"github.com/QuantumNous/new-api/relay/channel/task/hailuo"
	"github.com/QuantumNous/new-api/relay/channel/task/kling"
	"github.com/QuantumNous/new-api/relay/channel/task/sora"
	"github.com/QuantumNous/new-api/relay/channel/task/vertex"
	"github.com/QuantumNous/new-api/relay/channel/task/vidu"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVideoPricingReservationMatchesProviderRequest(t *testing.T) {
	before := ratio_setting.VideoPrice2JSONString()
	t.Cleanup(func() { _ = ratio_setting.UpdateVideoPriceByJSONString(before) })
	require.NoError(t, ratio_setting.UpdateVideoPriceByJSONString(`{"video-test":{"price_per_second":0.1}}`))
	for _, adaptor := range []videoUsageAdaptor{
		&sora.TaskAdaptor{}, &gemini.TaskAdaptor{}, &vertex.TaskAdaptor{}, &ali.TaskAdaptor{},
		&doubao.TaskAdaptor{}, &vidu.TaskAdaptor{}, &kling.TaskAdaptor{}, &hailuo.TaskAdaptor{},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/videos", nil)
		c.Set("task_request", relaycommon.TaskSubmitReq{Model: "video-test", Prompt: "test", Seconds: "8"})
		info := &relaycommon.RelayInfo{OriginModelName: "video-test", UsingGroup: "default", UserGroup: "default", ChannelMeta: &relaycommon.ChannelMeta{}, TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
		info.UpstreamModelName = "wan2.5-t2v-preview"
		price, err := taskPriceData(c, info, adaptor)
		require.NoError(t, err, "%T", adaptor)
		require.Equal(t, int(.8*common.QuotaPerUnit), price.Quota, "%T", adaptor)
		require.NotNil(t, price.VideoBilling)
		require.Equal(t, float64(8), price.VideoBilling.Usage.Seconds)
		// Gemini and Vertex previously priced Seconds without forwarding it.
		if builder, ok := adaptor.(interface {
			BuildRequestBody(*gin.Context, *relaycommon.RelayInfo) (io.Reader, error)
		}); ok {
			if _, isGemini := adaptor.(*gemini.TaskAdaptor); isGemini {
				info.PriceData = price
				body, err := builder.BuildRequestBody(c, info)
				require.NoError(t, err)
				var payload gemini.VeoRequestPayload
				require.NoError(t, common.DecodeJson(body, &payload))
				require.Equal(t, 8, payload.Parameters.DurationSeconds)
			}
		}
	}
}

func TestVideoPricingRejectsDurationBypasses(t *testing.T) {
	for _, raw := range []string{
		`{"duration":0}`, `{"duration":-1}`, `{"seconds":"999999999999999999999"}`,
		`{"metadata":{"parameters":{"duration":3601}}}`,
		`{"metadata":{"durationSeconds":0}}`, `{"seconds":"NaN"}`, `{"seconds":"8.5"}`,
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		var req relaycommon.TaskSubmitReq
		require.NoError(t, common.UnmarshalJsonStr(raw, &req))
		c.Set("task_request", req)
		require.Error(t, normalizeVideoDuration(c), raw)
	}
	// An unconfigured video keeps the legacy per-call pricing path.
	before := ratio_setting.ModelPrice2JSONString()
	t.Cleanup(func() { _ = ratio_setting.UpdateModelPriceByJSONString(before) })
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"legacy-video":0.25}`))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	price, err := taskPriceData(c, &relaycommon.RelayInfo{OriginModelName: "legacy-video", UsingGroup: "default"}, struct{}{})
	require.NoError(t, err)
	require.Nil(t, price.VideoBilling)
	require.Equal(t, int(.25*common.QuotaPerUnit), price.Quota)
	_, err = (types.VideoPricing{PricePerSecond: .1}).Quota(types.VideoUsage{Seconds: 8}, -1)
	require.Error(t, err)
}

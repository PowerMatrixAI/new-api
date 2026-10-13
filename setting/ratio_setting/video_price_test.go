package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVideoPriceValidationAndSnapshotIsolation(t *testing.T) {
	previous := VideoPrice2JSONString()
	t.Cleanup(func() { _ = UpdateVideoPriceByJSONString(previous) })
	valid := `{"video-test":{"price_per_second":0.1,"tiers":[{"resolution":"1080P","generate_audio":false,"price_per_second":0}]}}`
	require.NoError(t, UpdateVideoPriceByJSONString(valid))
	frozen, ok := GetVideoPrice("video-test")
	require.True(t, ok)
	for _, raw := range []string{
		`null`, `[]`, `{"video-test":{}}`, `{"video-test":{"price_per_second":null}}`,
		`{"video-test":{"price_per_second":-1}}`,
		`{"video-test":{"price_per_second":0.1,"tiers":[{"resolution":"1080p"}]}}`,
		`{"video-test":{"price_per_second":0.1,"tiers":[{"price_per_second":1}]}}`,
		`{"video-test":{"price_per_second":0.1,"tiers":[{"resolution":"1080P","price_per_second":1},{"resolution":"1080p","price_per_second":2}]}}`,
	} {
		require.Error(t, UpdateVideoPriceByJSONString(raw), raw)
		still, _ := GetVideoPrice("video-test")
		require.Equal(t, frozen, still)
	}
	copy, _ := GetVideoPrice("video-test")
	copy.Tiers[0].PricePerSecond = 99
	*copy.Tiers[0].GenerateAudio = true
	still, _ := GetVideoPrice("video-test")
	require.Equal(t, frozen, still)
	require.NoError(t, UpdateVideoPriceByJSONString(`{"video-test":{"price_per_second":9}}`))
	require.Equal(t, .1, frozen.PricePerSecond)
	free := `{"video-test":{"price_per_second":0}}`
	require.NoError(t, UpdateVideoPriceByJSONString(free))
}

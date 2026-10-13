package types

import (
	"math"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestVideoPricingUnitsAndTierPrecedence(t *testing.T) {
	audio, silent := true, false
	pricing := VideoPricing{PricePerSecond: .1, Tiers: []VideoPriceTier{
		{GenerateAudio: &audio, PricePerSecond: .2},
		{Resolution: "1080p", GenerateAudio: &audio, PricePerSecond: .4},
		{Resolution: "1080p", PricePerSecond: .3},
		{Resolution: "720p", GenerateAudio: &silent, PricePerSecond: 0},
	}}
	for _, tc := range []struct {
		usage VideoUsage
		price float64
	}{
		{VideoUsage{Seconds: 8}, .1},
		{VideoUsage{Seconds: 8, GenerateAudio: &audio}, .2},
		{VideoUsage{Seconds: 8, Resolution: "1920*1080", GenerateAudio: &silent}, .3},
		{VideoUsage{Seconds: 8, Resolution: "1080P", GenerateAudio: &audio}, .4},
		{VideoUsage{Seconds: 8, Resolution: "720p", GenerateAudio: &silent}, 0},
	} {
		quota, err := pricing.Quota(tc.usage, 1.5)
		require.NoError(t, err)
		require.Equal(t, int(math.Round(8*tc.price*common.QuotaPerUnit*1.5)), quota)
	}
	data, err := common.Marshal(pricing)
	require.NoError(t, err)
	require.Contains(t, string(data), `"generate_audio":false`)
	require.Contains(t, string(data), `"price_per_second":0`)
}

func TestVideoPricingRejectsUnsafeQuantities(t *testing.T) {
	p := VideoPricing{PricePerSecond: .1}
	for _, duration := range []float64{0, -1, 3601, math.NaN(), math.Inf(1), math.MaxFloat64} {
		_, err := p.Quota(VideoUsage{Seconds: duration}, 1)
		require.Error(t, err)
	}
	for _, ratio := range []float64{-1, math.NaN(), math.Inf(1), math.MaxFloat64} {
		_, err := p.Quota(VideoUsage{Seconds: 8}, ratio)
		require.Error(t, err)
	}
	_, err := (VideoPricing{PricePerSecond: math.MaxFloat64}).Quota(VideoUsage{Seconds: 8}, 1)
	require.Error(t, err)
}

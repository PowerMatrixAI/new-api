package ratio_setting

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
)

var videoPrices = struct {
	sync.RWMutex
	data map[string]types.VideoPricing
}{data: make(map[string]types.VideoPricing)}

func parseVideoPrices(raw string) (map[string]types.VideoPricing, error) {
	var rawPrices map[string]map[string]any
	if err := common.UnmarshalJsonStr(raw, &rawPrices); err != nil {
		return nil, err
	}
	for name, fields := range rawPrices {
		if _, ok := fields["price_per_second"].(float64); !ok {
			return nil, fmt.Errorf("model %q requires an explicit price_per_second", name)
		}
		if tiers, ok := fields["tiers"].([]any); ok {
			for _, rawTier := range tiers {
				tier, ok := rawTier.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid video price tier for model %q", name)
				}
				if _, ok := tier["price_per_second"].(float64); !ok {
					return nil, fmt.Errorf("video price tier for model %q requires an explicit price_per_second", name)
				}
			}
		}
	}
	var prices map[string]types.VideoPricing
	if err := common.UnmarshalJsonStr(raw, &prices); err != nil {
		return nil, err
	}
	if prices == nil {
		return nil, fmt.Errorf("video prices must be a JSON object")
	}
	validPrice := func(price float64) bool { return !math.IsNaN(price) && !math.IsInf(price, 0) && price >= 0 }
	for name, pricing := range prices {
		if strings.TrimSpace(name) == "" || !validPrice(pricing.PricePerSecond) || len(pricing.Tiers) > 64 {
			return nil, fmt.Errorf("invalid video pricing for model %q", name)
		}
		seen := make(map[string]bool)
		for i := range pricing.Tiers {
			tier := &pricing.Tiers[i]
			tier.Resolution = types.NormalizeVideoResolution(tier.Resolution)
			audio := "any"
			if tier.GenerateAudio != nil {
				audio = fmt.Sprint(*tier.GenerateAudio)
			}
			key := tier.Resolution + ":" + audio
			if (tier.Resolution == "" && tier.GenerateAudio == nil) || !validPrice(tier.PricePerSecond) || seen[key] {
				return nil, fmt.Errorf("invalid or duplicate video price tier for model %q", name)
			}
			seen[key] = true
		}
		prices[name] = pricing
	}
	return prices, nil
}

func CheckVideoPrice(raw string) error {
	_, err := parseVideoPrices(raw)
	return err
}

func UpdateVideoPriceByJSONString(raw string) error {
	prices, err := parseVideoPrices(raw)
	if err != nil {
		return err
	}
	videoPrices.Lock()
	videoPrices.data = prices
	videoPrices.Unlock()
	InvalidateExposedDataCache()
	return nil
}

func GetVideoPrice(name string) (types.VideoPricing, bool) {
	videoPrices.RLock()
	defer videoPrices.RUnlock()
	pricing, ok := videoPrices.data[FormatMatchingModelName(name)]
	return pricing.Clone(), ok
}

func GetVideoPriceCopy() map[string]types.VideoPricing {
	videoPrices.RLock()
	defer videoPrices.RUnlock()
	prices := make(map[string]types.VideoPricing, len(videoPrices.data))
	for name, pricing := range videoPrices.data {
		prices[name] = pricing.Clone()
	}
	return prices
}

func VideoPrice2JSONString() string {
	data, _ := common.Marshal(GetVideoPriceCopy())
	return string(data)
}

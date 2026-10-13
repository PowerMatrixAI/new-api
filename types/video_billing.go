package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

const MaxVideoBillingSeconds = 3600

// VideoPricing is an explicit USD/second tariff. It does not use legacy token
// or resolution multipliers; each matching tier is a complete unit price.
type VideoPricing struct {
	PricePerSecond float64          `json:"price_per_second"`
	Tiers          []VideoPriceTier `json:"tiers,omitempty"`
}

type VideoPriceTier struct {
	Resolution     string  `json:"resolution,omitempty"`
	GenerateAudio  *bool   `json:"generate_audio,omitempty"`
	PricePerSecond float64 `json:"price_per_second"`
}

// Completion fields are optional: missing facts retain the submission estimate.
type VideoUsage struct {
	Seconds       float64 `json:"seconds,omitempty"`
	Resolution    string  `json:"resolution,omitempty"`
	GenerateAudio *bool   `json:"generate_audio,omitempty"`
}

type VideoBillingSnapshot struct {
	Pricing VideoPricing `json:"pricing"`
	Usage   VideoUsage   `json:"usage"`
}

func (p VideoPricing) Clone() VideoPricing {
	p.Tiers = append([]VideoPriceTier(nil), p.Tiers...)
	for i := range p.Tiers {
		if p.Tiers[i].GenerateAudio != nil {
			value := *p.Tiers[i].GenerateAudio
			p.Tiers[i].GenerateAudio = &value
		}
	}
	return p
}

func NormalizeVideoResolution(resolution string) string {
	resolution = strings.ToLower(strings.TrimSpace(resolution))
	parts := strings.Split(strings.ReplaceAll(resolution, "*", "x"), "x")
	if len(parts) == 2 {
		w, e1 := strconv.Atoi(parts[0])
		h, e2 := strconv.Atoi(parts[1])
		if e1 == nil && e2 == nil && w > 0 && h > 0 {
			if w > h {
				w = h
			}
			return strconv.Itoa(w) + "p"
		}
	}
	return resolution
}

// Resolution-specific rows take precedence over audio-only rows. A row with
// both conditions is most specific, independent of its position in the list.
func (p VideoPricing) UnitPrice(usage VideoUsage) float64 {
	price, priority := p.PricePerSecond, 0
	for _, tier := range p.Tiers {
		score := 0
		if tier.Resolution != "" {
			if NormalizeVideoResolution(tier.Resolution) != NormalizeVideoResolution(usage.Resolution) {
				continue
			}
			score += 2
		}
		if tier.GenerateAudio != nil {
			if usage.GenerateAudio == nil || *tier.GenerateAudio != *usage.GenerateAudio {
				continue
			}
			score++
		}
		if score > priority {
			price, priority = tier.PricePerSecond, score
		}
	}
	return price
}

func ValidVideoSeconds(seconds float64) bool {
	return !math.IsNaN(seconds) && !math.IsInf(seconds, 0) && seconds > 0 && seconds <= MaxVideoBillingSeconds
}

func (p VideoPricing) Quota(usage VideoUsage, groupRatio float64) (int, error) {
	if !ValidVideoSeconds(usage.Seconds) {
		return 0, fmt.Errorf("video duration must be between 0 and %d seconds", MaxVideoBillingSeconds)
	}
	quota := usage.Seconds * p.UnitPrice(usage) * common.QuotaPerUnit * groupRatio
	// Match the upstream host's int32 quota bound; reject overflow before casting.
	if p.UnitPrice(usage) < 0 || groupRatio < 0 || math.IsNaN(quota) || math.IsInf(quota, 0) || quota < 0 || quota > math.MaxInt32 {
		return 0, fmt.Errorf("video billing quota is out of range")
	}
	return int(math.Round(quota)), nil
}

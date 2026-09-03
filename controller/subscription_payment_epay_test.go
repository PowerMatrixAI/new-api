package controller

import (
	"math"
	"testing"

	"github.com/QuantumNous/new-api/setting/operation_setting"
)

func TestGetSubscriptionEpayPayMoneyUsesDisplayedCNYAmount(t *testing.T) {
	generalSetting := operation_setting.GetGeneralSetting()
	originalDisplayType := generalSetting.QuotaDisplayType
	originalExchangeRate := operation_setting.USDExchangeRate
	t.Cleanup(func() {
		generalSetting.QuotaDisplayType = originalDisplayType
		operation_setting.USDExchangeRate = originalExchangeRate
	})

	generalSetting.QuotaDisplayType = operation_setting.QuotaDisplayTypeCNY
	operation_setting.USDExchangeRate = 6.8

	got := getSubscriptionEpayPayMoney(299)
	if math.Abs(got-2033.2) > 1e-9 {
		t.Fatalf("expected CNY payment amount 2033.2, got %v", got)
	}
}

func TestGetSubscriptionEpayPayMoneyKeepsUSDAmountOutsideCNYDisplay(t *testing.T) {
	generalSetting := operation_setting.GetGeneralSetting()
	originalDisplayType := generalSetting.QuotaDisplayType
	originalExchangeRate := operation_setting.USDExchangeRate
	t.Cleanup(func() {
		generalSetting.QuotaDisplayType = originalDisplayType
		operation_setting.USDExchangeRate = originalExchangeRate
	})

	generalSetting.QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	operation_setting.USDExchangeRate = 6.8

	if got := getSubscriptionEpayPayMoney(299); got != 299 {
		t.Fatalf("expected USD payment amount 299, got %v", got)
	}
}

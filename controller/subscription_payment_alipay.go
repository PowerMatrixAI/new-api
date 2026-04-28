package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
)

type SubscriptionAlipayPayRequest struct {
	PlanId int `json:"plan_id"`
}

// SubscriptionRequestAlipayPay 发起订阅支付宝支付
func SubscriptionRequestAlipayPay(c *gin.Context) {
	var req SubscriptionAlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if setting.AlipayAppId == "" || setting.AlipayPrivateKey == "" || setting.AlipayPublicKey == "" {
		common.ApiErrorMsg(c, "当前管理员未配置支付宝信息")
		return
	}

	plan, err := model.GetSubscriptionPlanById(req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if !plan.Enabled {
		common.ApiErrorMsg(c, "套餐未启用")
		return
	}
	if plan.PriceAmount < 0.01 {
		common.ApiErrorMsg(c, "套餐金额过低")
		return
	}

	userId := c.GetInt("id")
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			common.ApiErrorMsg(c, "已达到该套餐购买上限")
			return
		}
	}

	tradeNo := fmt.Sprintf("SUBUSR%dNO%s%s", userId, common.GetRandomString(6), randstr.String(4))
	notifyUrl := system_setting.ServerAddress + "/api/subscription/alipay/notify"
	returnUrl := system_setting.ServerAddress + "/console/topup?pay=success"

	// 套餐价格单位是 USD，支付宝收人民币，按 AlipayUnitPrice 换算
	payMoney := plan.PriceAmount * setting.AlipayUnitPrice

	payLink, err := genAlipayPagePayLink(tradeNo, payMoney, notifyUrl, returnUrl)
	if err != nil {
		common.SysLog(fmt.Sprintf("生成订阅支付宝支付链接失败: %v", err))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	order := &model.SubscriptionOrder{
		UserId:        userId,
		PlanId:        plan.Id,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodAlipay,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_url": payLink,
		},
	})
}

// SubscriptionAlipayNotify 支付宝订阅异步回调
func SubscriptionAlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		common.SysLog(fmt.Sprintf("支付宝订阅回调解析失败: %v", err))
		_, _ = c.Writer.WriteString("fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if !alipayVerifySign(params) {
		common.SysLog("支付宝订阅回调验签失败")
		_, _ = c.Writer.WriteString("fail")
		return
	}

	tradeStatus := params["trade_status"]
	outTradeNo := params["out_trade_no"]

	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		_, _ = c.Writer.WriteString("success")
		return
	}

	LockOrder(outTradeNo)
	defer UnlockOrder(outTradeNo)

	if err := model.CompleteSubscriptionOrder(outTradeNo, common.GetJsonString(params)); err != nil {
		common.SysLog(fmt.Sprintf("支付宝订阅回调完成订单失败: %v", err))
		_, _ = c.Writer.WriteString("fail")
		return
	}

	_, _ = c.Writer.WriteString("success")
}

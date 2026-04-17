package controller

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/thanhpk/randstr"
)

const PaymentMethodAlipay = "alipay_native"

// alipayGateway 支付宝网关地址
func alipayGateway() string {
	if setting.AlipaySandbox {
		return "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
	}
	return "https://openapi.alipay.com/gateway.do"
}

// AlipayPayRequest 支付宝充值请求
type AlipayPayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

func RequestAlipayAmount(c *gin.Context) {
	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	minTopup := getAlipayMinTopup()
	if req.Amount < minTopup {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", minTopup)})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getAlipayPayMoney(float64(req.Amount), group)
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}
	c.JSON(200, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

func RequestAlipayPay(c *gin.Context) {
	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	if req.PaymentMethod != PaymentMethodAlipay {
		c.JSON(200, gin.H{"message": "error", "data": "不支持的支付渠道"})
		return
	}

	minTopup := getAlipayMinTopup()
	if req.Amount < minTopup {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", minTopup)})
		return
	}

	if setting.AlipayAppId == "" || setting.AlipayPrivateKey == "" || setting.AlipayPublicKey == "" {
		c.JSON(200, gin.H{"message": "error", "data": "当前管理员未配置支付宝信息"})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getAlipayPayMoney(float64(req.Amount), group)
	if payMoney <= 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	tradeNo := fmt.Sprintf("USR%dNO%s%s", id, common.GetRandomString(6), randstr.String(4))
	notifyUrl := system_setting.ServerAddress + "/api/user/alipay/notify"
	returnUrl := system_setting.ServerAddress + "/console/log"

	payLink, err := genAlipayPagePayLink(tradeNo, payMoney, notifyUrl, returnUrl)
	if err != nil {
		common.SysLog(fmt.Sprintf("生成支付宝支付链接失败: %v", err))
		c.JSON(200, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(amount)
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	topUp := &model.TopUp{
		UserId:        id,
		Amount:        amount,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodAlipay,
		CreateTime:    time.Now().Unix(),
		Status:        common.TopUpStatusPending,
	}
	if err = topUp.Insert(); err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"pay_link": payLink,
		},
	})
}

// AlipayNotify 支付宝异步回调
func AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		common.SysLog(fmt.Sprintf("支付宝回调解析失败: %v", err))
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
		common.SysLog("支付宝回调验签失败")
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

	topUp := model.GetTopUpByTradeNo(outTradeNo)
	if topUp == nil {
		common.SysLog(fmt.Sprintf("支付宝回调未找到订单: %s", outTradeNo))
		_, _ = c.Writer.WriteString("fail")
		return
	}

	if topUp.Status == common.TopUpStatusPending {
		topUp.Status = common.TopUpStatusSuccess
		if err := topUp.Update(); err != nil {
			common.SysLog(fmt.Sprintf("支付宝回调更新订单失败: %v", topUp))
			_, _ = c.Writer.WriteString("fail")
			return
		}
		dAmount := decimal.NewFromInt(int64(topUp.Amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())
		if err := model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true); err != nil {
			common.SysLog(fmt.Sprintf("支付宝回调更新用户配额失败: %v", topUp))
			_, _ = c.Writer.WriteString("fail")
			return
		}
		common.SysLog(fmt.Sprintf("支付宝回调充值成功 %v", topUp))
		model.RecordLog(topUp.UserId, model.LogTypeTopup,
			fmt.Sprintf("使用支付宝充值成功，充值金额: %v，支付金额：%.2f", logger.LogQuota(quotaToAdd), topUp.Money))
	}

	_, _ = c.Writer.WriteString("success")
}

// genAlipayPagePayLink 生成支付宝电脑网站支付跳转链接（alipay.trade.page.pay）
func genAlipayPagePayLink(outTradeNo string, totalAmount float64, notifyUrl, returnUrl string) (string, error) {
	bizContent := fmt.Sprintf(
		`{"out_trade_no":"%s","product_code":"FAST_INSTANT_TRADE_PAY","total_amount":"%.2f","subject":"在线充值"}`,
		outTradeNo, totalAmount,
	)

	params := map[string]string{
		"app_id":      setting.AlipayAppId,
		"method":      "alipay.trade.page.pay",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notifyUrl,
		"return_url":  returnUrl,
		"biz_content": bizContent,
	}

	sign, err := alipaySign(params)
	if err != nil {
		return "", err
	}
	params["sign"] = sign

	// 构建 GET 跳转链接
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	return alipayGateway() + "?" + query.Encode(), nil
}

// alipaySign 对参数按字典序排序后用 RSA2 签名
func alipaySign(params map[string]string) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	signStr := strings.Join(parts, "&")

	privateKey, err := parseRSAPrivateKey(setting.AlipayPrivateKey)
	if err != nil {
		return "", fmt.Errorf("解析支付宝私钥失败: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)

	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	if err != nil {
		return "", fmt.Errorf("RSA2 签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// alipayVerifySign 验证支付宝回调签名
func alipayVerifySign(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}

	// 过滤 sign 和 sign_type
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	signStr := strings.Join(parts, "&")

	publicKey, err := parseRSAPublicKey(setting.AlipayPublicKey)
	if err != nil {
		common.SysLog(fmt.Sprintf("解析支付宝公钥失败: %v", err))
		return false
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	h := sha256.New()
	h.Write([]byte(signStr))
	digest := h.Sum(nil)

	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest, sigBytes) == nil
}

func parseRSAPrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	keyStr = strings.TrimSpace(keyStr)
	if !strings.HasPrefix(keyStr, "-----") {
		keyStr = "-----BEGIN PRIVATE KEY-----\n" + keyStr + "\n-----END PRIVATE KEY-----"
	}
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, fmt.Errorf("无法解析 PEM 块")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("不是 RSA 私钥")
	}
	return rsaKey, nil
}

func parseRSAPublicKey(keyStr string) (*rsa.PublicKey, error) {
	keyStr = strings.TrimSpace(keyStr)
	if !strings.HasPrefix(keyStr, "-----") {
		keyStr = "-----BEGIN PUBLIC KEY-----\n" + keyStr + "\n-----END PUBLIC KEY-----"
	}
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, fmt.Errorf("无法解析 PEM 块")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是 RSA 公钥")
	}
	return rsaPub, nil
}

func getAlipayPayMoney(amount float64, group string) float64 {
	dAmount := decimal.NewFromFloat(amount)
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		dAmount = dAmount.Div(dQuotaPerUnit)
	}
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	discount := 1.0
	if ds, ok := setting.GetAlipayAmountDiscount()[int(amount)]; ok && ds > 0 {
		discount = ds
	}

	dUnitPrice := decimal.NewFromFloat(setting.AlipayUnitPrice)
	dTopupGroupRatio := decimal.NewFromFloat(topupGroupRatio)
	dDiscount := decimal.NewFromFloat(discount)

	payMoney := dAmount.Mul(dUnitPrice).Mul(dTopupGroupRatio).Mul(dDiscount)
	return payMoney.InexactFloat64()
}

func getAlipayMinTopup() int64 {
	minTopup := setting.AlipayMinTopUp
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dMinTopup := decimal.NewFromInt(int64(minTopup))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		minTopup = int(dMinTopup.Mul(dQuotaPerUnit).IntPart())
	}
	return int64(minTopup)
}

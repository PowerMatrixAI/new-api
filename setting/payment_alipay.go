package setting

import "github.com/QuantumNous/new-api/common"

// AlipayAppId 支付宝应用 APPID
var AlipayAppId = ""

// AlipayPrivateKey 应用私钥（RSA2 PKCS8 格式，不含头尾）
var AlipayPrivateKey = ""

// AlipayPublicKey 支付宝公钥（用于验签，不含头尾）
var AlipayPublicKey = ""

// AlipaySandbox 是否使用沙箱环境
var AlipaySandbox = false

// AlipayUnitPrice 充值价格(x 元)
var AlipayUnitPrice = 1.0

// AlipayMinTopUp 最低充值美元数量
var AlipayMinTopUp = 1

// AlipayAmountOptions 充值额度选项（逗号分隔，例如 "10,20,50,100"），为空则使用默认生成
var AlipayAmountOptions = ""

// AlipayAmountDiscount 充值金额折扣（JSON 格式，例如 {"100":0.9,"500":0.85}）
var AlipayAmountDiscount = "{}"

// GetAlipayAmountOptions 解析 AlipayAmountOptions 字符串为 []int
func GetAlipayAmountOptions() []int {
	if AlipayAmountOptions == "" {
		return nil
	}
	var result []int
	_ = common.UnmarshalJsonStr("["+AlipayAmountOptions+"]", &result)
	return result
}

// GetAlipayAmountDiscount 解析 AlipayAmountDiscount JSON 为 map[int]float64
func GetAlipayAmountDiscount() map[int]float64 {
	result := map[int]float64{}
	_ = common.UnmarshalJsonStr(AlipayAmountDiscount, &result)
	return result
}

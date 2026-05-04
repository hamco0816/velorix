package provider

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

// 虎皮椒（Xunhupay）支付服务商接入。
// 接口文档：https://www.xunhupay.com/doc/api/pay.html
//
// 用途：
//   - 个人微信、支付宝聚合收款，无需企业资质。
//   - 与 EasyPay 类似，账号在虎皮椒控制台开通后可立即使用。

// 虎皮椒接口地址、签名相关常量。
const (
	xunhupayDefaultAPIBase    = "https://api.xunhupay.com"
	xunhupayPathPay           = "/payment/do.html"
	xunhupayPathQuery         = "/payment/query.html"
	xunhupayHTTPTimeout       = 10 * time.Second
	xunhupayMaxResponseSize   = 1 << 20
	xunhupayAPIVersion        = "1.1"
	xunhupayWxPayChannel      = "wechat"
	xunhupayAlipayChannel     = "alipay"
	xunhupaySuccessACK        = "success"
	xunhupayStatusPaid        = "OD"
	xunhupayStatusRefunded    = "CD"
	xunhupayStatusRefunding   = "RD"
	xunhupayStatusRefundFail  = "UD"
	xunhupayResponseSummaryMx = 512
)

// Xunhupay 实现 payment.Provider，对接虎皮椒聚合支付平台。
type Xunhupay struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
}

// NewXunhupay 创建一个新的虎皮椒 Provider 实例。
// 必填 config：appid、appsecret、notifyUrl
// 可选 config：returnUrl、callbackUrl、apiBase
func NewXunhupay(instanceID string, config map[string]string) (*Xunhupay, error) {
	for _, k := range []string{"appid", "appsecret", "notifyUrl"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("xunhupay config missing required key: %s", k)
		}
	}
	cfg := make(map[string]string, len(config))
	for k, v := range config {
		cfg[k] = v
	}
	cfg["apiBase"] = normalizeXunhupayAPIBase(cfg["apiBase"])
	return &Xunhupay{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: xunhupayHTTPTimeout},
	}, nil
}

func normalizeXunhupayAPIBase(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return xunhupayDefaultAPIBase
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return xunhupayDefaultAPIBase
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	for _, suffix := range []string{xunhupayPathPay, xunhupayPathQuery} {
		if strings.HasSuffix(parsed.Path, suffix) {
			parsed.Path = strings.TrimSuffix(parsed.Path, suffix)
			parsed.Path = strings.TrimRight(parsed.Path, "/")
			break
		}
	}
	return strings.TrimRight(parsed.String(), "/")
}

func (x *Xunhupay) apiBase() string {
	if x == nil {
		return xunhupayDefaultAPIBase
	}
	if base := normalizeXunhupayAPIBase(x.config["apiBase"]); base != "" {
		return base
	}
	return xunhupayDefaultAPIBase
}

func (x *Xunhupay) Name() string        { return "Xunhupay" }
func (x *Xunhupay) ProviderKey() string { return payment.TypeXunhupay }
func (x *Xunhupay) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay, payment.TypeWxpay}
}

// MerchantIdentityMetadata 返回当前商户的可对外识别信息（用于一致性校验）。
func (x *Xunhupay) MerchantIdentityMetadata() map[string]string {
	if x == nil {
		return nil
	}
	appid := strings.TrimSpace(x.config["appid"])
	if appid == "" {
		return nil
	}
	return map[string]string{"appid": appid}
}

// CreatePayment 调用虎皮椒下单接口，返回跳转链接或扫码地址。
func (x *Xunhupay) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	notifyURL, returnURL := x.resolveURLs(req)
	channel := xunhupayPaymentChannel(req.PaymentType)
	if channel == "" {
		return nil, fmt.Errorf("xunhupay unsupported payment type: %s", req.PaymentType)
	}

	nonceStr, err := xunhupayNonce()
	if err != nil {
		return nil, fmt.Errorf("xunhupay nonce: %w", err)
	}

	params := map[string]string{
		"version":        xunhupayAPIVersion,
		"appid":          x.config["appid"],
		"trade_order_id": req.OrderID,
		"total_fee":      req.Amount,
		"title":          req.Subject,
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"notify_url":     notifyURL,
		"nonce_str":      nonceStr,
		"payment":        channel,
		"type":           xunhupayDeviceType(req.IsMobile),
	}
	if returnURL != "" {
		params["return_url"] = returnURL
	}
	if cb := strings.TrimSpace(x.config["callbackUrl"]); cb != "" {
		params["callback_url"] = cb
	}
	if req.ClientIP != "" {
		params["wap_url"] = req.ClientIP
	}
	params["hash"] = xunhupaySign(params, x.config["appsecret"])

	body, err := x.post(ctx, x.apiBase()+xunhupayPathPay, params)
	if err != nil {
		return nil, fmt.Errorf("xunhupay create: %w", err)
	}

	var resp struct {
		OpenID    string          `json:"openid"`
		URL       string          `json:"url"`
		URLQrcode string          `json:"url_qrcode"`
		Hash      string          `json:"hash"`
		ErrCode   json.RawMessage `json:"errcode"`
		ErrMsg    string          `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("xunhupay parse: %s", xunhupayResponseSummary(body))
	}
	if !xunhupayCodeIsSuccess(resp.ErrCode) {
		return nil, fmt.Errorf("xunhupay error: %s", strings.TrimSpace(resp.ErrMsg))
	}

	payURL := resp.URL
	qrCode := resp.URLQrcode
	if req.IsMobile && payURL == "" {
		payURL = qrCode
	}
	tradeNo := strings.TrimSpace(resp.OpenID)
	return &payment.CreatePaymentResponse{
		TradeNo: tradeNo,
		PayURL:  payURL,
		QRCode:  qrCode,
	}, nil
}

// QueryOrder 通过 query.html 接口查询订单状态。
func (x *Xunhupay) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	nonceStr, err := xunhupayNonce()
	if err != nil {
		return nil, fmt.Errorf("xunhupay nonce: %w", err)
	}
	params := map[string]string{
		"appid":          x.config["appid"],
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"nonce_str":      nonceStr,
		"trade_order_id": tradeNo,
	}
	params["hash"] = xunhupaySign(params, x.config["appsecret"])

	body, err := x.post(ctx, x.apiBase()+xunhupayPathQuery, params)
	if err != nil {
		return nil, fmt.Errorf("xunhupay query: %w", err)
	}
	var resp struct {
		Data struct {
			Status        string `json:"status"`
			TotalFee      string `json:"total_fee"`
			TransactionID string `json:"transaction_id"`
			OpenOrderID   string `json:"open_order_id"`
			Time          string `json:"time"`
		} `json:"data"`
		ErrCode json.RawMessage `json:"errcode"`
		ErrMsg  string          `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("xunhupay parse query: %s", xunhupayResponseSummary(body))
	}
	if !xunhupayCodeIsSuccess(resp.ErrCode) {
		return nil, fmt.Errorf("xunhupay query failed: %s", strings.TrimSpace(resp.ErrMsg))
	}
	amount, _ := strconv.ParseFloat(resp.Data.TotalFee, 64)
	return &payment.QueryOrderResponse{
		TradeNo:  resp.Data.OpenOrderID,
		Status:   xunhupayMapStatus(resp.Data.Status),
		Amount:   amount,
		Metadata: x.MerchantIdentityMetadata(),
	}, nil
}

// VerifyNotification 校验异步通知签名并解析订单。
func (x *Xunhupay) VerifyNotification(_ context.Context, rawBody string, _ map[string]string) (*payment.PaymentNotification, error) {
	params, err := xunhupayParseNotifyBody(rawBody)
	if err != nil {
		return nil, fmt.Errorf("xunhupay parse notify: %w", err)
	}
	sign := strings.TrimSpace(params["hash"])
	if sign == "" {
		return nil, fmt.Errorf("xunhupay missing hash")
	}
	expected := xunhupaySign(params, x.config["appsecret"])
	if !strings.EqualFold(sign, expected) {
		return nil, fmt.Errorf("xunhupay invalid signature")
	}

	status := payment.ProviderStatusFailed
	switch params["status"] {
	case xunhupayStatusPaid:
		status = payment.ProviderStatusSuccess
	case xunhupayStatusRefunded:
		status = payment.ProviderStatusRefunded
	}

	amount, _ := strconv.ParseFloat(params["total_fee"], 64)
	metadata := x.MerchantIdentityMetadata()
	if appid := strings.TrimSpace(params["appid"]); appid != "" {
		if metadata == nil {
			metadata = map[string]string{}
		}
		metadata["appid"] = appid
	}
	return &payment.PaymentNotification{
		TradeNo:  params["transaction_id"],
		OrderID:  params["trade_order_id"],
		Amount:   amount,
		Status:   status,
		RawData:  rawBody,
		Metadata: metadata,
	}, nil
}

// Refund 暂未对接虎皮椒的退款 API（虎皮椒的退款仅在管理后台手工处理）。
func (x *Xunhupay) Refund(_ context.Context, _ payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("xunhupay refund not supported by API; please refund in xunhupay merchant dashboard")
}

func (x *Xunhupay) resolveURLs(req payment.CreatePaymentRequest) (string, string) {
	notifyURL := req.NotifyURL
	if notifyURL == "" {
		notifyURL = x.config["notifyUrl"]
	}
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = x.config["returnUrl"]
	}
	return notifyURL, returnURL
}

// xunhupayPaymentChannel 把内部支付类型映射为虎皮椒支持的支付通道。
func xunhupayPaymentChannel(paymentType string) string {
	switch payment.GetBasePaymentType(paymentType) {
	case payment.TypeAlipay:
		return xunhupayAlipayChannel
	case payment.TypeWxpay:
		return xunhupayWxPayChannel
	}
	return ""
}

func xunhupayDeviceType(isMobile bool) string {
	if isMobile {
		return "WAP"
	}
	return "PC"
}

func xunhupayMapStatus(raw string) string {
	switch raw {
	case xunhupayStatusPaid:
		return payment.ProviderStatusPaid
	case xunhupayStatusRefunded:
		return payment.ProviderStatusRefunded
	case xunhupayStatusRefunding, xunhupayStatusRefundFail:
		return payment.ProviderStatusPending
	}
	return payment.ProviderStatusPending
}

// xunhupaySign 按虎皮椒规则生成签名：
//
//	1. 取出全部非空参数，剔除 hash 字段；
//	2. 按 key 的 ASCII 升序排序；
//	3. 用 "k1=v1&k2=v2" 形式拼接；
//	4. 末尾直接拼接 appsecret；
//	5. 取 32 位小写 MD5。
func xunhupaySign(params map[string]string, appsecret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "hash" || strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			_ = buf.WriteByte('&')
		}
		_, _ = buf.WriteString(k)
		_ = buf.WriteByte('=')
		_, _ = buf.WriteString(params[k])
	}
	_, _ = buf.WriteString(appsecret)
	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}

func xunhupayNonce() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// xunhupayCodeIsSuccess 适配 errcode 既可能是数字也可能是字符串的场景。
func xunhupayCodeIsSuccess(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	trimmed := strings.TrimSpace(string(raw))
	switch trimmed {
	case "0", `"0"`, `""`:
		return true
	}
	if n, err := strconv.Atoi(strings.Trim(trimmed, `"`)); err == nil && n == 0 {
		return true
	}
	return false
}

// xunhupayParseNotifyBody 同时兼容表单与 JSON 两种回调形式（虎皮椒后台可配置）。
func xunhupayParseNotifyBody(rawBody string) (map[string]string, error) {
	body := strings.TrimSpace(rawBody)
	params := make(map[string]string)
	if body == "" {
		return params, nil
	}
	if strings.HasPrefix(body, "{") {
		var raw map[string]any
		if err := json.Unmarshal([]byte(body), &raw); err != nil {
			return nil, err
		}
		for k, v := range raw {
			params[k] = xunhupayStringify(v)
		}
		return params, nil
	}
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}
	for k := range values {
		params[k] = values.Get(k)
	}
	return params, nil
}

func xunhupayStringify(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case json.Number:
		return val.String()
	default:
		out, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(out)
	}
}

// post 以 application/x-www-form-urlencoded 提交参数。
func (x *Xunhupay) post(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := x.httpClient
	if client == nil {
		client = &http.Client{Timeout: xunhupayHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, xunhupayMaxResponseSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("xunhupay HTTP %d: %s", resp.StatusCode, xunhupayResponseSummary(body))
	}
	return body, nil
}

func xunhupayResponseSummary(body []byte) string {
	summary := strings.Join(strings.Fields(string(body)), " ")
	if summary == "" {
		return "<empty>"
	}
	if len(summary) > xunhupayResponseSummaryMx {
		return summary[:xunhupayResponseSummaryMx] + "..."
	}
	return summary
}

// XunhupaySuccessResponse 异步通知验签成功后回写给虎皮椒的固定响应。
const XunhupaySuccessResponse = xunhupaySuccessACK

// 接口实现保证。
var (
	_ payment.Provider                 = (*Xunhupay)(nil)
	_ payment.MerchantIdentityProvider = (*Xunhupay)(nil)
)

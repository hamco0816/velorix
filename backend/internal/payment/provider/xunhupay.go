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
	xunhupayDefaultAPIBase       = "https://api.xunhupay.com"
	xunhupayPathPay              = "/payment/do.html"
	xunhupayPathQuery            = "/payment/query.html"
	xunhupayHTTPTimeout          = 10 * time.Second
	xunhupayMaxResponseSize      = 1 << 20
	xunhupayAPIVersion           = "1.1"
	xunhupayWxPayChannel         = "wechat"
	xunhupayAlipayChannel        = "alipay"
	xunhupaySuccessACK           = "success"
	xunhupayNotifyStatusPaid     = "OD"
	xunhupayNotifyStatusRefunded = "CD"
	xunhupayQueryStatusPaid      = "OD"
	xunhupayQueryStatusPending   = "WP"
	xunhupayQueryStatusClosed    = "CD"
	xunhupayStatusRefunding      = "RD"
	xunhupayStatusRefundFail     = "UD"
	xunhupayResponseSummaryMx    = 4096
)

type xunhupayCredential struct {
	AppID     string
	AppSecret string
	Scope     string
}

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
	if strings.TrimSpace(config["notifyUrl"]) == "" {
		return nil, fmt.Errorf("xunhupay config missing required key: notifyUrl")
	}
	if len(xunhupayCredentials(config, "")) == 0 {
		return nil, fmt.Errorf("xunhupay config requires at least one appid/appsecret pair")
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

func xunhupayConfigValue(config map[string]string, keys ...string) string {
	for _, key := range keys {
		for actualKey, value := range config {
			if strings.EqualFold(actualKey, key) {
				if trimmed := strings.TrimSpace(value); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}

func xunhupayCredentialFrom(config map[string]string, scope string, appKeys []string, secretKeys []string) (xunhupayCredential, bool) {
	cred := xunhupayCredential{
		AppID:     xunhupayConfigValue(config, appKeys...),
		AppSecret: xunhupayConfigValue(config, secretKeys...),
		Scope:     scope,
	}
	return cred, cred.AppID != "" && cred.AppSecret != ""
}

func xunhupayCredentials(config map[string]string, paymentType string) []xunhupayCredential {
	seen := map[string]struct{}{}
	add := func(out *[]xunhupayCredential, cred xunhupayCredential) {
		if cred.AppID == "" || cred.AppSecret == "" {
			return
		}
		key := cred.Scope + "\x00" + cred.AppID
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		*out = append(*out, cred)
	}

	var credentials []xunhupayCredential
	baseType := payment.GetBasePaymentType(strings.TrimSpace(paymentType))
	alipayCred, hasAlipay := xunhupayCredentialFrom(config, payment.TypeAlipay,
		[]string{"alipayAppId", "alipayAppID", "alipayAppid", "appidAlipay", "appIdAlipay"},
		[]string{"alipayAppSecret", "alipayAppsecret", "appsecretAlipay", "appSecretAlipay"},
	)
	wxpayCred, hasWxpay := xunhupayCredentialFrom(config, payment.TypeWxpay,
		[]string{"wxpayAppId", "wxpayAppID", "wxpayAppid", "wechatAppId", "wechatAppID", "appidWxpay", "appIdWxpay"},
		[]string{"wxpayAppSecret", "wxpayAppsecret", "wechatAppSecret", "appsecretWxpay", "appSecretWxpay"},
	)
	legacyCred, hasLegacy := xunhupayCredentialFrom(config, "legacy",
		[]string{"appid", "appId", "appID"},
		[]string{"appsecret", "appSecret", "appSECRET"},
	)

	switch baseType {
	case payment.TypeAlipay:
		if hasAlipay {
			add(&credentials, alipayCred)
		}
		if hasLegacy {
			add(&credentials, legacyCred)
		}
	case payment.TypeWxpay:
		if hasWxpay {
			add(&credentials, wxpayCred)
		}
		if hasLegacy {
			add(&credentials, legacyCred)
		}
	default:
		if hasLegacy {
			add(&credentials, legacyCred)
		}
		if hasAlipay {
			add(&credentials, alipayCred)
		}
		if hasWxpay {
			add(&credentials, wxpayCred)
		}
	}
	return credentials
}

func xunhupayCredentialForPayment(config map[string]string, paymentType string) (xunhupayCredential, error) {
	credentials := xunhupayCredentials(config, paymentType)
	if len(credentials) == 0 {
		return xunhupayCredential{}, fmt.Errorf("xunhupay credentials missing for payment type: %s", paymentType)
	}
	return credentials[0], nil
}

func xunhupayCredentialsForAppID(config map[string]string, appID string) []xunhupayCredential {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return nil
	}
	var out []xunhupayCredential
	for _, cred := range xunhupayCredentials(config, "") {
		if strings.EqualFold(cred.AppID, appID) {
			out = append(out, cred)
		}
	}
	return out
}

// ResolveXunhupayAppID returns the merchant appid used for a payment type.
// It keeps old single-appid configs working while allowing separate Alipay and
// WeChat credentials in one Xunhupay instance.
func ResolveXunhupayAppID(config map[string]string, paymentType string) string {
	cred, err := xunhupayCredentialForPayment(config, paymentType)
	if err != nil {
		return ""
	}
	return cred.AppID
}

func ValidateXunhupayPaymentTypeConfig(config map[string]string, paymentType string) error {
	if xunhupayPaymentChannel(paymentType) == "" {
		return nil
	}
	_, err := xunhupayCredentialForPayment(config, paymentType)
	return err
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
	credentials := xunhupayCredentials(x.config, "")
	if len(credentials) != 1 {
		return nil
	}
	return map[string]string{"appid": credentials[0].AppID}
}

// CreatePayment 调用虎皮椒下单接口，返回跳转链接或扫码地址。
func (x *Xunhupay) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	notifyURL, returnURL := x.resolveURLs(req)
	channel := xunhupayPaymentChannel(req.PaymentType)
	if channel == "" {
		return nil, fmt.Errorf("xunhupay unsupported payment type: %s", req.PaymentType)
	}
	cred, err := xunhupayCredentialForPayment(x.config, req.PaymentType)
	if err != nil {
		return nil, err
	}

	nonceStr, err := xunhupayNonce()
	if err != nil {
		return nil, fmt.Errorf("xunhupay nonce: %w", err)
	}

	params := map[string]string{
		"version":        xunhupayAPIVersion,
		"appid":          cred.AppID,
		"trade_order_id": req.OrderID,
		"payment":        channel,
		"total_fee":      req.Amount,
		"title":          req.Subject,
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"notify_url":     notifyURL,
		"nonce_str":      nonceStr,
	}
	if returnURL != "" {
		params["return_url"] = returnURL
	}
	if cb := strings.TrimSpace(x.config["callbackUrl"]); cb != "" {
		params["callback_url"] = cb
	}
	// 移动端 H5 支付需要补充 wap_url / wap_name，否则虎皮椒会拒绝下单。
	if req.IsMobile {
		params["type"] = "WAP"
		if wapURL := xunhupayWapURL(returnURL, x.config); wapURL != "" {
			params["wap_url"] = wapURL
		}
		params["wap_name"] = xunhupayWapName(req.Subject, x.config)
	}
	params["plugins"] = "sub2api"
	params["hash"] = xunhupaySign(params, cred.AppSecret)

	body, err := x.post(ctx, x.apiBase()+xunhupayPathPay, params)
	if err != nil {
		return nil, fmt.Errorf("xunhupay create: %w", err)
	}
	if err := xunhupayVerifyResponseHash(body, cred.AppSecret); err != nil {
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
		return nil, fmt.Errorf("xunhupay parse: %w (body=%s)", err, xunhupayResponseSummary(body))
	}
	if !xunhupayCodeIsSuccess(resp.ErrCode) {
		return nil, fmt.Errorf("xunhupay error: %s (body=%s)", strings.TrimSpace(resp.ErrMsg), xunhupayResponseSummary(body))
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
	var lastErr error
	for _, cred := range xunhupayCredentials(x.config, "") {
		for _, orderField := range []string{"out_trade_order", "trade_order_id"} {
			resp, err := x.queryOrderWithCredential(ctx, tradeNo, cred, orderField)
			if err == nil {
				return resp, nil
			}
			lastErr = err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("xunhupay query: no credentials configured")
}

func (x *Xunhupay) queryOrderWithCredential(ctx context.Context, tradeNo string, cred xunhupayCredential, orderField string) (*payment.QueryOrderResponse, error) {
	nonceStr, err := xunhupayNonce()
	if err != nil {
		return nil, fmt.Errorf("xunhupay nonce: %w", err)
	}
	params := map[string]string{
		"appid":     cred.AppID,
		"time":      strconv.FormatInt(time.Now().Unix(), 10),
		"nonce_str": nonceStr,
		orderField:  tradeNo,
	}
	params["hash"] = xunhupaySign(params, cred.AppSecret)

	body, err := x.post(ctx, x.apiBase()+xunhupayPathQuery, params)
	if err != nil {
		return nil, fmt.Errorf("xunhupay query: %w", err)
	}
	if err := xunhupayVerifyResponseHash(body, cred.AppSecret); err != nil {
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
		return nil, fmt.Errorf("xunhupay parse query: %w (body=%s)", err, xunhupayResponseSummary(body))
	}
	if !xunhupayCodeIsSuccess(resp.ErrCode) {
		return nil, fmt.Errorf("xunhupay query failed via %s/%s: %s", cred.Scope, orderField, strings.TrimSpace(resp.ErrMsg))
	}
	amount, err := strconv.ParseFloat(resp.Data.TotalFee, 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("xunhupay query invalid total_fee: %q", resp.Data.TotalFee)
	}
	return &payment.QueryOrderResponse{
		TradeNo:  resp.Data.OpenOrderID,
		Status:   xunhupayMapStatus(resp.Data.Status),
		Amount:   amount,
		Metadata: map[string]string{"appid": cred.AppID},
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
	credentials := xunhupayCredentialsForAppID(x.config, params["appid"])
	if len(credentials) == 0 {
		return nil, fmt.Errorf("xunhupay unknown appid")
	}
	valid := false
	for _, cred := range credentials {
		if strings.EqualFold(sign, xunhupaySign(params, cred.AppSecret)) {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("xunhupay invalid signature")
	}

	status := payment.ProviderStatusFailed
	switch params["status"] {
	case xunhupayNotifyStatusPaid:
		status = payment.ProviderStatusSuccess
	case xunhupayNotifyStatusRefunded:
		status = payment.ProviderStatusRefunded
	}

	amount, err := strconv.ParseFloat(params["total_fee"], 64)
	if err != nil || amount <= 0 {
		return nil, fmt.Errorf("xunhupay invalid total_fee: %q", params["total_fee"])
	}
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

// xunhupayWapURL 推导虎皮椒 H5 支付所需的 wap_url（商户网站访问地址）。
// 优先使用 config 显式配置，其次从 return_url 取 scheme://host。
func xunhupayWapURL(returnURL string, config map[string]string) string {
	if v := xunhupayConfigValue(config, "wapUrl", "wap_url"); v != "" {
		return v
	}
	if returnURL == "" {
		return ""
	}
	parsed, err := url.Parse(returnURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

// xunhupayWapName 推导虎皮椒 H5 支付所需的 wap_name（商户网站名称）。
// 优先使用 config 显式配置，其次回退到订单标题。
func xunhupayWapName(subject string, config map[string]string) string {
	if v := xunhupayConfigValue(config, "wapName", "wap_name"); v != "" {
		return v
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "Sub2API"
	}
	return subject
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

func xunhupayMapStatus(raw string) string {
	switch raw {
	case xunhupayQueryStatusPaid:
		return payment.ProviderStatusPaid
	case xunhupayQueryStatusClosed:
		return payment.ProviderStatusFailed
	case xunhupayQueryStatusPending:
		return payment.ProviderStatusPending
	case xunhupayStatusRefunding, xunhupayStatusRefundFail:
		return payment.ProviderStatusPending
	}
	return payment.ProviderStatusPending
}

// xunhupaySign 按虎皮椒规则生成签名：
//
//  1. 取出全部非空参数，剔除 hash 字段；
//  2. 按 key 的 ASCII 升序排序；
//  3. 用 "k1=v1&k2=v2" 形式拼接；
//  4. 末尾直接拼接 appsecret；
//  5. 取 32 位小写 MD5。
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

func xunhupayVerifyResponseHash(body []byte, appsecret string) error {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}
	hashValue, ok := raw["hash"]
	if !ok {
		return nil
	}
	hash := strings.TrimSpace(xunhupayStringify(hashValue))
	if hash == "" {
		return nil
	}
	for _, params := range []map[string]string{
		xunhupayResponseSignParams(raw, false),
		xunhupayResponseSignParams(raw, true),
	} {
		if strings.EqualFold(hash, xunhupaySign(params, appsecret)) {
			return nil
		}
	}
	return fmt.Errorf("invalid response signature")
}

func xunhupayResponseSignParams(raw map[string]any, includeScalarValues bool) map[string]string {
	params := make(map[string]string, len(raw))
	for k, v := range raw {
		if k == "hash" {
			continue
		}
		switch val := v.(type) {
		case string:
			params[k] = val
		case json.Number:
			if includeScalarValues {
				params[k] = val.String()
			}
		case float64:
			if includeScalarValues {
				params[k] = strconv.FormatFloat(val, 'f', -1, 64)
			}
		case bool:
			if includeScalarValues {
				params[k] = xunhupayStringify(val)
			}
		}
	}
	return params
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

// post 以 application/x-www-form-urlencoded 形式提交参数，匹配虎皮椒
// /payment/do.html 与 /payment/query.html 接口文档要求。
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
	req.Header.Set("Accept", "application/json")
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

package provider

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

// 校验签名拼接顺序（ASCII 升序）与末尾直拼 appsecret 行为符合虎皮椒文档。
func TestXunhupaySignMatchesDocSpec(t *testing.T) {
	params := map[string]string{
		"appid":          "abc",
		"trade_order_id": "order-1",
		"total_fee":      "1.00",
		"title":          "vip",
		"time":           "1700000000",
		"notify_url":     "https://example.com/notify",
		"nonce_str":      "nonce123",
		"hash":           "ignored",
		"empty":          "",
	}
	got := xunhupaySign(params, "secret-x")

	keys := []string{"appid", "trade_order_id", "total_fee", "title", "time", "notify_url", "nonce_str"}
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
	_, _ = buf.WriteString("secret-x")
	expected := hex.EncodeToString(func() []byte {
		h := md5.Sum([]byte(buf.String()))
		return h[:]
	}())

	if got != expected {
		t.Fatalf("xunhupaySign mismatch:\n  got      %s\n  expected %s", got, expected)
	}
}

func TestXunhupayParseFormNotify(t *testing.T) {
	body := "trade_order_id=ord_1&transaction_id=tx_1&total_fee=9.90&status=OD&appid=A&hash=H"
	params, err := xunhupayParseNotifyBody(body)
	if err != nil {
		t.Fatalf("parse form notify: %v", err)
	}
	if params["status"] != "OD" || params["trade_order_id"] != "ord_1" {
		t.Fatalf("unexpected parsed params: %+v", params)
	}
}

func TestXunhupayParseJSONNotify(t *testing.T) {
	body := `{"trade_order_id":"ord_2","total_fee":"1.50","status":"CD","hash":"H"}`
	params, err := xunhupayParseNotifyBody(body)
	if err != nil {
		t.Fatalf("parse json notify: %v", err)
	}
	if params["status"] != "CD" || params["trade_order_id"] != "ord_2" {
		t.Fatalf("unexpected parsed params: %+v", params)
	}
}

// 验签 + 状态映射端到端：构造合法回调，校验解析得到的订单字段。
func TestXunhupayVerifyNotificationSuccess(t *testing.T) {
	x, err := NewXunhupay("inst-1", map[string]string{
		"appid":     "abc",
		"appsecret": "secret-x",
		"notifyUrl": "https://example.com/notify",
	})
	if err != nil {
		t.Fatalf("create xunhupay: %v", err)
	}

	form := url.Values{}
	form.Set("appid", "abc")
	form.Set("trade_order_id", "ord_3")
	form.Set("transaction_id", "tx_3")
	form.Set("open_order_id", "open_3")
	form.Set("total_fee", "9.99")
	form.Set("time", "1700000000")
	form.Set("nonce_str", "nonce")
	form.Set("status", xunhupayNotifyStatusPaid)
	plain := map[string]string{}
	for k := range form {
		plain[k] = form.Get(k)
	}
	form.Set("hash", xunhupaySign(plain, "secret-x"))

	ntf, err := x.VerifyNotification(context.Background(), form.Encode(), nil)
	if err != nil {
		t.Fatalf("verify notification: %v", err)
	}
	if ntf.Status != payment.ProviderStatusSuccess {
		t.Fatalf("expected success status, got %s", ntf.Status)
	}
	if ntf.OrderID != "ord_3" || ntf.TradeNo != "tx_3" {
		t.Fatalf("unexpected order/trade no: %+v", ntf)
	}
	if ntf.Amount < 9.989 || ntf.Amount > 9.991 {
		t.Fatalf("amount mismatch: %v", ntf.Amount)
	}
	if ntf.Metadata["appid"] != "abc" {
		t.Fatalf("metadata appid mismatch: %+v", ntf.Metadata)
	}
}

func TestXunhupayChannelSpecificCredentials(t *testing.T) {
	x, err := NewXunhupay("inst-1", map[string]string{
		"alipayAppId":     "ali-app",
		"alipayAppSecret": "ali-secret",
		"wxpayAppId":      "wx-app",
		"wxpayAppSecret":  "wx-secret",
		"notifyUrl":       "https://example.com/notify",
	})
	if err != nil {
		t.Fatalf("create xunhupay: %v", err)
	}
	if got := ResolveXunhupayAppID(x.config, payment.TypeAlipay); got != "ali-app" {
		t.Fatalf("alipay appid = %q", got)
	}
	if got := ResolveXunhupayAppID(x.config, payment.TypeWxpay); got != "wx-app" {
		t.Fatalf("wxpay appid = %q", got)
	}

	form := url.Values{}
	form.Set("appid", "wx-app")
	form.Set("trade_order_id", "ord_4")
	form.Set("transaction_id", "tx_4")
	form.Set("total_fee", "8.80")
	form.Set("time", "1700000000")
	form.Set("nonce_str", "nonce")
	form.Set("status", xunhupayNotifyStatusPaid)
	plain := map[string]string{}
	for k := range form {
		plain[k] = form.Get(k)
	}
	form.Set("hash", xunhupaySign(plain, "wx-secret"))

	ntf, err := x.VerifyNotification(context.Background(), form.Encode(), nil)
	if err != nil {
		t.Fatalf("verify notification with wx credential: %v", err)
	}
	if ntf.Metadata["appid"] != "wx-app" {
		t.Fatalf("metadata appid mismatch: %+v", ntf.Metadata)
	}
}

func TestXunhupayQueryUsesOfficialFieldAndResponseHash(t *testing.T) {
	var sawOfficialField bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != xunhupayPathQuery {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		var requestParams map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
			t.Fatalf("decode json request: %v", err)
		}
		if requestParams["out_trade_order"] == "ord_query" {
			sawOfficialField = true
		}
		if requestParams["appid"] != "ali-app" {
			t.Fatalf("appid = %q", requestParams["appid"])
		}
		resp := map[string]any{
			"errcode": 0,
			"errmsg":  "",
			"data": map[string]any{
				"status":        xunhupayQueryStatusPaid,
				"total_fee":     "12.30",
				"open_order_id": "open-query",
			},
		}
		resp["hash"] = xunhupaySign(map[string]string{}, "ali-secret")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	x, err := NewXunhupay("inst-1", map[string]string{
		"alipayAppId":     "ali-app",
		"alipayAppSecret": "ali-secret",
		"notifyUrl":       "https://example.com/notify",
		"apiBase":         server.URL,
	})
	if err != nil {
		t.Fatalf("create xunhupay: %v", err)
	}
	resp, err := x.QueryOrder(context.Background(), "ord_query")
	if err != nil {
		t.Fatalf("query order: %v", err)
	}
	if !sawOfficialField {
		t.Fatalf("expected query to use out_trade_order")
	}
	if resp.Status != payment.ProviderStatusPaid || resp.Amount != 12.30 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestXunhupayVerifyNotificationBadSignRejected(t *testing.T) {
	x, err := NewXunhupay("inst-1", map[string]string{
		"appid":     "abc",
		"appsecret": "secret-x",
		"notifyUrl": "https://example.com/notify",
	})
	if err != nil {
		t.Fatalf("create xunhupay: %v", err)
	}
	body := "appid=abc&trade_order_id=ord&status=OD&hash=deadbeef"
	if _, err := x.VerifyNotification(context.Background(), body, nil); err == nil {
		t.Fatalf("expected signature error, got nil")
	}
}

func TestXunhupayVerifyNotificationInvalidAmountRejected(t *testing.T) {
	x, err := NewXunhupay("inst-1", map[string]string{
		"appid":     "abc",
		"appsecret": "secret-x",
		"notifyUrl": "https://example.com/notify",
	})
	if err != nil {
		t.Fatalf("create xunhupay: %v", err)
	}

	form := url.Values{}
	form.Set("appid", "abc")
	form.Set("trade_order_id", "ord_bad_amount")
	form.Set("transaction_id", "tx_bad_amount")
	form.Set("total_fee", "not-a-number")
	form.Set("status", xunhupayNotifyStatusPaid)
	plain := map[string]string{}
	for k := range form {
		plain[k] = form.Get(k)
	}
	form.Set("hash", xunhupaySign(plain, "secret-x"))

	if _, err := x.VerifyNotification(context.Background(), form.Encode(), nil); err == nil {
		t.Fatalf("expected invalid amount error, got nil")
	}
}

func TestXunhupayPaymentChannelRouting(t *testing.T) {
	cases := map[string]string{
		payment.TypeAlipay:       xunhupayAlipayChannel,
		payment.TypeAlipayDirect: xunhupayAlipayChannel,
		payment.TypeWxpay:        xunhupayWxPayChannel,
		payment.TypeWxpayDirect:  xunhupayWxPayChannel,
		payment.TypeStripe:       "",
	}
	for in, want := range cases {
		got := xunhupayPaymentChannel(in)
		if got != want {
			t.Errorf("xunhupayPaymentChannel(%q) = %q, want %q", in, got, want)
		}
	}
}

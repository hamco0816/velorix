package service

import "testing"

func TestNormalizeContactMethods(t *testing.T) {
	methods := NormalizeContactMethods([]ContactMethod{
		{Type: "qq", Label: "", Value: " 123456 ", URL: "javascript:alert(1)"},
		{Type: "wx", Label: "", Value: "velorix", URL: "https://example.com/wechat"},
		{Type: "custom", Label: "Telegram", Value: "", URL: "ftp://example.com"},
	}, "")

	if len(methods) != 2 {
		t.Fatalf("expected 2 valid methods, got %d", len(methods))
	}
	if methods[0].Type != ContactMethodTypeQQ || methods[0].Label != "QQ" || methods[0].Value != "123456" || methods[0].URL != "" {
		t.Fatalf("unexpected QQ normalization: %#v", methods[0])
	}
	if methods[1].Type != ContactMethodTypeWeChat || methods[1].Label != "\u5fae\u4fe1" || methods[1].URL != "https://example.com/wechat" {
		t.Fatalf("unexpected WeChat normalization: %#v", methods[1])
	}
}

func TestNormalizeContactMethodsFallsBackToLegacyContactInfo(t *testing.T) {
	methods := NormalizeContactMethods(nil, " support@example.com ")
	if len(methods) != 1 {
		t.Fatalf("expected one legacy fallback method, got %d", len(methods))
	}
	if methods[0].Type != ContactMethodTypeCustom || methods[0].Label != "\u5ba2\u670d" || methods[0].Value != "support@example.com" {
		t.Fatalf("unexpected legacy fallback: %#v", methods[0])
	}
}

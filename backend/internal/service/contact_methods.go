package service

import (
	"encoding/json"
	"net/url"
	"strings"
)

const (
	ContactMethodTypeQQ     = "qq"
	ContactMethodTypeWeChat = "wechat"
	ContactMethodTypeCustom = "custom"

	maxContactMethods  = 8
	maxContactLabelLen = 32
	maxContactValueLen = 128
	maxContactURLLen   = 512
)

// ContactMethod represents one public customer-service contact channel.
type ContactMethod struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	URL   string `json:"url,omitempty"`
}

func ParseContactMethods(raw string, legacyContactInfo string) []ContactMethod {
	var methods []ContactMethod
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &methods); err == nil {
			return NormalizeContactMethods(methods, legacyContactInfo)
		}
	}
	return NormalizeContactMethods(nil, legacyContactInfo)
}

func NormalizeContactMethods(methods []ContactMethod, legacyContactInfo string) []ContactMethod {
	out := make([]ContactMethod, 0, min(len(methods), maxContactMethods))
	for _, method := range methods {
		normalized := ContactMethod{
			Type:  normalizeContactMethodType(method.Type),
			Label: truncateContactString(strings.TrimSpace(method.Label), maxContactLabelLen),
			Value: truncateContactString(strings.TrimSpace(method.Value), maxContactValueLen),
			URL:   sanitizeContactMethodURL(method.URL),
		}
		if normalized.Value == "" && normalized.URL == "" {
			continue
		}
		if normalized.Label == "" {
			normalized.Label = defaultContactMethodLabel(normalized.Type)
		}
		out = append(out, normalized)
		if len(out) >= maxContactMethods {
			break
		}
	}
	if len(out) == 0 {
		legacy := truncateContactString(strings.TrimSpace(legacyContactInfo), maxContactValueLen)
		if legacy != "" {
			out = append(out, ContactMethod{
				Type:  ContactMethodTypeCustom,
				Label: defaultContactMethodLabel(ContactMethodTypeCustom),
				Value: legacy,
			})
		}
	}
	return out
}

func ContactMethodsJSON(methods []ContactMethod) string {
	normalized := NormalizeContactMethods(methods, "")
	data, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func ContactMethodsSummary(methods []ContactMethod, fallback string) string {
	normalized := NormalizeContactMethods(methods, "")
	if len(normalized) == 0 {
		return strings.TrimSpace(fallback)
	}
	parts := make([]string, 0, len(normalized))
	for _, method := range normalized {
		value := method.Value
		if value == "" {
			value = method.URL
		}
		if value == "" {
			continue
		}
		parts = append(parts, method.Label+": "+value)
	}
	return strings.Join(parts, " / ")
}

func normalizeContactMethodType(methodType string) string {
	switch strings.ToLower(strings.TrimSpace(methodType)) {
	case ContactMethodTypeQQ:
		return ContactMethodTypeQQ
	case ContactMethodTypeWeChat, "weixin", "wx":
		return ContactMethodTypeWeChat
	default:
		return ContactMethodTypeCustom
	}
}

func defaultContactMethodLabel(methodType string) string {
	switch methodType {
	case ContactMethodTypeQQ:
		return "QQ"
	case ContactMethodTypeWeChat:
		return "\u5fae\u4fe1"
	default:
		return "\u5ba2\u670d"
	}
}

func sanitizeContactMethodURL(raw string) string {
	value := truncateContactString(strings.TrimSpace(raw), maxContactURLLen)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if parsed.Host == "" {
		return ""
	}
	return value
}

func truncateContactString(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}
	return string(runes[:maxLen])
}

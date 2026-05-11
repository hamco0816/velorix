package service

import (
	"encoding/json"
	"net/url"
	"strings"
)

const (
	ContactMethodTypeQQ      = "qq"
	ContactMethodTypeQQGroup = "qq_group"
	ContactMethodTypeWeChat  = "wechat"
	ContactMethodTypeCustom  = "custom"

	maxContactMethods  = 8
	maxContactLabelLen = 32
	maxContactValueLen = 128
	maxContactURLLen   = 512
	// 单张二维码 base64 上限 64KB（QQ 群码 PNG 原图通常 3-10KB，base64 后 4-14KB，留 4x 余量足够）
	// 超过的图 admin 应该先压缩。整张联系方式 JSON 最大 ≈ 8 × 64KB = 512KB，对 settings 表无压力。
	maxContactImageDataLen = 65536
)

// ContactMethod represents one public customer-service contact channel.
// ImageData 用于 qq_group 这类需要扫二维码的场景，dataURL 格式 data:image/png;base64,xxxx。
// 直接 inline 在 settings JSON 不另开文件存储，与现有 site_logo 同款方案。
type ContactMethod struct {
	Type      string `json:"type"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	URL       string `json:"url,omitempty"`
	ImageData string `json:"image_data,omitempty"`
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
			Type:      normalizeContactMethodType(method.Type),
			Label:     truncateContactString(strings.TrimSpace(method.Label), maxContactLabelLen),
			Value:     truncateContactString(strings.TrimSpace(method.Value), maxContactValueLen),
			URL:       sanitizeContactMethodURL(method.URL),
			ImageData: sanitizeContactImageData(method.ImageData),
		}
		// qq_group 通常以二维码 + 群号为主，允许 Value 为空但有 ImageData 时仍保留
		if normalized.Value == "" && normalized.URL == "" && normalized.ImageData == "" {
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
	case ContactMethodTypeQQGroup, "qqgroup", "qq-group":
		return ContactMethodTypeQQGroup
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
	case ContactMethodTypeQQGroup:
		return "QQ\u7fa4" // QQ\u7fa4
	case ContactMethodTypeWeChat:
		return "\u5fae\u4fe1"
	default:
		return "\u5ba2\u670d"
	}
}

// sanitizeContactImageData \u6821\u9a8c dataURL\uff08\u4ec5 png/jpeg/webp/gif/svg+xml\uff09\uff0c\u8d85 maxContactImageDataLen \u4e22\u5f03\u3002
// \u4e0d\u518d\u4e8c\u6b21\u89e3\u7801 base64\uff0cadmin \u7aef\u5df2\u7ecf\u9884\u538b\u7f29\uff1b\u540e\u7aef\u53ea\u505a\u683c\u5f0f\u767d\u540d\u5355 + \u957f\u5ea6\u4e0a\u9650\u907f\u514d\u6c61\u67d3 settings \u8868\u3002
func sanitizeContactImageData(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if len(value) > maxContactImageDataLen {
		return ""
	}
	if !strings.HasPrefix(value, "data:image/") {
		return ""
	}
	// \u53ea\u5141\u8bb8\u5df2\u77e5 MIME \u524d\u7f00\uff0c\u9632\u6b62\u628a\u522b\u7684 dataURL \u585e\u8fdb\u6765
	allowed := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/jpg;base64,",
		"data:image/webp;base64,",
		"data:image/gif;base64,",
		"data:image/svg+xml;base64,",
	}
	for _, p := range allowed {
		if strings.HasPrefix(value, p) {
			return value
		}
	}
	return ""
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

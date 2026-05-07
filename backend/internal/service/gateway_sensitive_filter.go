package service

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"golang.org/x/sync/singleflight"
)

const (
	gatewaySensitiveFilterCacheKey  = "gateway_sensitive_filter"
	gatewaySensitiveFilterCacheTTL  = 60 * time.Second
	gatewaySensitiveFilterErrorTTL  = 5 * time.Second
	gatewaySensitiveFilterDBTimeout = 5 * time.Second
	gatewaySensitiveFilterMaxWords  = 2000
	gatewaySensitiveFilterMaxRunes  = 200
	gatewaySensitiveFilterPathRoot  = "$"
)

type GatewaySensitiveFilterSettings struct {
	Enabled bool
	Words   []string
}

type GatewaySensitiveMatch struct {
	Word    string
	Path    string
	Source  string
	Preview string
}

type gatewaySensitiveRule struct {
	Word   string
	Source string
}

type cachedGatewaySensitiveFilterSettings struct {
	settings  *GatewaySensitiveFilterSettings
	expiresAt int64
}

var gatewaySensitiveFilterCache atomic.Value // *cachedGatewaySensitiveFilterSettings
var gatewaySensitiveFilterSF singleflight.Group

var gatewaySensitiveContainerKeys = map[string]struct{}{
	"messages":     {},
	"message":      {},
	"contents":     {},
	"content":      {},
	"parts":        {},
	"input":        {},
	"prompt":       {},
	"instructions": {},
	"system":       {},
	"text":         {},
	"query":        {},
}

var gatewaySensitiveStringKeys = map[string]struct{}{
	"content":      {},
	"text":         {},
	"input":        {},
	"prompt":       {},
	"instructions": {},
	"system":       {},
	"query":        {},
}

var gatewaySensitiveSkipKeys = map[string]struct{}{
	"model":        {},
	"role":         {},
	"type":         {},
	"name":         {},
	"id":           {},
	"metadata":     {},
	"user":         {},
	"url":          {},
	"image_url":    {},
	"b64_json":     {},
	"mask":         {},
	"file":         {},
	"filename":     {},
	"mime_type":    {},
	"content_type": {},
}

var gatewaySensitiveBuiltInWords = []string{
	"ignore previous instructions",
	"ignore all previous instructions",
	"ignore system prompt",
	"ignore safety policy",
	"bypass safety",
	"bypass restrictions",
	"jailbreak mode",
	"developer mode",
	"do anything now",
	"prompt injection",
	"system prompt leak",
	"reveal your system prompt",
	"act as dan",
	"无视之前的指令",
	"忽略之前的指令",
	"忽略所有指令",
	"忽略系统提示",
	"绕过安全策略",
	"绕过限制",
	"解除限制",
	"越狱模式",
	"开发者模式",
	"提示词注入",
	"泄露系统提示词",
	"输出系统提示词",
	"钓鱼网站",
	"钓鱼邮件模板",
	"盗取密码",
	"窃取凭据",
	"木马程序",
	"勒索软件",
	"后门程序",
	"免杀",
	"绕过杀软",
	"恶意软件",
	"远控木马",
	"c2 server",
	"command and control",
	"credential stealer",
	"password stealer",
	"phishing kit",
	"ransomware",
	"malware loader",
	"反编译源码",
	"逆向工程",
	"绕过授权",
	"破解软件",
	"注册码生成器",
	"exploit chain",
	"sql injection payload",
	"xss payload",
	"privilege escalation exploit",
}

func ParseGatewaySensitiveWords(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	seen := make(map[string]struct{})
	words := make([]string, 0)
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case '\n', '\r', ',', ';', '\uFF0C', '\uFF1B':
			return true
		default:
			return false
		}
	}) {
		word := strings.TrimSpace(part)
		if word == "" {
			continue
		}
		runes := []rune(word)
		if len(runes) > gatewaySensitiveFilterMaxRunes {
			word = string(runes[:gatewaySensitiveFilterMaxRunes])
		}
		key := compactGatewaySensitiveText(word)
		if key == "" {
			key = strings.ToLower(word)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		words = append(words, word)
		if len(words) >= gatewaySensitiveFilterMaxWords {
			break
		}
	}
	return words
}

func NormalizeGatewaySensitiveWordsText(raw string) string {
	return strings.Join(ParseGatewaySensitiveWords(raw), "\n")
}

func (s *SettingService) GetGatewaySensitiveFilterSettings(ctx context.Context) *GatewaySensitiveFilterSettings {
	if s == nil || s.settingRepo == nil {
		return &GatewaySensitiveFilterSettings{Enabled: true}
	}
	if cached, ok := gatewaySensitiveFilterCache.Load().(*cachedGatewaySensitiveFilterSettings); ok && cached != nil {
		if time.Now().UnixNano() < cached.expiresAt {
			return cloneGatewaySensitiveFilterSettings(cached.settings)
		}
	}

	val, _, _ := gatewaySensitiveFilterSF.Do(gatewaySensitiveFilterCacheKey, func() (any, error) {
		if cached, ok := gatewaySensitiveFilterCache.Load().(*cachedGatewaySensitiveFilterSettings); ok && cached != nil {
			if time.Now().UnixNano() < cached.expiresAt {
				return cloneGatewaySensitiveFilterSettings(cached.settings), nil
			}
		}

		dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gatewaySensitiveFilterDBTimeout)
		defer cancel()
		values, err := s.settingRepo.GetMultiple(dbCtx, []string{
			SettingKeyGatewaySensitiveFilterEnabled,
			SettingKeyGatewaySensitiveFilterWords,
		})
		if err != nil {
			slog.Warn("failed to get gateway sensitive filter settings", "error", err)
			if cached, ok := gatewaySensitiveFilterCache.Load().(*cachedGatewaySensitiveFilterSettings); ok && cached != nil && cached.settings != nil {
				gatewaySensitiveFilterCache.Store(&cachedGatewaySensitiveFilterSettings{
					settings:  cached.settings,
					expiresAt: time.Now().Add(gatewaySensitiveFilterErrorTTL).UnixNano(),
				})
				return cloneGatewaySensitiveFilterSettings(cached.settings), nil
			}
			fallback := &GatewaySensitiveFilterSettings{Enabled: true}
			gatewaySensitiveFilterCache.Store(&cachedGatewaySensitiveFilterSettings{
				settings:  fallback,
				expiresAt: time.Now().Add(gatewaySensitiveFilterErrorTTL).UnixNano(),
			})
			return fallback, nil
		}

		settings := &GatewaySensitiveFilterSettings{
			Enabled: !strings.EqualFold(strings.TrimSpace(values[SettingKeyGatewaySensitiveFilterEnabled]), "false"),
			Words:   ParseGatewaySensitiveWords(values[SettingKeyGatewaySensitiveFilterWords]),
		}
		gatewaySensitiveFilterCache.Store(&cachedGatewaySensitiveFilterSettings{
			settings:  settings,
			expiresAt: time.Now().Add(gatewaySensitiveFilterCacheTTL).UnixNano(),
		})
		return cloneGatewaySensitiveFilterSettings(settings), nil
	})
	if settings, ok := val.(*GatewaySensitiveFilterSettings); ok {
		return settings
	}
	return &GatewaySensitiveFilterSettings{Enabled: true}
}

func storeGatewaySensitiveFilterSettings(settings *SystemSettings) {
	if settings == nil {
		return
	}
	gatewaySensitiveFilterSF.Forget(gatewaySensitiveFilterCacheKey)
	gatewaySensitiveFilterCache.Store(&cachedGatewaySensitiveFilterSettings{
		settings: &GatewaySensitiveFilterSettings{
			Enabled: settings.GatewaySensitiveFilterEnabled,
			Words:   ParseGatewaySensitiveWords(settings.GatewaySensitiveFilterWords),
		},
		expiresAt: time.Now().Add(gatewaySensitiveFilterCacheTTL).UnixNano(),
	})
}

func CheckGatewaySensitiveJSON(body []byte, settings *GatewaySensitiveFilterSettings) (*GatewaySensitiveMatch, error) {
	if settings == nil || !settings.Enabled || len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}
	rules := gatewaySensitiveEffectiveRules(settings.Words)
	if len(rules) == 0 {
		return nil, nil
	}
	var payload any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, nil
	}
	return findGatewaySensitiveMatch(payload, rules, gatewaySensitiveFilterPathRoot, false), nil
}

func findGatewaySensitiveMatch(value any, rules []gatewaySensitiveRule, path string, forced bool) *GatewaySensitiveMatch {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			normalizedKey := strings.ToLower(strings.TrimSpace(key))
			if _, skip := gatewaySensitiveSkipKeys[normalizedKey]; skip {
				continue
			}
			_, container := gatewaySensitiveContainerKeys[normalizedKey]
			_, stringKey := gatewaySensitiveStringKeys[normalizedKey]
			if match := findGatewaySensitiveMatch(item, rules, appendGatewaySensitivePath(path, key), forced || container || stringKey); match != nil {
				return match
			}
		}
	case []any:
		for i, item := range v {
			if match := findGatewaySensitiveMatch(item, rules, appendGatewaySensitiveIndex(path, i), forced); match != nil {
				return match
			}
		}
	case string:
		if forced {
			return matchGatewaySensitiveText(v, rules, path)
		}
	}
	return nil
}

func matchGatewaySensitiveText(text string, rules []gatewaySensitiveRule, path string) *GatewaySensitiveMatch {
	if text == "" {
		return nil
	}
	lowerText := strings.ToLower(text)
	compactText := compactGatewaySensitiveText(text)
	for _, rule := range rules {
		word := strings.TrimSpace(rule.Word)
		if word == "" {
			continue
		}
		lowerWord := strings.ToLower(word)
		if lowerWord != "" && strings.Contains(lowerText, lowerWord) {
			return &GatewaySensitiveMatch{Word: word, Path: path, Source: rule.Source, Preview: sanitizeGatewaySensitivePreview(text)}
		}
		compactWord := compactGatewaySensitiveText(word)
		if compactWord != "" && strings.Contains(compactText, compactWord) {
			return &GatewaySensitiveMatch{Word: word, Path: path, Source: rule.Source, Preview: sanitizeGatewaySensitivePreview(text)}
		}
	}
	return nil
}

func compactGatewaySensitiveText(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range strings.ToLower(value) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || isGatewaySensitiveIgnoredRune(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func gatewaySensitiveEffectiveRules(custom []string) []gatewaySensitiveRule {
	if len(custom) == 0 {
		rules := make([]gatewaySensitiveRule, 0, len(gatewaySensitiveBuiltInWords))
		for _, word := range gatewaySensitiveBuiltInWords {
			rules = append(rules, gatewaySensitiveRule{Word: word, Source: "builtin"})
		}
		return rules
	}
	seen := make(map[string]struct{}, len(gatewaySensitiveBuiltInWords)+len(custom))
	rules := make([]gatewaySensitiveRule, 0, len(gatewaySensitiveBuiltInWords)+len(custom))
	for _, word := range gatewaySensitiveBuiltInWords {
		word = strings.TrimSpace(word)
		key := compactGatewaySensitiveText(word)
		if word == "" || key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		rules = append(rules, gatewaySensitiveRule{Word: word, Source: "builtin"})
	}
	for _, word := range custom {
		word = strings.TrimSpace(word)
		key := compactGatewaySensitiveText(word)
		if word == "" || key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		rules = append(rules, gatewaySensitiveRule{Word: word, Source: "custom"})
	}
	return rules
}

func sanitizeGatewaySensitivePreview(text string) string {
	preview := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if preview == "" {
		return ""
	}
	const maxRunes = 240
	runes := []rune(preview)
	if len(runes) <= maxRunes {
		return preview
	}
	return string(runes[:maxRunes])
}

func isGatewaySensitiveIgnoredRune(r rune) bool {
	switch r {
	case '\u200b', '\u200c', '\u200d', '\u200e', '\u200f', '\u2060', '\ufeff':
		return true
	default:
		return false
	}
}

func cloneGatewaySensitiveFilterSettings(settings *GatewaySensitiveFilterSettings) *GatewaySensitiveFilterSettings {
	if settings == nil {
		return &GatewaySensitiveFilterSettings{}
	}
	clone := &GatewaySensitiveFilterSettings{Enabled: settings.Enabled}
	if len(settings.Words) > 0 {
		clone.Words = append([]string(nil), settings.Words...)
	}
	return clone
}

func appendGatewaySensitivePath(path, key string) string {
	if path == "" {
		path = gatewaySensitiveFilterPathRoot
	}
	return path + "." + key
}

func appendGatewaySensitiveIndex(path string, index int) string {
	if path == "" {
		path = gatewaySensitiveFilterPathRoot
	}
	return path + "[" + strconv.Itoa(index) + "]"
}

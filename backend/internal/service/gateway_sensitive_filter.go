package service

import (
	"bytes"
	"context"
	"encoding/json"
	"hash/fnv"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	goahocorasick "github.com/anknown/ahocorasick"
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
	"messages": {},
	"message":  {},
	"contents": {},
	"content":  {},
	"parts":    {},
	"input":    {},
	"prompt":   {},
	"text":     {},
	"query":    {},
}

var gatewaySensitiveStringKeys = map[string]struct{}{
	"content": {},
	"text":    {},
	"input":   {},
	"prompt":  {},
	"query":   {},
}

// system / instructions 这类字段属于开发者 / IDE 配置（Claude Code、VSCode、Cursor 等都会
// 自带固定 system prompt），用户无法在 user message 之外注入。把它们彻底跳过避免误报：
// 真正的 prompt injection 攻击载体只能落在 user role 的 content 里，那才是要查的。
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
	"system":       {}, // Anthropic 顶层 system 字段 + OpenAI messages[].role=system 容器
	"instructions": {}, // OpenAI Responses API instructions 字段
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
		// 跳过 OpenAI / Anthropic 的 system / developer / assistant role 消息：
		// role=system/developer 是 IDE 或开发者配置，不是用户能注入的攻击载体（避免 VSCode/Claude
		// Code 等工具自带的 system prompt 触发"prompt injection"误报）；assistant 是模型上一轮
		// 的输出，admin 风控也不应该针对它。只查用户能直接控制的 user role 内容。
		if roleVal, ok := v["role"].(string); ok {
			role := strings.ToLower(strings.TrimSpace(roleVal))
			if role == "system" || role == "developer" || role == "assistant" {
				return nil
			}
		}
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

// matchGatewaySensitiveText 双层匹配：
//  1. AC 自动机扫一次 lowerText（O(M+hits)），命中直接返回
//  2. AC 自动机扫一次 compactText（去空格/标点/零宽 后再匹配 compactWord 词典），防绕过
//
// 词典随 settings 变化，按 fnv64 hash 缓存 AC 实例（builtin+custom 词集合不变时复用）。
func matchGatewaySensitiveText(text string, rules []gatewaySensitiveRule, path string) *GatewaySensitiveMatch {
	if text == "" || len(rules) == 0 {
		return nil
	}
	lowerText := strings.ToLower(text)
	if hit := acFirstHit(getOrBuildSensitiveACForRules(rules, false), lowerText); hit != "" {
		return buildSensitiveMatch(hit, rules, path, text)
	}
	compactText := compactGatewaySensitiveText(text)
	if compactText == "" {
		return nil
	}
	if hit := acFirstHit(getOrBuildSensitiveACForRules(rules, true), compactText); hit != "" {
		return buildSensitiveMatch(hit, rules, path, text)
	}
	return nil
}

// buildSensitiveMatch 命中后回填 Source（区分 builtin / custom）：AC 扫的是规范化后的 word，
// 通过 lowerWord 或 compactWord 反查原 rule，保留原 word 字面给上游告警。
func buildSensitiveMatch(hitWord string, rules []gatewaySensitiveRule, path, original string) *GatewaySensitiveMatch {
	preview := sanitizeGatewaySensitivePreview(original)
	for _, rule := range rules {
		word := strings.TrimSpace(rule.Word)
		if word == "" {
			continue
		}
		if strings.EqualFold(word, hitWord) ||
			strings.EqualFold(strings.ToLower(word), hitWord) ||
			compactGatewaySensitiveText(word) == hitWord {
			return &GatewaySensitiveMatch{Word: word, Path: path, Source: rule.Source, Preview: preview}
		}
	}
	// 兜底（理论上 rules 必含此词）
	return &GatewaySensitiveMatch{Word: hitWord, Path: path, Source: "builtin", Preview: preview}
}

// --- AC 自动机缓存（按词典+模式 fnv hash 复用实例）---

type sensitiveACCacheEntry struct {
	machine *goahocorasick.Machine
}

var sensitiveACCache sync.Map // key: string -> *sensitiveACCacheEntry

func sensitiveACCacheKey(rules []gatewaySensitiveRule, compact bool) (string, []string) {
	if len(rules) == 0 {
		return "", nil
	}
	words := make([]string, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for _, r := range rules {
		var w string
		if compact {
			w = compactGatewaySensitiveText(r.Word)
		} else {
			w = strings.ToLower(strings.TrimSpace(r.Word))
		}
		if w == "" {
			continue
		}
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		words = append(words, w)
	}
	if len(words) == 0 {
		return "", nil
	}
	sort.Strings(words)
	hasher := fnv.New64a()
	if compact {
		_, _ = hasher.Write([]byte("c|"))
	} else {
		_, _ = hasher.Write([]byte("l|"))
	}
	for _, w := range words {
		_, _ = hasher.Write([]byte{0})
		_, _ = hasher.Write([]byte(w))
	}
	return strconv.FormatUint(hasher.Sum64(), 16), words
}

func getOrBuildSensitiveACForRules(rules []gatewaySensitiveRule, compact bool) *goahocorasick.Machine {
	key, words := sensitiveACCacheKey(rules, compact)
	if key == "" || len(words) == 0 {
		return nil
	}
	if v, ok := sensitiveACCache.Load(key); ok {
		if e, ok2 := v.(*sensitiveACCacheEntry); ok2 && e != nil {
			return e.machine
		}
	}
	machine := buildAhoCorasick(words)
	if machine == nil {
		return nil
	}
	if actual, loaded := sensitiveACCache.LoadOrStore(key, &sensitiveACCacheEntry{machine: machine}); loaded {
		if e, ok := actual.(*sensitiveACCacheEntry); ok && e != nil {
			return e.machine
		}
	}
	return machine
}

func buildAhoCorasick(words []string) *goahocorasick.Machine {
	if len(words) == 0 {
		return nil
	}
	patterns := make([][]rune, 0, len(words))
	for _, w := range words {
		if w == "" {
			continue
		}
		patterns = append(patterns, []rune(w))
	}
	if len(patterns) == 0 {
		return nil
	}
	m := new(goahocorasick.Machine)
	if err := m.Build(patterns); err != nil {
		slog.Warn("sensitive: build aho-corasick failed", "error", err, "patterns", len(patterns))
		return nil
	}
	return m
}

// acFirstHit 用 AC 扫一遍文本，返回第一个命中的词；未命中返回空串。
func acFirstHit(m *goahocorasick.Machine, text string) string {
	if m == nil || text == "" {
		return ""
	}
	hits := m.MultiPatternSearch([]rune(text), true)
	if len(hits) == 0 {
		return ""
	}
	return string(hits[0].Word)
}

func compactGatewaySensitiveText(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range strings.ToLower(value) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || isGatewaySensitiveIgnoredRune(r) {
			continue
		}
		_, _ = builder.WriteRune(r)
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

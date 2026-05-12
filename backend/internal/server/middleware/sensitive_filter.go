package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const gatewaySensitiveFilterUserMessage = "请求包含可能触发上游账号风控的内容，已被本地安全策略拦截。请改写为防御、学习或合规用途后重试。如确认为误报，请联系管理员并提供下方 Request ID。"

const gatewaySensitiveFormFieldMaxBytes = 256 * 1024

// GatewaySensitiveFilter blocks configured sensitive prompt content before proxying upstream.
func GatewaySensitiveFilter(settingService *service.SettingService, opsService *service.OpsService, writeError GatewayErrorWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if settingService == nil || c.Request == nil || c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		// 白名单用户跳过敏感词检测：admin 在风控页面对合法用户加白名单后，
		// 该用户的所有后续请求直接放行，不再走 sensitive_filter（防止反复误拦）。
		if apiKey, ok := GetAPIKeyFromContext(c); ok && apiKey != nil {
			var userID int64
			if apiKey.User != nil && apiKey.User.ID > 0 {
				userID = apiKey.User.ID
			} else if apiKey.UserID > 0 {
				userID = apiKey.UserID
			}
			if userID > 0 && settingService.IsUserSafetyAllowlisted(c.Request.Context(), userID) {
				c.Next()
				return
			}
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			writeError(c, http.StatusRequestEntityTooLarge, "request body is too large")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		settings := settingService.GetGatewaySensitiveFilterSettings(c.Request.Context())
		match, err := checkGatewaySensitiveRequestBody(c.GetHeader("Content-Type"), body, settings)
		if err != nil || match == nil {
			c.Next()
			return
		}

		// Action 分级：
		//  - "blocked"：硬拦截（高危 builtin + custom 词）
		//  - "warned"：软命中，仅记录事件不拦截请求（容易误报的 builtin 词）
		// 这样合法用户讨论"prompt injection / 逆向工程"等学术话题不被误伤，
		// admin 仍能在风控页面看到事件、必要时升级为 block 规则。
		action := match.Action
		if action == "" {
			action = service.SafetyRiskActionBlocked
		}
		slog.Warn("gateway sensitive filter triggered",
			"path", c.Request.URL.Path,
			"field", match.Path,
			"word", match.Word,
			"source", match.Source,
			"action", action,
		)
		recordGatewaySensitiveRisk(c, opsService, match)
		if action == service.SafetyRiskActionWarned {
			c.Next()
			return
		}
		// 给用户面向的错误信息加 Request ID，方便用户联系 admin 复核时引用具体事件
		userMsg := gatewaySensitiveFilterUserMessage
		if reqID := contextString(c, ctxkey.RequestID); reqID != "" {
			userMsg += " Request ID: " + reqID
		}
		writeError(c, http.StatusForbidden, userMsg)
		c.Abort()
	}
}

func checkGatewaySensitiveRequestBody(contentType string, body []byte, settings *service.GatewaySensitiveFilterSettings) (*service.GatewaySensitiveMatch, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err == nil {
		switch strings.ToLower(strings.TrimSpace(mediaType)) {
		case "multipart/form-data":
			return checkGatewaySensitiveMultipart(body, params["boundary"], settings)
		case "application/x-www-form-urlencoded":
			return checkGatewaySensitiveURLEncodedForm(body, settings)
		}
	}
	return service.CheckGatewaySensitiveJSON(body, settings)
}

func checkGatewaySensitiveMultipart(body []byte, boundary string, settings *service.GatewaySensitiveFilterSettings) (*service.GatewaySensitiveMatch, error) {
	boundary = strings.TrimSpace(boundary)
	if boundary == "" {
		return nil, nil
	}
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	fields := make(map[string]any)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil
		}
		if part.FileName() != "" {
			continue
		}
		name := strings.TrimSpace(part.FormName())
		if !isGatewaySensitiveFormField(name) {
			continue
		}
		valueBytes, err := io.ReadAll(io.LimitReader(part, gatewaySensitiveFormFieldMaxBytes+1))
		if err != nil {
			return nil, nil
		}
		if len(valueBytes) > gatewaySensitiveFormFieldMaxBytes {
			valueBytes = valueBytes[:gatewaySensitiveFormFieldMaxBytes]
		}
		addGatewaySensitiveFormField(fields, name, string(valueBytes))
	}
	return checkGatewaySensitiveFields(fields, settings)
}

func checkGatewaySensitiveURLEncodedForm(body []byte, settings *service.GatewaySensitiveFilterSettings) (*service.GatewaySensitiveMatch, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, nil
	}
	fields := make(map[string]any)
	for name, list := range values {
		if !isGatewaySensitiveFormField(name) {
			continue
		}
		for _, value := range list {
			addGatewaySensitiveFormField(fields, name, value)
		}
	}
	return checkGatewaySensitiveFields(fields, settings)
}

func checkGatewaySensitiveFields(fields map[string]any, settings *service.GatewaySensitiveFilterSettings) (*service.GatewaySensitiveMatch, error) {
	if len(fields) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, nil
	}
	return service.CheckGatewaySensitiveJSON(encoded, settings)
}

func addGatewaySensitiveFormField(fields map[string]any, name string, value string) {
	if fields == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if existing, ok := fields[name]; ok {
		switch list := existing.(type) {
		case []string:
			fields[name] = append(list, value)
		case string:
			fields[name] = []string{list, value}
		default:
			fields[name] = value
		}
		return
	}
	fields[name] = value
}

func isGatewaySensitiveFormField(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "messages", "message", "contents", "content", "parts", "input", "prompt", "instructions", "system", "text", "query":
		return true
	default:
		return false
	}
}

func recordGatewaySensitiveRisk(c *gin.Context, opsService *service.OpsService, match *service.GatewaySensitiveMatch) {
	if c == nil || c.Request == nil || opsService == nil || match == nil {
		return
	}

	// Action 沿用 match.Action（block / warn），让 admin 在事件列表能区分硬拦截还是软提醒
	action := match.Action
	if action == "" {
		action = service.SafetyRiskActionBlocked
	}
	severity := "warning"
	if action == service.SafetyRiskActionBlocked {
		severity = "danger"
	}
	input := &service.SafetyRiskEventInput{
		Method:          c.Request.Method,
		Path:            c.Request.URL.Path,
		ClientIP:        ip.GetTrustedClientIP(c),
		UserAgent:       c.GetHeader("User-Agent"),
		RuleSource:      match.Source,
		RuleWord:        match.Word,
		RulePath:        match.Path,
		Category:        "content_safety",
		Severity:        severity,
		Action:          action,
		AIReviewed:      false,
		AIReviewResult:  "not_used",
		Status:          service.SafetyRiskStatusPending,
		PromptPreview:   match.Preview,
		RequestID:       contextString(c, ctxkey.RequestID),
		ClientRequestID: contextString(c, ctxkey.ClientRequestID),
	}

	if apiKey, ok := GetAPIKeyFromContext(c); ok && apiKey != nil {
		if apiKey.ID > 0 {
			id := apiKey.ID
			input.APIKeyID = &id
		}
		input.APIKeyName = apiKey.Name
		if apiKey.User != nil && apiKey.User.ID > 0 {
			uid := apiKey.User.ID
			input.UserID = &uid
		} else if apiKey.UserID > 0 {
			uid := apiKey.UserID
			input.UserID = &uid
		}
		if apiKey.Group != nil {
			if apiKey.Group.ID > 0 {
				gid := apiKey.Group.ID
				input.GroupID = &gid
			}
			input.GroupName = apiKey.Group.Name
		} else if apiKey.GroupID != nil && *apiKey.GroupID > 0 {
			gid := *apiKey.GroupID
			input.GroupID = &gid
		}
	}

	eventID, err := opsService.RecordSafetyRiskEvent(c.Request.Context(), input)
	if err != nil {
		slog.Warn("failed to record gateway sensitive risk event",
			"error", err,
			"path", input.Path,
			"request_id", input.RequestID,
		)
		return
	}
	// 触发异步 AI 审核（admin 在 settings 配了 enabled + api_key_id + model 才会真跑）。
	// fire-and-forget；不阻塞用户请求；失败仅记日志，事件保持"未经过 AI"。
	if eventID > 0 {
		opsService.TriggerAIReviewAsync(eventID, input.PromptPreview)
	}
}

func contextString(c *gin.Context, key any) string {
	if c == nil || c.Request == nil {
		return ""
	}
	value, _ := c.Request.Context().Value(key).(string)
	return strings.TrimSpace(value)
}

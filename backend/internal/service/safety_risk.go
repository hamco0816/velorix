package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SafetyRiskStatusPending  = "pending"
	SafetyRiskStatusReviewed = "reviewed"
	SafetyRiskStatusCleared  = "cleared"

	SafetyRiskActionBlocked = "blocked"
	SafetyRiskActionWarned  = "warned"

	SafetyRiskReviewSourceLocal = "local"
	SafetyRiskReviewSourceAI    = "ai"
)

type SafetyRiskEvent struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	UserID    *int64 `json:"user_id,omitempty"`
	UserEmail string `json:"user_email"`

	APIKeyID   *int64 `json:"api_key_id,omitempty"`
	APIKeyName string `json:"api_key_name"`

	GroupID   *int64 `json:"group_id,omitempty"`
	GroupName string `json:"group_name"`

	RequestID       string `json:"request_id"`
	ClientRequestID string `json:"client_request_id"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	ClientIP        string `json:"client_ip"`
	UserAgent       string `json:"user_agent"`

	RuleSource string `json:"rule_source"`
	RuleWord   string `json:"rule_word"`
	RulePath   string `json:"rule_path"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Action     string `json:"action"`

	AIReviewed       bool   `json:"ai_reviewed"`
	AIReviewProvider string `json:"ai_review_provider"`
	AIReviewResult   string `json:"ai_review_result"`

	Status        string     `json:"status"`
	PromptPreview string     `json:"prompt_preview"`
	ReviewedByID  *int64     `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote    string     `json:"review_note"`
	ClearedAt     *time.Time `json:"cleared_at,omitempty"`
}

type SafetyRiskEventInput struct {
	CreatedAt time.Time

	UserID     *int64
	APIKeyID   *int64
	APIKeyName string
	GroupID    *int64
	GroupName  string

	RequestID       string
	ClientRequestID string
	Method          string
	Path            string
	ClientIP        string
	UserAgent       string

	RuleSource string
	RuleWord   string
	RulePath   string
	Category   string
	Severity   string
	Action     string

	AIReviewed       bool
	AIReviewProvider string
	AIReviewResult   string

	Status        string
	PromptPreview string
	ReviewNote    string
}

type SafetyRiskEventFilter struct {
	StartTime *time.Time
	EndTime   *time.Time

	Status   string
	Action   string
	Severity string
	Source   string
	Query    string

	UserID   *int64
	APIKeyID *int64
	GroupID  *int64

	Page     int
	PageSize int
}

type SafetyRiskEventList struct {
	Events   []*SafetyRiskEvent `json:"events"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type SafetyRiskReviewInput struct {
	Status           string
	ReviewedByUserID int64
	ReviewNote       string
}

type SafetyRiskClearUserInput struct {
	UserID           int64
	ReviewedByUserID int64
	ReviewNote       string
}

type SafetyRiskClearUserResult struct {
	Cleared int64 `json:"cleared"`
}

// SafetyRiskRuleStat 单条规则的聚合统计，admin 用于判断哪条规则误报多。
//   - TotalHits / BlockedCount / WarnedCount：命中数与拦截动作分布
//   - ClearedCount：admin 或 AI 已归档（多数说明误报）
//   - AIPassCount / AIRejectCount / AIFlagCount：AI 判定分布
type SafetyRiskRuleStat struct {
	RuleWord      string `json:"rule_word"`
	RuleSource    string `json:"rule_source"`
	TotalHits     int64  `json:"total_hits"`
	BlockedCount  int64  `json:"blocked_count"`
	WarnedCount   int64  `json:"warned_count"`
	ClearedCount  int64  `json:"cleared_count"`
	AIPassCount   int64  `json:"ai_pass_count"`
	AIRejectCount int64  `json:"ai_reject_count"`
	AIFlagCount   int64  `json:"ai_flag_count"`
	LastHitAt     string `json:"last_hit_at,omitempty"`
}

// GetSafetyRiskRuleStats 返回时间窗口内的规则命中聚合统计。
// sinceHours <=0 默认 168（7 天）；limit <=0 默认 50；按 TotalHits 降序。
func (s *OpsService) GetSafetyRiskRuleStats(ctx context.Context, sinceHours, limit int) ([]*SafetyRiskRuleStat, error) {
	if s == nil || s.opsRepo == nil {
		return nil, infraerrors.ServiceUnavailable("OPS_REPO_UNAVAILABLE", "Ops repository not available")
	}
	if sinceHours <= 0 {
		sinceHours = 168
	}
	if sinceHours > 24*30 {
		sinceHours = 24 * 30 // 上限 30 天
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	return s.opsRepo.SafetyRiskRuleStats(ctx, sinceHours, limit)
}

func (s *OpsService) RecordSafetyRiskEvent(ctx context.Context, input *SafetyRiskEventInput) (int64, error) {
	if s == nil || s.opsRepo == nil || input == nil {
		return 0, nil
	}
	prepareSafetyRiskEventInput(input)
	id, err := s.opsRepo.InsertSafetyRiskEvent(ctx, input)
	if err != nil {
		log.Printf("[Ops] RecordSafetyRiskEvent failed: %v", err)
		return 0, err
	}
	return id, nil
}

func (s *OpsService) ListSafetyRiskEvents(ctx context.Context, filter *SafetyRiskEventFilter) (*SafetyRiskEventList, error) {
	if s == nil || s.opsRepo == nil {
		return nil, infraerrors.ServiceUnavailable("OPS_REPO_UNAVAILABLE", "Ops repository not available")
	}
	normalizeSafetyRiskFilter(filter)
	result, err := s.opsRepo.ListSafetyRiskEvents(ctx, filter)
	if err != nil {
		return nil, infraerrors.InternalServer("SAFETY_RISK_LIST_FAILED", "Failed to list safety risk events").WithCause(err)
	}
	return result, nil
}

func (s *OpsService) ReviewSafetyRiskEvent(ctx context.Context, eventID int64, input *SafetyRiskReviewInput) error {
	if s == nil || s.opsRepo == nil {
		return infraerrors.ServiceUnavailable("OPS_REPO_UNAVAILABLE", "Ops repository not available")
	}
	if eventID <= 0 {
		return infraerrors.BadRequest("SAFETY_RISK_INVALID_ID", "invalid safety risk event id")
	}
	if input == nil {
		input = &SafetyRiskReviewInput{}
	}
	if input.ReviewedByUserID <= 0 {
		return infraerrors.Unauthorized("UNAUTHORIZED", "Unauthorized")
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = SafetyRiskStatusReviewed
	}
	if !isValidSafetyRiskStatus(status) {
		return infraerrors.BadRequest("SAFETY_RISK_INVALID_STATUS", "invalid safety risk status")
	}
	note := truncateString(strings.TrimSpace(input.ReviewNote), 1000)
	err := s.opsRepo.UpdateSafetyRiskEventStatus(ctx, eventID, status, &input.ReviewedByUserID, note)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return infraerrors.NotFound("SAFETY_RISK_EVENT_NOT_FOUND", "safety risk event not found")
		}
		return infraerrors.InternalServer("SAFETY_RISK_REVIEW_FAILED", "Failed to update safety risk event").WithCause(err)
	}
	return nil
}

func (s *OpsService) ClearSafetyRiskEventsForUser(ctx context.Context, input *SafetyRiskClearUserInput) (*SafetyRiskClearUserResult, error) {
	if s == nil || s.opsRepo == nil {
		return nil, infraerrors.ServiceUnavailable("OPS_REPO_UNAVAILABLE", "Ops repository not available")
	}
	if input == nil || input.UserID <= 0 {
		return nil, infraerrors.BadRequest("SAFETY_RISK_INVALID_USER", "invalid user id")
	}
	if input.ReviewedByUserID <= 0 {
		return nil, infraerrors.Unauthorized("UNAUTHORIZED", "Unauthorized")
	}
	note := strings.TrimSpace(input.ReviewNote)
	if note == "" {
		note = "manual clear"
	}
	note = truncateString(note, 1000)
	cleared, err := s.opsRepo.ClearSafetyRiskEventsForUser(ctx, input.UserID, input.ReviewedByUserID, note)
	if err != nil {
		return nil, infraerrors.InternalServer("SAFETY_RISK_CLEAR_FAILED", "Failed to clear safety risk events").WithCause(err)
	}
	return &SafetyRiskClearUserResult{Cleared: cleared}, nil
}

func prepareSafetyRiskEventInput(input *SafetyRiskEventInput) {
	if input.CreatedAt.IsZero() {
		input.CreatedAt = time.Now()
	}
	input.APIKeyName = truncateString(strings.TrimSpace(input.APIKeyName), 128)
	input.GroupName = truncateString(strings.TrimSpace(input.GroupName), 128)
	input.RequestID = truncateString(strings.TrimSpace(input.RequestID), 128)
	input.ClientRequestID = truncateString(strings.TrimSpace(input.ClientRequestID), 128)
	input.Method = truncateString(strings.TrimSpace(input.Method), 16)
	input.Path = truncateString(strings.TrimSpace(input.Path), 256)
	input.ClientIP = truncateString(strings.TrimSpace(input.ClientIP), 64)
	input.UserAgent = truncateString(strings.TrimSpace(input.UserAgent), 512)
	input.RuleSource = truncateString(strings.TrimSpace(input.RuleSource), 32)
	input.RuleWord = truncateString(strings.TrimSpace(input.RuleWord), 200)
	input.RulePath = truncateString(strings.TrimSpace(input.RulePath), 256)
	input.Category = truncateString(strings.TrimSpace(input.Category), 64)
	input.Severity = truncateString(strings.TrimSpace(input.Severity), 32)
	input.Action = truncateString(strings.TrimSpace(input.Action), 32)
	input.AIReviewProvider = truncateString(strings.TrimSpace(input.AIReviewProvider), 64)
	input.AIReviewResult = truncateString(strings.TrimSpace(input.AIReviewResult), 256)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.PromptPreview = truncateString(strings.TrimSpace(input.PromptPreview), 1000)
	input.ReviewNote = truncateString(strings.TrimSpace(input.ReviewNote), 1000)

	if input.RuleSource == "" {
		input.RuleSource = SafetyRiskReviewSourceLocal
	}
	if input.Category == "" {
		input.Category = "content_safety"
	}
	if input.Severity == "" {
		input.Severity = "warning"
	}
	if input.Action == "" {
		input.Action = SafetyRiskActionBlocked
	}
	if input.AIReviewResult == "" {
		input.AIReviewResult = "not_used"
	}
	if !isValidSafetyRiskStatus(input.Status) {
		input.Status = SafetyRiskStatusPending
	}
}

func normalizeSafetyRiskFilter(filter *SafetyRiskEventFilter) {
	if filter == nil {
		return
	}
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.Action = strings.ToLower(strings.TrimSpace(filter.Action))
	filter.Severity = strings.ToLower(strings.TrimSpace(filter.Severity))
	filter.Source = strings.ToLower(strings.TrimSpace(filter.Source))
	filter.Query = strings.TrimSpace(filter.Query)
	if filter.Status == "all" {
		filter.Status = ""
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 500 {
		filter.PageSize = 500
	}
}

func isValidSafetyRiskStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case SafetyRiskStatusPending, SafetyRiskStatusReviewed, SafetyRiskStatusCleared:
		return true
	default:
		return false
	}
}

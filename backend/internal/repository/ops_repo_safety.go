package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const insertSafetyRiskEventSQL = `
INSERT INTO safety_risk_events (
  created_at,
  user_id,
  api_key_id,
  api_key_name,
  group_id,
  group_name,
  request_id,
  client_request_id,
  method,
  path,
  client_ip,
  user_agent,
  rule_source,
  rule_word,
  rule_path,
  category,
  severity,
  action,
  ai_reviewed,
  ai_review_provider,
  ai_review_result,
  status,
  prompt_preview,
  review_note
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
)`

func (r *opsRepository) InsertSafetyRiskEvent(ctx context.Context, input *service.SafetyRiskEventInput) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil ops repository")
	}
	if input == nil {
		return 0, fmt.Errorf("nil input")
	}
	var id int64
	err := r.db.QueryRowContext(
		ctx,
		insertSafetyRiskEventSQL+" RETURNING id",
		input.CreatedAt,
		opsNullInt64(input.UserID),
		opsNullInt64(input.APIKeyID),
		input.APIKeyName,
		opsNullInt64(input.GroupID),
		input.GroupName,
		input.RequestID,
		input.ClientRequestID,
		input.Method,
		input.Path,
		input.ClientIP,
		input.UserAgent,
		input.RuleSource,
		input.RuleWord,
		input.RulePath,
		input.Category,
		input.Severity,
		input.Action,
		input.AIReviewed,
		input.AIReviewProvider,
		input.AIReviewResult,
		input.Status,
		input.PromptPreview,
		input.ReviewNote,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *opsRepository) ListSafetyRiskEvents(ctx context.Context, filter *service.SafetyRiskEventFilter) (*service.SafetyRiskEventList, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil {
		filter = &service.SafetyRiskEventFilter{}
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 500 {
		pageSize = 500
	}

	where, args := buildSafetyRiskEventsWhere(filter)
	countSQL := "SELECT COUNT(*) FROM safety_risk_events e LEFT JOIN users u ON e.user_id = u.id " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	argsWithLimit := append(args, pageSize, offset)
	selectSQL := `
SELECT
  e.id,
  e.created_at,
  e.user_id,
  COALESCE(u.email, ''),
  e.api_key_id,
  COALESCE(e.api_key_name, ''),
  e.group_id,
  COALESCE(e.group_name, ''),
  COALESCE(e.request_id, ''),
  COALESCE(e.client_request_id, ''),
  COALESCE(e.method, ''),
  COALESCE(e.path, ''),
  COALESCE(e.client_ip, ''),
  COALESCE(e.user_agent, ''),
  COALESCE(e.rule_source, ''),
  COALESCE(e.rule_word, ''),
  COALESCE(e.rule_path, ''),
  COALESCE(e.category, ''),
  COALESCE(e.severity, ''),
  COALESCE(e.action, ''),
  COALESCE(e.ai_reviewed, false),
  COALESCE(e.ai_review_provider, ''),
  COALESCE(e.ai_review_result, ''),
  COALESCE(e.status, ''),
  COALESCE(e.prompt_preview, ''),
  e.reviewed_by_user_id,
  e.reviewed_at,
  COALESCE(e.review_note, ''),
  e.cleared_at
FROM safety_risk_events e
LEFT JOIN users u ON e.user_id = u.id
` + where + `
ORDER BY e.created_at DESC
LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)

	rows, err := r.db.QueryContext(ctx, selectSQL, argsWithLimit...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.SafetyRiskEvent, 0, pageSize)
	for rows.Next() {
		var item service.SafetyRiskEvent
		var userID sql.NullInt64
		var apiKeyID sql.NullInt64
		var groupID sql.NullInt64
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime
		var clearedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.CreatedAt,
			&userID,
			&item.UserEmail,
			&apiKeyID,
			&item.APIKeyName,
			&groupID,
			&item.GroupName,
			&item.RequestID,
			&item.ClientRequestID,
			&item.Method,
			&item.Path,
			&item.ClientIP,
			&item.UserAgent,
			&item.RuleSource,
			&item.RuleWord,
			&item.RulePath,
			&item.Category,
			&item.Severity,
			&item.Action,
			&item.AIReviewed,
			&item.AIReviewProvider,
			&item.AIReviewResult,
			&item.Status,
			&item.PromptPreview,
			&reviewedBy,
			&reviewedAt,
			&item.ReviewNote,
			&clearedAt,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			v := userID.Int64
			item.UserID = &v
		}
		if apiKeyID.Valid {
			v := apiKeyID.Int64
			item.APIKeyID = &v
		}
		if groupID.Valid {
			v := groupID.Int64
			item.GroupID = &v
		}
		if reviewedBy.Valid {
			v := reviewedBy.Int64
			item.ReviewedByID = &v
		}
		if reviewedAt.Valid {
			t := reviewedAt.Time
			item.ReviewedAt = &t
		}
		if clearedAt.Valid {
			t := clearedAt.Time
			item.ClearedAt = &t
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &service.SafetyRiskEventList{
		Events:   items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *opsRepository) UpdateSafetyRiskEventStatus(ctx context.Context, id int64, status string, reviewedByUserID *int64, reviewNote string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil ops repository")
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE safety_risk_events
SET
  status = $2,
  reviewed_by_user_id = $3,
  reviewed_at = NOW(),
  review_note = $4,
  cleared_at = CASE WHEN $2 = 'cleared' THEN NOW() ELSE cleared_at END
WHERE id = $1
`, id, status, opsNullInt64(reviewedByUserID), reviewNote)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *opsRepository) ClearSafetyRiskEventsForUser(ctx context.Context, userID int64, reviewedByUserID int64, reviewNote string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil ops repository")
	}
	res, err := r.db.ExecContext(ctx, `
UPDATE safety_risk_events
SET
  status = 'cleared',
  reviewed_by_user_id = $2,
  reviewed_at = NOW(),
  review_note = $3,
  cleared_at = NOW()
WHERE user_id = $1
  AND status <> 'cleared'
`, userID, reviewedByUserID, reviewNote)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func buildSafetyRiskEventsWhere(filter *service.SafetyRiskEventFilter) (string, []any) {
	conditions := make([]string, 0, 8)
	args := make([]any, 0, 8)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if filter != nil {
		if filter.StartTime != nil && !filter.StartTime.IsZero() {
			add("e.created_at >= $%d", *filter.StartTime)
		}
		if filter.EndTime != nil && !filter.EndTime.IsZero() {
			add("e.created_at <= $%d", *filter.EndTime)
		}
		if status := strings.TrimSpace(filter.Status); status != "" && !strings.EqualFold(status, "all") {
			add("e.status = $%d", strings.ToLower(status))
		}
		if action := strings.TrimSpace(filter.Action); action != "" {
			add("e.action = $%d", strings.ToLower(action))
		}
		if severity := strings.TrimSpace(filter.Severity); severity != "" {
			add("e.severity = $%d", strings.ToLower(severity))
		}
		if source := strings.TrimSpace(filter.Source); source != "" {
			add("e.rule_source = $%d", strings.ToLower(source))
		}
		if filter.UserID != nil && *filter.UserID > 0 {
			add("e.user_id = $%d", *filter.UserID)
		}
		if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
			add("e.api_key_id = $%d", *filter.APIKeyID)
		}
		if filter.GroupID != nil && *filter.GroupID > 0 {
			add("e.group_id = $%d", *filter.GroupID)
		}
		if q := strings.TrimSpace(filter.Query); q != "" {
			args = append(args, q)
			idx := len(args)
			conditions = append(conditions, fmt.Sprintf(`(
  e.request_id ILIKE '%%' || $%d || '%%'
  OR e.client_request_id ILIKE '%%' || $%d || '%%'
  OR e.path ILIKE '%%' || $%d || '%%'
  OR e.rule_word ILIKE '%%' || $%d || '%%'
  OR e.prompt_preview ILIKE '%%' || $%d || '%%'
  OR e.api_key_name ILIKE '%%' || $%d || '%%'
  OR COALESCE(u.email, '') ILIKE '%%' || $%d || '%%'
)`, idx, idx, idx, idx, idx, idx, idx))
		}
	}

	if len(conditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

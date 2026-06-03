package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	apperrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	SupportConversationStatusOpen   = "open"
	SupportConversationStatusClosed = "closed"

	SupportSenderUser   = "user"
	SupportSenderAdmin  = "admin"
	SupportSenderSystem = "system"

	supportMessageMaxLen = 4000
)

var (
	ErrSupportMessageEmpty      = apperrors.BadRequest("SUPPORT_MESSAGE_EMPTY", "Message cannot be empty")
	ErrSupportMessageTooLong    = apperrors.BadRequest("SUPPORT_MESSAGE_TOO_LONG", "Message is too long")
	ErrSupportConversationID    = apperrors.BadRequest("INVALID_SUPPORT_CONVERSATION", "Invalid conversation ID")
	ErrSupportConversationNF    = apperrors.NotFound("SUPPORT_CONVERSATION_NOT_FOUND", "Support conversation not found")
	ErrSupportConversationState = apperrors.Conflict("SUPPORT_CONVERSATION_CONFLICT", "Conversation state has changed")
	ErrSupportNotInitialized    = apperrors.ServiceUnavailable("SUPPORT_NOT_INITIALIZED", "Support service is not initialized")
)

type SupportService struct {
	db  *sql.DB
	hub *SupportHub
}

func NewSupportService(db *sql.DB, hub *SupportHub) *SupportService {
	return &SupportService{db: db, hub: hub}
}

func (s *SupportService) Hub() *SupportHub {
	if s == nil {
		return nil
	}
	return s.hub
}

type SupportConversation struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	UserEmail        string     `json:"user_email,omitempty"`
	Username         string     `json:"username,omitempty"`
	Status           string     `json:"status"`
	Subject          string     `json:"subject"`
	LastMessage      string     `json:"last_message"`
	LastMessageAt    *time.Time `json:"last_message_at,omitempty"`
	UserUnreadCount  int        `json:"user_unread_count"`
	AdminUnreadCount int        `json:"admin_unread_count"`
	UserLastReadAt   *time.Time `json:"user_last_read_at,omitempty"`
	AdminLastReadAt  *time.Time `json:"admin_last_read_at,omitempty"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type SupportMessage struct {
	ID             int64     `json:"id"`
	ConversationID int64     `json:"conversation_id"`
	SenderType     string    `json:"sender_type"`
	SenderID       *int64    `json:"sender_id,omitempty"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type SupportConversationDetail struct {
	Conversation *SupportConversation `json:"conversation"`
	Messages     []SupportMessage     `json:"messages"`
}

type SupportMessageResult struct {
	Conversation *SupportConversation `json:"conversation"`
	Message      *SupportMessage      `json:"message"`
}

type SupportListFilters struct {
	Status string
	Search string
}

type supportQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *SupportService) ensureReady() error {
	if s == nil || s.db == nil {
		return ErrSupportNotInitialized
	}
	return nil
}

func (s *SupportService) GetUserConversationDetail(ctx context.Context, userID int64, limit int) (*SupportConversationDetail, error) {
	conv, err := s.GetOpenUserConversation(ctx, userID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return &SupportConversationDetail{Messages: []SupportMessage{}}, nil
	}
	messages, err := s.ListMessages(ctx, conv.ID, 0, limit)
	if err != nil {
		return nil, err
	}
	return &SupportConversationDetail{Conversation: conv, Messages: messages}, nil
}

func (s *SupportService) GetOpenUserConversation(ctx context.Context, userID int64) (*SupportConversation, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrSupportConversationID
	}

	conv, err := s.getOpenConversationForUser(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get open support conversation: %w", err)
	}
	return conv, nil
}

func (s *SupportService) GetOrCreateUserConversation(ctx context.Context, userID int64) (*SupportConversation, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrSupportConversationID
	}

	conv, err := s.getOpenConversationForUser(ctx, userID)
	if err == nil {
		return conv, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("get open support conversation: %w", err)
	}

	var id int64
	if err := s.db.QueryRowContext(ctx, `
		INSERT INTO support_conversations (user_id, status, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, userID, SupportConversationStatusOpen).Scan(&id); err != nil {
		// The partial unique index may win a race; re-read before surfacing the error.
		conv, retryErr := s.getOpenConversationForUser(ctx, userID)
		if retryErr == nil {
			return conv, nil
		}
		return nil, fmt.Errorf("create support conversation: %w", err)
	}

	return s.getConversationByID(ctx, id)
}

func (s *SupportService) ListUserMessages(ctx context.Context, userID int64, beforeID int64, limit int) (*SupportConversationDetail, error) {
	conv, err := s.GetOpenUserConversation(ctx, userID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return &SupportConversationDetail{Messages: []SupportMessage{}}, nil
	}
	messages, err := s.ListMessages(ctx, conv.ID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	return &SupportConversationDetail{Conversation: conv, Messages: messages}, nil
}

func (s *SupportService) SendUserMessage(ctx context.Context, userID int64, content string) (*SupportMessageResult, error) {
	content, err := normalizeSupportMessage(content)
	if err != nil {
		return nil, err
	}
	conv, err := s.GetOrCreateUserConversation(ctx, userID)
	if err != nil {
		return nil, err
	}

	result, err := s.insertMessageAndUpdateConversation(ctx, conv.ID, SupportSenderUser, userID, content)
	if err != nil {
		return nil, err
	}

	s.broadcastMessage(result)
	return result, nil
}

func (s *SupportService) SendAdminMessage(ctx context.Context, adminID int64, conversationID int64, content string) (*SupportMessageResult, error) {
	content, err := normalizeSupportMessage(content)
	if err != nil {
		return nil, err
	}
	if conversationID <= 0 {
		return nil, ErrSupportConversationID
	}

	result, err := s.insertMessageAndUpdateConversation(ctx, conversationID, SupportSenderAdmin, adminID, content)
	if err != nil {
		return nil, err
	}

	s.broadcastMessage(result)
	return result, nil
}

func (s *SupportService) insertMessageAndUpdateConversation(ctx context.Context, conversationID int64, senderType string, senderID int64, content string) (*SupportMessageResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin support message tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	conv, err := querySupportConversationByID(ctx, tx, conversationID)
	if err == sql.ErrNoRows {
		return nil, ErrSupportConversationNF
	}
	if err != nil {
		return nil, fmt.Errorf("get support conversation: %w", err)
	}

	now := time.Now().UTC()
	var msg SupportMessage
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO support_messages (conversation_id, sender_type, sender_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, conversation_id, sender_type, sender_id, content, created_at
	`, conversationID, senderType, senderID, content, now).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderType,
		nullableInt64Scanner(&msg.SenderID),
		&msg.Content,
		&msg.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert support message: %w", err)
	}

	var updateErr error
	switch senderType {
	case SupportSenderUser:
		_, updateErr = tx.ExecContext(ctx, `
			UPDATE support_conversations
			SET status = $1,
			    closed_at = NULL,
			    last_message = $2,
			    last_message_at = $3,
			    admin_unread_count = admin_unread_count + 1,
			    updated_at = $3
			WHERE id = $4 AND deleted_at IS NULL
		`, SupportConversationStatusOpen, content, now, conversationID)
	case SupportSenderAdmin:
		_, updateErr = tx.ExecContext(ctx, `
			UPDATE support_conversations
			SET status = $1,
			    closed_at = NULL,
			    last_message = $2,
			    last_message_at = $3,
			    user_unread_count = user_unread_count + 1,
			    updated_at = $3
			WHERE id = $4 AND deleted_at IS NULL
		`, SupportConversationStatusOpen, content, now, conversationID)
	default:
		return nil, ErrSupportConversationState
	}
	if updateErr != nil {
		return nil, fmt.Errorf("update support conversation: %w", updateErr)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support message tx: %w", err)
	}

	updated, err := s.getConversationByID(ctx, conv.ID)
	if err != nil {
		return nil, err
	}
	return &SupportMessageResult{Conversation: updated, Message: &msg}, nil
}

func (s *SupportService) ListMessages(ctx context.Context, conversationID int64, beforeID int64, limit int) ([]SupportMessage, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if conversationID <= 0 {
		return nil, ErrSupportConversationID
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, conversation_id, sender_type, sender_id, content, created_at
		FROM support_messages
		WHERE conversation_id = $1 AND deleted_at IS NULL
	`
	args := []any{conversationID}
	if beforeID > 0 {
		args = append(args, beforeID)
		query += fmt.Sprintf(" AND id < $%d", len(args))
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d", len(args))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list support messages: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	messages := make([]SupportMessage, 0, limit)
	for rows.Next() {
		var msg SupportMessage
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderType,
			nullableInt64Scanner(&msg.SenderID),
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan support message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support messages: %w", err)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (s *SupportService) ListAdminConversations(ctx context.Context, params pagination.PaginationParams, filters SupportListFilters) ([]SupportConversation, *pagination.PaginationResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, nil, err
	}

	limit := params.Limit()
	offset := params.Offset()
	where := []string{"c.deleted_at IS NULL"}
	args := make([]any, 0, 4)

	status := strings.ToLower(strings.TrimSpace(filters.Status))
	if status != "" && status != "all" {
		if status != SupportConversationStatusOpen && status != SupportConversationStatusClosed {
			status = SupportConversationStatusOpen
		}
		args = append(args, status)
		where = append(where, fmt.Sprintf("c.status = $%d", len(args)))
	}

	search := strings.TrimSpace(filters.Search)
	if search != "" {
		if len(search) > 200 {
			search = search[:200]
		}
		args = append(args, "%"+search+"%")
		idx := len(args)
		where = append(where, fmt.Sprintf("(u.email ILIKE $%d OR u.username ILIKE $%d OR c.last_message ILIKE $%d OR CAST(c.id AS TEXT) ILIKE $%d)", idx, idx, idx, idx))
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")
	countQuery := `
		SELECT COUNT(*)
		FROM support_conversations c
		LEFT JOIN users u ON u.id = c.user_id
	` + " " + whereSQL

	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count support conversations: %w", err)
	}

	dataArgs := append([]any{}, args...)
	dataArgs = append(dataArgs, limit, offset)
	limitParam := len(dataArgs) - 1
	offsetParam := len(dataArgs)
	query := `
		SELECT c.id, c.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''),
		       c.status, c.subject, c.last_message, c.last_message_at,
		       c.user_unread_count, c.admin_unread_count,
		       c.user_last_read_at, c.admin_last_read_at, c.closed_at,
		       c.created_at, c.updated_at
		FROM support_conversations c
		LEFT JOIN users u ON u.id = c.user_id
	` + " " + whereSQL + fmt.Sprintf(`
		ORDER BY c.admin_unread_count DESC,
		         COALESCE(c.last_message_at, c.updated_at) DESC,
		         c.id DESC
		LIMIT $%d OFFSET $%d
	`, limitParam, offsetParam)

	rows, err := s.db.QueryContext(ctx, query, dataArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("list support conversations: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	items := make([]SupportConversation, 0, limit)
	for rows.Next() {
		conv, err := scanSupportConversation(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *conv)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate support conversations: %w", err)
	}

	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages < 1 {
		pages = 1
	}
	return items, &pagination.PaginationResult{
		Total:    total,
		Page:     params.Page,
		PageSize: limit,
		Pages:    pages,
	}, nil
}

func (s *SupportService) GetAdminConversationDetail(ctx context.Context, conversationID int64, limit int) (*SupportConversationDetail, error) {
	if conversationID <= 0 {
		return nil, ErrSupportConversationID
	}
	conv, err := s.getConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	messages, err := s.ListMessages(ctx, conversationID, 0, limit)
	if err != nil {
		return nil, err
	}
	return &SupportConversationDetail{Conversation: conv, Messages: messages}, nil
}

func (s *SupportService) MarkUserRead(ctx context.Context, userID int64) (*SupportConversation, error) {
	conv, err := s.GetOrCreateUserConversation(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		UPDATE support_conversations
		SET user_unread_count = 0, user_last_read_at = $1, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`, now, conv.ID, userID); err != nil {
		return nil, fmt.Errorf("mark support user read: %w", err)
	}
	updated, err := s.getConversationByID(ctx, conv.ID)
	if err != nil {
		return nil, err
	}
	if s.hub != nil {
		s.hub.BroadcastToUser(userID, SupportRealtimeEvent{
			Type:           SupportEventUnread,
			ConversationID: updated.ID,
			UserID:         userID,
			Conversation:   updated,
			UnreadCount:    0,
		})
	}
	return updated, nil
}

func (s *SupportService) MarkAdminRead(ctx context.Context, conversationID int64) (*SupportConversation, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if conversationID <= 0 {
		return nil, ErrSupportConversationID
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		UPDATE support_conversations
		SET admin_unread_count = 0, admin_last_read_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, now, conversationID); err != nil {
		return nil, fmt.Errorf("mark support admin read: %w", err)
	}
	updated, err := s.getConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	s.broadcastConversationUpdated(updated)
	return updated, nil
}

func (s *SupportService) CloseConversation(ctx context.Context, conversationID int64) (*SupportConversation, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if conversationID <= 0 {
		return nil, ErrSupportConversationID
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		UPDATE support_conversations
		SET status = $1, closed_at = $2, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, SupportConversationStatusClosed, now, conversationID)
	if err != nil {
		return nil, fmt.Errorf("close support conversation: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return nil, ErrSupportConversationNF
	}
	updated, err := s.getConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	s.broadcastConversationUpdated(updated)
	return updated, nil
}

func (s *SupportService) ReopenConversation(ctx context.Context, conversationID int64) (*SupportConversation, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if conversationID <= 0 {
		return nil, ErrSupportConversationID
	}

	conv, err := s.getConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	open, err := s.getOpenConversationForUser(ctx, conv.UserID)
	if err == nil && open.ID != conversationID {
		return nil, ErrSupportConversationState
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("check open support conversation: %w", err)
	}

	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		UPDATE support_conversations
		SET status = $1, closed_at = NULL, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, SupportConversationStatusOpen, now, conversationID); err != nil {
		return nil, fmt.Errorf("reopen support conversation: %w", err)
	}
	updated, err := s.getConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	s.broadcastConversationUpdated(updated)
	return updated, nil
}

func (s *SupportService) GetAdminUnreadCount(ctx context.Context) (int, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}
	var count sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(admin_unread_count), 0)
		FROM support_conversations
		WHERE status = $1 AND deleted_at IS NULL
	`, SupportConversationStatusOpen).Scan(&count); err != nil {
		return 0, fmt.Errorf("get support admin unread: %w", err)
	}
	if !count.Valid {
		return 0, nil
	}
	return int(count.Int64), nil
}

func (s *SupportService) getOpenConversationForUser(ctx context.Context, userID int64) (*SupportConversation, error) {
	return queryOpenSupportConversationForUser(ctx, s.db, userID)
}

func (s *SupportService) getConversationByID(ctx context.Context, id int64) (*SupportConversation, error) {
	conv, err := querySupportConversationByID(ctx, s.db, id)
	if err == sql.ErrNoRows {
		return nil, ErrSupportConversationNF
	}
	return conv, err
}

func (s *SupportService) broadcastMessage(result *SupportMessageResult) {
	if s == nil || s.hub == nil || result == nil || result.Conversation == nil || result.Message == nil {
		return
	}
	event := SupportRealtimeEvent{
		Type:           SupportEventMessage,
		ConversationID: result.Conversation.ID,
		UserID:         result.Conversation.UserID,
		Message:        result.Message,
		Conversation:   result.Conversation,
	}
	s.hub.BroadcastToUser(result.Conversation.UserID, event)
	s.hub.BroadcastAdmins(event)
}

func (s *SupportService) broadcastConversationUpdated(conv *SupportConversation) {
	if s == nil || s.hub == nil || conv == nil {
		return
	}
	event := SupportRealtimeEvent{
		Type:           SupportEventConversationUpdated,
		ConversationID: conv.ID,
		UserID:         conv.UserID,
		Conversation:   conv,
	}
	s.hub.BroadcastToUser(conv.UserID, event)
	s.hub.BroadcastAdmins(event)
}

func queryOpenSupportConversationForUser(ctx context.Context, q supportQueryer, userID int64) (*SupportConversation, error) {
	row := q.QueryRowContext(ctx, `
		SELECT c.id, c.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''),
		       c.status, c.subject, c.last_message, c.last_message_at,
		       c.user_unread_count, c.admin_unread_count,
		       c.user_last_read_at, c.admin_last_read_at, c.closed_at,
		       c.created_at, c.updated_at
		FROM support_conversations c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.user_id = $1 AND c.status = $2 AND c.deleted_at IS NULL
		ORDER BY c.id DESC
		LIMIT 1
	`, userID, SupportConversationStatusOpen)
	return scanSupportConversationRow(row)
}

func querySupportConversationByID(ctx context.Context, q supportQueryer, id int64) (*SupportConversation, error) {
	row := q.QueryRowContext(ctx, `
		SELECT c.id, c.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''),
		       c.status, c.subject, c.last_message, c.last_message_at,
		       c.user_unread_count, c.admin_unread_count,
		       c.user_last_read_at, c.admin_last_read_at, c.closed_at,
		       c.created_at, c.updated_at
		FROM support_conversations c
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`, id)
	return scanSupportConversationRow(row)
}

type supportRowScanner interface {
	Scan(dest ...any) error
}

func scanSupportConversationRow(row supportRowScanner) (*SupportConversation, error) {
	var conv SupportConversation
	var lastMessageAt sql.NullTime
	var userLastReadAt sql.NullTime
	var adminLastReadAt sql.NullTime
	var closedAt sql.NullTime
	if err := row.Scan(
		&conv.ID,
		&conv.UserID,
		&conv.UserEmail,
		&conv.Username,
		&conv.Status,
		&conv.Subject,
		&conv.LastMessage,
		&lastMessageAt,
		&conv.UserUnreadCount,
		&conv.AdminUnreadCount,
		&userLastReadAt,
		&adminLastReadAt,
		&closedAt,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	); err != nil {
		return nil, err
	}
	conv.LastMessageAt = timePtrFromNull(lastMessageAt)
	conv.UserLastReadAt = timePtrFromNull(userLastReadAt)
	conv.AdminLastReadAt = timePtrFromNull(adminLastReadAt)
	conv.ClosedAt = timePtrFromNull(closedAt)
	return &conv, nil
}

func scanSupportConversation(rows *sql.Rows) (*SupportConversation, error) {
	conv, err := scanSupportConversationRow(rows)
	if err != nil {
		return nil, fmt.Errorf("scan support conversation: %w", err)
	}
	return conv, nil
}

func timePtrFromNull(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func normalizeSupportMessage(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ErrSupportMessageEmpty
	}
	if len([]rune(content)) > supportMessageMaxLen {
		return "", ErrSupportMessageTooLong
	}
	return content, nil
}

type nullableInt64PtrScanner struct {
	target **int64
}

func nullableInt64Scanner(target **int64) *nullableInt64PtrScanner {
	return &nullableInt64PtrScanner{target: target}
}

func (s *nullableInt64PtrScanner) Scan(src any) error {
	if s == nil || s.target == nil {
		return nil
	}
	if src == nil {
		*s.target = nil
		return nil
	}
	var v int64
	switch value := src.(type) {
	case int64:
		v = value
	case int:
		v = int64(value)
	case int32:
		v = int64(value)
	default:
		return fmt.Errorf("unsupported int64 scan type %T", src)
	}
	*s.target = &v
	return nil
}

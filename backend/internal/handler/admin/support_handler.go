package admin

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var supportAdminWSUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return isAllowedSupportAdminWSOrigin(r)
	},
	Subprotocols: []string{"sub2api-support-admin"},
}

type SupportHandler struct {
	supportService *service.SupportService
}

func NewSupportHandler(supportService *service.SupportService) *SupportHandler {
	return &SupportHandler{supportService: supportService}
}

type SendSupportReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *SupportHandler) ListConversations(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortOrder: pagination.SortOrderDesc,
	}
	filters := service.SupportListFilters{
		Status: strings.TrimSpace(c.DefaultQuery("status", service.SupportConversationStatusOpen)),
		Search: strings.TrimSpace(c.Query("search")),
	}
	items, pageResult, err := h.supportService.ListAdminConversations(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, pageResult.Total, page, pageSize)
}

func (h *SupportHandler) GetMessages(c *gin.Context) {
	conversationID, ok := parseSupportConversationID(c)
	if !ok {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}
	messages, err := h.supportService.ListMessages(c.Request.Context(), conversationID, parseInt64Query(c, "before_id"), parseSupportLimit(c, 50))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, messages)
}

func (h *SupportHandler) SendMessage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	conversationID, ok := parseSupportConversationID(c)
	if !ok {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}
	var req SendSupportReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.supportService.SendAdminMessage(c.Request.Context(), subject.UserID, conversationID, req.Content)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportHandler) MarkRead(c *gin.Context) {
	conversationID, ok := parseSupportConversationID(c)
	if !ok {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}
	conv, err := h.supportService.MarkAdminRead(c.Request.Context(), conversationID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *SupportHandler) CloseConversation(c *gin.Context) {
	conversationID, ok := parseSupportConversationID(c)
	if !ok {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}
	conv, err := h.supportService.CloseConversation(c.Request.Context(), conversationID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *SupportHandler) ReopenConversation(c *gin.Context) {
	conversationID, ok := parseSupportConversationID(c)
	if !ok {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}
	conv, err := h.supportService.ReopenConversation(c.Request.Context(), conversationID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *SupportHandler) Unread(c *gin.Context) {
	count, err := h.supportService.GetAdminUnreadCount(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"unread_count": count})
}

func (h *SupportHandler) WebSocket(c *gin.Context) {
	if h == nil || h.supportService == nil || h.supportService.Hub() == nil {
		response.ErrorFrom(c, service.ErrSupportNotInitialized)
		return
	}

	conn, err := supportAdminWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LegacyPrintf("handler.admin.support", "[SupportAdminWS] upgrade failed: %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	sub := service.NewSupportSubscriber(0, true)
	h.supportService.Hub().Register(sub)
	defer h.supportService.Hub().Unregister(sub)

	unread, _ := h.supportService.GetAdminUnreadCount(c.Request.Context())
	h.supportService.Hub().SendToSubscriber(sub, service.SupportRealtimeEvent{
		Type:        service.SupportEventReady,
		UnreadCount: unread,
	})

	serveAdminSupportWS(c.Request.Context(), conn, sub)
}

func serveAdminSupportWS(parentCtx context.Context, conn *websocket.Conn, sub *service.SupportSubscriber) {
	if conn == nil || sub == nil {
		return
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var closeOnce sync.Once
	closeConn := func() {
		closeOnce.Do(func() {
			_ = conn.Close()
		})
	}

	closeFrameCh := make(chan []byte, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		conn.SetReadLimit(2048)
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			logger.LegacyPrintf("handler.admin.support", "[SupportAdminWS] set read deadline failed: %v", err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		})
		conn.SetCloseHandler(func(code int, text string) error {
			select {
			case closeFrameCh <- websocket.FormatCloseMessage(code, text):
			default:
			}
			cancel()
			return nil
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					logger.LegacyPrintf("handler.admin.support", "[SupportAdminWS] read failed: %v", err)
				}
				return
			}
		}
	}()

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()
	writeWithTimeout := func(messageType int, data []byte) error {
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return err
		}
		return conn.WriteMessage(messageType, data)
	}
	sendClose := func(closeFrame []byte) {
		if closeFrame == nil {
			closeFrame = websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		}
		_ = writeWithTimeout(websocket.CloseMessage, closeFrame)
	}

	for {
		select {
		case payload, ok := <-sub.Send():
			if !ok {
				closeConn()
				wg.Wait()
				return
			}
			if err := writeWithTimeout(websocket.TextMessage, payload); err != nil {
				logger.LegacyPrintf("handler.admin.support", "[SupportAdminWS] write failed: %v", err)
				cancel()
				closeConn()
				wg.Wait()
				return
			}
		case <-pingTicker.C:
			if err := writeWithTimeout(websocket.PingMessage, nil); err != nil {
				logger.LegacyPrintf("handler.admin.support", "[SupportAdminWS] ping failed: %v", err)
				cancel()
				closeConn()
				wg.Wait()
				return
			}
		case closeFrame := <-closeFrameCh:
			sendClose(closeFrame)
			closeConn()
			wg.Wait()
			return
		case <-ctx.Done():
			var closeFrame []byte
			select {
			case closeFrame = <-closeFrameCh:
			default:
			}
			sendClose(closeFrame)
			closeConn()
			wg.Wait()
			return
		}
	}
}

func parseSupportConversationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	return id, err == nil && id > 0
}

func parseSupportLimit(c *gin.Context, fallback int) int {
	limit := fallback
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func parseInt64Query(c *gin.Context, key string) int64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func isAllowedSupportAdminWSOrigin(r *http.Request) bool {
	if r == nil {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Hostname() == "" {
		return false
	}
	return strings.EqualFold(parsed.Hostname(), hostWithoutPort(r.Host))
}

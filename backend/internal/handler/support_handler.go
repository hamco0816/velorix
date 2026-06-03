package handler

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	supportWSWriteTimeout = 10 * time.Second
	supportWSPongWait     = 60 * time.Second
	supportWSPingInterval = 30 * time.Second
	supportWSMaxReadBytes = 2048
)

var supportWSUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return isAllowedSupportWSOrigin(r)
	},
	Subprotocols: []string{"sub2api-support"},
}

type SupportHandler struct {
	supportService *service.SupportService
}

func NewSupportHandler(supportService *service.SupportService) *SupportHandler {
	return &SupportHandler{supportService: supportService}
}

type SendSupportMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *SupportHandler) GetConversation(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	detail, err := h.supportService.GetUserConversationDetail(c.Request.Context(), subject.UserID, parseSupportLimit(c, 50))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
}

func (h *SupportHandler) ListMessages(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	detail, err := h.supportService.ListUserMessages(c.Request.Context(), subject.UserID, parseInt64Query(c, "before_id"), parseSupportLimit(c, 50))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, detail)
}

func (h *SupportHandler) SendMessage(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req SendSupportMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.supportService.SendUserMessage(c.Request.Context(), subject.UserID, req.Content)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *SupportHandler) MarkRead(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	conv, err := h.supportService.MarkUserRead(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, conv)
}

func (h *SupportHandler) WebSocket(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	if h == nil || h.supportService == nil || h.supportService.Hub() == nil {
		response.ErrorFrom(c, service.ErrSupportNotInitialized)
		return
	}

	conn, err := supportWSUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LegacyPrintf("handler.support", "[SupportWS] user upgrade failed: %v", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	sub := service.NewSupportSubscriber(subject.UserID, false)
	h.supportService.Hub().Register(sub)
	defer h.supportService.Hub().Unregister(sub)

	if conv, err := h.supportService.GetOpenUserConversation(c.Request.Context(), subject.UserID); err == nil && conv != nil {
		h.supportService.Hub().SendToSubscriber(sub, service.SupportRealtimeEvent{
			Type:           service.SupportEventReady,
			ConversationID: conv.ID,
			UserID:         subject.UserID,
			Conversation:   conv,
			UnreadCount:    conv.UserUnreadCount,
		})
	} else {
		h.supportService.Hub().SendToSubscriber(sub, service.SupportRealtimeEvent{
			Type:        service.SupportEventReady,
			UserID:      subject.UserID,
			UnreadCount: 0,
		})
	}

	serveSupportWS(c.Request.Context(), conn, sub, "user")
}

func serveSupportWS(parentCtx context.Context, conn *websocket.Conn, sub *service.SupportSubscriber, logScope string) {
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

		conn.SetReadLimit(supportWSMaxReadBytes)
		if err := conn.SetReadDeadline(time.Now().Add(supportWSPongWait)); err != nil {
			logger.LegacyPrintf("handler.support", "[SupportWS] %s set read deadline failed: %v", logScope, err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(supportWSPongWait))
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
					logger.LegacyPrintf("handler.support", "[SupportWS] %s read failed: %v", logScope, err)
				}
				return
			}
		}
	}()

	pingTicker := time.NewTicker(supportWSPingInterval)
	defer pingTicker.Stop()

	writeWithTimeout := func(messageType int, data []byte) error {
		if err := conn.SetWriteDeadline(time.Now().Add(supportWSWriteTimeout)); err != nil {
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
				logger.LegacyPrintf("handler.support", "[SupportWS] %s write failed: %v", logScope, err)
				cancel()
				closeConn()
				wg.Wait()
				return
			}
		case <-pingTicker.C:
			if err := writeWithTimeout(websocket.PingMessage, nil); err != nil {
				logger.LegacyPrintf("handler.support", "[SupportWS] %s ping failed: %v", logScope, err)
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

func isAllowedSupportWSOrigin(r *http.Request) bool {
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

func hostWithoutPort(hostport string) string {
	hostport = strings.TrimSpace(hostport)
	if hostport == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return host
	}
	if strings.HasPrefix(hostport, "[") && strings.HasSuffix(hostport, "]") {
		return strings.Trim(hostport, "[]")
	}
	return strings.Split(hostport, ":")[0]
}

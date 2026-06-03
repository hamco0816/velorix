package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	SupportEventReady               = "support.ready"
	SupportEventMessage             = "support.message"
	SupportEventConversationUpdated = "support.conversation_updated"
	SupportEventUnread              = "support.unread"
)

type SupportRealtimeEvent struct {
	Type           string               `json:"type"`
	ConversationID int64                `json:"conversation_id,omitempty"`
	UserID         int64                `json:"user_id,omitempty"`
	Message        *SupportMessage      `json:"message,omitempty"`
	Conversation   *SupportConversation `json:"conversation,omitempty"`
	UnreadCount    int                  `json:"unread_count,omitempty"`
	Timestamp      time.Time            `json:"timestamp"`
}

type SupportSubscriber struct {
	UserID int64
	Admin  bool

	send chan []byte
}

func NewSupportSubscriber(userID int64, admin bool) *SupportSubscriber {
	return &SupportSubscriber{
		UserID: userID,
		Admin:  admin,
		send:   make(chan []byte, 32),
	}
}

func (s *SupportSubscriber) Send() <-chan []byte {
	if s == nil {
		return nil
	}
	return s.send
}

type SupportHub struct {
	mu     sync.RWMutex
	admins map[*SupportSubscriber]struct{}
	byUser map[int64]map[*SupportSubscriber]struct{}
}

func NewSupportHub() *SupportHub {
	return &SupportHub{
		admins: make(map[*SupportSubscriber]struct{}),
		byUser: make(map[int64]map[*SupportSubscriber]struct{}),
	}
}

func (h *SupportHub) Register(s *SupportSubscriber) {
	if h == nil || s == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if s.Admin {
		h.admins[s] = struct{}{}
		return
	}
	if _, ok := h.byUser[s.UserID]; !ok {
		h.byUser[s.UserID] = make(map[*SupportSubscriber]struct{})
	}
	h.byUser[s.UserID][s] = struct{}{}
}

func (h *SupportHub) Unregister(s *SupportSubscriber) {
	if h == nil || s == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if s.Admin {
		if _, ok := h.admins[s]; ok {
			delete(h.admins, s)
			close(s.send)
		}
		return
	}

	subs := h.byUser[s.UserID]
	if _, ok := subs[s]; ok {
		delete(subs, s)
		if len(subs) == 0 {
			delete(h.byUser, s.UserID)
		}
		close(s.send)
	}
}

func (h *SupportHub) SendToSubscriber(s *SupportSubscriber, event SupportRealtimeEvent) {
	if h == nil || s == nil {
		return
	}
	payload, err := marshalSupportEvent(event)
	if err != nil {
		logger.LegacyPrintf("service.support", "[SupportHub] marshal event failed: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	h.enqueueLocked(s, payload)
}

func (h *SupportHub) BroadcastToUser(userID int64, event SupportRealtimeEvent) {
	if h == nil || userID <= 0 {
		return
	}
	payload, err := marshalSupportEvent(event)
	if err != nil {
		logger.LegacyPrintf("service.support", "[SupportHub] marshal user event failed: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.byUser[userID] {
		h.enqueueLocked(sub, payload)
	}
}

func (h *SupportHub) BroadcastAdmins(event SupportRealtimeEvent) {
	if h == nil {
		return
	}
	payload, err := marshalSupportEvent(event)
	if err != nil {
		logger.LegacyPrintf("service.support", "[SupportHub] marshal admin event failed: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.admins {
		h.enqueueLocked(sub, payload)
	}
}

func (h *SupportHub) enqueueLocked(sub *SupportSubscriber, payload []byte) {
	if sub == nil || payload == nil {
		return
	}
	select {
	case sub.send <- payload:
	default:
		// Keep the connection alive but drop stale realtime payloads when a browser
		// stops reading. HTTP APIs still carry the durable history.
	}
}

func marshalSupportEvent(event SupportRealtimeEvent) ([]byte, error) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	return json.Marshal(event)
}

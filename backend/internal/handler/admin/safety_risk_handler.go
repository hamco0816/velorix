package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type safetyRiskReviewRequest struct {
	Status     string `json:"status"`
	ReviewNote string `json:"review_note"`
}

type safetyRiskClearUserRequest struct {
	UserID     int64  `json:"user_id"`
	ReviewNote string `json:"review_note"`
}

// ListSafetyRiskEvents lists gateway safety risk logs.
// GET /api/v1/admin/ops/safety-risk-events
func (h *OpsHandler) ListSafetyRiskEvents(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}

	page, pageSize := response.ParsePagination(c)
	if pageSize > 500 {
		pageSize = 500
	}
	startTime, endTime, err := parseOpsTimeRange(c, "24h")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filter := &service.SafetyRiskEventFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
		Action:   strings.TrimSpace(c.Query("action")),
		Severity: strings.TrimSpace(c.Query("severity")),
		Source:   strings.TrimSpace(c.Query("source")),
		Query:    strings.TrimSpace(c.Query("q")),
	}
	if !startTime.IsZero() {
		filter.StartTime = &startTime
	}
	if !endTime.IsZero() {
		filter.EndTime = &endTime
	}
	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if v := strings.TrimSpace(c.Query("api_key_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid api_key_id")
			return
		}
		filter.APIKeyID = &id
	}
	if v := strings.TrimSpace(c.Query("group_id")); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group_id")
			return
		}
		filter.GroupID = &id
	}

	result, err := h.opsService.ListSafetyRiskEvents(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Events, int64(result.Total), result.Page, result.PageSize)
}

// ReviewSafetyRiskEvent marks a safety risk event as reviewed/pending/cleared.
// PUT /api/v1/admin/ops/safety-risk-events/:id/review
func (h *OpsHandler) ReviewSafetyRiskEvent(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid event id")
		return
	}

	var req safetyRiskReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	input := &service.SafetyRiskReviewInput{
		Status:           req.Status,
		ReviewNote:       req.ReviewNote,
		ReviewedByUserID: subject.UserID,
	}
	if err := h.opsService.ReviewSafetyRiskEvent(c.Request.Context(), id, input); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// ClearSafetyRiskEventsForUser clears active warning records for one user.
// POST /api/v1/admin/ops/safety-risk-events/clear-user
func (h *OpsHandler) ClearSafetyRiskEventsForUser(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req safetyRiskClearUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.opsService.ClearSafetyRiskEventsForUser(c.Request.Context(), &service.SafetyRiskClearUserInput{
		UserID:           req.UserID,
		ReviewNote:       req.ReviewNote,
		ReviewedByUserID: subject.UserID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

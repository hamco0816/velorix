package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// 把 plan 的 (validity_days, validity_unit) 归一化为天数。
// 与 service.psComputeValidityDays 保持一致：week=×7、month=×30、其他原值。
// 独立实现一份，避免跨包暴露 unexported 函数。
func adminNormalizeValidityDays(days int, unit string) int {
	switch unit {
	case "week", "weeks":
		return days * 7
	case "month", "months":
		return days * 30
	default:
		return days
	}
}

// SeatHandler 管理员视角的独享池/独享名额操作。
type SeatHandler struct {
	paymentService *service.PaymentService
	configService  *service.PaymentConfigService
}

// NewSeatHandler 创建管理员 SeatHandler。
func NewSeatHandler(paymentService *service.PaymentService, configService *service.PaymentConfigService) *SeatHandler {
	return &SeatHandler{paymentService: paymentService, configService: configService}
}

// AdminSeatView 管理员看到的 seat 详情（含完整 account_id）。
type AdminSeatView struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	GroupID       int64   `json:"group_id"`
	PlanID        int64   `json:"plan_id"`
	AccountID     int64   `json:"account_id"`
	Status        string  `json:"status"`
	StartsAt      string  `json:"starts_at"`
	ExpiresAt     string  `json:"expires_at"`
	AssignedAt    string  `json:"assigned_at"`
	LastRenewalAt string  `json:"last_renewal_at,omitempty"`
	UsageUsd      float64 `json:"usage_usd"`
	AssignedBy    int64   `json:"assigned_by,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

func toAdminSeatView(s *dbent.ExclusiveSubscription) AdminSeatView {
	v := AdminSeatView{
		ID: s.ID, UserID: s.UserID, GroupID: s.GroupID,
		PlanID: s.PlanID, AccountID: s.AccountID, Status: s.Status,
		StartsAt:   s.StartsAt.Format(time.RFC3339),
		ExpiresAt:  s.ExpiresAt.Format(time.RFC3339),
		AssignedAt: s.AssignedAt.Format(time.RFC3339),
		UsageUsd:   s.UsageUsd,
	}
	if s.LastRenewalAt != nil {
		v.LastRenewalAt = s.LastRenewalAt.Format(time.RFC3339)
	}
	if s.AssignedBy != nil {
		v.AssignedBy = *s.AssignedBy
	}
	if s.Notes != nil {
		v.Notes = *s.Notes
	}
	return v
}

// ListSeats GET /api/v1/admin/seats?user_id=&group_id=&status=&limit=&offset=
//
// 按 user_id / group_id / status 过滤；不传则返回最新 100 条。
func (h *SeatHandler) ListSeats(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	groupID, _ := strconv.ParseInt(c.Query("group_id"), 10, 64)
	status := strings.TrimSpace(c.Query("status"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	seatSvc := h.paymentService.ExclusiveSeatService()
	seats, err := seatSvc.AdminListSeats(c.Request.Context(), service.AdminListSeatsFilter{
		UserID: userID, GroupID: groupID, Status: status,
		Limit: limit, Offset: offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	views := make([]AdminSeatView, 0, len(seats))
	for _, s := range seats {
		views = append(views, toAdminSeatView(s))
	}
	response.Success(c, gin.H{"items": views})
}

// GetGroupInventory GET /api/v1/admin/exclusive-pools/:groupId/inventory
//
// 返回独享池库存统计：总数 / 空闲 / 在用 / 即将到期。
func (h *SeatHandler) GetGroupInventory(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("groupId"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group ID")
		return
	}
	inv, err := h.paymentService.ExclusiveSeatService().GetGroupInventory(c.Request.Context(), groupID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"group_id":      groupID,
		"total":         inv.Total,
		"free":          inv.Free,
		"used":          inv.Used,
		"schedulable":   inv.Schedulable, // 当下可立即分配（剔除限流/过载/临时不可用）
		"expiring_in_7": inv.ExpiringIn7,
	})
}

// GrantSeat POST /api/v1/admin/seats/grant
//
// 请求体：{"user_id": 1, "plan_id": 2, "validity_days": 30, "notes": "..."}
// 由管理员从独享池里手动分一份名额给用户（不走支付）。
func (h *SeatHandler) GrantSeat(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var req struct {
		UserID       int64  `json:"user_id"`
		PlanID       int64  `json:"plan_id"`
		// GroupID 可选：前端"赠送座席"页选中的独享池 ID。后端用它和 plan.GroupID 做一致性校验，
		// 防止"在 A 池页面手填 B 池套餐"导致管理员看到 A 池库存却实际从 B 池发放（GPT round 19 #3）。
		GroupID      int64  `json:"group_id"`
		ValidityDays int    `json:"validity_days"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if req.UserID <= 0 || req.PlanID <= 0 {
		response.BadRequest(c, "user_id / plan_id required")
		return
	}
	plan, err := h.configService.GetPlan(c.Request.Context(), req.PlanID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	// 强校验 plan.kind=exclusive：admin 赠送独享名额时若误选了共享套餐，会创建出语义不一致的 seat
	if plan.Kind != domain.PlanKindExclusive {
		response.ErrorFrom(c, infraerrors.BadRequest("PLAN_KIND_INVALID", "selected plan is not an exclusive plan; choose a plan with kind='exclusive'"))
		return
	}
	// 一致性校验：前端选了独享池 A，提交的 plan 必须属于 A 池
	if req.GroupID > 0 && plan.GroupID != req.GroupID {
		response.ErrorFrom(c, infraerrors.BadRequest("PLAN_GROUP_MISMATCH",
			"plan does not belong to the selected exclusive pool; pick a plan from this pool"))
		return
	}
	// 有效期归一化：plan.ValidityDays 单位由 plan.ValidityUnit 决定（day/week/month）
	// 直接传 ValidityDays 不做单位换算会让 month/week 套餐被截到几天
	days := req.ValidityDays
	if days <= 0 {
		days = adminNormalizeValidityDays(plan.ValidityDays, plan.ValidityUnit)
	}
	seat, err := h.paymentService.ExclusiveSeatService().AssignSeat(c.Request.Context(), service.AssignSeatInput{
		UserID:       req.UserID,
		GroupID:      plan.GroupID,
		PlanID:       plan.ID,
		ValidityDays: days,
		AssignedBy:   subject.UserID,
		Notes:        req.Notes,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, toAdminSeatView(seat))
}

// ReleaseSeat POST /api/v1/admin/seats/:id/release
//
// 请求体：{"reason": "..."}
// 强制把 seat 状态改为 cancelled，账号回池。
func (h *SeatHandler) ReleaseSeat(c *gin.Context) {
	seatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid seat ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.paymentService.ExclusiveSeatService().ReleaseSeat(c.Request.Context(), seatID, "cancelled", req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": seatID, "status": "cancelled"})
}

// SwapSeatAccount POST /api/v1/admin/seats/:id/swap
//
// 强制把 seat 绑到另一个空闲账号（管理员救场）。
func (h *SeatHandler) SwapSeatAccount(c *gin.Context) {
	seatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid seat ID")
		return
	}
	updated, err := h.paymentService.ExclusiveSeatService().SwapSeatAccount(c.Request.Context(), seatID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, toAdminSeatView(updated))
}

// ExtendSeat POST /api/v1/admin/seats/:id/extend
//
// 请求体：{"days": 7}  正数为延期、负数为提前结束（拒绝结果在过去）。
func (h *SeatHandler) ExtendSeat(c *gin.Context) {
	seatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid seat ID")
		return
	}
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	updated, err := h.paymentService.ExclusiveSeatService().ExtendSeat(c.Request.Context(), seatID, req.Days)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, toAdminSeatView(updated))
}

// ListSeatReleaseFailures GET /api/v1/admin/seats/release-failures?since_hours=24&limit=200
//
// 列出"退款已成功但 seat 释放失败、且尚未补救成功"的订单。
// 后台 SeatReleaseRetryService 会每 5 分钟自动重试，但管理员可以通过此接口排查长期未闭环的项。
func (h *SeatHandler) ListSeatReleaseFailures(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	sinceHours, _ := strconv.Atoi(c.DefaultQuery("since_hours", "24"))
	if sinceHours <= 0 {
		sinceHours = 24
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	if limit <= 0 {
		limit = 200
	}
	pending, err := h.paymentService.ListPendingSeatReleaseFailures(c.Request.Context(),
		time.Duration(sinceHours)*time.Hour, limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": pending, "total": len(pending)})
}

// RetrySeatReleaseFailure POST /api/v1/admin/seats/release-failures/:orderID/retry
//
// 管理员手动触发某条退款订单的 seat 释放重试。订单必须处于"完全退款"状态（OrderStatusRefunded）。
// 部分退款（PartiallyRefunded）会被服务层 RetrySeatReleaseForOrder 显式拒绝并返回
// ORDER_NOT_FULLY_REFUNDED——保留独享号权益的语义在 GPT round 23 #2 / round 24 #2 已统一。
func (h *SeatHandler) RetrySeatReleaseFailure(c *gin.Context) {
	if _, ok := requireAuth(c); !ok {
		return
	}
	orderID, err := strconv.ParseInt(c.Param("orderID"), 10, 64)
	if err != nil || orderID <= 0 {
		response.BadRequest(c, "invalid orderID")
		return
	}
	if err := h.paymentService.RetrySeatReleaseForOrder(c.Request.Context(), orderID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"order_id": orderID, "retried": true})
}

// requireAuth 从 admin 鉴权中间件获取调用者（管理员）的 user ID。
type authSubject struct {
	UserID int64
}

func requireAuth(c *gin.Context) (*authSubject, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.ErrorFrom(c, infraerrors.Unauthorized("UNAUTHORIZED", "missing user context"))
		return nil, false
	}
	return &authSubject{UserID: subject.UserID}, true
}

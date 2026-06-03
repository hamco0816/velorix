package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// SeatHandler 用户视角的独享名额（seat）操作。
type SeatHandler struct {
	paymentService *service.PaymentService
	configService  *service.PaymentConfigService
}

// NewSeatHandler 创建 SeatHandler。
func NewSeatHandler(paymentService *service.PaymentService, configService *service.PaymentConfigService) *SeatHandler {
	return &SeatHandler{paymentService: paymentService, configService: configService}
}

// SeatView 是 seat 暴露给"用户端"的脱敏视图：只保留 account_label，原始 account_id
// 不下发（避免用户拼接调度/账号管理类内部 API）。管理员路径有独立的 AdminSeatView。
type SeatView struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	GroupID       int64   `json:"group_id"`
	GroupName     string  `json:"group_name,omitempty"`
	GroupPlatform string  `json:"group_platform,omitempty"`
	PlanID        int64   `json:"plan_id"`
	PlanName      string  `json:"plan_name,omitempty"`
	AccountLabel  string  `json:"account_label"` // 脱敏标识：池名 + accountID 后 4 位 hash
	Status        string  `json:"status"`
	StartsAt      string  `json:"starts_at"`
	ExpiresAt     string  `json:"expires_at"`
	AssignedAt    string  `json:"assigned_at"`
	LastRenewalAt string  `json:"last_renewal_at,omitempty"`
	UsageUsd      float64 `json:"usage_usd"`
	Notes         string  `json:"notes,omitempty"`

	// 日/周/月窗口用量与上限，用于用户端展示进度条（nil 表示该窗口无上限，
	// 前端不绘制对应进度条；用量字段始终下发，便于用户看到累计消耗）
	DailyUsageUsd   float64  `json:"daily_usage_usd"`
	WeeklyUsageUsd  float64  `json:"weekly_usage_usd"`
	MonthlyUsageUsd float64  `json:"monthly_usage_usd"`
	DailyLimitUsd   *float64 `json:"daily_limit_usd,omitempty"`
	WeeklyLimitUsd  *float64 `json:"weekly_limit_usd,omitempty"`
	MonthlyLimitUsd *float64 `json:"monthly_limit_usd,omitempty"`
}

// MyExclusiveSeats GET /api/v1/seats —— 当前用户的所有独享名额（含历史）。
func (h *SeatHandler) MyExclusiveSeats(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	seatSvc := h.paymentService.ExclusiveSeatService()
	seats, err := seatSvc.ListAllByUser(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	views := h.toSeatViews(c, seats)
	response.Success(c, gin.H{"items": views, "total": len(views)})
}

// PreviewRenewal GET /api/v1/seats/:id/renewal-preview ——
// 返回续费一份独享名额需要支付的金额、有效期天数、关联的 plan 信息。
// 前端拿到后调 CreateOrder 创建续费订单走支付流程。
//
// 设计说明：v1 曾提供直接 RenewSeat 的接口，会被恶意用户无限免费续费。
// v2 强制续费走 CreateOrder（带 renewal_seat_id）+ 支付 + fulfillment.RenewSeat。
func (h *SeatHandler) PreviewRenewal(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	seatID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid seat ID")
		return
	}
	seatSvc := h.paymentService.ExclusiveSeatService()
	seat, err := seatSvc.GetSeat(c.Request.Context(), seatID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if seat.UserID != subject.UserID {
		response.ErrorFrom(c, infraerrors.Forbidden("SEAT_NOT_OWNED", "seat does not belong to current user"))
		return
	}
	// 前置校验：seat 当下是否真的能续费（状态 + 宽限期 + 原账号未被别人占）
	// 不能续费时直接报错，前端不应展示"立即续费"按钮，避免用户进入支付后才发现退款
	if err := seatSvc.CheckSeatRenewable(c.Request.Context(), seat, 7*24*time.Hour); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	plan, err := h.configService.GetPlan(c.Request.Context(), seat.PlanID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	// 上次实付价（用于「上次 ¥X → 本次 ¥Y」对比），失败仅记 0，不阻塞主流程
	lastPaidPrice, _ := seatSvc.LastPaidPriceForSeat(c.Request.Context(), seat)
	// 续费允许已下架（for_sale=false）的 plan，但 GetCheckoutInfo 只返回在售套餐 → 前端
	// `checkout.plans.find(p => p.id === planId)` 会找不到下架套餐，无法进入续费支付页（GPT round 28 #1）。
	// 解法：本接口同时返回完整 plan DTO，前端续费分支优先使用此 plan 而不是从 checkout.plans 查找。
	//
	// 注意（GPT round 29 #3 + round 30 #3）：不能直接返回 ent.SubscriptionPlan：
	//   - ent.Features 是 \n 分隔字符串，前端 v-for 会按字符迭代
	//   - 需要补 group_platform / group_name / supported_model_scopes，并在 plan 限额覆盖为 nil 时
	//     回落到 group 默认值，否则确认页可能错误显示"通用平台 / 不限额"
	groupInfo := h.configService.GetGroupInfoMap(c.Request.Context(), []*dbent.SubscriptionPlan{plan})
	gi := groupInfo[plan.GroupID]
	response.Success(c, gin.H{
		"seat_id":            seat.ID,
		"plan_id":            plan.ID,
		"plan_name":          plan.Name,
		"price":              plan.Price,
		"validity_days":      plan.ValidityDays,
		"validity_unit":      plan.ValidityUnit,
		"current_expires_at": seat.ExpiresAt,
		"last_paid_price":    lastPaidPrice,
		"plan":               renderRenewalPlanDTO(plan, gi),
	})
	_ = time.Now() // keep import; Time 在结构体里被序列化
}

// renderRenewalPlanDTO 把 ent.SubscriptionPlan 转成与 /payment/checkout-info 同款的 plan DTO。
// 复用 parseFeatures（payment_handler.go 同包 helper）让 \n 分隔的 features 字符串变成 []string，
// 并按 group 信息补齐 group_platform / group_name / supported_model_scopes，
// 在 plan 限额/倍率为 nil 时回落 group 默认值（与正常 checkout 行为一致）。
func renderRenewalPlanDTO(plan *dbent.SubscriptionPlan, gi service.PlanGroupInfo) gin.H {
	if plan == nil {
		return nil
	}
	// 限额 / 倍率：plan 覆盖优先，nil 回落 group 默认（与 payment_handler.GetCheckoutInfo 同款）
	rateMultiplier := gi.RateMultiplier
	if plan.RateMultiplier != nil && *plan.RateMultiplier > 0 {
		rateMultiplier = *plan.RateMultiplier
	}
	dailyLimit := gi.DailyLimitUSD
	if plan.DailyLimitUsd != nil && *plan.DailyLimitUsd > 0 {
		v := *plan.DailyLimitUsd
		dailyLimit = &v
	}
	weeklyLimit := gi.WeeklyLimitUSD
	if plan.WeeklyLimitUsd != nil && *plan.WeeklyLimitUsd > 0 {
		v := *plan.WeeklyLimitUsd
		weeklyLimit = &v
	}
	monthlyLimit := gi.MonthlyLimitUSD
	if plan.MonthlyLimitUsd != nil && *plan.MonthlyLimitUsd > 0 {
		v := *plan.MonthlyLimitUsd
		monthlyLimit = &v
	}
	dto := gin.H{
		"id":                      int64(plan.ID),
		"group_id":                plan.GroupID,
		"group_platform":          gi.Platform,
		"group_name":              gi.Name,
		"name":                    plan.Name,
		"description":             plan.Description,
		"price":                   plan.Price,
		"validity_days":           plan.ValidityDays,
		"validity_unit":           plan.ValidityUnit,
		"features":                parseFeatures(plan.Features),
		"product_name":            plan.ProductName,
		"for_sale":                plan.ForSale,
		"sort_order":              plan.SortOrder,
		"is_popular":              plan.IsPopular,
		"badge_text":              plan.BadgeText,
		"badge_color":             plan.BadgeColor,
		"kind":                    plan.Kind,
		"rate_multiplier":         rateMultiplier,
		"daily_limit_usd":         dailyLimit,
		"weekly_limit_usd":        weeklyLimit,
		"monthly_limit_usd":       monthlyLimit,
		"has_plan_limit_override": plan.DailyLimitUsd != nil || plan.WeeklyLimitUsd != nil || plan.MonthlyLimitUsd != nil || plan.RateMultiplier != nil,
		"supported_model_scopes":  gi.ModelScopes,
	}
	if plan.OriginalPrice != nil && *plan.OriginalPrice > 0 {
		dto["original_price"] = *plan.OriginalPrice
	}
	return dto
}

// toSeatViews 把 ent 实体批量转换成对外视图，并 join group/plan 简单信息。
func (h *SeatHandler) toSeatViews(c *gin.Context, seats []*dbent.ExclusiveSubscription) []SeatView {
	out := make([]SeatView, 0, len(seats))
	for _, s := range seats {
		out = append(out, h.toSeatView(c, s))
	}
	return out
}

// toSeatView 单个 seat 视图：脱敏 account_label = "<group_name>-#<id_suffix>"。
func (h *SeatHandler) toSeatView(c *gin.Context, s *dbent.ExclusiveSubscription) SeatView {
	if s == nil {
		return SeatView{}
	}
	v := SeatView{
		ID:              s.ID,
		UserID:          s.UserID,
		GroupID:         s.GroupID,
		PlanID:          s.PlanID,
		AccountLabel:    maskAccountLabel(s.AccountID),
		Status:          s.Status,
		StartsAt:        s.StartsAt.Format(time.RFC3339),
		ExpiresAt:       s.ExpiresAt.Format(time.RFC3339),
		AssignedAt:      s.AssignedAt.Format(time.RFC3339),
		UsageUsd:        s.UsageUsd,
		DailyUsageUsd:   s.DailyUsageUsd,
		WeeklyUsageUsd:  s.WeeklyUsageUsd,
		MonthlyUsageUsd: s.MonthlyUsageUsd,
		DailyLimitUsd:   s.DailyLimitUsd,
		WeeklyLimitUsd:  s.WeeklyLimitUsd,
		MonthlyLimitUsd: s.MonthlyLimitUsd,
	}
	if s.LastRenewalAt != nil {
		v.LastRenewalAt = s.LastRenewalAt.Format(time.RFC3339)
	}
	if s.Notes != nil {
		v.Notes = *s.Notes
	}
	// 加载 group / plan 简易元信息（best-effort，失败仅留空）
	if h.configService != nil {
		if plan, err := h.configService.GetPlan(c.Request.Context(), s.PlanID); err == nil {
			v.PlanName = plan.Name
			v.AccountLabel = maskAccountLabelWithGroup(plan.Name, s.AccountID)
		}
	}
	return v
}

// maskAccountLabel 默认形式：#A1B2（account_id 的 16 进制后 4 位）。
func maskAccountLabel(accountID int64) string {
	const hex = "0123456789ABCDEF"
	v := uint16(accountID & 0xFFFF)
	return "#" + string([]byte{
		hex[(v>>12)&0xF], hex[(v>>8)&0xF], hex[(v>>4)&0xF], hex[v&0xF],
	})
}

// maskAccountLabelWithGroup 形式：<group_or_plan>-#A1B2 ——更友好的客服沟通标识。
func maskAccountLabelWithGroup(groupOrPlan string, accountID int64) string {
	if groupOrPlan == "" {
		return maskAccountLabel(accountID)
	}
	return groupOrPlan + "-" + maskAccountLabel(accountID)
}

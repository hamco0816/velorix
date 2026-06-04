package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/exclusivesubscription"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// 续费宽限期：与 PaymentService.precheckPlanLockWithRenewal 中允许 expired 续费的窗口保持一致。
// 修改时请同步更新两处。
const seatRenewalGraceWindow = 7 * 24 * time.Hour

const (
	maxPlanBadgeTextRunes = 12
	maxPlanLabelRunes     = 24
	defaultPlanBadgeColor = "gold"
	defaultPlanTierStyle  = "basic"
)

// 档位样式升级阶梯 key（与前端 utils/tierStyle.ts 保持一致）。
var allowedPlanTierStyles = map[string]struct{}{
	"basic":    {}, // 简约（Lite）
	"standard": {}, // 标准（Plus）
	"advanced": {}, // 进阶（Pro）
	"flagship": {}, // 旗舰（Ultra）
	"luxury":   {}, // 豪华（Max）
	"supreme":  {}, // 至尊（更高档）
}

// 套餐角标预设色板 key（尊贵色调）；admin 只能从这组里选，避免随意配色。
// 与前端 utils/badgeTone.ts 的 key 保持一致。
var allowedPlanBadgeColors = map[string]struct{}{
	"gold":     {}, // 鎏金（默认）
	"obsidian": {}, // 黑金至尊
	"purple":   {}, // 帝王紫
	"emerald":  {}, // 翡翠
	"sapphire": {}, // 宝石蓝
	"rose":     {}, // 玫瑰金
}

// countProtectedSeatsByPlan 统计该 plan 当前被"用户业务还在持有"的 seat 数：
//   - active：明确还在用
//   - 已 expired 但仍在续费宽限期（≤7 天）：用户随时会发起续费，删除/换 group/换 kind 会让续费失败
//
// UpdatePlan / DeletePlan 共用此口径，避免与续费窗口口径漂移。
func (s *PaymentConfigService) countProtectedSeatsByPlan(ctx context.Context, planID int64) (int, error) {
	graceCutoff := time.Now().Add(-seatRenewalGraceWindow)
	withinGrace := exclusivesubscription.And(
		exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusExpired),
		exclusivesubscription.ExpiresAtGTE(graceCutoff),
	)
	stillProtected := exclusivesubscription.Or(
		exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
		withinGrace,
	)
	return s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.PlanIDEQ(planID),
			exclusivesubscription.DeletedAtIsNil(),
			stillProtected,
		).Count(ctx)
}

// 校验套餐 kind：必须是 shared 或 exclusive，空字符串视为 shared（向后兼容）
func normalizePlanKind(kind string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(kind))
	if v == "" {
		return domain.PlanKindShared, nil
	}
	if v != domain.PlanKindShared && v != domain.PlanKindExclusive {
		return "", infraerrors.BadRequest("PLAN_KIND_INVALID", "kind must be 'shared' or 'exclusive'")
	}
	return v, nil
}

// validatePlanRequired checks that all required fields for a plan are provided.
func validatePlanRequired(name string, groupID int64, price float64, validityDays int, validityUnit string, originalPrice *float64) error {
	if strings.TrimSpace(name) == "" {
		return infraerrors.BadRequest("PLAN_NAME_REQUIRED", "plan name is required")
	}
	if groupID <= 0 {
		return infraerrors.BadRequest("PLAN_GROUP_REQUIRED", "group is required")
	}
	if price <= 0 {
		return infraerrors.BadRequest("PLAN_PRICE_INVALID", "price must be > 0")
	}
	if validityDays <= 0 {
		return infraerrors.BadRequest("PLAN_VALIDITY_REQUIRED", "validity days must be > 0")
	}
	if strings.TrimSpace(validityUnit) == "" {
		return infraerrors.BadRequest("PLAN_VALIDITY_UNIT_REQUIRED", "validity unit is required")
	}
	if originalPrice != nil && *originalPrice < 0 {
		return infraerrors.BadRequest("PLAN_ORIGINAL_PRICE_INVALID", "original price must be >= 0")
	}
	return nil
}

// validatePlanPatch validates only the non-nil fields in a patch update.
func validatePlanPatch(req UpdatePlanRequest) error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return infraerrors.BadRequest("PLAN_NAME_REQUIRED", "plan name is required")
	}
	if req.GroupID != nil && *req.GroupID <= 0 {
		return infraerrors.BadRequest("PLAN_GROUP_REQUIRED", "group is required")
	}
	if req.Price != nil && *req.Price <= 0 {
		return infraerrors.BadRequest("PLAN_PRICE_INVALID", "price must be > 0")
	}
	if req.ValidityDays != nil && *req.ValidityDays <= 0 {
		return infraerrors.BadRequest("PLAN_VALIDITY_REQUIRED", "validity days must be > 0")
	}
	if req.ValidityUnit != nil && strings.TrimSpace(*req.ValidityUnit) == "" {
		return infraerrors.BadRequest("PLAN_VALIDITY_UNIT_REQUIRED", "validity unit is required")
	}
	if req.OriginalPrice != nil && *req.OriginalPrice < 0 {
		return infraerrors.BadRequest("PLAN_ORIGINAL_PRICE_INVALID", "original price must be >= 0")
	}
	if req.BadgeText != nil {
		if _, err := normalizePlanBadgeText(*req.BadgeText); err != nil {
			return err
		}
	}
	if req.BadgeColor != nil {
		if _, err := normalizePlanBadgeColor(*req.BadgeColor); err != nil {
			return err
		}
	}
	if req.PlanLabel != nil {
		if _, err := normalizePlanLabel(*req.PlanLabel); err != nil {
			return err
		}
	}
	if req.TierStyle != nil {
		if _, err := normalizePlanTierStyle(*req.TierStyle); err != nil {
			return err
		}
	}
	return nil
}

func normalizePlanBadgeText(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	if utf8.RuneCountInString(v) > maxPlanBadgeTextRunes {
		return "", infraerrors.BadRequest("PLAN_BADGE_TEXT_TOO_LONG", fmt.Sprintf("badge_text can contain at most %d characters", maxPlanBadgeTextRunes))
	}
	return v, nil
}

func planBadgeTextForCreate(req CreatePlanRequest) (string, error) {
	badgeText, err := normalizePlanBadgeText(req.BadgeText)
	if err != nil {
		return "", err
	}
	return badgeText, nil
}

// 校验档位名：去空白；超长报错。空字符串合法（前端自动推导）。
func normalizePlanLabel(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if utf8.RuneCountInString(v) > maxPlanLabelRunes {
		return "", infraerrors.BadRequest("PLAN_LABEL_TOO_LONG", fmt.Sprintf("plan_label can contain at most %d characters", maxPlanLabelRunes))
	}
	return v, nil
}

// 校验档位样式：空字符串回落到默认 basic；非法 key 报错。
func normalizePlanTierStyle(raw string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return defaultPlanTierStyle, nil
	}
	if _, ok := allowedPlanTierStyles[v]; !ok {
		return "", infraerrors.BadRequest("PLAN_TIER_STYLE_INVALID", fmt.Sprintf("tier_style %q is not a supported preset", v))
	}
	return v, nil
}

// 校验角标配色：空字符串回落到默认 gold；非法 key 报错。
func normalizePlanBadgeColor(raw string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	if v == "" {
		return defaultPlanBadgeColor, nil
	}
	if _, ok := allowedPlanBadgeColors[v]; !ok {
		return "", infraerrors.BadRequest("PLAN_BADGE_COLOR_INVALID", fmt.Sprintf("badge_color %q is not a supported preset", v))
	}
	return v, nil
}

// --- Plan CRUD ---

// PlanGroupInfo holds the group details needed for subscription plan display.
type PlanGroupInfo struct {
	Platform        string   `json:"platform"`
	Name            string   `json:"name"`
	RateMultiplier  float64  `json:"rate_multiplier"`
	DailyLimitUSD   *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD  *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD *float64 `json:"monthly_limit_usd"`
	ModelScopes     []string `json:"supported_model_scopes"`
}

// GetGroupPlatformMap returns a map of group_id → platform for the given plans.
func (s *PaymentConfigService) GetGroupPlatformMap(ctx context.Context, plans []*dbent.SubscriptionPlan) map[int64]string {
	info := s.GetGroupInfoMap(ctx, plans)
	m := make(map[int64]string, len(info))
	for id, gi := range info {
		m[id] = gi.Platform
	}
	return m
}

// GetGroupInfoMap returns a map of group_id → PlanGroupInfo for the given plans.
func (s *PaymentConfigService) GetGroupInfoMap(ctx context.Context, plans []*dbent.SubscriptionPlan) map[int64]PlanGroupInfo {
	ids := make([]int64, 0, len(plans))
	seen := make(map[int64]bool)
	for _, p := range plans {
		if !seen[p.GroupID] {
			seen[p.GroupID] = true
			ids = append(ids, p.GroupID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	groups, err := s.entClient.Group.Query().Where(group.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil
	}
	m := make(map[int64]PlanGroupInfo, len(groups))
	for _, g := range groups {
		m[int64(g.ID)] = PlanGroupInfo{
			Platform:        g.Platform,
			Name:            g.Name,
			RateMultiplier:  g.RateMultiplier,
			DailyLimitUSD:   g.DailyLimitUsd,
			WeeklyLimitUSD:  g.WeeklyLimitUsd,
			MonthlyLimitUSD: g.MonthlyLimitUsd,
			ModelScopes:     g.SupportedModelScopes,
		}
	}
	return m
}

func (s *PaymentConfigService) ListPlans(ctx context.Context) ([]*dbent.SubscriptionPlan, error) {
	return s.entClient.SubscriptionPlan.Query().Order(subscriptionplan.BySortOrder()).All(ctx)
}

func (s *PaymentConfigService) ListPlansForSale(ctx context.Context) ([]*dbent.SubscriptionPlan, error) {
	return s.entClient.SubscriptionPlan.Query().Where(subscriptionplan.ForSaleEQ(true)).Order(subscriptionplan.BySortOrder()).All(ctx)
}

func (s *PaymentConfigService) CreatePlan(ctx context.Context, req CreatePlanRequest) (*dbent.SubscriptionPlan, error) {
	if err := validatePlanRequired(req.Name, req.GroupID, req.Price, req.ValidityDays, req.ValidityUnit, req.OriginalPrice); err != nil {
		return nil, err
	}
	kind, err := normalizePlanKind(req.Kind)
	if err != nil {
		return nil, err
	}
	badgeText, err := planBadgeTextForCreate(req)
	if err != nil {
		return nil, err
	}
	badgeColor, err := normalizePlanBadgeColor(req.BadgeColor)
	if err != nil {
		return nil, err
	}
	planLabel, err := normalizePlanLabel(req.PlanLabel)
	if err != nil {
		return nil, err
	}
	tierStyle, err := normalizePlanTierStyle(req.TierStyle)
	if err != nil {
		return nil, err
	}
	// is_popular（推荐档，整列高亮）与 badge_text（角标文字）解耦，各自独立控制
	b := s.entClient.SubscriptionPlan.Create().
		SetGroupID(req.GroupID).SetName(req.Name).SetDescription(req.Description).
		SetPrice(req.Price).SetValidityDays(req.ValidityDays).SetValidityUnit(req.ValidityUnit).
		SetFeatures(req.Features).SetProductName(req.ProductName).
		SetForSale(req.ForSale).SetSortOrder(req.SortOrder).SetIsPopular(req.IsPopular).SetBadgeText(badgeText).SetBadgeColor(badgeColor).SetPlanLabel(planLabel).SetTierStyle(tierStyle).SetKind(kind)
	if req.OriginalPrice != nil {
		b.SetOriginalPrice(*req.OriginalPrice)
	}
	// 套餐级限额/倍率覆盖：传 0 / nil 都视为"不覆盖"，保持 NULL 让调度回落到 group
	if req.DailyLimitUSD != nil && *req.DailyLimitUSD > 0 {
		b.SetDailyLimitUsd(*req.DailyLimitUSD)
	}
	if req.WeeklyLimitUSD != nil && *req.WeeklyLimitUSD > 0 {
		b.SetWeeklyLimitUsd(*req.WeeklyLimitUSD)
	}
	if req.MonthlyLimitUSD != nil && *req.MonthlyLimitUSD > 0 {
		b.SetMonthlyLimitUsd(*req.MonthlyLimitUSD)
	}
	if req.RateMultiplier != nil && *req.RateMultiplier > 0 {
		b.SetRateMultiplier(*req.RateMultiplier)
	}
	return b.Save(ctx)
}

// UpdatePlan updates a subscription plan by ID (patch semantics).
// NOTE: This function exceeds 30 lines due to per-field nil-check patch update boilerplate
// plus a validation guard for non-nil fields.
func (s *PaymentConfigService) UpdatePlan(ctx context.Context, id int64, req UpdatePlanRequest) (*dbent.SubscriptionPlan, error) {
	if err := validatePlanPatch(req); err != nil {
		return nil, err
	}
	// 已有受保护独享 seat 时禁止修改 kind 和 group_id：
	//   - kind 改成 shared 后，续费校验/退款释放 seat 流程都会因找不到独享语义而失败
	//   - group_id 变更后，原有 seat 仍绑在旧 group/旧账号，但用户续费走新 group，状态错乱
	// 注意：管理端编辑套餐保存时 payload 永远会带 group_id 和 kind，要先与 current 对比确认"真改了"
	// 才触发拦截，避免误伤改价/改名/下架等无关字段的保存请求。
	current, err := s.entClient.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	kindChanged := false
	if req.Kind != nil {
		normalizedKind, kErr := normalizePlanKind(*req.Kind)
		if kErr != nil {
			return nil, kErr
		}
		kindChanged = normalizedKind != current.Kind
	}
	groupChanged := req.GroupID != nil && *req.GroupID != current.GroupID
	if kindChanged || groupChanged {
		protected, err := s.countProtectedSeatsByPlan(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("check protected exclusive seats: %w", err)
		}
		if protected > 0 {
			return nil, infraerrors.Conflict("ACTIVE_SEATS_BOUND",
				fmt.Sprintf("this plan has %d active or recently-expired (within renewal grace) exclusive seats; cannot change kind or group_id while seats are bound (mark for_sale=false to take it offline first)", protected))
		}
	}
	u := s.entClient.SubscriptionPlan.UpdateOneID(id)
	if req.GroupID != nil {
		u.SetGroupID(*req.GroupID)
	}
	if req.Name != nil {
		u.SetName(*req.Name)
	}
	if req.Description != nil {
		u.SetDescription(*req.Description)
	}
	if req.Price != nil {
		u.SetPrice(*req.Price)
	}
	if req.OriginalPrice != nil {
		u.SetOriginalPrice(*req.OriginalPrice)
	}
	if req.ValidityDays != nil {
		u.SetValidityDays(*req.ValidityDays)
	}
	if req.ValidityUnit != nil {
		u.SetValidityUnit(*req.ValidityUnit)
	}
	if req.Features != nil {
		u.SetFeatures(*req.Features)
	}
	if req.ProductName != nil {
		u.SetProductName(*req.ProductName)
	}
	if req.ForSale != nil {
		u.SetForSale(*req.ForSale)
	}
	if req.SortOrder != nil {
		u.SetSortOrder(*req.SortOrder)
	}
	// is_popular（推荐档高亮）与 badge_text（角标文字）解耦，各自独立 patch
	if req.IsPopular != nil {
		u.SetIsPopular(*req.IsPopular)
	}
	if req.BadgeText != nil {
		badgeText, err := normalizePlanBadgeText(*req.BadgeText)
		if err != nil {
			return nil, err
		}
		u.SetBadgeText(badgeText)
	}
	if req.PlanLabel != nil {
		planLabel, err := normalizePlanLabel(*req.PlanLabel)
		if err != nil {
			return nil, err
		}
		u.SetPlanLabel(planLabel)
	}
	if req.TierStyle != nil {
		tierStyle, err := normalizePlanTierStyle(*req.TierStyle)
		if err != nil {
			return nil, err
		}
		u.SetTierStyle(tierStyle)
	}
	if req.BadgeColor != nil {
		badgeColor, err := normalizePlanBadgeColor(*req.BadgeColor)
		if err != nil {
			return nil, err
		}
		u.SetBadgeColor(badgeColor)
	}
	if req.Kind != nil {
		kind, err := normalizePlanKind(*req.Kind)
		if err != nil {
			return nil, err
		}
		u.SetKind(kind)
	}
	// 套餐级限额/倍率覆盖：> 0 = 设置；= 0 = 清空（回落到 group）；nil = 不修改
	if req.DailyLimitUSD != nil {
		if *req.DailyLimitUSD > 0 {
			u.SetDailyLimitUsd(*req.DailyLimitUSD)
		} else {
			u.ClearDailyLimitUsd()
		}
	}
	if req.WeeklyLimitUSD != nil {
		if *req.WeeklyLimitUSD > 0 {
			u.SetWeeklyLimitUsd(*req.WeeklyLimitUSD)
		} else {
			u.ClearWeeklyLimitUsd()
		}
	}
	if req.MonthlyLimitUSD != nil {
		if *req.MonthlyLimitUSD > 0 {
			u.SetMonthlyLimitUsd(*req.MonthlyLimitUSD)
		} else {
			u.ClearMonthlyLimitUsd()
		}
	}
	if req.RateMultiplier != nil {
		if *req.RateMultiplier > 0 {
			u.SetRateMultiplier(*req.RateMultiplier)
		} else {
			u.ClearRateMultiplier()
		}
	}
	return u.Save(ctx)
}

func (s *PaymentConfigService) DeletePlan(ctx context.Context, id int64) error {
	count, err := s.countPendingOrdersByPlan(ctx, id)
	if err != nil {
		return fmt.Errorf("check pending orders: %w", err)
	}
	if count > 0 {
		return infraerrors.Conflict("PENDING_ORDERS",
			fmt.Sprintf("this plan has %d in-progress orders and cannot be deleted — wait for orders to complete first", count))
	}
	// 活跃 + 7 天宽限期内 expired 的 seat 都要保护：seat.plan_id 锁定到 plan，
	// 删 plan 后 RenewSeat / PreviewRenewal / 退款流程全部失败
	// （共享池 user_subscription 不存 plan_id，不受影响）
	protected, err := s.countProtectedSeatsByPlan(ctx, id)
	if err != nil {
		return fmt.Errorf("check protected exclusive seats: %w", err)
	}
	if protected > 0 {
		return infraerrors.Conflict("ACTIVE_SEATS_BOUND",
			fmt.Sprintf("this plan has %d active or recently-expired (within renewal grace) exclusive seats bound to it; mark for_sale=false to take it offline instead of deleting", protected))
	}
	return s.entClient.SubscriptionPlan.DeleteOneID(id).Exec(ctx)
}

// GetPlan returns a subscription plan by ID.
func (s *PaymentConfigService) GetPlan(ctx context.Context, id int64) (*dbent.SubscriptionPlan, error) {
	plan, err := s.entClient.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	return plan, nil
}

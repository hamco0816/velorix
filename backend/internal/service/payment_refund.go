package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/paymentproviderinstance"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// --- Refund Flow ---

// getOrderProviderInstance looks up the provider instance that processed this order.
// For legacy orders without provider_instance_id, it resolves only when the
// historical instance is uniquely identifiable from the stored order fields.
func (s *PaymentService) getOrderProviderInstance(ctx context.Context, o *dbent.PaymentOrder) (*dbent.PaymentProviderInstance, error) {
	if s == nil || s.entClient == nil || o == nil {
		return nil, nil
	}

	if snapshot := psOrderProviderSnapshot(o); snapshot != nil {
		return s.resolveSnapshotOrderProviderInstance(ctx, o, snapshot)
	}

	instIDStr := strings.TrimSpace(psStringValue(o.ProviderInstanceID))
	if instIDStr == "" {
		return s.resolveUniqueLegacyOrderProviderInstance(ctx, o)
	}

	instID, err := strconv.ParseInt(instIDStr, 10, 64)
	if err != nil {
		return nil, nil
	}
	return s.entClient.PaymentProviderInstance.Get(ctx, instID)
}

// getRefundOrderProviderInstance resolves the provider instance for refund paths.
// Refunds must be pinned to an explicit historical binding, so legacy
// "best-effort" provider guessing is intentionally not allowed here.
func (s *PaymentService) getRefundOrderProviderInstance(ctx context.Context, o *dbent.PaymentOrder) (*dbent.PaymentProviderInstance, error) {
	if s == nil || s.entClient == nil || o == nil {
		return nil, nil
	}

	if snapshot := psOrderProviderSnapshot(o); snapshot != nil {
		return s.resolveSnapshotOrderProviderInstance(ctx, o, snapshot)
	}

	instIDStr := strings.TrimSpace(psStringValue(o.ProviderInstanceID))
	if instIDStr == "" {
		return nil, nil
	}

	instID, err := strconv.ParseInt(instIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("order %d refund provider instance id is invalid: %s", o.ID, instIDStr)
	}
	inst, err := s.entClient.PaymentProviderInstance.Get(ctx, instID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("order %d refund provider instance %s is missing", o.ID, instIDStr)
		}
		return nil, err
	}
	return inst, nil
}

func (s *PaymentService) resolveUniqueLegacyOrderProviderInstance(ctx context.Context, o *dbent.PaymentOrder) (*dbent.PaymentProviderInstance, error) {
	paymentType := payment.GetBasePaymentType(strings.TrimSpace(o.PaymentType))
	providerKey := strings.TrimSpace(psStringValue(o.ProviderKey))
	if providerKey != "" {
		instances, err := s.entClient.PaymentProviderInstance.Query().
			Where(paymentproviderinstance.ProviderKeyEQ(providerKey)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		matched := psFilterLegacyOrderProviderInstances(paymentType, instances)
		if len(matched) == 1 {
			return matched[0], nil
		}
		return nil, nil
	}

	if paymentType == "" {
		return nil, nil
	}

	instances, err := s.entClient.PaymentProviderInstance.Query().
		All(ctx)
	if err != nil {
		return nil, err
	}

	matched := psFilterLegacyOrderProviderInstances(paymentType, instances)
	if len(matched) == 1 {
		return matched[0], nil
	}
	return nil, nil
}

func psFilterLegacyOrderProviderInstances(orderPaymentType string, instances []*dbent.PaymentProviderInstance) []*dbent.PaymentProviderInstance {
	if len(instances) == 0 {
		return nil
	}
	if strings.TrimSpace(orderPaymentType) == "" {
		return instances
	}
	var matched []*dbent.PaymentProviderInstance
	for _, inst := range instances {
		if psLegacyOrderMatchesInstance(orderPaymentType, inst) {
			matched = append(matched, inst)
		}
	}
	return matched
}

func psLegacyOrderMatchesInstance(orderPaymentType string, inst *dbent.PaymentProviderInstance) bool {
	if inst == nil {
		return false
	}

	baseType := payment.GetBasePaymentType(strings.TrimSpace(orderPaymentType))
	instanceProviderKey := strings.TrimSpace(inst.ProviderKey)
	if baseType == "" {
		return false
	}

	if baseType == payment.TypeStripe {
		return instanceProviderKey == payment.TypeStripe
	}
	if instanceProviderKey == payment.TypeStripe {
		return false
	}
	if instanceProviderKey == baseType {
		return true
	}
	return payment.InstanceSupportsType(inst.SupportedTypes, baseType)
}

func (s *PaymentService) RequestRefund(ctx context.Context, oid, uid int64, reason string) error {
	o, err := s.validateRefundRequest(ctx, oid, uid)
	if err != nil {
		return err
	}
	u, err := s.userRepo.GetByID(ctx, o.UserID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if u.Balance < o.Amount {
		return infraerrors.BadRequest("BALANCE_NOT_ENOUGH", "refund amount exceeds balance")
	}
	nr := strings.TrimSpace(reason)
	now := time.Now()
	by := fmt.Sprintf("%d", uid)
	c, err := s.entClient.PaymentOrder.Update().Where(paymentorder.IDEQ(oid), paymentorder.UserIDEQ(uid), paymentorder.StatusEQ(OrderStatusCompleted), paymentorder.OrderTypeEQ(payment.OrderTypeBalance)).SetStatus(OrderStatusRefundRequested).SetRefundRequestedAt(now).SetRefundRequestReason(nr).SetRefundRequestedBy(by).SetRefundAmount(o.Amount).Save(ctx)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	if c == 0 {
		return infraerrors.Conflict("CONFLICT", "order status changed")
	}
	s.writeAuditLog(ctx, oid, "REFUND_REQUESTED", fmt.Sprintf("user:%d", uid), map[string]any{"amount": o.Amount, "reason": nr})
	return nil
}

func (s *PaymentService) validateRefundRequest(ctx context.Context, oid, uid int64) (*dbent.PaymentOrder, error) {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.UserID != uid {
		return nil, infraerrors.Forbidden("FORBIDDEN", "no permission")
	}
	if o.OrderType != payment.OrderTypeBalance {
		return nil, infraerrors.BadRequest("INVALID_ORDER_TYPE", "only balance orders can request refund")
	}
	if o.Status != OrderStatusCompleted {
		return nil, infraerrors.BadRequest("INVALID_STATUS", "only completed orders can request refund")
	}
	// Check provider instance allows user refund
	inst, err := s.getRefundOrderProviderInstance(ctx, o)
	if err != nil || inst == nil {
		return nil, infraerrors.Forbidden("USER_REFUND_DISABLED", "refund is not available for this order")
	}
	if !inst.AllowUserRefund {
		return nil, infraerrors.Forbidden("USER_REFUND_DISABLED", "user refund is not enabled for this provider")
	}
	return o, nil
}

func (s *PaymentService) PrepareRefund(ctx context.Context, oid int64, amt float64, reason string, force, deduct bool) (*RefundPlan, *RefundResult, error) {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return nil, nil, infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	// OrderStatusFailed 也允许退款：fulfillment 失败但钱已收（如独享池库存抢光场景）
	// 必须给用户退款，否则收钱不发货等于黑产
	// OrderStatusPartiallyRefunded 允许继续退款：管理员前置 UI 已支持二次退款（GPT round 23 #3），
	// 后端 allow 列表必须同步包含，否则界面能点但接口拒绝
	ok := []string{OrderStatusCompleted, OrderStatusRefundRequested, OrderStatusRefundFailed, OrderStatusFailed, OrderStatusPartiallyRefunded}
	if !psSliceContains(ok, o.Status) {
		return nil, nil, infraerrors.BadRequest("INVALID_STATUS", "order status does not allow refund")
	}
	// 已部分退款时，剩余可退金额 = order.Amount - order.RefundAmount
	if o.Status == OrderStatusPartiallyRefunded {
		remaining := o.Amount - o.RefundAmount
		if remaining <= 0 {
			return nil, nil, infraerrors.BadRequest("FULLY_REFUNDED", "this order has already been fully refunded")
		}
	}
	// Check provider instance allows admin refund
	inst, instErr := s.getRefundOrderProviderInstance(ctx, o)
	if instErr != nil {
		slog.Warn("refund: provider instance lookup failed", "orderID", oid, "error", instErr)
		return nil, nil, infraerrors.InternalServer("PROVIDER_LOOKUP_FAILED", "failed to look up payment provider for this order")
	}
	if inst == nil {
		// Legacy order without provider_instance_id — block refund
		return nil, nil, infraerrors.Forbidden("REFUND_DISABLED", "refund is not available for this order")
	}
	if !inst.RefundEnabled {
		return nil, nil, infraerrors.Forbidden("REFUND_DISABLED", "refund is not enabled for this provider")
	}
	if math.IsNaN(amt) || math.IsInf(amt, 0) {
		return nil, nil, infraerrors.BadRequest("INVALID_AMOUNT", "invalid refund amount")
	}
	// 已部分退款的订单：本次最多再退"剩余可退金额"= order.Amount - order.RefundAmount。
	// 注意 REFUND_REQUESTED：用户申请退款时 RequestRefund 把"申请金额"暂存在 refund_amount，
	// 这阶段它语义是"申请金额"，不是"已退金额"。审批时必须按 0 已退处理，否则用户全额申请后
	// 管理员审批会被算成"剩余可退 = 0"直接拒绝（GPT round 24 #1）。
	alreadyRefunded := o.RefundAmount
	if o.Status == OrderStatusRefundRequested {
		alreadyRefunded = 0
	}
	maxRefundable := o.Amount - alreadyRefunded
	if amt <= 0 {
		amt = maxRefundable
	}
	if amt-maxRefundable > amountToleranceCNY {
		return nil, nil, infraerrors.BadRequest("REFUND_AMOUNT_EXCEEDED", "refund amount exceeds remaining refundable")
	}
	ga := calculateGatewayRefundAmount(o.Amount, o.PayAmount, amt)
	rr := strings.TrimSpace(reason)
	if rr == "" && o.RefundRequestReason != nil {
		rr = *o.RefundRequestReason
	}
	if rr == "" {
		rr = fmt.Sprintf("refund order:%d", o.ID)
	}
	p := &RefundPlan{OrderID: oid, Order: o, RefundAmount: amt, GatewayAmount: ga, Reason: rr, Force: force, DeductBalance: deduct, DeductionType: payment.DeductionTypeNone}
	if deduct {
		if er := s.prepDeduct(ctx, o, p, force); er != nil {
			return nil, er, nil
		}
	}
	return p, nil, nil
}

func (s *PaymentService) prepDeduct(ctx context.Context, o *dbent.PaymentOrder, p *RefundPlan, force bool) *RefundResult {
	if o.OrderType == payment.OrderTypeSubscription {
		// 独享套餐订单不扣 user_subscription：seat 是独立实体，
		// 退款时由 releaseSeatForRefundedOrder 在 ExecuteRefund 末尾释放/缩短具体 seat
		// （扣共享 sub 会影响同 group 其他独享 seat 的鉴权期限，跨 seat 串扰）
		if o.PlanID != nil {
			if plan, err := s.entClient.SubscriptionPlan.Get(ctx, *o.PlanID); err == nil && plan != nil && plan.Kind == domain.PlanKindExclusive {
				p.DeductionType = payment.DeductionTypeNone
				return nil
			}
		}
		p.DeductionType = payment.DeductionTypeSubscription
		if o.SubscriptionGroupID != nil && o.SubscriptionDays != nil {
			p.SubDaysToDeduct = *o.SubscriptionDays
			sub, err := s.subscriptionSvc.GetActiveSubscription(ctx, o.UserID, *o.SubscriptionGroupID)
			if err == nil && sub != nil {
				p.SubscriptionID = sub.ID
			} else if !force {
				return &RefundResult{Success: false, Warning: "cannot find active subscription for deduction, use force", RequireForce: true}
			}
		}
		return nil
	}
	u, err := s.userRepo.GetByID(ctx, o.UserID)
	if err != nil {
		if !force {
			return &RefundResult{Success: false, Warning: "cannot fetch user balance, use force", RequireForce: true}
		}
		return nil
	}
	p.DeductionType = payment.DeductionTypeBalance
	p.BalanceToDeduct = math.Min(p.RefundAmount, u.Balance)
	return nil
}

func (s *PaymentService) ExecuteRefund(ctx context.Context, p *RefundPlan) (*RefundResult, error) {
	c, err := s.entClient.PaymentOrder.Update().Where(paymentorder.IDEQ(p.OrderID), paymentorder.StatusIn(OrderStatusCompleted, OrderStatusRefundRequested, OrderStatusRefundFailed, OrderStatusFailed, OrderStatusPartiallyRefunded)).SetStatus(OrderStatusRefunding).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("lock: %w", err)
	}
	if c == 0 {
		return nil, infraerrors.Conflict("CONFLICT", "order status changed")
	}
	if p.DeductionType == payment.DeductionTypeBalance && p.BalanceToDeduct > 0 {
		// Skip balance deduction on retry if previous attempt already deducted
		// but failed to roll back (REFUND_ROLLBACK_FAILED in audit log).
		if !s.hasAuditLog(ctx, p.OrderID, "REFUND_ROLLBACK_FAILED") {
			if err := s.userRepo.DeductBalance(ctx, p.Order.UserID, p.BalanceToDeduct); err != nil {
				s.restoreStatus(ctx, p)
				return nil, fmt.Errorf("deduction: %w", err)
			}
		} else {
			slog.Warn("skipping balance deduction on retry (previous rollback failed)", "orderID", p.OrderID)
			p.BalanceToDeduct = 0
		}
	}
	if p.DeductionType == payment.DeductionTypeSubscription && p.SubDaysToDeduct > 0 && p.SubscriptionID > 0 {
		if !s.hasAuditLog(ctx, p.OrderID, "REFUND_ROLLBACK_FAILED") {
			_, err := s.subscriptionSvc.ExtendSubscription(ctx, p.SubscriptionID, -p.SubDaysToDeduct)
			if err != nil {
				if errors.Is(err, ErrAdjustWouldExpire) {
					// Deduction would expire the subscription — revoke it entirely
					slog.Info("subscription deduction would expire, revoking", "orderID", p.OrderID, "subID", p.SubscriptionID, "days", p.SubDaysToDeduct)
					if revokeErr := s.subscriptionSvc.RevokeSubscription(ctx, p.SubscriptionID); revokeErr != nil {
						s.restoreStatus(ctx, p)
						return nil, fmt.Errorf("revoke subscription: %w", revokeErr)
					}
				} else {
					// Other errors (DB failure, not found) — abort refund
					s.restoreStatus(ctx, p)
					return nil, fmt.Errorf("deduct subscription days: %w", err)
				}
			}
		} else {
			slog.Warn("skipping subscription deduction on retry (previous rollback failed)", "orderID", p.OrderID)
			p.SubDaysToDeduct = 0
		}
	}
	if err := s.gwRefund(ctx, p); err != nil {
		return s.handleGwFail(ctx, p, err)
	}
	return s.markRefundOk(ctx, p)
}

func (s *PaymentService) gwRefund(ctx context.Context, p *RefundPlan) error {
	if p.Order.PaymentTradeNo == "" {
		s.writeAuditLog(ctx, p.Order.ID, "REFUND_NO_TRADE_NO", "admin", map[string]any{"detail": "skipped"})
		return nil
	}

	// Use the exact provider instance that created this order, not a random one
	// from the registry. Each instance has its own merchant credentials.
	prov, err := s.getRefundProvider(ctx, p.Order)
	if err != nil {
		return fmt.Errorf("get refund provider: %w", err)
	}
	if err := validateProviderSnapshotMetadata(p.Order, prov.ProviderKey(), providerMerchantIdentityMetadata(prov)); err != nil {
		s.writeAuditLog(ctx, p.Order.ID, "REFUND_PROVIDER_METADATA_MISMATCH", "admin", map[string]any{
			"detail": err.Error(),
		})
		return err
	}
	_, err = prov.Refund(ctx, payment.RefundRequest{
		TradeNo: p.Order.PaymentTradeNo,
		OrderID: p.Order.OutTradeNo,
		Amount:  strconv.FormatFloat(p.GatewayAmount, 'f', 2, 64),
		Reason:  p.Reason,
	})
	return err
}

// getRefundProvider creates a provider using the order's original instance config.
// Delegates to getOrderProvider which handles instance lookup and fallback.
func (s *PaymentService) getRefundProvider(ctx context.Context, o *dbent.PaymentOrder) (payment.Provider, error) {
	inst, err := s.getRefundOrderProviderInstance(ctx, o)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, fmt.Errorf("refund provider instance is unavailable for order %d", o.ID)
	}
	return s.createProviderFromInstance(ctx, inst)
}

func (s *PaymentService) handleGwFail(ctx context.Context, p *RefundPlan, gErr error) (*RefundResult, error) {
	if s.RollbackRefund(ctx, p, gErr) {
		s.restoreStatus(ctx, p)
		s.writeAuditLog(ctx, p.OrderID, "REFUND_GATEWAY_FAILED", "admin", map[string]any{"detail": psErrMsg(gErr)})
		return &RefundResult{Success: false, Warning: "gateway failed: " + psErrMsg(gErr) + ", rolled back"}, nil
	}
	now := time.Now()
	_, _ = s.entClient.PaymentOrder.UpdateOneID(p.OrderID).SetStatus(OrderStatusRefundFailed).SetFailedAt(now).SetFailedReason(psErrMsg(gErr)).Save(ctx)
	s.writeAuditLog(ctx, p.OrderID, "REFUND_FAILED", "admin", map[string]any{"detail": psErrMsg(gErr)})
	return nil, infraerrors.InternalServer("REFUND_FAILED", psErrMsg(gErr))
}

func (s *PaymentService) markRefundOk(ctx context.Context, p *RefundPlan) (*RefundResult, error) {
	// 二次退款累加：order.RefundAmount = 之前已退 + 本次新退（GPT round 23 #3 续）。
	// REFUND_REQUESTED 阶段的 refund_amount 是"申请金额"占位（GPT round 24 #1），
	// 与 PrepareRefund 同款处理：累计已退按 0 计算，不要把"申请额"误算进去。
	prevRefunded := p.Order.RefundAmount
	if p.Order.Status == OrderStatusRefundRequested {
		prevRefunded = 0
	}
	totalRefunded := prevRefunded + p.RefundAmount
	fs := OrderStatusRefunded
	isPartial := totalRefunded+amountToleranceCNY < p.Order.Amount
	if isPartial {
		fs = OrderStatusPartiallyRefunded
	}
	now := time.Now()
	_, err := s.entClient.PaymentOrder.UpdateOneID(p.OrderID).SetStatus(fs).SetRefundAmount(totalRefunded).SetRefundReason(p.Reason).SetRefundAt(now).SetForceRefund(p.Force).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark refund: %w", err)
	}
	s.writeAuditLog(ctx, p.OrderID, "REFUND_SUCCESS", "admin", map[string]any{
		"refundAmount":    p.RefundAmount,
		"reason":          p.Reason,
		"balanceDeducted": p.BalanceToDeduct,
		"force":           p.Force,
		"partial":         isPartial,
	})
	// 独享套餐订单退款：仅在全额退款时释放 seat / 撤销整单续费天数（GPT round 23 #2）。
	// 部分退款只退一部分钱给用户，不应同时取消独享名额——否则用户体验是"只退 30% 但服务没了"。
	// 部分退款的实际服务损失靠后续运营或全额退款时再处理。
	if !isPartial {
		s.releaseSeatForRefundedOrder(ctx, p.Order, p.Reason)
	} else {
		s.writeAuditLog(ctx, p.OrderID, "PARTIAL_REFUND_SEAT_RETAINED", "system", map[string]any{
			"refundAmount": p.RefundAmount,
			"orderAmount":  p.Order.Amount,
			"reason":       "partial refund: keeping exclusive seat / renewal days intact",
		})
	}
	return &RefundResult{Success: true, BalanceDeducted: p.BalanceToDeduct, SubDaysDeducted: p.SubDaysToDeduct}, nil
}

// releaseSeatForRefundedOrder 在订单退款完成时尝试释放对应的 seat。
// 仅对 order_type=subscription 且 plan.kind=exclusive 的订单生效；其他订单直接 noop。
// 失败仅记录日志，不向上抛错，避免影响退款主流程。
//
// 处理优先级：
//  0. 续费订单（renewal_seat_id != null）：缩短 seat.expires_at 撤销续费效果
//  1. 新购订单：通过 source_order_id 精确反查 → 整个 seat 释放为 refunded
//  2. 找不到时回退到 (user_id, plan_id) 反查最近一份 active seat（向后兼容旧数据）
func (s *PaymentService) releaseSeatForRefundedOrder(ctx context.Context, o *dbent.PaymentOrder, reason string) {
	if o == nil || o.OrderType != payment.OrderTypeSubscription || o.PlanID == nil {
		return
	}
	plan, err := s.entClient.SubscriptionPlan.Get(ctx, *o.PlanID)
	if err != nil || plan.Kind != domain.PlanKindExclusive {
		return
	}

	// 0) 续费订单退款：撤销续费天数（不释放整个 seat），避免用户白嫖续费时长。
	// 关键：必须先确认续费真的发生过（doSub 写过 EXCLUSIVE_SEAT_RENEWED 审计），
	// 否则 fulfillment 失败的自动退款会把一个根本没续费成功的 seat 凭空减天数。
	if o.RenewalSeatID != nil && *o.RenewalSeatID > 0 && o.SubscriptionDays != nil && *o.SubscriptionDays > 0 {
		if !s.hasAuditLog(ctx, o.ID, "EXCLUSIVE_SEAT_RENEWED") {
			slog.Info("[ExclusiveSeat] skip revoke-renewal: renewal never succeeded", "seat_id", *o.RenewalSeatID, "order_id", o.ID)
			return
		}
		seat, revErr := s.exclusiveSeatSvc.RevokeRenewal(ctx, *o.RenewalSeatID, *o.SubscriptionDays)
		if revErr != nil {
			slog.Error("[ExclusiveSeat] revoke-renewal-on-refund failed", "seat_id", *o.RenewalSeatID, "order_id", o.ID, "days", *o.SubscriptionDays, "error", revErr)
			s.recordSeatReleaseFailure(ctx, o.ID, *o.RenewalSeatID, "revoke_renewal_failed", reason, revErr)
			return
		}
		s.writeAuditLog(ctx, o.ID, "EXCLUSIVE_SEAT_RENEWAL_REVOKED", "system", map[string]any{
			"seat_id":        fmt.Sprintf("%d", seat.ID),
			"days_reverted":  fmt.Sprintf("%d", *o.SubscriptionDays),
			"new_expires_at": seat.ExpiresAt.Format(time.RFC3339),
			"reason":         reason,
		})
		return
	}

	// 1) 精确匹配：通过 source_order_id 反查
	if seat, err := s.exclusiveSeatSvc.FindActiveSeatByOrder(ctx, o.ID); err == nil && seat != nil {
		s.releaseAndAudit(ctx, o.ID, seat, reason)
		return
	}

	// "无库存自动退款"场景（GPT round 29 #2）：fulfillment 失败，根本没创建过 seat。
	// 跳过 fallback，否则 (user_id, plan_id) 反查会误释放用户已有的迁移前旧 seat。
	// 同样，未写过 SUBSCRIPTION_SUCCESS 审计的订单都视为"没履约成功"，没什么 seat 需要回收。
	if s.hasAuditLog(ctx, o.ID, "EXCLUSIVE_SEAT_NO_STOCK") || !s.hasAuditLog(ctx, o.ID, "SUBSCRIPTION_SUCCESS") {
		slog.Info("[ExclusiveSeat] skip release-on-refund: order never produced a seat",
			"order_id", o.ID, "user_id", o.UserID, "plan_id", *o.PlanID)
		return
	}

	// 2) 回退：旧数据（migration 136 之前的 seat 没有 source_order_id）按 (user_id, plan_id) 反查。
	// 已经被前一步 fulfillment 守卫过滤掉"未真正发货"的失败单，不会误伤用户原有 seat。
	seats, err := s.exclusiveSeatSvc.ListActiveByUser(ctx, o.UserID, plan.GroupID)
	if err != nil {
		slog.Error("[ExclusiveSeat] release-on-refund: list active failed", "order_id", o.ID, "error", err)
		s.recordSeatReleaseFailure(ctx, o.ID, 0, "list_active_failed", reason, err)
		return
	}
	for _, seat := range seats {
		if seat.PlanID != *o.PlanID {
			continue
		}
		// 优先选不带 source_order_id 的旧 seat（避免误释放新数据）
		if seat.SourceOrderID != nil {
			continue
		}
		s.releaseAndAudit(ctx, o.ID, seat, reason)
		return
	}
}

// releaseAndAudit 把 seat 释放为 refunded 并写一条审计日志。
// 失败时除日志外另写一条 EXCLUSIVE_SEAT_RELEASE_FAILED 审计，让告警系统/管理后台能识别
// "退款已成功但 seat 仍可用"的危险状态，便于人工或重试任务介入修复。
func (s *PaymentService) releaseAndAudit(ctx context.Context, orderID int64, seat *dbent.ExclusiveSubscription, reason string) {
	if err := s.exclusiveSeatSvc.ReleaseSeat(ctx, seat.ID, domain.ExclusiveSeatStatusRefunded, reason); err != nil {
		slog.Error("[ExclusiveSeat] release-on-refund failed", "seat_id", seat.ID, "order_id", orderID, "error", err)
		s.recordSeatReleaseFailure(ctx, orderID, seat.ID, "release_seat_failed", reason, err)
		return
	}
	s.writeAuditLog(ctx, orderID, "EXCLUSIVE_SEAT_RELEASED_ON_REFUND", "admin", formatSeatAuditFields(seat))
}

// SeatReleaseFailureItem 待修复的 seat 释放失败项视图（admin 列表 / 重试 worker 共用）。
type SeatReleaseFailureItem struct {
	OrderID    int64     `json:"order_id"`
	SeatID     int64     `json:"seat_id"`
	Stage      string    `json:"stage"`
	Reason     string    `json:"reason"`
	Error      string    `json:"error"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ListPendingSeatReleaseFailures 列出"退款已成功但 seat 释放失败、且尚未补救成功"的订单。
// 取最近 since 时间窗内的 EXCLUSIVE_SEAT_RELEASE_FAILED 审计；同 order_id 已经补到
// EXCLUSIVE_SEAT_RELEASED_ON_REFUND 或 EXCLUSIVE_SEAT_RENEWAL_REVOKED 的视为已闭环跳过。
//
// 用于：(1) 后台 SeatReleaseRetryService 自动扫描重试 (2) admin API 排查列表
func (s *PaymentService) ListPendingSeatReleaseFailures(ctx context.Context, since time.Duration, limit int) ([]SeatReleaseFailureItem, error) {
	if s == nil || s.entClient == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	cutoff := time.Now().Add(-since)
	logs, err := s.queryAuditLogsSince(ctx, "EXCLUSIVE_SEAT_RELEASE_FAILED", cutoff, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SeatReleaseFailureItem, 0, len(logs))
	for _, lg := range logs {
		// 已经有后续闭环动作就跳过（避免重试已修复的）
		if s.hasAuditLog(ctx, parsePendingOrderID(lg.OrderID), "EXCLUSIVE_SEAT_RELEASED_ON_REFUND") {
			continue
		}
		if s.hasAuditLog(ctx, parsePendingOrderID(lg.OrderID), "EXCLUSIVE_SEAT_RENEWAL_REVOKED") {
			continue
		}
		item := SeatReleaseFailureItem{OccurredAt: lg.CreatedAt}
		fillSeatReleaseFailureFromDetail(lg.Detail, &item)
		out = append(out, item)
	}
	return out, nil
}

// RetrySeatReleaseForOrder 重新跑一次 releaseSeatForRefundedOrder。
// 提供给 admin API（手动按订单触发）和 SeatReleaseRetryService（后台周期扫描）。
// 找不到 order / 已闭环时直接 nil；重试成功会写新的 RELEASED_ON_REFUND/REVOKED 审计自动闭环。
func (s *PaymentService) RetrySeatReleaseForOrder(ctx context.Context, orderID int64) error {
	if s == nil || s.entClient == nil || orderID <= 0 {
		return nil
	}
	// 已闭环 → 直接返回
	if s.hasAuditLog(ctx, orderID, "EXCLUSIVE_SEAT_RELEASED_ON_REFUND") ||
		s.hasAuditLog(ctx, orderID, "EXCLUSIVE_SEAT_RENEWAL_REVOKED") {
		return nil
	}
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return fmt.Errorf("load order: %w", err)
	}
	// 必须是"全额已成功退款"的订单才允许重试 seat 释放（GPT round 24 #2）：
	// PartiallyRefunded 在主路径已经被 markRefundOk 显式跳过 release（"只退一部分钱不取消独享号"），
	// 这里的重试入口也必须排除，否则后台 worker / admin 手动重试会绕过保护把 seat 完整释放掉。
	if order.Status != OrderStatusRefunded {
		return infraerrors.BadRequest("ORDER_NOT_FULLY_REFUNDED",
			"only fully-refunded orders are eligible for seat-release retry; partial refunds intentionally keep the seat")
	}
	reason := psStringValue(order.RefundReason)
	if strings.TrimSpace(reason) == "" {
		reason = "retry: seat release failure repair"
	}
	s.releaseSeatForRefundedOrder(ctx, order, reason)
	return nil
}

// queryAuditLogsSince 取最近 cutoff 时间窗内某 action 的审计记录（限制 limit 条），按时间倒序。
func (s *PaymentService) queryAuditLogsSince(ctx context.Context, action string, cutoff time.Time, limit int) ([]*dbent.PaymentAuditLog, error) {
	return s.entClient.PaymentAuditLog.Query().
		Where(
			paymentauditlog.ActionEQ(action),
			paymentauditlog.CreatedAtGTE(cutoff),
		).
		Order(paymentauditlog.ByCreatedAt()).
		Limit(limit).All(ctx)
}

// parsePendingOrderID 把审计日志的 string order_id 转回 int64，失败返回 0
func parsePendingOrderID(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// fillSeatReleaseFailureFromDetail 把 EXCLUSIVE_SEAT_RELEASE_FAILED 的 detail JSON 解析到 item 上。
func fillSeatReleaseFailureFromDetail(detail string, item *SeatReleaseFailureItem) {
	if item == nil || strings.TrimSpace(detail) == "" {
		return
	}
	var raw map[string]string
	if err := json.Unmarshal([]byte(detail), &raw); err != nil {
		return
	}
	item.OrderID = parsePendingOrderID(raw["order_id"])
	item.SeatID = parsePendingOrderID(raw["seat_id"])
	item.Stage = raw["stage"]
	item.Reason = raw["reason"]
	item.Error = raw["error"]
}

// recordSeatReleaseFailure 在退款已经写成功但 seat 释放/续费撤销失败时留显式痕迹。
// 仅 best-effort 写审计 + ERROR 级日志（带 'seat_release_failed' 关键词便于告警系统抓取）。
// 不抛错，避免给已经成功的退款流程引入回滚副作用。
func (s *PaymentService) recordSeatReleaseFailure(ctx context.Context, orderID, seatID int64, stage, reason string, cause error) {
	errMsg := ""
	if cause != nil {
		errMsg = cause.Error()
	}
	slog.Error("seat_release_failed",
		"order_id", orderID, "seat_id", seatID, "stage", stage, "reason", reason, "error", errMsg)
	s.writeAuditLog(ctx, orderID, "EXCLUSIVE_SEAT_RELEASE_FAILED", "system", map[string]any{
		"order_id": fmt.Sprintf("%d", orderID),
		"seat_id":  fmt.Sprintf("%d", seatID),
		"stage":    stage,
		"reason":   reason,
		"error":    errMsg,
	})
}

func (s *PaymentService) RollbackRefund(ctx context.Context, p *RefundPlan, gErr error) bool {
	if p.DeductionType == payment.DeductionTypeBalance && p.BalanceToDeduct > 0 {
		if err := s.userRepo.UpdateBalance(ctx, p.Order.UserID, p.BalanceToDeduct); err != nil {
			slog.Error("[CRITICAL] rollback failed", "orderID", p.OrderID, "amount", p.BalanceToDeduct, "error", err)
			s.writeAuditLog(ctx, p.OrderID, "REFUND_ROLLBACK_FAILED", "admin", map[string]any{"gatewayError": psErrMsg(gErr), "rollbackError": psErrMsg(err), "balanceDeducted": p.BalanceToDeduct})
			return false
		}
	}
	if p.DeductionType == payment.DeductionTypeSubscription && p.SubDaysToDeduct > 0 && p.SubscriptionID > 0 {
		if _, err := s.subscriptionSvc.ExtendSubscription(ctx, p.SubscriptionID, p.SubDaysToDeduct); err != nil {
			slog.Error("[CRITICAL] subscription rollback failed", "orderID", p.OrderID, "subID", p.SubscriptionID, "days", p.SubDaysToDeduct, "error", err)
			s.writeAuditLog(ctx, p.OrderID, "REFUND_ROLLBACK_FAILED", "admin", map[string]any{"gatewayError": psErrMsg(gErr), "rollbackError": psErrMsg(err), "subDaysDeducted": p.SubDaysToDeduct})
			return false
		}
	}
	return true
}

func (s *PaymentService) restoreStatus(ctx context.Context, p *RefundPlan) {
	rs := OrderStatusCompleted
	switch p.Order.Status {
	case OrderStatusRefundRequested:
		rs = OrderStatusRefundRequested
	case OrderStatusFailed:
		// 来自 fulfillment 失败的订单（如独享池抢光），退款也失败时回到 FAILED
		// 不要错误回到 COMPLETED——那意味着"已发货"，跟实际不符
		rs = OrderStatusFailed
	case OrderStatusPartiallyRefunded:
		// 二次退款失败时必须保留 PARTIALLY_REFUNDED：之前已经退过一部分钱，order.RefundAmount 也不是 0，
		// 默认回 COMPLETED 会让前端"已退款/可退金额"判断完全错误（GPT round 25 #1）。
		rs = OrderStatusPartiallyRefunded
	case OrderStatusRefundFailed:
		// 之前已经"退款失败"过的订单：再次失败仍保留 REFUND_FAILED 语义
		rs = OrderStatusRefundFailed
	}
	_, _ = s.entClient.PaymentOrder.UpdateOneID(p.OrderID).SetStatus(rs).Save(ctx)
}

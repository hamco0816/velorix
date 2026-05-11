// Package service 中独享池（独享名额）相关业务逻辑。
//
// 独享名额代表「一个用户独占一个上游账号一段时间」：
//   - 支付完成时通过 AssignSeat 在事务内挑一个空闲账号绑定（FOR UPDATE SKIP LOCKED 防超卖）
//   - 调度器查询用户的活跃名额来跳过共享池负载均衡，直接使用绑定账号
//   - 过期回收任务、退款、用户/管理员主动取消时通过 ReleaseSeat 释放
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/accountgroup"
	"github.com/Wei-Shaw/sub2api/ent/exclusivesubscription"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// ErrNoFreeAccount 独享池库存不足（没有空闲账号可分配）。
// 调用方根据具体场景决定后续动作（前台限购 / 自动退款 / 提示管理员补货）。
var ErrNoFreeAccount = errors.New("exclusive_pool: no free account in pool")

// ExclusiveSeatService 提供独享名额的全生命周期管理。
// SchedulerOutboxNotifier 在独享 seat 操作改变 account.assigned_seat_id 后触发，
// 让调度快照失效并刷新（避免共享池快照里残留刚被独享占用/刚释放的账号）。
// 由 wire 阶段从 accountRepo 适配注入；nil 时为 noop（兼容旧代码与测试）
type SchedulerOutboxNotifier interface {
	NotifyAccountChanged(ctx context.Context, accountID int64) error
}

type ExclusiveSeatService struct {
	entClient *dbent.Client
	outbox    SchedulerOutboxNotifier
}

// NewExclusiveSeatService 创建一个 ExclusiveSeatService 实例。
func NewExclusiveSeatService(entClient *dbent.Client) *ExclusiveSeatService {
	return &ExclusiveSeatService{entClient: entClient}
}

// SetSchedulerOutbox 注入 outbox 通知器（避免构造期循环依赖）。
func (s *ExclusiveSeatService) SetSchedulerOutbox(outbox SchedulerOutboxNotifier) {
	if s != nil {
		s.outbox = outbox
	}
}

// notifySchedulerAccountChanged 安全调用 outbox（nil 时 noop，错误仅记日志不阻塞主流程）。
func (s *ExclusiveSeatService) notifySchedulerAccountChanged(ctx context.Context, accountID int64) {
	if s == nil || s.outbox == nil || accountID <= 0 {
		return
	}
	if err := s.outbox.NotifyAccountChanged(ctx, accountID); err != nil {
		slog.Warn("[ExclusiveSeat] scheduler outbox notify failed", "account_id", accountID, "error", err)
	}
}

// AssignSeatInput 独享名额分配的输入参数。
type AssignSeatInput struct {
	UserID       int64
	GroupID      int64
	PlanID       int64
	ValidityDays int
	// AssignedBy 管理员赠送时填管理员 ID；正常支付分配为 0
	AssignedBy int64
	// SourceOrderID 关联的支付订单 ID；管理员赠送或换号时为 0（不填）
	SourceOrderID int64
	Notes         string
}

// AssignSeat 为用户在指定独享池中分配一份名额。
//
// 关键步骤（事务内）：
//  1. 通过 SELECT ... FOR UPDATE SKIP LOCKED 在该 group 池子里挑一个空闲账号
//  2. 没有空闲账号则返回 ErrNoFreeAccount（调用方处理：通常是自动退款）
//  3. 创建 exclusive_subscription 记录
//  4. 更新 account.assigned_seat_id 反向索引
//
// 并发安全：FOR UPDATE SKIP LOCKED 保证两个并发请求不会拿到同一个账号；
// 同时 exclusive_subscriptions.account_id 上的部分唯一索引兜底防止脏数据。
func (s *ExclusiveSeatService) AssignSeat(ctx context.Context, in AssignSeatInput) (*dbent.ExclusiveSubscription, error) {
	if in.UserID <= 0 || in.GroupID <= 0 || in.PlanID <= 0 || in.ValidityDays <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "user_id/group_id/plan_id/validity_days must be > 0")
	}

	// 幂等保护（与 migration 139 唯一约束配合）：同一 source_order_id 已经创建过 active seat 时直接返回，
	// 避免 fulfillment 链路被重试时重复消耗库存
	if in.SourceOrderID > 0 {
		existing, err := s.entClient.ExclusiveSubscription.Query().
			Where(
				exclusivesubscription.SourceOrderIDEQ(in.SourceOrderID),
				exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
				exclusivesubscription.DeletedAtIsNil(),
			).Only(ctx)
		if err == nil && existing != nil {
			slog.Info("[ExclusiveSeat] AssignSeat idempotent return", "source_order_id", in.SourceOrderID, "seat_id", existing.ID)
			return existing, nil
		}
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 通过原始 SQL 用 FOR UPDATE SKIP LOCKED 挑一个空闲账号 ID
	accountID, err := pickFreeAccountForUpdate(ctx, tx, in.GroupID)
	if err != nil {
		return nil, err
	}

	// 创建独享名额记录
	now := time.Now()
	expiresAt := now.AddDate(0, 0, in.ValidityDays)
	create := tx.ExclusiveSubscription.Create().
		SetUserID(in.UserID).
		SetGroupID(in.GroupID).
		SetPlanID(in.PlanID).
		SetAccountID(accountID).
		SetStatus(domain.ExclusiveSeatStatusActive).
		SetStartsAt(now).
		SetExpiresAt(expiresAt).
		SetAssignedAt(now)
	if in.Notes != "" {
		create.SetNotes(in.Notes)
	}
	if in.AssignedBy > 0 {
		create.SetAssignedBy(in.AssignedBy)
	}
	if in.SourceOrderID > 0 {
		create.SetSourceOrderID(in.SourceOrderID)
	}
	// 把 plan 当前的限额/倍率快照到 seat（NULL 时调度回落到 group）
	if plan, err := tx.SubscriptionPlan.Get(ctx, in.PlanID); err == nil && plan != nil {
		if plan.DailyLimitUsd != nil {
			create.SetDailyLimitUsd(*plan.DailyLimitUsd)
		}
		if plan.WeeklyLimitUsd != nil {
			create.SetWeeklyLimitUsd(*plan.WeeklyLimitUsd)
		}
		if plan.MonthlyLimitUsd != nil {
			create.SetMonthlyLimitUsd(*plan.MonthlyLimitUsd)
		}
		if plan.RateMultiplier != nil {
			create.SetRateMultiplier(*plan.RateMultiplier)
		}
	}
	seat, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create exclusive_subscription: %w", err)
	}

	// 更新 account 反向索引
	if _, err = tx.Account.UpdateOneID(accountID).SetAssignedSeatID(seat.ID).Save(ctx); err != nil {
		return nil, fmt.Errorf("update account.assigned_seat_id: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	// account.assigned_seat_id 变了，通知调度快照刷新（共享池快照不能再列入此账号）
	s.notifySchedulerAccountChanged(ctx, accountID)
	return seat, nil
}

// pickFreeAccountForUpdate 在指定 group 池子里挑一个未被独享名额占用的账号 ID。
// 使用 FOR UPDATE SKIP LOCKED 防止并发抢购时多个事务拿到同一个账号。
//
// 入选条件：
//   - 在该 group 的 account_groups 关联里
//   - account.status = 'active' 且 schedulable = true
//   - account.deleted_at IS NULL
//   - 不在任何 active exclusive_subscription 中
//
// 排序：按 priority 升序（高优先级优先），其次按 last_used_at 升序（最近未使用的优先）
func pickFreeAccountForUpdate(ctx context.Context, tx *dbent.Tx, groupID int64) (int64, error) {
	// SQL: SELECT id FROM accounts a
	//      JOIN account_groups ag ON ag.account_id = a.id
	//      WHERE ag.group_id = $1 AND a.status = 'active' AND a.schedulable = true
	//        AND a.deleted_at IS NULL
	//        AND NOT EXISTS (SELECT 1 FROM exclusive_subscriptions e
	//                        WHERE e.account_id = a.id AND e.status = 'active' AND e.deleted_at IS NULL)
	//      ORDER BY a.priority ASC, a.last_used_at ASC NULLS FIRST
	//      LIMIT 1 FOR UPDATE OF a SKIP LOCKED
	driver, ok := tx.Client().Driver().(interface {
		Dialect() string
	})
	if !ok {
		return 0, fmt.Errorf("ent driver missing Dialect()")
	}
	if driver.Dialect() != dialect.Postgres {
		// SKIP LOCKED 仅在 PG 下有效；非 PG 退化为普通查询（仍由唯一索引兜底）
		return pickFreeAccountFallback(ctx, tx, groupID)
	}

	// 加完整可调度过滤（与 service.Account.IsSchedulable 对齐）：
	//   - rate_limit_reset_at / overload_until / temp_unschedulable_until 必须已过期或未设
	//   - auto_pause_on_expired + expires_at 已到 → 跳过
	// 全池都临时不可用时返回 ErrNoFreeAccount → 触发 autoRefund，比"买到 5h 限流账号"体验更好
	rows, err := tx.QueryContext(ctx, `
		SELECT a.id
		FROM accounts a
		JOIN account_groups ag ON ag.account_id = a.id
		WHERE ag.group_id = $1
		  AND a.status = 'active'
		  AND a.schedulable = TRUE
		  AND a.deleted_at IS NULL
		  AND (a.rate_limit_reset_at IS NULL OR a.rate_limit_reset_at < NOW())
		  AND (a.overload_until IS NULL OR a.overload_until < NOW())
		  AND (a.temp_unschedulable_until IS NULL OR a.temp_unschedulable_until < NOW())
		  AND (a.auto_pause_on_expired = FALSE OR a.expires_at IS NULL OR a.expires_at > NOW())
		  AND NOT EXISTS (
			SELECT 1 FROM exclusive_subscriptions e
			WHERE e.account_id = a.id AND e.status = 'active' AND e.deleted_at IS NULL
		  )
		ORDER BY a.priority ASC, a.last_used_at ASC NULLS FIRST
		LIMIT 1
		FOR UPDATE OF a SKIP LOCKED
	`, groupID)
	if err != nil {
		return 0, fmt.Errorf("pick free account: %w", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return 0, ErrNoFreeAccount
	}
	var id int64
	if err := rows.Scan(&id); err != nil {
		return 0, fmt.Errorf("scan account id: %w", err)
	}
	return id, nil
}

// lockSeatRowForUpdate 在 PG 下用 SELECT ... FOR UPDATE 行锁串行化并发写者；
// 在 SQLite（unit test 环境）等不支持 FOR UPDATE 语法的 dialect 上自动降级
// 为普通 Only 查询，避免测试环境因 dialect 不兼容直接报错 SEAT_NOT_FOUND。
//
// 生产 PG 行为不变（仍然 ForUpdate 锁行），仅 SQLite 测试时退化为快照读
// （单连接 SQLite 本身没有真并发，无需 FOR UPDATE 也安全）。
func lockSeatRowForUpdate(ctx context.Context, tx *dbent.Tx, seatID int64) (*dbent.ExclusiveSubscription, error) {
	q := tx.ExclusiveSubscription.Query().Where(exclusivesubscription.IDEQ(seatID))
	if d, ok := tx.Client().Driver().(interface{ Dialect() string }); ok && d.Dialect() == dialect.Postgres {
		q = q.ForUpdate()
	}
	return q.Only(ctx)
}

// pickFreeAccountFallback 非 PG 数据库（如 SQLite 测试环境）的回退实现。
// 没有 SKIP LOCKED，并发安全靠 exclusive_subscriptions.account_id 部分唯一索引兜底。
func pickFreeAccountFallback(ctx context.Context, tx *dbent.Tx, groupID int64) (int64, error) {
	var accountIDs []int64
	err := tx.AccountGroup.Query().
		Where(accountgroup.GroupIDEQ(groupID)).
		Select(accountgroup.FieldAccountID).Scan(ctx, &accountIDs)
	if err != nil {
		return 0, fmt.Errorf("list group accounts: %w", err)
	}
	if len(accountIDs) == 0 {
		return 0, ErrNoFreeAccount
	}
	var occupiedIDs []int64
	err = tx.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.AccountIDIn(accountIDs...),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Select(exclusivesubscription.FieldAccountID).Scan(ctx, &occupiedIDs)
	if err != nil {
		return 0, fmt.Errorf("list occupied accounts: %w", err)
	}
	occupied := make(map[int64]struct{}, len(occupiedIDs))
	for _, id := range occupiedIDs {
		occupied[id] = struct{}{}
	}
	free := accountIDs[:0]
	for _, id := range accountIDs {
		if _, ok := occupied[id]; !ok {
			free = append(free, id)
		}
	}
	if len(free) == 0 {
		return 0, ErrNoFreeAccount
	}
	// 从空闲列表里挑 priority 最高（数值最小）的那个；同步排除当前不可用的账号（与 PG 路径对齐）
	now := time.Now()
	accs, err := tx.Account.Query().
		Where(
			account.IDIn(free...),
			account.StatusEQ(domain.StatusActive),
			account.SchedulableEQ(true),
			account.DeletedAtIsNil(),
			account.Or(account.RateLimitResetAtIsNil(), account.RateLimitResetAtLT(now)),
			account.Or(account.OverloadUntilIsNil(), account.OverloadUntilLT(now)),
			account.Or(account.TempUnschedulableUntilIsNil(), account.TempUnschedulableUntilLT(now)),
			// 与 pickFreeAccountForUpdate (PG 路径) 和 GetGroupInventory (库存口径) 完全对齐：
			// auto_pause_on_expired=true 且账号已过期的不能新分配，否则 fallback 路径会发出"该账号"，
			// 但 PG 路径和库存都会跳过它，造成口径分裂
			account.Or(account.AutoPauseOnExpiredEQ(false), account.ExpiresAtIsNil(), account.ExpiresAtGT(now)),
		).
		Order(account.ByPriority(), account.ByLastUsedAt()).
		Limit(1).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query free accounts: %w", err)
	}
	if len(accs) == 0 {
		return 0, ErrNoFreeAccount
	}
	return accs[0].ID, nil
}

// ListActiveByUser 查询用户在指定 group 下的所有活跃独享名额。
// 调度器进入时优先调用此方法，命中则跳过共享池负载均衡。
func (s *ExclusiveSeatService) ListActiveByUser(ctx context.Context, userID, groupID int64) ([]*dbent.ExclusiveSubscription, error) {
	now := time.Now()
	return s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.UserIDEQ(userID),
			exclusivesubscription.GroupIDEQ(groupID),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.ExpiresAtGT(now),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Order(exclusivesubscription.ByAssignedAt()).All(ctx)
}

// ListAllByUser 查询用户的所有独享名额（含历史 expired / refunded / cancelled），
// 按 assigned_at 倒序。用于用户中心「我的独享号」页。
func (s *ExclusiveSeatService) ListAllByUser(ctx context.Context, userID int64) ([]*dbent.ExclusiveSubscription, error) {
	return s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.UserIDEQ(userID),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Order(exclusivesubscription.ByAssignedAt(sql.OrderDesc())).All(ctx)
}

// GetSeat 根据 ID 取一份独享名额。
func (s *ExclusiveSeatService) GetSeat(ctx context.Context, seatID int64) (*dbent.ExclusiveSubscription, error) {
	seat, err := s.entClient.ExclusiveSubscription.Get(ctx, seatID)
	if err != nil {
		return nil, infraerrors.NotFound("SEAT_NOT_FOUND", "exclusive subscription not found")
	}
	return seat, nil
}

// IncrementSeatUsage 累加独享名额的累计 usage_usd 与日/周/月窗口用量字段。
// 调度器命中独享 seat 后产生的 cost 通过这里回写，让用户中心「累计用量」显示真实数据，
// 同时让 CheckBillingEligibility 能在请求前按窗口判断超额。
//
// 入参 seatID 为 0 时直接 noop；amount <= 0 时也 noop。失败仅记日志。
//
// 窗口策略与 UserSubscription 一致：window_start 为 nil 时初始化为 now；
// 距离当前已超过 24h / 7d / 30d 时重置（lazy reset），避免另起 cron。
//
// 并发安全（GPT round 22 #3）：原始实现"先 Get 再 Update"在两个并发请求同时跨过窗口边界时，
// 都会把窗口 SetXxxUsageUsd(amount) 重置为自己的 amount，后提交的会覆盖前一个，造成用量低计 / 限额延后触发。
// 修复：把 Get + Update 包在事务里，并用 SELECT ... FOR UPDATE 锁住该 seat 行，
// 后到的请求会等前一个事务提交、读到已重置的窗口，然后正确累加而非覆盖。
func (s *ExclusiveSeatService) IncrementSeatUsage(ctx context.Context, seatID int64, amount float64) error {
	if seatID <= 0 || amount <= 0 {
		return nil
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for seat usage increment: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	seat, err := lockSeatRowForUpdate(ctx, tx, seatID)
	if err != nil {
		return fmt.Errorf("lock seat for usage increment: %w", err)
	}
	now := time.Now()
	upd := tx.ExclusiveSubscription.UpdateOneID(seatID).
		AddUsageUsd(amount)

	// 日窗口：nil 或满 24h → 重置，否则累加
	if seat.DailyWindowStart == nil || now.Sub(*seat.DailyWindowStart) >= 24*time.Hour {
		upd = upd.SetDailyWindowStart(now).SetDailyUsageUsd(amount)
	} else {
		upd = upd.AddDailyUsageUsd(amount)
	}
	if seat.WeeklyWindowStart == nil || now.Sub(*seat.WeeklyWindowStart) >= 7*24*time.Hour {
		upd = upd.SetWeeklyWindowStart(now).SetWeeklyUsageUsd(amount)
	} else {
		upd = upd.AddWeeklyUsageUsd(amount)
	}
	if seat.MonthlyWindowStart == nil || now.Sub(*seat.MonthlyWindowStart) >= 30*24*time.Hour {
		upd = upd.SetMonthlyWindowStart(now).SetMonthlyUsageUsd(amount)
	} else {
		upd = upd.AddMonthlyUsageUsd(amount)
	}
	if err := upd.Exec(ctx); err != nil {
		return fmt.Errorf("update seat usage: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seat usage: %w", err)
	}
	committed = true
	return nil
}

// SeatWindowFallback 表示 seat 限额字段为 NULL 时回落到的 group 默认值。
// 任一字段 nil 表示该窗口不启用 group 兜底。
type SeatWindowFallback struct {
	DailyLimitUSD   *float64
	WeeklyLimitUSD  *float64
	MonthlyLimitUSD *float64
}

// LoadGroupWindowFallback 拉取 group 的 daily/weekly/monthly 限额作为 seat 兜底来源。
// 失败时返回 nil（fail-open；调用方按 seat 自带限额判断）。
// 调度器、CheckActiveSeatUsageLimit 共用，保证两处口径一致。
func (s *ExclusiveSeatService) LoadGroupWindowFallback(ctx context.Context, groupID int64) *SeatWindowFallback {
	if s == nil || s.entClient == nil || groupID <= 0 {
		return nil
	}
	g, err := s.entClient.Group.Get(ctx, groupID)
	if err != nil || g == nil {
		return nil
	}
	return &SeatWindowFallback{
		DailyLimitUSD:   g.DailyLimitUsd,
		WeeklyLimitUSD:  g.WeeklyLimitUsd,
		MonthlyLimitUSD: g.MonthlyLimitUsd,
	}
}

// effectiveSeatLimit 计算 seat 在某窗口的有效限额：seat 自身值优先，nil 时回落 fallback。
// 返回 (limit, hasLimit)；limit <= 0 视作"该窗口不启用限额"。
func effectiveSeatLimit(seatLimit *float64, fallback *float64) (float64, bool) {
	if seatLimit != nil && *seatLimit > 0 {
		return *seatLimit, true
	}
	if fallback != nil && *fallback > 0 {
		return *fallback, true
	}
	return 0, false
}

// SeatWindowExceeded 判断 seat 在 now 时刻是否有任一窗口已超额。
// 抽出来供两处使用，保证"前置 billing 检查"和"调度候选过滤"用同一份判定，
// 避免 GPT round 17 #2 的"放行的 seat 不被调度命中、被调度命中的 seat 已超额"漂移。
//
// 调度路径调用，不传 fallback。
func SeatWindowExceeded(seat *dbent.ExclusiveSubscription, now time.Time) (bool, string) {
	return SeatWindowExceededWithFallback(seat, nil, now)
}

// SeatWindowExceededWithFallback 在 SeatWindowExceeded 基础上支持 group 限额兜底。
// seat 限额字段为 NULL 时使用 fallback（来自 group.daily/weekly/monthly_limit_usd），与共享订阅
// 限额检查的 effectiveLimit 行为对齐（GPT round 22 #1）。
func SeatWindowExceededWithFallback(seat *dbent.ExclusiveSubscription, fb *SeatWindowFallback, now time.Time) (bool, string) {
	if seat == nil {
		return false, ""
	}
	var dailyFB, weeklyFB, monthlyFB *float64
	if fb != nil {
		dailyFB, weeklyFB, monthlyFB = fb.DailyLimitUSD, fb.WeeklyLimitUSD, fb.MonthlyLimitUSD
	}
	if limit, has := effectiveSeatLimit(seat.DailyLimitUsd, dailyFB); has &&
		seat.DailyWindowStart != nil && now.Sub(*seat.DailyWindowStart) < 24*time.Hour &&
		seat.DailyUsageUsd >= limit {
		return true, "daily"
	}
	if limit, has := effectiveSeatLimit(seat.WeeklyLimitUsd, weeklyFB); has &&
		seat.WeeklyWindowStart != nil && now.Sub(*seat.WeeklyWindowStart) < 7*24*time.Hour &&
		seat.WeeklyUsageUsd >= limit {
		return true, "weekly"
	}
	if limit, has := effectiveSeatLimit(seat.MonthlyLimitUsd, monthlyFB); has &&
		seat.MonthlyWindowStart != nil && now.Sub(*seat.MonthlyWindowStart) < 30*24*time.Hour &&
		seat.MonthlyUsageUsd >= limit {
		return true, "monthly"
	}
	return false, ""
}

// CheckActiveSeatUsageLimit 在请求前按 seat 的 daily/weekly/monthly_limit_usd 拦截超额。
// limit 字段 nullable，NULL = 不启用该窗口限制（回落到 group 行为，由调用方处理）。
// 返回 ErrDailyLimitExceeded / ErrWeeklyLimitExceeded / ErrMonthlyLimitExceeded 之一，
// 没有超额时返回 nil。窗口起点处理与 IncrementSeatUsage 同款 lazy reset 语义。
//
// 注意：此处只做"任一 seat 还有额度即放行"的快速通道判断；调度层 trySelectExclusiveSeatAccount
// 会再次按 SeatWindowExceeded 过滤候选，确保被命中的 seat 一定有额度（口径一致）。
func (s *ExclusiveSeatService) CheckActiveSeatUsageLimit(ctx context.Context, userID, groupID int64) error {
	if s == nil || s.entClient == nil || userID <= 0 || groupID <= 0 {
		return nil
	}
	seats, err := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.UserIDEQ(userID),
			exclusivesubscription.GroupIDEQ(groupID),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.ExpiresAtGT(time.Now()),
			exclusivesubscription.DeletedAtIsNil(),
		).All(ctx)
	if err != nil {
		return fmt.Errorf("query active seats: %w", err)
	}
	if len(seats) == 0 {
		return nil
	}
	// 拉一次 group 限额作为 seat 限额为 NULL 时的兜底（GPT round 22 #1）。
	// 失败时 fallback=nil，共享订阅分支也是这种"读不到 group 就当无限额"的语义对齐
	fallback := s.LoadGroupWindowFallback(ctx, groupID)
	now := time.Now()
	dailyExceeded, weeklyExceeded, monthlyExceeded := false, false, false
	for _, seat := range seats {
		exceeded, dim := SeatWindowExceededWithFallback(seat, fallback, now)
		if !exceeded {
			// 任一 seat 仍有额度 → 放行（DP2C：调度层挑这个还有额度的 seat）
			return nil
		}
		switch dim {
		case "daily":
			dailyExceeded = true
		case "weekly":
			weeklyExceeded = true
		case "monthly":
			monthlyExceeded = true
		}
	}
	// 全部 seat 都超额：返回最严格的那个（按业务通常优先级 daily > weekly > monthly）
	if dailyExceeded {
		return ErrDailyLimitExceeded
	}
	if weeklyExceeded {
		return ErrWeeklyLimitExceeded
	}
	if monthlyExceeded {
		return ErrMonthlyLimitExceeded
	}
	return nil
}

// FindActiveSeatByOrder 根据订单 ID 精确反查该订单分配的活跃 seat。
// 退款流程优先用此精确匹配；找不到时调用方应回退到旧的 (user_id, plan_id) 反查。
func (s *ExclusiveSeatService) FindActiveSeatByOrder(ctx context.Context, orderID int64) (*dbent.ExclusiveSubscription, error) {
	if orderID <= 0 {
		return nil, nil
	}
	seat, err := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.SourceOrderIDEQ(orderID),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		// 不存在或多条都视为"未精确命中"，让调用方回退反查
		return nil, nil
	}
	return seat, nil
}

// LastPaidPriceForSeat 查找该 seat 最近一笔已成功支付的订单金额（首购或续费的那一笔）。
// 用于续费 preview 给用户展示「上次价格 vs 本次价格」对比，方便涨/降价感知。
//
// 查找顺序：
//  1. 该 seat 的续费订单（renewal_seat_id = seatID 且 status = COMPLETED）— 最新一笔
//  2. 没续费记录则回退到首购订单（seat.source_order_id 对应的订单）
//
// 返回 0 表示没有可用历史价格（管理员手工赠送的 seat 等）。
func (s *ExclusiveSeatService) LastPaidPriceForSeat(ctx context.Context, seat *dbent.ExclusiveSubscription) (float64, error) {
	if seat == nil {
		return 0, nil
	}
	// 续费订单优先（更"近"的支付价）
	renewal, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.RenewalSeatIDEQ(seat.ID),
			paymentorder.StatusEQ(OrderStatusCompleted),
		).
		Order(paymentorder.ByCreatedAt(sql.OrderDesc())).
		Limit(1).All(ctx)
	if err != nil {
		return 0, fmt.Errorf("query renewal orders: %w", err)
	}
	if len(renewal) > 0 {
		return renewal[0].Amount, nil
	}
	// 回退到首购订单
	if seat.SourceOrderID == nil || *seat.SourceOrderID <= 0 {
		return 0, nil
	}
	first, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.IDEQ(*seat.SourceOrderID),
			paymentorder.StatusEQ(OrderStatusCompleted),
		).Only(ctx)
	if err != nil {
		// 首购单不存在或被退款都返回 0（视为无历史价）
		return 0, nil
	}
	return first.Amount, nil
}

// RenewSeat 续费一份独享名额：保留绑定账号，把 expires_at 延长。
//
// 续费规则：
//   - 名额 status=active：直接延期
//   - 已 expired 但在宽限期内（默认 7 天）：恢复为 active 并延期
//   - 已 expired 且超过宽限期 / 已 refunded / 已 cancelled：拒绝
//
// dbAccountSchedulableForSeatRenewal 判断 seat 想"恢复"的原账号是否还能正常调度。
// 复用 service.Account.IsSchedulable 的字段口径（status、schedulable、自动暂停、过载、限流、临停），
// 让"过期 seat 续费"与"实际请求调度"用同一份判定，避免续费成功后请求路径上还是不可用。
//
// 不做配额检查（quota_limit）：账号 quota 是请求时计算的运行态，续费时短暂超额属于正常波动，
// 让用户先恢复 seat、运行时由 IsSchedulable 拦截即可。
func dbAccountSchedulableForSeatRenewal(a *dbent.Account, now time.Time) bool {
	if a == nil {
		return false
	}
	if a.Status != domain.StatusActive || !a.Schedulable {
		return false
	}
	if a.AutoPauseOnExpired && a.ExpiresAt != nil && !now.Before(*a.ExpiresAt) {
		return false
	}
	if a.OverloadUntil != nil && now.Before(*a.OverloadUntil) {
		return false
	}
	if a.RateLimitResetAt != nil && now.Before(*a.RateLimitResetAt) {
		return false
	}
	if a.TempUnschedulableUntil != nil && now.Before(*a.TempUnschedulableUntil) {
		return false
	}
	return true
}

// dbAccountStillInGroup 检查账号是否仍归属于 seat 当时绑定的 group。
// 管理员把账号"移出独享池"或"换池"后，原账号即便 IsSchedulable 也不应被恢复——网关不会再把它
// 当作该 group 的可调度账号，续费成功后用户仍然用不了。
func (s *ExclusiveSeatService) dbAccountStillInGroup(ctx context.Context, accountID, groupID int64) bool {
	if s == nil || s.entClient == nil {
		return false
	}
	exists, err := s.entClient.AccountGroup.Query().
		Where(
			accountgroup.AccountIDEQ(accountID),
			accountgroup.GroupIDEQ(groupID),
		).Exist(ctx)
	if err != nil {
		return false
	}
	return exists
}

// CheckSeatRenewable 在 PreviewRenewal / CreateOrder 等"前置环节"判断该 seat 当下是否可续费。
// 与 RenewSeat 内部的状态/账号校验同款逻辑，提到外面的目的是把"用户付完款才发现不能续费"
// 变成"用户根本无法进入支付"，避免无意义的下单 → 自动退款回路。
//
//	ok = nil   该 seat 当前可续费
//	BadRequest SEAT_BEYOND_GRACE / SEAT_NOT_RENEWABLE   状态/宽限期不允许
//	Conflict   SEAT_ACCOUNT_TAKEN                       原账号已被别人占用 / 已不可调度 / 已不在原池
func (s *ExclusiveSeatService) CheckSeatRenewable(ctx context.Context, seat *dbent.ExclusiveSubscription, gracePeriod time.Duration) error {
	if seat == nil {
		return infraerrors.BadRequest("SEAT_NOT_RENEWABLE", "seat is nil")
	}
	now := time.Now()
	switch seat.Status {
	case domain.ExclusiveSeatStatusActive:
		// active 也要校验账号当下是否仍可调度（GPT round 28 #2）：
		// 管理员可能在 seat 还活跃时把账号禁用 / 移出 group / 改了反向索引指向别处。
		// 这种情况下续费成功用户也用不了，必须前置拦截避免"付完款用不了再退款"。
		return s.assertSeatAccountStillUsable(ctx, seat, now)
	case domain.ExclusiveSeatStatusExpired:
		if gracePeriod <= 0 || now.Sub(seat.ExpiresAt) > gracePeriod {
			return infraerrors.BadRequest("SEAT_BEYOND_GRACE", "renewal beyond grace period; please purchase a new plan")
		}
		return s.assertSeatAccountStillUsable(ctx, seat, now)
	default:
		return infraerrors.BadRequest("SEAT_NOT_RENEWABLE", "seat is not renewable in current status")
	}
}

// assertSeatAccountStillUsable 抽出 active / expired 分支共用的账号校验：
// 账号存在 + 反向索引未被别人占 + 仍可调度 + 仍在 seat.GroupID 内。
// 任一条不满足都返回 SEAT_ACCOUNT_TAKEN 让调用方拒绝续费。
func (s *ExclusiveSeatService) assertSeatAccountStillUsable(ctx context.Context, seat *dbent.ExclusiveSubscription, now time.Time) error {
	acc, err := s.entClient.Account.Get(ctx, seat.AccountID)
	if err != nil || acc == nil {
		return infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account is no longer available; please purchase a new plan")
	}
	// 已被别人占用：直接拒绝
	if acc.AssignedSeatID != nil && *acc.AssignedSeatID != seat.ID {
		return infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account is no longer available; please purchase a new plan")
	}
	// 调度资格：账号被禁用 / 自动暂停 / 长时间限流等场景，续费了用户也用不了
	if !dbAccountSchedulableForSeatRenewal(acc, now) {
		return infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account is no longer schedulable; please purchase a new plan")
	}
	// 账号已被移出原 group（管理员调池）：续费后网关也不会把它作为该 group 的候选
	if !s.dbAccountStillInGroup(ctx, seat.AccountID, seat.GroupID) {
		return infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account no longer belongs to the seat's group; please purchase a new plan")
	}
	return nil
}

func (s *ExclusiveSeatService) RenewSeat(ctx context.Context, seatID int64, validityDays int, gracePeriod time.Duration) (*dbent.ExclusiveSubscription, error) {
	if validityDays <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "validity_days must be > 0")
	}
	seat, err := s.GetSeat(ctx, seatID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	switch seat.Status {
	case domain.ExclusiveSeatStatusActive:
		// 履约时再校验账号当下仍可恢复（GPT round 29 #4）：
		// 预览/下单前置已经校验过，但下单到付款回调有时间窗，管理员可能在此期间禁用/移走账号。
		// 这里再做一次防 race，确保续费成功后立即可用，不会出现"付款成功但马上 503"。
		if err := s.assertSeatAccountStillUsable(ctx, seat, now); err != nil {
			return nil, err
		}
		// 续费基准：取 max(now, current expires_at)，避免缩短时长
		base := seat.ExpiresAt
		if base.Before(now) {
			base = now
		}
		newExpiresAt := base.AddDate(0, 0, validityDays)
		// 乐观锁：仅在 status 仍为 active 时更新。防御与定时任务 ReleaseSeat 的 race
		// （否则可能出现 seat.status=expired + expires_at=未来的鬼状态，账号被误回收）。
		upd := s.entClient.ExclusiveSubscription.Update().
			Where(
				exclusivesubscription.IDEQ(seatID),
				exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			).
			SetExpiresAt(newExpiresAt).
			SetLastRenewalAt(now)
		// 续费时同步刷新 plan 的限额快照（plan 限额变更后续费用户按新值）
		s.applyPlanLimitsToSeatUpdate(ctx, upd, seat.PlanID)
		n, err := upd.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("renew active seat: %w", err)
		}
		if n == 0 {
			// 状态在 select 与 update 之间被并发改动（通常是定时任务把它转 expired）
			return nil, infraerrors.Conflict("SEAT_STATUS_CHANGED", "seat status changed during renewal; please retry")
		}
		return s.GetSeat(ctx, seatID)
	case domain.ExclusiveSeatStatusExpired:
		if gracePeriod <= 0 || now.Sub(seat.ExpiresAt) > gracePeriod {
			return nil, infraerrors.BadRequest("SEAT_BEYOND_GRACE", "renewal beyond grace period; please purchase a new plan")
		}
		// 检查原账号是否还能用：未被别人占 + 仍可调度 + 仍在原 group。
		// 与 CheckSeatRenewable 同款判断（fulfillment 兜底，前置已经过滤过；这里再走一遍防 race）。
		acc, accErr := s.entClient.Account.Get(ctx, seat.AccountID)
		if accErr != nil || acc == nil {
			return nil, infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account is no longer available; please purchase a new plan")
		}
		if acc.AssignedSeatID != nil && *acc.AssignedSeatID != seatID {
			return nil, infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account is no longer available; please purchase a new plan")
		}
		if !dbAccountSchedulableForSeatRenewal(acc, now) {
			return nil, infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account is no longer schedulable; please purchase a new plan")
		}
		if !s.dbAccountStillInGroup(ctx, seat.AccountID, seat.GroupID) {
			return nil, infraerrors.Conflict("SEAT_ACCOUNT_TAKEN", "the original account no longer belongs to the seat's group; please purchase a new plan")
		}
		newExpiresAt := now.AddDate(0, 0, validityDays)
		// 恢复 active 并续期；同时把 account 反向索引指回这个 seat
		tx, txErr := s.entClient.Tx(ctx)
		if txErr != nil {
			return nil, fmt.Errorf("begin tx: %w", txErr)
		}
		updateBuilder := tx.ExclusiveSubscription.UpdateOneID(seatID).
			SetStatus(domain.ExclusiveSeatStatusActive).
			SetExpiresAt(newExpiresAt).
			SetLastRenewalAt(now)
		// expired 分支也刷新 plan 限额快照（与 active 分支行为一致），续费按当前 plan 限额
		if plan, planErr := tx.SubscriptionPlan.Get(ctx, seat.PlanID); planErr == nil && plan != nil {
			if plan.DailyLimitUsd != nil {
				updateBuilder.SetDailyLimitUsd(*plan.DailyLimitUsd)
			} else {
				updateBuilder.ClearDailyLimitUsd()
			}
			if plan.WeeklyLimitUsd != nil {
				updateBuilder.SetWeeklyLimitUsd(*plan.WeeklyLimitUsd)
			} else {
				updateBuilder.ClearWeeklyLimitUsd()
			}
			if plan.MonthlyLimitUsd != nil {
				updateBuilder.SetMonthlyLimitUsd(*plan.MonthlyLimitUsd)
			} else {
				updateBuilder.ClearMonthlyLimitUsd()
			}
			if plan.RateMultiplier != nil {
				updateBuilder.SetRateMultiplier(*plan.RateMultiplier)
			} else {
				updateBuilder.ClearRateMultiplier()
			}
		}
		updated, err := updateBuilder.Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("update seat: %w", err)
		}
		if _, err = tx.Account.UpdateOneID(seat.AccountID).SetAssignedSeatID(seatID).Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("restore account.assigned_seat_id: %w", err)
		}
		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx: %w", err)
		}
		// account 反向索引被恢复，通知调度快照刷新
		s.notifySchedulerAccountChanged(ctx, seat.AccountID)
		return updated, nil
	default:
		return nil, infraerrors.BadRequest("SEAT_NOT_RENEWABLE", "seat is not renewable in status: "+seat.Status)
	}
}

// applyPlanLimitsToSeatUpdate 续费时把当前 plan 的限额/倍率快照刷到 seat update builder。
// plan 没改限额或查询失败时静默 noop（保留原 seat 上的快照值）。
func (s *ExclusiveSeatService) applyPlanLimitsToSeatUpdate(ctx context.Context, upd *dbent.ExclusiveSubscriptionUpdate, planID int64) {
	if planID <= 0 {
		return
	}
	plan, err := s.entClient.SubscriptionPlan.Get(ctx, planID)
	if err != nil || plan == nil {
		return
	}
	if plan.DailyLimitUsd != nil {
		upd.SetDailyLimitUsd(*plan.DailyLimitUsd)
	} else {
		upd.ClearDailyLimitUsd()
	}
	if plan.WeeklyLimitUsd != nil {
		upd.SetWeeklyLimitUsd(*plan.WeeklyLimitUsd)
	} else {
		upd.ClearWeeklyLimitUsd()
	}
	if plan.MonthlyLimitUsd != nil {
		upd.SetMonthlyLimitUsd(*plan.MonthlyLimitUsd)
	} else {
		upd.ClearMonthlyLimitUsd()
	}
	if plan.RateMultiplier != nil {
		upd.SetRateMultiplier(*plan.RateMultiplier)
	} else {
		upd.ClearRateMultiplier()
	}
}

// RevokeRenewal 撤销一次续费的效果：把 seat 的 expires_at 减去 days。
// 续费订单退款时调用，避免用户白嫖续费天数。
//
// 处理细节：
//   - 仅对 active / expired 状态生效；其他状态（refunded / cancelled）拒绝
//   - 减完 expires_at <= now 时，状态转为 expired 并清空 account 反向索引
//   - 减后 expires_at 仍 > now 时，保持 active 不变
func (s *ExclusiveSeatService) RevokeRenewal(ctx context.Context, seatID int64, days int) (*dbent.ExclusiveSubscription, error) {
	if days <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "days must be > 0")
	}
	seat, err := s.GetSeat(ctx, seatID)
	if err != nil {
		return nil, err
	}
	if seat.Status != domain.ExclusiveSeatStatusActive && seat.Status != domain.ExclusiveSeatStatusExpired {
		return nil, infraerrors.BadRequest("SEAT_NOT_REVOKABLE", "seat is not revokable in status: "+seat.Status)
	}
	newExpiresAt := seat.ExpiresAt.AddDate(0, 0, -days)
	now := time.Now()
	if newExpiresAt.After(now) {
		return s.entClient.ExclusiveSubscription.UpdateOneID(seatID).SetExpiresAt(newExpiresAt).Save(ctx)
	}
	// 减完已过期：转 expired 并释放账号反向索引（事务保护）
	tx, txErr := s.entClient.Tx(ctx)
	if txErr != nil {
		return nil, fmt.Errorf("begin tx: %w", txErr)
	}
	// 事务内 ForUpdate 重新读 seat（GPT round 31 #2）：与并发 SwapSeatAccount 互斥，
	// 拿到最新的 seat.AccountID 再清反向索引；不然事务外的旧 AccountID 会清错账号，
	// 让新换绑账号的 assigned_seat_id 残留指向已 expired 的 seat、库存卡住。
	latest, lerr := lockSeatRowForUpdate(ctx, tx, seatID)
	if lerr != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("lock seat: %w", lerr)
	}
	updated, err := tx.ExclusiveSubscription.UpdateOneID(seatID).
		SetExpiresAt(newExpiresAt).
		SetStatus(domain.ExclusiveSeatStatusExpired).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("update seat: %w", err)
	}
	// 清 account 反向索引：必须 where account.assigned_seat_id = seatID 才清。
	// 否则该 seat 已过期 → 原账号又被分配给新 active seat 时，撤销旧续费会误清新 seat 的绑定
	// （GPT round 21 #1）。account 取事务内最新读到的 latest.AccountID，与 SwapSeatAccount 互斥
	if _, err = tx.Account.Update().
		Where(account.IDEQ(latest.AccountID), account.AssignedSeatIDEQ(seatID)).
		ClearAssignedSeatID().Save(ctx); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("clear account.assigned_seat_id: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	// account 被释放，通知调度快照刷新（共享池可重新调度此账号）
	s.notifySchedulerAccountChanged(ctx, latest.AccountID)
	return updated, nil
}

// ExtendSeat 直接延长 seat 的过期时间（管理员补偿用），不要求 active 状态。
// 仅对 active 状态生效；其他状态拒绝（避免对已退款/已取消的 seat 再延期造成歧义）。
func (s *ExclusiveSeatService) ExtendSeat(ctx context.Context, seatID int64, days int) (*dbent.ExclusiveSubscription, error) {
	if days == 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "days must be != 0")
	}
	seat, err := s.GetSeat(ctx, seatID)
	if err != nil {
		return nil, err
	}
	if seat.Status != domain.ExclusiveSeatStatusActive {
		return nil, infraerrors.BadRequest("SEAT_NOT_ACTIVE", "only active seats can be extended")
	}
	newExpiresAt := seat.ExpiresAt.AddDate(0, 0, days)
	if newExpiresAt.Before(time.Now()) {
		return nil, infraerrors.BadRequest("EXTENSION_NEGATIVE", "resulting expires_at would be in the past")
	}
	return s.entClient.ExclusiveSubscription.UpdateOneID(seatID).SetExpiresAt(newExpiresAt).Save(ctx)
}

// SwapSeatAccount 强制把 seat 绑定到另一个空闲账号（管理员救场）。
// 流程：原账号清反向索引 → 在同 group 池子里挑新空闲账号 → 更新 seat.account_id + 新账号反向索引。
func (s *ExclusiveSeatService) SwapSeatAccount(ctx context.Context, seatID int64) (*dbent.ExclusiveSubscription, error) {
	seat, err := s.GetSeat(ctx, seatID)
	if err != nil {
		return nil, err
	}
	if seat.Status != domain.ExclusiveSeatStatusActive {
		return nil, infraerrors.BadRequest("SEAT_NOT_ACTIVE", "only active seats can swap account")
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	newAccountID, err := pickFreeAccountForUpdate(ctx, tx, seat.GroupID)
	if err != nil {
		return nil, err
	}
	oldAccountID := seat.AccountID
	// 乐观锁（GPT round 30 #2）：事务外读取 seat 后到事务内写入有间隔，
	// 期间到期任务 / 退款 / 管理员释放可能把 seat 转为 expired/refunded/cancelled。
	// 不带 status=active 条件直接 SetAccountID 会把新账号绑到已经终态的 seat，
	// 反向索引 + 库存口径都会错乱。这里用 Update().Where(IDEQ + StatusEQ active) 并校验 affected count。
	affected, err := tx.ExclusiveSubscription.Update().
		Where(
			exclusivesubscription.IDEQ(seatID),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
		).
		SetAccountID(newAccountID).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update seat: %w", err)
	}
	if affected == 0 {
		// 在事务内才发现 seat 已不是 active：直接放弃，不要破坏 newAccountID 的绑定状态
		// （pickFreeAccountForUpdate 没修改 account；事务回滚后 newAccountID 仍是空闲）
		err = infraerrors.Conflict("SEAT_STATUS_CHANGED", "seat status changed during swap; please refresh and retry")
		return nil, err
	}
	if _, err = tx.Account.Update().
		Where(account.IDEQ(oldAccountID), account.AssignedSeatIDEQ(seatID)).
		ClearAssignedSeatID().Save(ctx); err != nil {
		return nil, fmt.Errorf("clear old account: %w", err)
	}
	if _, err = tx.Account.UpdateOneID(newAccountID).SetAssignedSeatID(seatID).Save(ctx); err != nil {
		return nil, fmt.Errorf("set new account: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	// 重新读取一次 seat 拿最新数据返回（乐观锁版本不返回 updated 实体）
	updated, err := s.entClient.ExclusiveSubscription.Get(ctx, seatID)
	if err != nil {
		return nil, fmt.Errorf("reload seat after swap: %w", err)
	}
	// 旧账号释放、新账号占用，两边都要通知调度快照刷新
	s.notifySchedulerAccountChanged(ctx, oldAccountID)
	s.notifySchedulerAccountChanged(ctx, newAccountID)
	slog.Info("[ExclusiveSeat] swapped account",
		"seat_id", seatID, "old_account", oldAccountID, "new_account", newAccountID)
	return updated, nil
}

// AdminListSeatsFilter 后台 seat 列表的过滤参数。
type AdminListSeatsFilter struct {
	UserID  int64
	GroupID int64
	Status  string
	Limit   int
	Offset  int
}

// AdminListSeats 后台 seat 列表查询：按 user_id / group_id / status 过滤。
func (s *ExclusiveSeatService) AdminListSeats(ctx context.Context, f AdminListSeatsFilter) ([]*dbent.ExclusiveSubscription, error) {
	q := s.entClient.ExclusiveSubscription.Query().
		Where(exclusivesubscription.DeletedAtIsNil())
	if f.UserID > 0 {
		q = q.Where(exclusivesubscription.UserIDEQ(f.UserID))
	}
	if f.GroupID > 0 {
		q = q.Where(exclusivesubscription.GroupIDEQ(f.GroupID))
	}
	if f.Status != "" {
		q = q.Where(exclusivesubscription.StatusEQ(f.Status))
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q = q.Order(exclusivesubscription.ByAssignedAt(sql.OrderDesc())).Limit(limit)
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	return q.All(ctx)
}

// ListByGroup 查询某独享池下的所有名额（含历史，用于后台库存看板详情）。
func (s *ExclusiveSeatService) ListByGroup(ctx context.Context, groupID int64, limit, offset int) ([]*dbent.ExclusiveSubscription, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.GroupIDEQ(groupID),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Order(exclusivesubscription.ByAssignedAt(sql.OrderDesc())).
		Limit(limit)
	if offset > 0 {
		q = q.Offset(offset)
	}
	return q.All(ctx)
}

// ReleaseSeat 释放一份独享名额：清 account.assigned_seat_id 反向索引，
// 把 seat.status 改成给定终态（expired / refunded / cancelled）。
//
// 终态语义：
//   - expired：自然到期回收
//   - refunded：因退款释放
//   - cancelled：用户/管理员主动取消
//
// 已经是终态时直接返回（幂等）。
func (s *ExclusiveSeatService) ReleaseSeat(ctx context.Context, seatID int64, terminalStatus, reason string) error {
	if terminalStatus == domain.ExclusiveSeatStatusActive {
		return infraerrors.BadRequest("INVALID_STATUS", "release status cannot be 'active'")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// SELECT ... FOR UPDATE：阻塞与 SwapSeatAccount 等并发写者，确保读到的 AccountID
	// 是事务串行化后的最新值。否则 GPT round 31 #2 场景：换绑刚把 seat.AccountID 改到 B，
	// 这里仍按事务前快照里的旧 A 清反向索引，B 上的 assigned_seat_id 残留指向已 expired 的 seat。
	seat, err := lockSeatRowForUpdate(ctx, tx, seatID)
	if err != nil {
		return infraerrors.NotFound("SEAT_NOT_FOUND", "exclusive subscription not found")
	}
	if seat.Status != domain.ExclusiveSeatStatusActive {
		// 已经是终态，幂等返回
		return tx.Commit()
	}

	notesPatch := seat.Notes
	if reason != "" {
		extra := fmt.Sprintf("[release:%s] %s", terminalStatus, reason)
		if notesPatch != nil && *notesPatch != "" {
			combined := *notesPatch + " | " + extra
			notesPatch = &combined
		} else {
			notesPatch = &extra
		}
	}

	upd := tx.ExclusiveSubscription.UpdateOneID(seatID).SetStatus(terminalStatus)
	if notesPatch != nil {
		upd.SetNotes(*notesPatch)
	}
	if _, err = upd.Save(ctx); err != nil {
		return fmt.Errorf("update seat: %w", err)
	}
	// 清 account 反向索引（仅当当前指向这个 seat 时才清，避免误清）。
	// seat.AccountID 已在事务内 ForUpdate 读到最新值，与可能并发的 SwapSeatAccount 互斥
	if _, err = tx.Account.Update().
		Where(account.IDEQ(seat.AccountID), account.AssignedSeatIDEQ(seatID)).
		ClearAssignedSeatID().Save(ctx); err != nil {
		return fmt.Errorf("clear account assigned_seat_id: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	// account 被释放，通知调度快照刷新
	s.notifySchedulerAccountChanged(ctx, seat.AccountID)
	slog.Info("[ExclusiveSeat] released",
		"seat_id", seatID, "user_id", seat.UserID, "account_id", seat.AccountID,
		"status", terminalStatus, "reason", reason)
	return nil
}

// ScanExpiredSeats 扫描自然过期的活跃独享名额（expires_at < now），返回待回收的 seat ID 列表。
// 调用方一般会循环 ReleaseSeat，与现有 account_expiry_service 的扫描风格一致。
func (s *ExclusiveSeatService) ScanExpiredSeats(ctx context.Context, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 100
	}
	now := time.Now()
	var ids []int64
	err := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.ExpiresAtLT(now),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Limit(limit).
		Select(exclusivesubscription.FieldID).Scan(ctx, &ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// RunExpiryOnce 扫描并回收所有自然过期的独享名额。
// 由后台定时任务调用（间隔通常为 5 分钟）。返回成功回收数量与第一个错误。
func (s *ExclusiveSeatService) RunExpiryOnce(ctx context.Context, batchLimit int) (int, error) {
	ids, err := s.ScanExpiredSeats(ctx, batchLimit)
	if err != nil {
		return 0, fmt.Errorf("scan expired: %w", err)
	}
	released := 0
	var firstErr error
	for _, id := range ids {
		if err := s.ReleaseSeat(ctx, id, domain.ExclusiveSeatStatusExpired, "auto expiry"); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			slog.Error("[ExclusiveSeat] expire release failed", "seat_id", id, "error", err)
			continue
		}
		released++
	}
	return released, firstErr
}

// GroupInventory 返回独享池的库存统计（总数 / 空闲 / 在用 / 即将到期 / 当下可分配）。
// 用于后台库存看板。
//
// 字段语义：
//   - Total: 池子里所有账号数（含限流/过载/临时不可用账号；不含软删账号）
//   - Used:  已被 active seat 占用的账号数
//   - Free:  Total - Used，账号当下"未被卖出"的数量（恒非负）
//   - Schedulable: 当下"立即能新分配给用户"的账号数（剔除临时不可用 + 已占用）
//   - ExpiringIn7: 7 天内即将到期的 active seat 数（提示 admin 准备库存）
type GroupInventory struct {
	Total       int
	Free        int
	Used        int
	Schedulable int
	ExpiringIn7 int
}

// GetGroupInventory 查询指定 group 的独享池库存统计。
func (s *ExclusiveSeatService) GetGroupInventory(ctx context.Context, groupID int64) (*GroupInventory, error) {
	// 该 group 下所有可调度账号
	var accountIDs []int64
	err := s.entClient.AccountGroup.Query().
		Where(accountgroup.GroupIDEQ(groupID)).
		Select(accountgroup.FieldAccountID).Scan(ctx, &accountIDs)
	if err != nil {
		return nil, err
	}
	if len(accountIDs) == 0 {
		return &GroupInventory{}, nil
	}
	now := time.Now()
	// Total: 池子里所有非删除的账号（含临时不可用），保证与 Used 同一基数，Free 恒非负
	total, err := s.entClient.Account.Query().
		Where(account.IDIn(accountIDs...), account.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	used, err := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.AccountIDIn(accountIDs...),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.DeletedAtIsNil(),
		).Count(ctx)
	if err != nil {
		return nil, err
	}
	// Schedulable: 当下立即能分配给新用户的账号数（与 pickFreeAccountForUpdate 过滤条件一致）
	// 排除临时不可用账号 + 已被占用的账号，是 admin 判断"还能卖几个"的真实指标
	var occupiedIDs []int64
	if err := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.AccountIDIn(accountIDs...),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.DeletedAtIsNil(),
		).
		Select(exclusivesubscription.FieldAccountID).Scan(ctx, &occupiedIDs); err != nil {
		return nil, fmt.Errorf("list occupied account ids: %w", err)
	}
	schedulableQ := s.entClient.Account.Query().
		Where(
			account.IDIn(accountIDs...),
			account.StatusEQ(domain.StatusActive),
			account.SchedulableEQ(true),
			account.DeletedAtIsNil(),
			account.Or(account.RateLimitResetAtIsNil(), account.RateLimitResetAtLT(now)),
			account.Or(account.OverloadUntilIsNil(), account.OverloadUntilLT(now)),
			account.Or(account.TempUnschedulableUntilIsNil(), account.TempUnschedulableUntilLT(now)),
			// 与 pickFreeAccountForUpdate 保持一致：auto_pause_on_expired=true 且账号已过期的不能新分配，
			// 否则会出现"库存显示有货 → 付款 → fulfillment 走 pickFreeAccountForUpdate 找不到 → 自动退款"
			account.Or(account.AutoPauseOnExpiredEQ(false), account.ExpiresAtIsNil(), account.ExpiresAtGT(now)),
		)
	if len(occupiedIDs) > 0 {
		schedulableQ = schedulableQ.Where(account.IDNotIn(occupiedIDs...))
	}
	schedulable, err := schedulableQ.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count schedulable: %w", err)
	}
	expiringIn7, err := s.entClient.ExclusiveSubscription.Query().
		Where(
			exclusivesubscription.AccountIDIn(accountIDs...),
			exclusivesubscription.StatusEQ(domain.ExclusiveSeatStatusActive),
			exclusivesubscription.ExpiresAtLT(now.Add(7*24*time.Hour)),
			exclusivesubscription.DeletedAtIsNil(),
		).Count(ctx)
	if err != nil {
		return nil, err
	}
	free := total - used
	if free < 0 {
		free = 0
	}
	return &GroupInventory{
		Total:       total,
		Free:        free,
		Used:        used,
		Schedulable: schedulable,
		ExpiringIn7: expiringIn7,
	}, nil
}

// formatSeatAuditFields 把 seat 的核心字段拼成 audit log 的 details map。
func formatSeatAuditFields(seat *dbent.ExclusiveSubscription) map[string]any {
	return map[string]any{
		"seat_id":    strconv.FormatInt(seat.ID, 10),
		"user_id":    strconv.FormatInt(seat.UserID, 10),
		"group_id":   strconv.FormatInt(seat.GroupID, 10),
		"plan_id":    strconv.FormatInt(seat.PlanID, 10),
		"account_id": strconv.FormatInt(seat.AccountID, 10),
		"expires_at": seat.ExpiresAt.Format(time.RFC3339),
	}
}

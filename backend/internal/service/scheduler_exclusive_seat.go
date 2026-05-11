// 独享池调度逻辑：网关在挑账号前先看用户在该 group 下是否有活跃独享名额；
// 命中则跳过共享池负载均衡，直接在用户名下的多个独享号之间负载均衡。
//
// 设计要点：
//   - 调度入口（GatewayService / OpenAIGatewayService）只调用一个公开函数 trySelectExclusiveSeatAccount
//   - 不降级到共享池：忠于"独享"语义（DP2C）。当用户名下所有独享号都不可用时返回 ErrNoUsableExclusiveAccount
//     由上层处理（重试/报错），不会悄悄退到共享池
//   - 多名额时按 priority 升序 + last_used_at 升序选择（与共享池调度一致）
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

// ErrNoUsableExclusiveAccount 用户名下所有独享名额对应的账号当前都不可用
// （被限流 / 过载 / 临时不可调度等），但用户**确实有**独享名额。
// 调度器据此报错而非降级到共享池，避免悄悄破坏「独享」承诺。
var ErrNoUsableExclusiveAccount = errors.New("scheduler: all exclusive seats currently unavailable")

// SeatAccountResolver 负责把 account ID 列表读成 *Account 实体的最小接口，
// 由两个网关 service 实现（直接复用 accountRepo）。
type SeatAccountResolver interface {
	GetByIDs(ctx context.Context, ids []int64) ([]*Account, error)
}

// trySelectExclusiveSeatAccount 网关调度入口先调本函数。
//
// 返回值：
//   - account != nil：命中独享名额，调用方应直接使用此账号，跳过共享池负载均衡
//   - account == nil, hit == false, err == nil：用户没有独享名额（或无法识别用户）→ 走原共享池路径
//   - err == ErrNoUsableExclusiveAccount：用户有独享名额但全部不可用 → 调用方应报错而非降级
//   - err != nil：内部错误，调用方按原 selectAccount 错误处理流程
func trySelectExclusiveSeatAccount(
	ctx context.Context,
	seatSvc *ExclusiveSeatService,
	resolver SeatAccountResolver,
	groupID *int64,
	excludedIDs map[int64]struct{},
) (acc *Account, hit bool, err error) {
	return trySelectExclusiveSeatAccountForModelWithCheck(ctx, seatSvc, resolver, groupID, "", excludedIDs, nil)
}

// SeatEligibilityCheck 调用方注入的账号适配性回调，返回 false 时该账号会被排除。
// 用于把共享池原有的"账号是否支持该模型 / 是否在同平台 / 是否还有配额 / OpenAI compact 支持"等
// 适配性判断同样应用到独享路径，避免把不支持该模型的账号发给上游导致 404。
//
// 判定语义：
//   - 仅限"账号是否支持本次请求"维度，不要把窗口费用、RPM、sticky 等"共享池负载均衡"逻辑塞进来
//     （独享账号是用户独占的，没有 RPM 互抢概念）
type SeatEligibilityCheck func(ctx context.Context, account *Account) bool

// trySelectExclusiveSeatAccountForModelWithCheck 在模型感知基础上再叠加调用方注入的适配性检查。
// 比如 GatewayService 传入"模型映射 + 作用域 + 平台 + 配额"，OpenAIGatewayService 传入
// "OpenAI 平台 + 模型支持 + compact 支持"。eligibilityCheck 可为 nil 表示不做额外过滤。
func trySelectExclusiveSeatAccountForModelWithCheck(
	ctx context.Context,
	seatSvc *ExclusiveSeatService,
	resolver SeatAccountResolver,
	groupID *int64,
	requestedModel string,
	excludedIDs map[int64]struct{},
	eligibilityCheck SeatEligibilityCheck,
) (acc *Account, hit bool, err error) {
	if seatSvc == nil || resolver == nil || groupID == nil {
		return nil, false, nil
	}
	userID, ok := ctx.Value(ctxkey.Sub2APIUserID).(int64)
	if !ok || userID <= 0 {
		return nil, false, nil
	}
	seats, err := seatSvc.ListActiveByUser(ctx, userID, *groupID)
	if err != nil {
		return nil, false, err
	}
	if len(seats) == 0 {
		return nil, false, nil
	}
	// 命中至少一份独享名额，从这里起就属于"独享路径"，不再降级到共享池。
	//
	// 同时按 seat 自身窗口额度过滤（GPT round 17 #2 + round 23 #4）：
	// 否则会出现 "前置 CheckActiveSeatUsageLimit 看到 seat A 还有额度放行 → 但 A 账号不可调度
	// → 实际命中已超额的 seat B"。这里把"窗口超额"也作为不可调度条件，保证最终命中的 seat
	// 一定还有额度，与前置 fast check 口径一致。
	// fallback 来自 group 默认限额，与 CheckActiveSeatUsageLimit 同源，避免"前置带 fallback 放行
	// 但调度器只看 seat 自身限额"导致的口径漂移
	now := time.Now()
	fallback := seatSvc.LoadGroupWindowFallback(ctx, *groupID)
	candidateIDs := make([]int64, 0, len(seats))
	for _, seat := range seats {
		if _, excluded := excludedIDs[seat.AccountID]; excluded {
			continue
		}
		if exceeded, _ := SeatWindowExceededWithFallback(seat, fallback, now); exceeded {
			continue
		}
		candidateIDs = append(candidateIDs, seat.AccountID)
	}
	if len(candidateIDs) == 0 {
		return nil, true, ErrNoUsableExclusiveAccount
	}
	accounts, err := resolver.GetByIDs(ctx, candidateIDs)
	if err != nil {
		return nil, true, err
	}
	usable := filterSchedulableActiveAccounts(accounts)
	// 模型级限流：跳过对 requestedModel 已限流的账号
	// （单 seat 用户没得选还是会拿到但至少 trySelect 返回错误时上层能感知；多 seat 自动切到下一个）
	if requestedModel != "" {
		filtered := usable[:0]
		for _, acc := range usable {
			if !acc.isModelRateLimitedWithContext(ctx, requestedModel) {
				filtered = append(filtered, acc)
			}
		}
		usable = filtered
	}
	// 调用方注入的适配性检查（模型映射 / 作用域 / 平台 / OpenAI compact 等）
	if eligibilityCheck != nil {
		filtered := usable[:0]
		for _, acc := range usable {
			if eligibilityCheck(ctx, acc) {
				filtered = append(filtered, acc)
			}
		}
		usable = filtered
	}
	if len(usable) == 0 {
		return nil, true, ErrNoUsableExclusiveAccount
	}
	chosen := pickByPriorityThenLastUsed(usable)
	slog.Debug("[ExclusiveSeat] selected",
		"user_id", userID, "group_id", *groupID,
		"account_id", chosen.ID, "candidates", len(usable), "model", requestedModel)
	return chosen, true, nil
}

// filterSchedulableActiveAccounts 过滤当前可用于调度的账号：
//   - 委托给 Account.IsSchedulable，与共享池调度判定标准完全一致
//     （含状态、Schedulable 标志、过期暂停、过载/限流/临时不可调度、配额超额等）
func filterSchedulableActiveAccounts(accounts []*Account) []*Account {
	out := accounts[:0]
	for _, acc := range accounts {
		if acc == nil {
			continue
		}
		if !acc.IsSchedulable() {
			continue
		}
		out = append(out, acc)
	}
	return out
}

// pickByPriorityThenLastUsed 在候选列表中按 priority 升序、last_used_at 升序挑一个。
// 与共享池调度的优先级排序保持一致，让独享池行为符合预期。
func pickByPriorityThenLastUsed(candidates []*Account) *Account {
	if len(candidates) == 0 {
		return nil
	}
	best := candidates[0]
	for _, acc := range candidates[1:] {
		if acc.Priority < best.Priority {
			best = acc
			continue
		}
		if acc.Priority > best.Priority {
			continue
		}
		// priority 相同：last_used_at 早的优先（包括 nil 视为最早）
		if compareLastUsedAt(acc, best) < 0 {
			best = acc
		}
	}
	return best
}

// compareLastUsedAt 比较两个账号的 last_used_at，nil 视作最早（应优先使用）。
// 返回 <0 表示 a 比 b 更早。
func compareLastUsedAt(a, b *Account) int {
	if a.LastUsedAt == nil && b.LastUsedAt == nil {
		return 0
	}
	if a.LastUsedAt == nil {
		return -1
	}
	if b.LastUsedAt == nil {
		return 1
	}
	if a.LastUsedAt.Before(*b.LastUsedAt) {
		return -1
	}
	if a.LastUsedAt.After(*b.LastUsedAt) {
		return 1
	}
	return 0
}

// accountRepoSeatResolver 适配 AccountRepository 到 SeatAccountResolver。
// 不暴露给外部使用，只作为两个网关 service 内部的 helper。
type accountRepoSeatResolver struct {
	repo AccountRepository
}

func (r *accountRepoSeatResolver) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return r.repo.GetByIDs(ctx, ids)
}

// 编译期断言，确保 *Account 上的若干判断方法存在（如不存在编译会失败，提示作者去 Account 上加）。
var _ = func(a *dbent.Account) bool { return a.Schedulable }
var _ = account.Label

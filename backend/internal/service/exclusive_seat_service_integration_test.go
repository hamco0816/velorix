//go:build unit

package service

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

func accountIDInForTest(ids ...int64) predicate.Account {
	return account.IDIn(ids...)
}

// 独享池端到端集成测试：用真实 SQLite + 完整 ExclusiveSeatService 链路验证
//   - AssignSeat：基础分配 + 反向索引正确
//   - 库存为 0 时 ErrNoFreeAccount
//   - 并发抢购最后一个名额：最多一人成功
//   - 续费保留绑定账号
//   - 过期回收 + 释放
//   - SwapSeatAccount 换号
//   - 通过 source_order_id 精确反查

func newSeatServiceTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	db, err := sql.Open("sqlite", "file:seat_svc_test_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// seatTestEnv 准备一个池子（group + 若干账号），返回 group_id 与 account_id 列表。
func seatTestEnv(t *testing.T, client *dbent.Client, ctx context.Context, accountCount int) (int64, []int64) {
	t.Helper()
	g, err := client.Group.Create().
		SetName("exclusive-pool-" + t.Name()).
		SetStatus(domain.StatusActive).
		SetPlatform(domain.PlatformAnthropic).
		Save(ctx)
	require.NoError(t, err)

	accountIDs := make([]int64, 0, accountCount)
	for i := 0; i < accountCount; i++ {
		acc, err := client.Account.Create().
			SetName("acc").
			SetPlatform(domain.PlatformAnthropic).
			SetType("oauth").
			SetStatus(domain.StatusActive).
			SetSchedulable(true).
			SetPriority(50).
			Save(ctx)
		require.NoError(t, err)
		_, err = client.AccountGroup.Create().
			SetAccountID(acc.ID).
			SetGroupID(int64(g.ID)).
			Save(ctx)
		require.NoError(t, err)
		accountIDs = append(accountIDs, acc.ID)
	}
	return int64(g.ID), accountIDs
}

func TestExclusiveSeat_AssignAndRelease(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, accIDs := seatTestEnv(t, client, ctx, 2)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{
		UserID: 1, GroupID: groupID, PlanID: 100,
		ValidityDays: 30, SourceOrderID: 555,
	})
	require.NoError(t, err)
	require.Equal(t, domain.ExclusiveSeatStatusActive, seat.Status)
	require.Contains(t, accIDs, seat.AccountID)
	require.NotNil(t, seat.SourceOrderID)
	require.Equal(t, int64(555), *seat.SourceOrderID)

	// 反向索引应该指向这个 seat
	acc, err := client.Account.Get(ctx, seat.AccountID)
	require.NoError(t, err)
	require.NotNil(t, acc.AssignedSeatID)
	require.Equal(t, seat.ID, *acc.AssignedSeatID)

	// 释放后反向索引清空、状态变 expired
	require.NoError(t, svc.ReleaseSeat(ctx, seat.ID, domain.ExclusiveSeatStatusExpired, "test"))
	acc2, err := client.Account.Get(ctx, seat.AccountID)
	require.NoError(t, err)
	require.Nil(t, acc2.AssignedSeatID)
	got, err := svc.GetSeat(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, domain.ExclusiveSeatStatusExpired, got.Status)
}

func TestExclusiveSeat_NoFreeAccount(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	// 把唯一空闲账号占了
	_, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30})
	require.NoError(t, err)

	// 再分配应该 ErrNoFreeAccount
	_, err = svc.AssignSeat(ctx, AssignSeatInput{UserID: 2, GroupID: groupID, PlanID: 100, ValidityDays: 30})
	require.ErrorIs(t, err, ErrNoFreeAccount)
}

func TestExclusiveSeat_ConcurrentLastSeat(t *testing.T) {
	// SQLite 的并发能力差，但通过 exclusive_subscriptions.account_id 部分唯一索引兜底
	// 仍应保证只有一人成功（不会两人都拿到同一个 account）
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, accIDs := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	const racers = 5
	var success, noStock atomic.Int32
	var wg sync.WaitGroup
	wg.Add(racers)
	for i := 0; i < racers; i++ {
		userID := int64(i + 1)
		go func() {
			defer wg.Done()
			_, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: userID, GroupID: groupID, PlanID: 100, ValidityDays: 30})
			switch {
			case err == nil:
				success.Add(1)
			case err == ErrNoFreeAccount:
				noStock.Add(1)
			}
		}()
	}
	wg.Wait()
	require.EqualValues(t, 1, success.Load(), "exactly one buyer should win")
	require.GreaterOrEqual(t, noStock.Load(), int32(0))
	// account_id 部分唯一索引保证不会有两个 active seat 占同一个账号
	count, err := client.ExclusiveSubscription.Query().Count(ctx)
	require.NoError(t, err)
	require.LessOrEqual(t, count, 1)
	require.Len(t, accIDs, 1)
}

func TestExclusiveSeat_RenewKeepsAccount(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30})
	require.NoError(t, err)
	originalAcc := seat.AccountID
	originalExpires := seat.ExpiresAt

	renewed, err := svc.RenewSeat(ctx, seat.ID, 15, 7*24*time.Hour)
	require.NoError(t, err)
	require.Equal(t, originalAcc, renewed.AccountID, "account must stay the same")
	require.True(t, renewed.ExpiresAt.After(originalExpires))
}

func TestExclusiveSeat_RenewBeyondGraceDenied(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30})
	require.NoError(t, err)

	// 模拟 8 天前已过期
	expired := time.Now().AddDate(0, 0, -8)
	_, err = client.ExclusiveSubscription.UpdateOneID(seat.ID).
		SetStatus(domain.ExclusiveSeatStatusExpired).
		SetExpiresAt(expired).Save(ctx)
	require.NoError(t, err)

	_, err = svc.RenewSeat(ctx, seat.ID, 30, 7*24*time.Hour)
	require.Error(t, err, "renewal beyond 7-day grace must be rejected")
}

func TestExclusiveSeat_RunExpiryOnce(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30})
	require.NoError(t, err)

	// 把 seat 改成已过期
	_, err = client.ExclusiveSubscription.UpdateOneID(seat.ID).
		SetExpiresAt(time.Now().Add(-time.Hour)).Save(ctx)
	require.NoError(t, err)

	released, err := svc.RunExpiryOnce(ctx, 100)
	require.NoError(t, err)
	require.Equal(t, 1, released)

	// 状态应该是 expired
	got, err := svc.GetSeat(ctx, seat.ID)
	require.NoError(t, err)
	require.Equal(t, domain.ExclusiveSeatStatusExpired, got.Status)

	// 反向索引清空
	acc, err := client.Account.Get(ctx, seat.AccountID)
	require.NoError(t, err)
	require.Nil(t, acc.AssignedSeatID)
}

func TestExclusiveSeat_SwapAccount(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, accIDs := seatTestEnv(t, client, ctx, 3)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30})
	require.NoError(t, err)
	originalAccID := seat.AccountID

	swapped, err := svc.SwapSeatAccount(ctx, seat.ID)
	require.NoError(t, err)
	require.NotEqual(t, originalAccID, swapped.AccountID)
	require.Contains(t, accIDs, swapped.AccountID)

	// 原账号反向索引清空、新账号反向索引建立
	oldAcc, err := client.Account.Get(ctx, originalAccID)
	require.NoError(t, err)
	require.Nil(t, oldAcc.AssignedSeatID)
	newAcc, err := client.Account.Get(ctx, swapped.AccountID)
	require.NoError(t, err)
	require.NotNil(t, newAcc.AssignedSeatID)
	require.Equal(t, seat.ID, *newAcc.AssignedSeatID)
}

func TestExclusiveSeat_FindActiveSeatByOrder(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 2)
	svc := NewExclusiveSeatService(client)

	seatA, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30, SourceOrderID: 1001})
	require.NoError(t, err)
	_, err = svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30, SourceOrderID: 1002})
	require.NoError(t, err)

	got, err := svc.FindActiveSeatByOrder(ctx, 1001)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, seatA.ID, got.ID)

	notFound, err := svc.FindActiveSeatByOrder(ctx, 9999)
	require.NoError(t, err)
	require.Nil(t, notFound, "non-existent order id should return nil without error")
}

// 测试 plan 有限额时被快照到 seat：分配 seat 后 seat 上的限额字段应等于 plan 当时的值。
func TestExclusiveSeat_AssignSnapshotsPlanLimits(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	// 创建一个带限额的 plan
	dailyL, weeklyL, monthlyL := 5.0, 20.0, 50.0
	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).
		SetName("plan-with-limits").
		SetPrice(99).
		SetValidityDays(30).
		SetKind(domain.PlanKindExclusive).
		SetDailyLimitUsd(dailyL).
		SetWeeklyLimitUsd(weeklyL).
		SetMonthlyLimitUsd(monthlyL).
		Save(ctx)
	require.NoError(t, err)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{
		UserID: 1, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, seat.DailyLimitUsd)
	require.Equal(t, dailyL, *seat.DailyLimitUsd, "daily limit should be snapshotted from plan")
	require.NotNil(t, seat.WeeklyLimitUsd)
	require.Equal(t, weeklyL, *seat.WeeklyLimitUsd)
	require.NotNil(t, seat.MonthlyLimitUsd)
	require.Equal(t, monthlyL, *seat.MonthlyLimitUsd)
}

// 测试 plan 没限额时 seat 限额保持 nil（→ 调度时回落到 group）
func TestExclusiveSeat_AssignNoLimitWhenPlanHasNone(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).
		SetName("plan-no-limits").
		SetPrice(50).
		SetValidityDays(30).
		SetKind(domain.PlanKindExclusive).
		Save(ctx)
	require.NoError(t, err)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{
		UserID: 1, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 101,
	})
	require.NoError(t, err)
	require.Nil(t, seat.DailyLimitUsd, "no plan limit → seat keeps nil to fallback to group")
	require.Nil(t, seat.WeeklyLimitUsd)
	require.Nil(t, seat.MonthlyLimitUsd)
}

// 多 seat 自动切换回归测试：用户买 2 个独享名额，账号 A 不可用时调度自动切到账号 B
// 验证产品语义"独享但多名额自动切换；不降级共享"
func TestExclusiveSeat_MultiSeatAutoSwitchWhenOneUnavailable(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, accIDs := seatTestEnv(t, client, ctx, 2)
	svc := NewExclusiveSeatService(client)

	// 用户买 2 个独享名额，分别绑到 accA / accB
	userID := int64(1)
	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).SetName("p").SetPrice(99).SetValidityDays(30).
		SetKind(domain.PlanKindExclusive).Save(ctx)
	require.NoError(t, err)
	seatA, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: userID, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 1})
	require.NoError(t, err)
	seatB, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: userID, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 2})
	require.NoError(t, err)
	require.NotEqual(t, seatA.AccountID, seatB.AccountID, "两份 seat 应分别绑到不同账号")
	require.Subset(t, accIDs, []int64{seatA.AccountID, seatB.AccountID})

	// 模拟 accA 限额满（schedulable=false）—— 让 trySelectExclusiveSeatAccount 应该跳过 accA 选 accB
	_, err = client.Account.UpdateOneID(seatA.AccountID).SetSchedulable(false).Save(ctx)
	require.NoError(t, err)

	resolver := &accountRepoSeatResolverForTest{client: client}
	ctxWithUser := context.WithValue(ctx, ctxkey.Sub2APIUserID, userID)
	g := groupID
	chosen, hit, err := trySelectExclusiveSeatAccount(ctxWithUser, svc, resolver, &g, nil)
	require.NoError(t, err)
	require.True(t, hit)
	require.NotNil(t, chosen)
	require.Equal(t, seatB.AccountID, chosen.ID, "accA 不可用时必须自动选 accB")

	// 反过来：accB 也不可用 → 两个都满 → 返回 ErrNoUsableExclusiveAccount（不降级到共享池）
	_, err = client.Account.UpdateOneID(seatB.AccountID).SetSchedulable(false).Save(ctx)
	require.NoError(t, err)
	chosen2, hit2, err2 := trySelectExclusiveSeatAccount(ctxWithUser, svc, resolver, &g, nil)
	require.Nil(t, chosen2)
	require.True(t, hit2, "用户有独享名额，hit=true 表示走独享路径")
	require.ErrorIs(t, err2, ErrNoUsableExclusiveAccount, "全部不可用应明确报错而非降级")
}

// 测试用 resolver：避免引入 accountRepository 完整实现
type accountRepoSeatResolverForTest struct{ client *dbent.Client }

func (r *accountRepoSeatResolverForTest) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.client.Account.Query().Where(accountIDInForTest(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Account, 0, len(rows))
	for _, m := range rows {
		out = append(out, &Account{
			ID:          m.ID,
			Name:        m.Name,
			Platform:    m.Platform,
			Type:        m.Type,
			Status:      m.Status,
			Schedulable: m.Schedulable,
			Priority:    m.Priority,
		})
	}
	return out, nil
}

// Bug E 回归：DeletePlan 必须在有 active 独享 seat 关联时拒绝，避免 seat 孤儿化
func TestDeletePlan_BlocksWhenActiveSeatsExist(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).SetName("plan-active").SetPrice(99).SetValidityDays(30).
		SetKind(domain.PlanKindExclusive).Save(ctx)
	require.NoError(t, err)
	_, err = svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 1})
	require.NoError(t, err)

	// 直接构造 PaymentConfigService 试删（不经 admin handler，避免依赖 setting repo）
	pc := &PaymentConfigService{entClient: client}
	err = pc.DeletePlan(ctx, plan.ID)
	require.Error(t, err, "should block deletion when active seats are bound")
	require.Contains(t, err.Error(), "ACTIVE_SEATS_BOUND")
}

// Bug E 续：seat 不再活跃（refunded/expired/cancelled）时允许删除
func TestDeletePlan_AllowsWhenNoActiveSeats(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).SetName("plan-released").SetPrice(99).SetValidityDays(30).
		SetKind(domain.PlanKindExclusive).Save(ctx)
	require.NoError(t, err)
	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 2})
	require.NoError(t, err)
	require.NoError(t, svc.ReleaseSeat(ctx, seat.ID, domain.ExclusiveSeatStatusRefunded, "test"))

	pc := &PaymentConfigService{entClient: client}
	require.NoError(t, pc.DeletePlan(ctx, plan.ID), "should allow delete after seat is released")
}

// Bug G 回归：RenewSeat expired 分支续费时也要刷新 plan 限额快照（与 active 分支一致）
func TestExclusiveSeat_RenewExpiredRefreshesLimitSnapshot(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	// 初始 plan：日限额 30
	daily := 30.0
	plan, err := client.SubscriptionPlan.Create().
		SetGroupID(groupID).SetName("plan").SetPrice(99).SetValidityDays(30).
		SetKind(domain.PlanKindExclusive).SetDailyLimitUsd(daily).Save(ctx)
	require.NoError(t, err)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: plan.ID, ValidityDays: 30, SourceOrderID: 1})
	require.NoError(t, err)
	require.Equal(t, 30.0, *seat.DailyLimitUsd)

	// 先把 seat 改成 expired（模拟到期）
	_, err = client.ExclusiveSubscription.UpdateOneID(seat.ID).
		SetStatus(domain.ExclusiveSeatStatusExpired).
		SetExpiresAt(time.Now().Add(-1 * time.Hour)).Save(ctx)
	require.NoError(t, err)
	// 把 account 反向索引也对齐 expired 后清理过的状态
	_, err = client.Account.UpdateOneID(seat.AccountID).ClearAssignedSeatID().Save(ctx)
	require.NoError(t, err)

	// admin 升级 plan：日限额改成 80
	newDaily := 80.0
	_, err = client.SubscriptionPlan.UpdateOneID(plan.ID).SetDailyLimitUsd(newDaily).Save(ctx)
	require.NoError(t, err)

	// 宽限期内续费 → expired 分支
	renewed, err := svc.RenewSeat(ctx, seat.ID, 30, 7*24*time.Hour)
	require.NoError(t, err)
	require.Equal(t, domain.ExclusiveSeatStatusActive, renewed.Status)
	require.NotNil(t, renewed.DailyLimitUsd, "expired-renewal must refresh limit snapshot")
	require.Equal(t, 80.0, *renewed.DailyLimitUsd, "should pick up updated plan limit, not old snapshot")
}

// RevokeRenewal 测试：续费订单退款时撤销续费天数。
func TestExclusiveSeat_RevokeRenewalShortenOnly(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	// 初始 30 天，续 30 天 → 60 天；撤销 30 天 → 回到 30 天
	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30, SourceOrderID: 1})
	require.NoError(t, err)
	renewed, err := svc.RenewSeat(ctx, seat.ID, 30, 7*24*time.Hour)
	require.NoError(t, err)

	revoked, err := svc.RevokeRenewal(ctx, seat.ID, 30)
	require.NoError(t, err)
	require.Equal(t, domain.ExclusiveSeatStatusActive, revoked.Status, "缩短后仍 > now 应保持 active")
	require.True(t, renewed.ExpiresAt.Sub(revoked.ExpiresAt) > 29*24*time.Hour, "expires_at 应被减约 30 天")
}

// RevokeRenewal 缩短超过 now：seat 转 expired 并清账号反向索引
func TestExclusiveSeat_RevokeRenewalToExpired(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30, SourceOrderID: 1})
	require.NoError(t, err)

	// 撤销 60 天（> 当前剩余 30 天）→ 应转 expired
	revoked, err := svc.RevokeRenewal(ctx, seat.ID, 60)
	require.NoError(t, err)
	require.Equal(t, domain.ExclusiveSeatStatusExpired, revoked.Status, "缩短后 expires_at < now 应转 expired")
	acc, err := client.Account.Get(ctx, seat.AccountID)
	require.NoError(t, err)
	require.Nil(t, acc.AssignedSeatID, "expired 后账号反向索引应清空")
}

// RevokeRenewal 状态校验：refunded / cancelled 状态拒绝
func TestExclusiveSeat_RevokeRenewalRejectInvalidStatus(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 1)
	svc := NewExclusiveSeatService(client)

	seat, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: 1, GroupID: groupID, PlanID: 100, ValidityDays: 30, SourceOrderID: 1})
	require.NoError(t, err)
	require.NoError(t, svc.ReleaseSeat(ctx, seat.ID, domain.ExclusiveSeatStatusRefunded, "test"))

	_, err = svc.RevokeRenewal(ctx, seat.ID, 30)
	require.Error(t, err, "refunded seat 不应允许 revoke renewal")
}

func TestExclusiveSeat_GetGroupInventory(t *testing.T) {
	client := newSeatServiceTestClient(t)
	ctx := context.Background()
	groupID, _ := seatTestEnv(t, client, ctx, 5)
	svc := NewExclusiveSeatService(client)

	inv, err := svc.GetGroupInventory(ctx, groupID)
	require.NoError(t, err)
	require.Equal(t, 5, inv.Total)
	require.Equal(t, 5, inv.Free)
	require.Equal(t, 0, inv.Used)

	for i := 0; i < 3; i++ {
		_, err := svc.AssignSeat(ctx, AssignSeatInput{UserID: int64(i + 1), GroupID: groupID, PlanID: 100, ValidityDays: 30})
		require.NoError(t, err)
	}
	inv2, err := svc.GetGroupInventory(ctx, groupID)
	require.NoError(t, err)
	require.Equal(t, 5, inv2.Total)
	require.Equal(t, 2, inv2.Free)
	require.Equal(t, 3, inv2.Used)
}

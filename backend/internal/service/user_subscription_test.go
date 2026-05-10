//go:build unit

package service

import "testing"

// ptrFloat 已在 payment_config_plans_validation_test.go（同 build tag）定义，复用

// EffectiveDailyLimit：sub 快照优先于 group
func TestUserSubscription_EffectiveDailyLimit_SubOverridesGroup(t *testing.T) {
	groupLimit := 100.0
	g := &Group{DailyLimitUSD: &groupLimit}
	sub := &UserSubscription{DailyLimitUSD: ptrFloat(30)}
	got, has := sub.EffectiveDailyLimit(g)
	if !has || got != 30 {
		t.Fatalf("expected sub override = 30, got has=%v val=%v", has, got)
	}
}

// EffectiveDailyLimit：sub 无快照时回落到 group
func TestUserSubscription_EffectiveDailyLimit_FallbackToGroup(t *testing.T) {
	groupLimit := 100.0
	g := &Group{DailyLimitUSD: &groupLimit}
	sub := &UserSubscription{}
	got, has := sub.EffectiveDailyLimit(g)
	if !has || got != 100 {
		t.Fatalf("expected fallback to group = 100, got has=%v val=%v", has, got)
	}
}

// EffectiveDailyLimit：sub 和 group 都没限额 → has=false
func TestUserSubscription_EffectiveDailyLimit_NoLimitAtAll(t *testing.T) {
	g := &Group{}
	sub := &UserSubscription{}
	got, has := sub.EffectiveDailyLimit(g)
	if has || got != 0 {
		t.Fatalf("expected no limit, got has=%v val=%v", has, got)
	}
}

// EffectiveDailyLimit：sub 快照为 0 时也应回落到 group（防御无效快照）
func TestUserSubscription_EffectiveDailyLimit_ZeroSnapshotFallsBack(t *testing.T) {
	groupLimit := 100.0
	g := &Group{DailyLimitUSD: &groupLimit}
	sub := &UserSubscription{DailyLimitUSD: ptrFloat(0)}
	got, has := sub.EffectiveDailyLimit(g)
	if !has || got != 100 {
		t.Fatalf("expected zero snapshot to fallback to group, got has=%v val=%v", has, got)
	}
}

// EffectiveRateMultiplier：sub > group > default(1.0)
func TestUserSubscription_EffectiveRateMultiplier(t *testing.T) {
	cases := []struct {
		name string
		sub  *UserSubscription
		grp  *Group
		want float64
	}{
		{"sub override", &UserSubscription{RateMultiplier: ptrFloat(2.5)}, &Group{RateMultiplier: 1.5}, 2.5},
		{"fallback group", &UserSubscription{}, &Group{RateMultiplier: 1.5}, 1.5},
		{"default 1.0", &UserSubscription{}, &Group{}, 1.0},
		{"zero sub fallbacks", &UserSubscription{RateMultiplier: ptrFloat(0)}, &Group{RateMultiplier: 1.2}, 1.2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.sub.EffectiveRateMultiplier(c.grp)
			if got != c.want {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

// CheckDailyLimit 用 effective limit 后行为：sub 30 限额 + 已用 25 + 加 6 = 31 → 拒绝
func TestUserSubscription_CheckDailyLimit_UsesEffectiveLimit(t *testing.T) {
	groupLimit := 100.0
	g := &Group{DailyLimitUSD: &groupLimit}
	sub := &UserSubscription{
		DailyLimitUSD: ptrFloat(30),
		DailyUsageUSD: 25,
	}
	if sub.CheckDailyLimit(g, 6) {
		t.Fatal("25 + 6 = 31 > sub limit 30 → expected to be rejected")
	}
	if !sub.CheckDailyLimit(g, 4) {
		t.Fatal("25 + 4 = 29 <= sub limit 30 → expected to pass")
	}
}

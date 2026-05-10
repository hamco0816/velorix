package service

import "time"

type UserSubscription struct {
	ID      int64
	UserID  int64
	GroupID int64

	StartsAt  time.Time
	ExpiresAt time.Time
	Status    string

	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time

	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64

	AssignedBy *int64
	AssignedAt time.Time
	Notes      string

	// 限额/倍率快照（migration 138）：购买时从 plan 拷贝，nil 表示回落到 group
	DailyLimitUSD   *float64
	WeeklyLimitUSD  *float64
	MonthlyLimitUSD *float64
	RateMultiplier  *float64

	CreatedAt time.Time
	UpdatedAt time.Time

	User           *User
	Group          *Group
	AssignedByUser *User
}

// EffectiveDailyLimit 返回该订阅生效的日限额：优先 sub 快照，其次 group。
// 返回 (limit, hasLimit)；hasLimit=false 表示完全不限日限额。
func (s *UserSubscription) EffectiveDailyLimit(group *Group) (float64, bool) {
	if s.DailyLimitUSD != nil && *s.DailyLimitUSD > 0 {
		return *s.DailyLimitUSD, true
	}
	if group != nil && group.HasDailyLimit() {
		return *group.DailyLimitUSD, true
	}
	return 0, false
}

// EffectiveWeeklyLimit 同 EffectiveDailyLimit，周维度
func (s *UserSubscription) EffectiveWeeklyLimit(group *Group) (float64, bool) {
	if s.WeeklyLimitUSD != nil && *s.WeeklyLimitUSD > 0 {
		return *s.WeeklyLimitUSD, true
	}
	if group != nil && group.HasWeeklyLimit() {
		return *group.WeeklyLimitUSD, true
	}
	return 0, false
}

// EffectiveMonthlyLimit 同 EffectiveDailyLimit，月维度
func (s *UserSubscription) EffectiveMonthlyLimit(group *Group) (float64, bool) {
	if s.MonthlyLimitUSD != nil && *s.MonthlyLimitUSD > 0 {
		return *s.MonthlyLimitUSD, true
	}
	if group != nil && group.HasMonthlyLimit() {
		return *group.MonthlyLimitUSD, true
	}
	return 0, false
}

// EffectiveRateMultiplier 返回该订阅生效的倍率：优先 sub 快照，其次 group，最后默认 1.0
func (s *UserSubscription) EffectiveRateMultiplier(group *Group) float64 {
	if s.RateMultiplier != nil && *s.RateMultiplier > 0 {
		return *s.RateMultiplier
	}
	if group != nil && group.RateMultiplier > 0 {
		return group.RateMultiplier
	}
	return 1.0
}

func (s *UserSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && time.Now().Before(s.ExpiresAt)
}

func (s *UserSubscription) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *UserSubscription) DaysRemaining() int {
	if s.IsExpired() {
		return 0
	}
	return int(time.Until(s.ExpiresAt).Hours() / 24)
}

func (s *UserSubscription) IsWindowActivated() bool {
	return s.DailyWindowStart != nil || s.WeeklyWindowStart != nil || s.MonthlyWindowStart != nil
}

func (s *UserSubscription) NeedsDailyReset() bool {
	if s.DailyWindowStart == nil {
		return false
	}
	return time.Since(*s.DailyWindowStart) >= 24*time.Hour
}

func (s *UserSubscription) NeedsWeeklyReset() bool {
	if s.WeeklyWindowStart == nil {
		return false
	}
	return time.Since(*s.WeeklyWindowStart) >= 7*24*time.Hour
}

func (s *UserSubscription) NeedsMonthlyReset() bool {
	if s.MonthlyWindowStart == nil {
		return false
	}
	return time.Since(*s.MonthlyWindowStart) >= 30*24*time.Hour
}

func (s *UserSubscription) DailyResetTime() *time.Time {
	if s.DailyWindowStart == nil {
		return nil
	}
	t := s.DailyWindowStart.Add(24 * time.Hour)
	return &t
}

func (s *UserSubscription) WeeklyResetTime() *time.Time {
	if s.WeeklyWindowStart == nil {
		return nil
	}
	t := s.WeeklyWindowStart.Add(7 * 24 * time.Hour)
	return &t
}

func (s *UserSubscription) MonthlyResetTime() *time.Time {
	if s.MonthlyWindowStart == nil {
		return nil
	}
	t := s.MonthlyWindowStart.Add(30 * 24 * time.Hour)
	return &t
}

func (s *UserSubscription) CheckDailyLimit(group *Group, additionalCost float64) bool {
	limit, has := s.EffectiveDailyLimit(group)
	if !has {
		return true
	}
	return s.DailyUsageUSD+additionalCost <= limit
}

func (s *UserSubscription) CheckWeeklyLimit(group *Group, additionalCost float64) bool {
	limit, has := s.EffectiveWeeklyLimit(group)
	if !has {
		return true
	}
	return s.WeeklyUsageUSD+additionalCost <= limit
}

func (s *UserSubscription) CheckMonthlyLimit(group *Group, additionalCost float64) bool {
	limit, has := s.EffectiveMonthlyLimit(group)
	if !has {
		return true
	}
	return s.MonthlyUsageUSD+additionalCost <= limit
}

func (s *UserSubscription) CheckAllLimits(group *Group, additionalCost float64) (daily, weekly, monthly bool) {
	daily = s.CheckDailyLimit(group, additionalCost)
	weekly = s.CheckWeeklyLimit(group, additionalCost)
	monthly = s.CheckMonthlyLimit(group, additionalCost)
	return
}

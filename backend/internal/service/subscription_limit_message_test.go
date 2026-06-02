package service

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 验证套餐额度超限提示文案：维度、用量/上限、重置时间、Retry-After

func TestBuildSubscriptionLimitDetail_Daily(t *testing.T) {
	now := time.Now()
	windowStart := now.Add(-2 * time.Hour) // 重置时间在约 22 小时后
	group := &Group{DailyLimitUSD: ptrFloat64(5)}
	sub := &UserSubscription{
		DailyWindowStart: ptrTime(windowStart),
		DailyUsageUSD:    5,
	}

	detail := BuildSubscriptionLimitDetail(sub, group, ErrDailyLimitExceeded)

	assert.Contains(t, detail.Message, "今日额度已用完")
	assert.Contains(t, detail.Message, "$5.00")
	assert.Contains(t, detail.Message, "重置")
	assert.Contains(t, detail.Message, "并非临时限流")
	// 重置时间约 22 小时后，Retry-After 应为正且接近该值
	require.Positive(t, detail.RetryAfterSeconds)
	assert.InDelta(t, 22*3600, detail.RetryAfterSeconds, 120)
}

func TestBuildSubscriptionLimitDetail_WeeklyAndMonthly(t *testing.T) {
	now := time.Now()
	group := &Group{
		WeeklyLimitUSD:  ptrFloat64(20),
		MonthlyLimitUSD: ptrFloat64(80),
	}
	sub := &UserSubscription{
		WeeklyWindowStart:  ptrTime(now),
		WeeklyUsageUSD:     20,
		MonthlyWindowStart: ptrTime(now),
		MonthlyUsageUSD:    80,
	}

	weekly := BuildSubscriptionLimitDetail(sub, group, ErrWeeklyLimitExceeded)
	assert.Contains(t, weekly.Message, "本周额度已用完")
	assert.Contains(t, weekly.Message, "$20.00")

	monthly := BuildSubscriptionLimitDetail(sub, group, ErrMonthlyLimitExceeded)
	assert.Contains(t, monthly.Message, "本月额度已用完")
	assert.Contains(t, monthly.Message, "$80.00")
}

func TestBuildSubscriptionLimitDetail_NilSubscription(t *testing.T) {
	detail := BuildSubscriptionLimitDetail(nil, nil, ErrDailyLimitExceeded)
	assert.Contains(t, detail.Message, "额度已用完")
	assert.Equal(t, 0, detail.RetryAfterSeconds)
}

func TestIsUsageLimitError(t *testing.T) {
	assert.True(t, IsUsageLimitError(ErrDailyLimitExceeded))
	assert.True(t, IsUsageLimitError(ErrWeeklyLimitExceeded))
	assert.True(t, IsUsageLimitError(ErrMonthlyLimitExceeded))
	assert.False(t, IsUsageLimitError(ErrSubscriptionExpired))
	assert.False(t, IsUsageLimitError(nil))
}

func TestHumanizeResetDuration(t *testing.T) {
	assert.True(t, strings.Contains(humanizeResetDuration(3*24*3600), "天"))
	assert.True(t, strings.Contains(humanizeResetDuration(5*3600), "小时"))
	assert.True(t, strings.Contains(humanizeResetDuration(10*60), "分钟"))
	assert.Equal(t, "不到 1 分钟", humanizeResetDuration(30))
}

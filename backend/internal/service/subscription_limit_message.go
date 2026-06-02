package service

// 订阅套餐额度超限时，向网关客户端返回友好提示的文案构造逻辑。
// 目的：客户达到日/周/月限额时能明确知道"是自己套餐额度用完了"，而不是被笼统的 429 当成临时限流反复重试。

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// 提示文案中金额按美元保留两位小数展示
const limitMessageMoneyFormat = "$%.2f"

// SubscriptionLimitDetail 订阅额度超限的展示信息
type SubscriptionLimitDetail struct {
	Message           string // 面向客户的中文提示
	RetryAfterSeconds int    // 距离额度重置的秒数，用于 Retry-After 响应头（0 表示无明确重置时间）
}

// IsUsageLimitError 判断错误是否为订阅用量超限（日/周/月任一）
func IsUsageLimitError(err error) bool {
	return errors.Is(err, ErrDailyLimitExceeded) ||
		errors.Is(err, ErrWeeklyLimitExceeded) ||
		errors.Is(err, ErrMonthlyLimitExceeded)
}

// BuildSubscriptionLimitDetail 根据超限维度构造面向客户的提示与重试时间。
// err 应为 ErrDailyLimitExceeded / ErrWeeklyLimitExceeded / ErrMonthlyLimitExceeded 之一。
func BuildSubscriptionLimitDetail(sub *UserSubscription, group *Group, err error) SubscriptionLimitDetail {
	if sub == nil {
		return SubscriptionLimitDetail{
			Message: "您的订阅套餐额度已用完，请等待额度刷新后再试或升级套餐。",
		}
	}

	// 按超限维度取对应的用量、上限和重置时间
	dimensionName, used, limit, resetAt := resolveLimitDimension(sub, group, err)

	message := fmt.Sprintf("您的订阅套餐%s额度已用完", dimensionName)
	if limit > 0 {
		message += fmt.Sprintf("（已用 "+limitMessageMoneyFormat+" / 上限 "+limitMessageMoneyFormat+"）", used, limit)
	}

	retryAfter := 0
	if resetAt != nil {
		retryAfter = secondsUntil(*resetAt)
		message += fmt.Sprintf("，将于约 %s后（%s）重置", humanizeResetDuration(retryAfter), resetAt.Format("01-02 15:04 MST"))
	}

	message += "。这是套餐用量限制，并非临时限流，请等待额度刷新后再试或升级套餐。"

	return SubscriptionLimitDetail{Message: message, RetryAfterSeconds: retryAfter}
}

// resolveLimitDimension 根据超限错误返回维度名称、已用量、上限和重置时间
func resolveLimitDimension(sub *UserSubscription, group *Group, err error) (name string, used, limit float64, resetAt *time.Time) {
	switch {
	case errors.Is(err, ErrWeeklyLimitExceeded):
		limit, _ = sub.EffectiveWeeklyLimit(group)
		return "本周", sub.WeeklyUsageUSD, limit, sub.WeeklyResetTime()
	case errors.Is(err, ErrMonthlyLimitExceeded):
		limit, _ = sub.EffectiveMonthlyLimit(group)
		return "本月", sub.MonthlyUsageUSD, limit, sub.MonthlyResetTime()
	default:
		limit, _ = sub.EffectiveDailyLimit(group)
		return "今日", sub.DailyUsageUSD, limit, sub.DailyResetTime()
	}
}

// secondsUntil 返回距离目标时间的秒数，向上取整，最小为 1
func secondsUntil(t time.Time) int {
	secs := int(math.Ceil(time.Until(t).Seconds()))
	if secs < 1 {
		secs = 1
	}
	return secs
}

// humanizeResetDuration 把剩余秒数转成中文的粗粒度描述（天/小时/分钟）
func humanizeResetDuration(seconds int) string {
	switch {
	case seconds >= 24*3600:
		return fmt.Sprintf("%d 天", int(math.Round(float64(seconds)/(24*3600))))
	case seconds >= 3600:
		return fmt.Sprintf("%d 小时", int(math.Round(float64(seconds)/3600)))
	case seconds >= 60:
		return fmt.Sprintf("%d 分钟", int(math.Round(float64(seconds)/60)))
	default:
		return "不到 1 分钟"
	}
}

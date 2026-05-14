package service

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// PricingAdvisorService 提供"订阅定价助手"所需的统计聚合：
// - 按 (platform, subscription_tier) 聚合 5h / 7d 滚动窗口的实际成本
// - 输出 avg / P50 / P95 / max 分位，作为订阅档位定额的数据参考
// - 返回 30 天 daily cost 趋势，给前端画稳定性曲线
//
// 所有金额都以"账号成本（total_cost）"为准 —— 也就是上游实际扣费金额（USD），不应用
// 用户分组倍率。定价决策的输入是「我每月在上游花了多少」，跟向下游用户加价无关。
type PricingAdvisorService struct {
	db *sql.DB
}

// NewPricingAdvisorService 构造函数。
func NewPricingAdvisorService(db *sql.DB) *PricingAdvisorService {
	return &PricingAdvisorService{db: db}
}

// TierStats 单个 (platform, tier) 分位统计。
type TierStats struct {
	Platform         string  `json:"platform"`
	Tier             string  `json:"tier"`            // 空字符串表示未分类
	SampleAccounts   int     `json:"sample_accounts"` // 样本账号数
	Window5hAvg      float64 `json:"window_5h_avg"`   // 每账号过去 N 天 5h 滑窗成本峰值的均值（USD）
	Window5hP50      float64 `json:"window_5h_p50"`
	Window5hP95      float64 `json:"window_5h_p95"`
	Window5hMax      float64 `json:"window_5h_max"`
	Window7dAvg      float64 `json:"window_7d_avg"`
	Window7dP50      float64 `json:"window_7d_p50"`
	Window7dP95      float64 `json:"window_7d_p95"`
	Window7dMax      float64 `json:"window_7d_max"`
	Daily30dAvg      float64 `json:"daily_30d_avg"` // 30 天日均（按账号平均后取均值）
	HasEnoughSamples bool    `json:"has_enough_samples"`

	// 反推得到的"上游档位 cap"（USD）：基于账号 (当前窗口已耗费, 上游 utilization%) 反算 cap = cost / util。
	// 取该档位下所有有效样本的中位数。0 表示该档位下没有可用 utilization 采样，前端会回退用 7d_p95 估计。
	Cap5hUsd       float64 `json:"cap_5h_usd"`
	Cap7dUsd       float64 `json:"cap_7d_usd"`
	CapSampleCount int     `json:"cap_sample_count"` // 参与 cap 反推的账号数（util > 5% 才算有效）
}

// TrendPoint 单个时间点的趋势数据。
type TrendPoint struct {
	Date      string  `json:"date"`        // YYYY-MM-DD
	Cost      float64 `json:"cost"`        // 该日所有该 tier 账号的总成本
	Accounts  int     `json:"accounts"`    // 该日活跃账号数
	AvgPerAcc float64 `json:"avg_per_acc"` // 平均到每账号的日成本
}

// TierTrend 单个 (platform, tier) 的 30 天趋势。
type TierTrend struct {
	Platform string       `json:"platform"`
	Tier     string       `json:"tier"`
	Points   []TrendPoint `json:"points"`
}

// PricingAdvisorParams 查询参数。
type PricingAdvisorParams struct {
	DaysWindow int    // 默认 30
	Platform   string // 空表示全部
}

type hourBucket struct {
	AccountID int64
	Hour      time.Time
	Cost      float64
}

// GetTierStats 按 (platform, tier) 聚合滚动窗口成本统计。
//
// 算法思路：
// 1. 拉取过去 N 天的所有 usage_log，按 account_id × hour 累加 total_cost
// 2. 对每个账号：按小时构造序列后用滑窗算 5h / 7d 窗口的峰值成本
// 3. 按 (platform, tier) 分组，对组内账号的峰值算 avg/P50/P95/max
//
// 用每个账号的"峰值窗口"作样本，避免被低峰期数据稀释，得到的是高峰用量上限参考。
func (s *PricingAdvisorService) GetTierStats(ctx context.Context, params PricingAdvisorParams) ([]TierStats, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("pricing advisor service unavailable")
	}
	days := params.DaysWindow
	if days <= 0 {
		days = 30
	}
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	rows, err := s.db.QueryContext(ctx, `
		SELECT u.account_id,
			   a.platform,
			   COALESCE(a.subscription_tier, '') AS tier,
			   date_trunc('hour', u.created_at) AS hour,
			   COALESCE(SUM(u.total_cost), 0) AS cost
		FROM usage_logs u
		JOIN accounts a ON a.id = u.account_id
		WHERE u.created_at >= $1
		  AND ($2::text = '' OR a.platform = $2)
		GROUP BY u.account_id, a.platform, a.subscription_tier, hour
		ORDER BY u.account_id, hour
	`, since, params.Platform)
	if err != nil {
		return nil, fmt.Errorf("query hourly buckets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type accountMeta struct {
		Platform string
		Tier     string
	}
	accountMetaMap := make(map[int64]accountMeta)
	byAccount := make(map[int64][]hourBucket)

	for rows.Next() {
		var (
			accID    int64
			platform string
			tier     string
			hour     time.Time
			cost     float64
		)
		if err := rows.Scan(&accID, &platform, &tier, &hour, &cost); err != nil {
			return nil, fmt.Errorf("scan bucket: %w", err)
		}
		accountMetaMap[accID] = accountMeta{Platform: platform, Tier: tier}
		byAccount[accID] = append(byAccount[accID], hourBucket{AccountID: accID, Hour: hour, Cost: cost})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iter: %w", err)
	}

	type accStat struct {
		Peak5h   float64
		Peak7d   float64
		Total30d float64
		Platform string
		Tier     string
		Cap5h    float64 // 反推 cap：>0 才参与中位数
		Cap7d    float64
	}

	// 拉每个账号的上游窗口反推所需数据（util + 当前窗口已消耗）。
	// util 来自上游 /usage 被动采样（缓存在 accounts.extra），cost 直接从 usage_logs 的对应窗口求和。
	caps5h, caps7d, capErr := s.collectAccountCaps(ctx, byAccount, params.Platform)
	if capErr != nil {
		// cap 反推失败不阻塞主流程：缺失就走老路（前端回退到 7d_p95 估计）
		caps5h = make(map[int64]float64)
		caps7d = make(map[int64]float64)
	}

	accStats := make([]accStat, 0, len(byAccount))
	for accID, bs := range byAccount {
		meta := accountMetaMap[accID]
		peak5h := slidingWindowPeak(bs, 5)
		peak7d := slidingWindowPeak(bs, 24*7)
		total := 0.0
		for _, b := range bs {
			total += b.Cost
		}
		accStats = append(accStats, accStat{
			Peak5h:   peak5h,
			Peak7d:   peak7d,
			Total30d: total,
			Platform: meta.Platform,
			Tier:     meta.Tier,
			Cap5h:    caps5h[accID],
			Cap7d:    caps7d[accID],
		})
	}

	type groupKey struct{ Platform, Tier string }
	groups := make(map[groupKey][]accStat)
	for _, st := range accStats {
		k := groupKey{Platform: st.Platform, Tier: st.Tier}
		groups[k] = append(groups[k], st)
	}

	result := make([]TierStats, 0, len(groups))
	for k, list := range groups {
		w5 := make([]float64, len(list))
		w7 := make([]float64, len(list))
		// cap 反推的样本：跳过 0 值（无有效 util 采样）
		validCap5h := make([]float64, 0, len(list))
		validCap7d := make([]float64, 0, len(list))
		var total30 float64
		for i, st := range list {
			w5[i] = st.Peak5h
			w7[i] = st.Peak7d
			total30 += st.Total30d
			if st.Cap5h > 0 {
				validCap5h = append(validCap5h, st.Cap5h)
			}
			if st.Cap7d > 0 {
				validCap7d = append(validCap7d, st.Cap7d)
			}
		}
		// cap 用中位数：避免单个异常账号（util 接近 0、cost 也很小，反推值噪声大）拉偏均值
		capSampleCount := len(validCap7d)
		if len(validCap5h) > capSampleCount {
			capSampleCount = len(validCap5h)
		}
		stats := TierStats{
			Platform:         k.Platform,
			Tier:             k.Tier,
			SampleAccounts:   len(list),
			Window5hAvg:      avgFloat(w5),
			Window5hP50:      percentile(w5, 0.50),
			Window5hP95:      percentile(w5, 0.95),
			Window5hMax:      maxFloat(w5),
			Window7dAvg:      avgFloat(w7),
			Window7dP50:      percentile(w7, 0.50),
			Window7dP95:      percentile(w7, 0.95),
			Window7dMax:      maxFloat(w7),
			Daily30dAvg:      total30 / float64(len(list)) / float64(days),
			HasEnoughSamples: len(list) >= 3,
			Cap5hUsd:         percentile(validCap5h, 0.50),
			Cap7dUsd:         percentile(validCap7d, 0.50),
			CapSampleCount:   capSampleCount,
		}
		result = append(result, stats)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Platform != result[j].Platform {
			return result[i].Platform < result[j].Platform
		}
		return result[i].Tier < result[j].Tier
	})
	return result, nil
}

// GetTierTrend 返回过去 N 天每天每个 (platform, tier) 的总成本曲线，供前端画稳定性图。
func (s *PricingAdvisorService) GetTierTrend(ctx context.Context, params PricingAdvisorParams) ([]TierTrend, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("pricing advisor service unavailable")
	}
	days := params.DaysWindow
	if days <= 0 {
		days = 30
	}
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	rows, err := s.db.QueryContext(ctx, `
		SELECT a.platform,
			   COALESCE(a.subscription_tier, '') AS tier,
			   date_trunc('day', u.created_at)::date AS day,
			   COALESCE(SUM(u.total_cost), 0) AS cost,
			   COUNT(DISTINCT u.account_id) AS accounts
		FROM usage_logs u
		JOIN accounts a ON a.id = u.account_id
		WHERE u.created_at >= $1
		  AND ($2::text = '' OR a.platform = $2)
		GROUP BY a.platform, a.subscription_tier, day
		ORDER BY a.platform, a.subscription_tier, day
	`, since, params.Platform)
	if err != nil {
		return nil, fmt.Errorf("query trend: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type key struct{ Platform, Tier string }
	trendMap := make(map[key][]TrendPoint)
	for rows.Next() {
		var (
			platform string
			tier     string
			day      time.Time
			cost     float64
			accs     int
		)
		if err := rows.Scan(&platform, &tier, &day, &cost, &accs); err != nil {
			return nil, fmt.Errorf("scan trend: %w", err)
		}
		k := key{Platform: platform, Tier: tier}
		avgPerAcc := 0.0
		if accs > 0 {
			avgPerAcc = cost / float64(accs)
		}
		trendMap[k] = append(trendMap[k], TrendPoint{
			Date:      day.Format("2006-01-02"),
			Cost:      cost,
			Accounts:  accs,
			AvgPerAcc: avgPerAcc,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iter: %w", err)
	}

	result := make([]TierTrend, 0, len(trendMap))
	for k, points := range trendMap {
		result = append(result, TierTrend{Platform: k.Platform, Tier: k.Tier, Points: points})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Platform != result[j].Platform {
			return result[i].Platform < result[j].Platform
		}
		return result[i].Tier < result[j].Tier
	})
	return result, nil
}

// slidingWindowPeak 在按小时升序的 buckets 上找出 windowHours 小时滑窗的最大成本和。
// 如果数据少于窗口长度，返回所有数据之和（避免短数据被低估）。
func slidingWindowPeak(buckets []hourBucket, windowHours int) float64 {
	if len(buckets) == 0 {
		return 0
	}
	start := buckets[0].Hour
	end := buckets[len(buckets)-1].Hour
	totalHours := int(end.Sub(start).Hours()) + 1
	if totalHours < windowHours {
		var sum float64
		for _, b := range buckets {
			sum += b.Cost
		}
		return sum
	}
	hourly := make([]float64, totalHours)
	for _, b := range buckets {
		idx := int(b.Hour.Sub(start).Hours())
		if idx >= 0 && idx < totalHours {
			hourly[idx] += b.Cost
		}
	}
	var windowSum float64
	for i := 0; i < windowHours && i < totalHours; i++ {
		windowSum += hourly[i]
	}
	peak := windowSum
	for i := windowHours; i < totalHours; i++ {
		windowSum += hourly[i] - hourly[i-windowHours]
		if windowSum > peak {
			peak = windowSum
		}
	}
	return peak
}

func avgFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := make([]float64, len(xs))
	copy(cp, xs)
	sort.Float64s(cp)
	n := len(cp)
	// 单元素直接返回，避免 P95 被 int() 截断到下标 0
	if n == 1 {
		return cp[0]
	}
	if p <= 0 {
		return cp[0]
	}
	if p >= 1 {
		return cp[n-1]
	}
	// 线性插值：例如 n=2 + P95，rank=0.95，floor=0，frac=0.95 → cp[0]*0.05 + cp[1]*0.95
	// 比 int() 截断更稳，对样本少时 P95 也能接近最大值
	rank := p * float64(n-1)
	floor := int(rank)
	if floor >= n-1 {
		return cp[n-1]
	}
	frac := rank - float64(floor)
	return cp[floor]*(1-frac) + cp[floor+1]*frac
}

func maxFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// collectAccountCaps 反推每个账号的上游 cap（USD）：cap = 当前窗口已消耗 / utilization。
//
// utilization 来自 accounts.extra 里的被动采样（session_window_utilization、passive_usage_7d_utilization），
// 当前窗口已消耗直接从 usage_logs 按时间求和。
//
// 算法要点：
//   - 只对 util ≥ 0.05 的账号反推：util 太小（接近 0）反推值噪声极大，例如 util=0.01,cost=0.5
//     得出 cap=$50，但其实可能是新窗口刚 reset 才用了 0.5 美元，cap 真值未知
//   - 5h 窗口直接用 now-5h 到 now 的 cost；7d 窗口同理用 now-7d 到 now
//   - 严格说应按 reset_at 反推窗口起点，但被动采样有滞后，固定回看反而更稳
//
// 返回 (cap5h_per_account, cap7d_per_account, error)。错误时上层会把整张 cap 表当空处理。
func (s *PricingAdvisorService) collectAccountCaps(ctx context.Context, byAccount map[int64][]hourBucket, platformFilter string) (map[int64]float64, map[int64]float64, error) {
	cap5h := make(map[int64]float64, len(byAccount))
	cap7d := make(map[int64]float64, len(byAccount))
	if len(byAccount) == 0 {
		return cap5h, cap7d, nil
	}

	now := time.Now()
	since5h := now.Add(-5 * time.Hour)
	since7d := now.Add(-7 * 24 * time.Hour)

	// 一次拉所有账号的 util + 5h/7d 窗口实际消耗。把窗口求和放进 SQL 比 Go 侧再循环一遍 buckets 更省。
	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id,
			   COALESCE((a.extra->>'session_window_utilization')::float, 0)     AS util_5h,
			   COALESCE((a.extra->>'passive_usage_7d_utilization')::float, 0)   AS util_7d,
			   COALESCE((SELECT SUM(total_cost) FROM usage_logs
						 WHERE account_id = a.id AND created_at >= $1), 0)      AS cost_5h,
			   COALESCE((SELECT SUM(total_cost) FROM usage_logs
						 WHERE account_id = a.id AND created_at >= $2), 0)      AS cost_7d
		FROM accounts a
		WHERE a.deleted_at IS NULL
		  AND ($3::text = '' OR a.platform = $3)
	`, since5h, since7d, platformFilter)
	if err != nil {
		return cap5h, cap7d, fmt.Errorf("query account caps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	const minValidUtil = 0.05 // util < 5% 不参与反推（噪声太大）
	for rows.Next() {
		var (
			accID          int64
			util5h, util7d float64
			cost5h, cost7d float64
		)
		if err := rows.Scan(&accID, &util5h, &util7d, &cost5h, &cost7d); err != nil {
			return cap5h, cap7d, fmt.Errorf("scan account caps: %w", err)
		}
		// 只关心当前 30 天窗口里有 usage 的账号；其它账号不在 byAccount 里就忽略
		if _, ok := byAccount[accID]; !ok {
			continue
		}
		if util5h >= minValidUtil && cost5h > 0 {
			cap5h[accID] = cost5h / util5h
		}
		if util7d >= minValidUtil && cost7d > 0 {
			cap7d[accID] = cost7d / util7d
		}
	}
	if err := rows.Err(); err != nil {
		return cap5h, cap7d, fmt.Errorf("rows iter caps: %w", err)
	}
	return cap5h, cap7d, nil
}

/**
 * 订阅定价助手 API
 *
 * 后端接口提供两个数据视图：
 * 1. tier-stats：按 (platform, tier) 聚合的 5h/7d 滚动窗口峰值成本统计（avg/P50/P95/max）
 * 2. tier-trend：按 (platform, tier) 的过去 N 天每日成本曲线
 *
 * 所有金额都以"账号成本（usage_log.total_cost）"为准，即上游实际扣费（USD）。
 */

import { apiClient } from '../client'

export interface TierStats {
  platform: string
  tier: string // 空字符串 = 未分类
  sample_accounts: number
  window_5h_avg: number
  window_5h_p50: number
  window_5h_p95: number
  window_5h_max: number
  window_7d_avg: number
  window_7d_p50: number
  window_7d_p95: number
  window_7d_max: number
  daily_30d_avg: number
  has_enough_samples: boolean

  // 反推得到的"上游档位 cap"（USD）：cap = 当前窗口已耗费 / utilization。0 表示无可用样本，前端会回退用 7d_p95。
  cap_5h_usd: number
  cap_7d_usd: number
  cap_sample_count: number // 参与 cap 反推的账号数（util > 5% 才算有效）
}

export interface TrendPoint {
  date: string // YYYY-MM-DD
  cost: number
  accounts: number
  avg_per_acc: number
}

export interface TierTrend {
  platform: string
  tier: string
  points: TrendPoint[]
}

export interface PricingAdvisorParams {
  days?: number
  platform?: string
}

interface TierStatsResponse {
  items: TierStats[]
  days_window: number
}

interface TierTrendResponse {
  items: TierTrend[]
  days_window: number
}

export const pricingAdvisorAPI = {
  async tierStats(params: PricingAdvisorParams = {}): Promise<TierStatsResponse> {
    const { data } = await apiClient.get<TierStatsResponse>('/admin/pricing-advisor/tier-stats', {
      params,
    })
    return data
  },
  async tierTrend(params: PricingAdvisorParams = {}): Promise<TierTrendResponse> {
    const { data } = await apiClient.get<TierTrendResponse>('/admin/pricing-advisor/tier-trend', {
      params,
    })
    return data
  },
}

export default pricingAdvisorAPI

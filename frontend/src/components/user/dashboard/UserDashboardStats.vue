<!--
  仪表盘统计区：账户英雄卡（余额/充值/订阅配额）+ 核心 KPI 卡 + Token/性能指标卡。
  数据系列配色收敛为品牌组合：brand 橙为主、emerald 表成本/积极、中性灰辅助。
-->
<template>
  <div class="space-y-8">
    <!-- 账户英雄区：简单模式（无计费）不展示余额 -->
    <UserDashboardHero
      v-if="!isSimple"
      :balance="balance"
      :today-cost="stats?.today_actual_cost || 0"
      :today-requests="stats?.today_requests || 0"
    />

    <!-- 核心数据：今日请求 / 今日消费 / API 密钥 -->
    <section class="space-y-4">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">
        {{ t('dashboard.coreStats') }}
      </h2>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <!-- Today Requests：品牌橙 sparkline + trend -->
        <div class="kpi-card">
          <div class="kpi-card-header">
            <div class="metric-icon metric-icon-brand">
              <Icon name="chart" size="sm" :stroke-width="1.75" />
            </div>
            <TrendChip :value="requestsTrend" />
          </div>
          <p class="kpi-card-label">{{ t('dashboard.todayRequests') }}</p>
          <p class="kpi-card-value">{{ formatNumber(stats?.today_requests || 0) }}</p>
          <p class="kpi-card-hint">
            {{ t('common.total') }}
            <span class="font-medium tabular-nums text-gray-700 dark:text-gray-300">{{ formatNumber(stats?.total_requests || 0) }}</span>
          </p>
          <div v-if="hasSpark(requestsSeries)" class="kpi-card-spark">
            <SparklineMini :data="requestsSeries" :color="SERIES_COLORS.brand" :height="40" />
          </div>
        </div>

        <!-- Today Cost：emerald 表成本 -->
        <div class="kpi-card">
          <div class="kpi-card-header">
            <div class="metric-icon metric-icon-emerald">
              <Icon name="dollar" size="sm" :stroke-width="1.75" />
            </div>
            <TrendChip :value="costTrend" />
          </div>
          <p class="kpi-card-label">{{ t('dashboard.todayCost') }}</p>
          <p class="kpi-card-value">${{ formatCost(stats?.today_actual_cost || 0) }}</p>
          <p class="kpi-card-hint tabular-nums">
            <span :title="t('dashboard.standard')">${{ formatCost(stats?.today_cost || 0) }}</span>
            <span class="ml-1 text-gray-400 dark:text-dark-500">{{ t('dashboard.standard') }}</span>
          </p>
          <div v-if="hasSpark(costSeries)" class="kpi-card-spark">
            <SparklineMini :data="costSeries" :color="SERIES_COLORS.emerald" :height="40" />
          </div>
        </div>

        <!-- API Keys：当前快照值，无 sparkline -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-neutral">
            <Icon name="key" size="sm" :stroke-width="1.75" />
          </div>
          <div class="metric-body">
            <p class="metric-label">{{ t('dashboard.apiKeys') }}</p>
            <p class="metric-value">{{ stats?.total_api_keys || 0 }}</p>
            <p class="metric-hint">
              <span class="inline-flex items-center gap-1">
                <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                <span class="font-medium tabular-nums text-emerald-600 dark:text-emerald-400">{{ stats?.active_api_keys || 0 }}</span>
                {{ t('common.active') }}
              </span>
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Token / 性能数据 -->
    <section class="space-y-4">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">
        {{ t('dashboard.tokenStats') }}
      </h2>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Today Tokens：带 sparkline -->
        <div class="kpi-card">
          <div class="kpi-card-header">
            <div class="metric-icon metric-icon-brand">
              <Icon name="cube" size="sm" :stroke-width="1.75" />
            </div>
            <TrendChip :value="tokensTrend" />
          </div>
          <p class="kpi-card-label">{{ t('dashboard.todayTokens') }}</p>
          <p class="kpi-card-value">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <!-- 输入/输出 + 缓存读取/创建 分两行展示，保证四项加起来等于 total_tokens；每项 title 释义图标含义 -->
          <div class="kpi-card-hint space-y-0.5 tabular-nums">
            <p>
              <span :title="t('usage.tokenIconHint.input')" class="cursor-help">
                <Icon name="arrowDown" size="xs" class="mr-0.5 inline-block text-emerald-500" />
                <span>{{ formatTokens(stats?.today_input_tokens || 0) }}</span>
              </span>
              <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
              <span :title="t('usage.tokenIconHint.output')" class="cursor-help">
                <Icon name="arrowUp" size="xs" class="mr-0.5 inline-block text-brand-500" />
                <span>{{ formatTokens(stats?.today_output_tokens || 0) }}</span>
              </span>
            </p>
            <p v-if="hasCacheToday">
              <span :title="t('usage.tokenIconHint.cacheRead')" class="cursor-help">
                <Icon name="inbox" size="xs" class="mr-0.5 inline-block text-sky-500" />
                <span>{{ formatTokens(stats?.today_cache_read_tokens || 0) }}</span>
              </span>
              <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
              <span :title="t('usage.tokenIconHint.cacheWrite')" class="cursor-help">
                <Icon name="edit" size="xs" class="mr-0.5 inline-block text-amber-500" />
                <span>{{ formatTokens(stats?.today_cache_creation_tokens || 0) }}</span>
              </span>
            </p>
          </div>
          <div v-if="hasSpark(tokensSeries)" class="kpi-card-spark">
            <SparklineMini :data="tokensSeries" :color="SERIES_COLORS.brandSoft" :height="40" />
          </div>
        </div>

        <!-- Total Tokens -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-neutral">
            <Icon name="database" size="sm" :stroke-width="1.75" />
          </div>
          <div class="metric-body">
            <p class="metric-label">{{ t('dashboard.totalTokens') }}</p>
            <p class="metric-value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
            <div class="metric-hint space-y-0.5 tabular-nums">
              <p>
                <span :title="t('usage.tokenIconHint.input')" class="cursor-help">
                  <Icon name="arrowDown" size="xs" class="mr-0.5 inline-block text-emerald-500" />
                  <span>{{ formatTokens(stats?.total_input_tokens || 0) }}</span>
                </span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <span :title="t('usage.tokenIconHint.output')" class="cursor-help">
                  <Icon name="arrowUp" size="xs" class="mr-0.5 inline-block text-brand-500" />
                  <span>{{ formatTokens(stats?.total_output_tokens || 0) }}</span>
                </span>
              </p>
              <p v-if="hasCacheTotal">
                <span :title="t('usage.tokenIconHint.cacheRead')" class="cursor-help">
                  <Icon name="inbox" size="xs" class="mr-0.5 inline-block text-sky-500" />
                  <span>{{ formatTokens(stats?.total_cache_read_tokens || 0) }}</span>
                </span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <span :title="t('usage.tokenIconHint.cacheWrite')" class="cursor-help">
                  <Icon name="edit" size="xs" class="mr-0.5 inline-block text-amber-500" />
                  <span>{{ formatTokens(stats?.total_cache_creation_tokens || 0) }}</span>
                </span>
              </p>
            </div>
          </div>
        </div>

        <!-- Performance -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-neutral">
            <Icon name="bolt" size="sm" :stroke-width="1.75" />
          </div>
          <div class="metric-body">
            <p class="metric-label">{{ t('dashboard.performance') }}</p>
            <p class="metric-value">
              {{ formatTokens(stats?.rpm || 0) }}
              <span class="text-xs font-normal text-gray-500 dark:text-dark-400">RPM</span>
            </p>
            <p class="metric-hint tabular-nums">
              <span class="font-medium text-gray-700 dark:text-gray-300">{{ formatTokens(stats?.tpm || 0) }}</span>
              TPM
            </p>
          </div>
        </div>

        <!-- Avg Response -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-neutral">
            <Icon name="clock" size="sm" :stroke-width="1.75" />
          </div>
          <div class="metric-body">
            <p class="metric-label">{{ t('dashboard.avgResponse') }}</p>
            <p class="metric-value">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
            <p class="metric-hint">{{ t('dashboard.averageTime') }}</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import SparklineMini from '@/components/charts/SparklineMini.vue'
import TrendChip from '@/components/charts/TrendChip.vue'
import UserDashboardHero from '@/components/user/dashboard/UserDashboardHero.vue'
import { formatCompactNumber } from '@/utils/format'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

const props = withDefaults(
  defineProps<{
    stats: UserStatsType
    balance: number
    isSimple: boolean
    requestsSeries?: number[]
    tokensSeries?: number[]
    costSeries?: number[]
    requestsTrend?: number | null
    tokensTrend?: number | null
    costTrend?: number | null
  }>(),
  {
    requestsSeries: () => [],
    tokensSeries: () => [],
    costSeries: () => [],
    requestsTrend: null,
    tokensTrend: null,
    costTrend: null
  }
)
const { t } = useI18n()

// sparkline 品牌化系列色：brand-500 / brand-400 / emerald-500（对应 Tailwind 令牌值）
const SERIES_COLORS = {
  brand: '#f97316',
  brandSoft: '#fb923c',
  emerald: '#10b981'
} as const

// 至少 2 个点 + 不全为 0 才显示 sparkline，避免无数据时出现空白条
const hasSpark = (series?: number[]) => Array.isArray(series) && series.length >= 2 && series.some((v) => v > 0)

// 是否有缓存数据（任一字段 > 0）— 没有就不渲染 cache 行，避免出现一整行 0
const hasCacheToday = computed(() =>
  (props.stats?.today_cache_read_tokens || 0) > 0 || (props.stats?.today_cache_creation_tokens || 0) > 0,
)
const hasCacheTotal = computed(() =>
  (props.stats?.total_cache_read_tokens || 0) > 0 || (props.stats?.total_cache_creation_tokens || 0) > 0,
)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
// 走全局 formatCompactNumber，统一支持到 T/P 单位
const formatTokens = (n: number) => formatCompactNumber(n, { decimals: 1 })
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>

<style scoped>
/* 用户仪表盘 KPI 卡：白底中性卡身，仅以 metric-icon 主题色做克制的功能区分 */
.kpi-card {
  @apply relative flex flex-col gap-1 overflow-hidden rounded-2xl border border-gray-200/70 bg-white p-5 transition-all duration-200;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
  box-shadow: 0 1px 1px rgb(16 24 40 / 0.03), 0 2px 5px -1px rgb(16 24 40 / 0.05);
}
.kpi-card:hover {
  border-color: rgb(212 212 216);
  transform: translateY(-2px);
  box-shadow: 0 10px 28px -8px rgb(16 24 40 / 0.14), 0 3px 8px -3px rgb(16 24 40 / 0.07);
}
:root.dark .kpi-card:hover {
  border-color: rgb(82 82 91);
}

.kpi-card-header { @apply flex items-start justify-between gap-2; }
.kpi-card-label { @apply mt-3 text-sm font-medium text-gray-500 dark:text-dark-400; }
.kpi-card-value { @apply mt-1 flex items-baseline gap-1.5 text-[30px] font-semibold leading-tight tabular-nums tracking-tight text-gray-900 dark:text-white; }
.kpi-card-hint { @apply mt-1 flex flex-wrap items-center gap-x-1 text-xs text-gray-500 dark:text-dark-400; }
.kpi-card-spark { @apply -mx-5 -mb-5 mt-4 h-10; }
.kpi-card-spark > svg { @apply h-full w-full; }

/* 用户仪表盘次级 metric 卡：白底，与 kpi-card 同节奏 */
.metric-card {
  @apply flex items-start gap-3 rounded-2xl border border-gray-200/70 bg-white p-5 transition-all duration-200;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
  box-shadow: 0 1px 1px rgb(16 24 40 / 0.03), 0 2px 5px -1px rgb(16 24 40 / 0.05);
}
.metric-card:hover {
  border-color: rgb(212 212 216);
  transform: translateY(-2px);
  box-shadow: 0 10px 28px -8px rgb(16 24 40 / 0.14), 0 3px 8px -3px rgb(16 24 40 / 0.07);
}
:root.dark .metric-card:hover {
  border-color: rgb(82 82 91);
}

/* 图标主题：brand=主指标、emerald=成本/积极、neutral=中性辅助（收敛历史紫/玫/青/靛配色） */
.metric-icon { @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl; }
.metric-icon-brand   { @apply bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300; }
.metric-icon-emerald { @apply bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300; }
.metric-icon-neutral { @apply bg-gray-100 text-gray-600 dark:bg-dark-700/70 dark:text-dark-300; }

.metric-body { @apply min-w-0 flex-1; }
.metric-label { @apply text-sm font-medium text-gray-500 dark:text-dark-400; }
.metric-value { @apply mt-1 flex items-baseline gap-1.5 text-[26px] font-semibold leading-tight tabular-nums tracking-tight text-gray-900 dark:text-white; }
.metric-hint { @apply mt-1.5 flex flex-wrap items-center gap-x-1 text-xs text-gray-500 dark:text-dark-400; }
</style>

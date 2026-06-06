<template>
  <AppLayout wide>
    <div class="space-y-8">
      <!-- 工具栏：sm 起两端对齐，小屏 stack 避免溢出 -->
      <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
        <span class="inline-flex w-fit items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
          {{ t('admin.dashboard.liveUpdated') }} {{ lastUpdatedLabel }}
        </span>
        <div class="flex flex-wrap items-center gap-2">
          <DateRangePicker
            v-model:start-date="startDate"
            v-model:end-date="endDate"
            @change="onDateRangeChange"
          />
          <div class="w-28">
            <Select
              v-model="granularity"
              :options="granularityOptions"
              @change="loadChartData"
            />
          </div>
          <button @click="loadDashboardStats" :disabled="chartsLoading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
            <Icon name="refresh" size="sm" :class="chartsLoading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- 今日运营快照 -->
        <section class="space-y-4">
          <div class="flex items-baseline justify-between">
            <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dashboard.todaySnapshot') }}
            </h2>
            <span class="text-xs text-gray-400 dark:text-dark-500">{{ todayLabel }}</span>
          </div>
          <!-- 移动端 2 列：4 张卡叠成 2×2 比单列长长一条更紧凑 -->
          <div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
            <!-- 今日请求 -->
            <div class="kpi-card">
              <div class="kpi-card-header">
                <div class="metric-icon metric-icon-brand">
                  <Icon name="chart" size="sm" :stroke-width="1.75" />
                </div>
                <TrendChip :value="requestsTrend" />
              </div>
              <p class="kpi-card-label">{{ t('admin.dashboard.todayRequests') }}</p>
              <p class="kpi-card-value">{{ formatNumber(stats.today_requests) }}</p>
              <p class="kpi-card-hint">
                {{ t('common.total') }}
                <span class="font-medium text-gray-700 dark:text-gray-300">{{ formatNumber(stats.total_requests) }}</span>
              </p>
              <div v-if="hasSparkData(requestsSeries)" class="kpi-card-spark">
                <SparklineMini :data="requestsSeries" color="#f97316" :height="40" />
              </div>
            </div>

            <!-- 今日 Token：拆解显示 input / output / cache R/W，让总数验证一致 -->
            <div class="kpi-card">
              <div class="kpi-card-header">
                <div class="metric-icon metric-icon-violet">
                  <Icon name="cube" size="sm" :stroke-width="1.75" />
                </div>
                <TrendChip :value="tokensTrend" />
              </div>
              <p class="kpi-card-label">{{ t('admin.dashboard.todayTokens') }}</p>
              <p class="kpi-card-value">{{ formatTokens(stats.today_tokens) }}</p>
              <div class="kpi-card-hint space-y-0.5 tabular-nums">
                <p>
                  <span :title="t('usage.tokenIconHint.input')" class="cursor-help">
                    <Icon name="arrowDown" size="xs" class="mr-0.5 inline-block text-emerald-500" />
                    <span>{{ formatTokens(stats.today_input_tokens) }}</span>
                  </span>
                  <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                  <span :title="t('usage.tokenIconHint.output')" class="cursor-help">
                    <Icon name="arrowUp" size="xs" class="mr-0.5 inline-block text-violet-500" />
                    <span>{{ formatTokens(stats.today_output_tokens) }}</span>
                  </span>
                </p>
                <p v-if="hasTodayCache">
                  <span :title="t('usage.tokenIconHint.cacheRead')" class="cursor-help">
                    <Icon name="inbox" size="xs" class="mr-0.5 inline-block text-sky-500" />
                    <span>{{ formatTokens(stats.today_cache_read_tokens) }}</span>
                  </span>
                  <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                  <span :title="t('usage.tokenIconHint.cacheWrite')" class="cursor-help">
                    <Icon name="edit" size="xs" class="mr-0.5 inline-block text-amber-500" />
                    <span>{{ formatTokens(stats.today_cache_creation_tokens) }}</span>
                  </span>
                </p>
              </div>
              <div v-if="hasSparkData(tokensSeries)" class="kpi-card-spark">
                <SparklineMini :data="tokensSeries" color="#8b5cf6" :height="40" />
              </div>
            </div>

            <!-- 今日消费 -->
            <div class="kpi-card">
              <div class="kpi-card-header">
                <div class="metric-icon metric-icon-emerald">
                  <Icon name="dollar" size="sm" :stroke-width="1.75" />
                </div>
                <TrendChip :value="costTrend" />
              </div>
              <p class="kpi-card-label">{{ t('admin.dashboard.todayCost') }}</p>
              <p class="kpi-card-value">${{ formatCost(stats.today_actual_cost) }}</p>
              <p class="kpi-card-hint tabular-nums">
                <span :title="t('admin.dashboard.accountCost')">${{ formatCost(stats.today_account_cost) }}</span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <span :title="t('admin.dashboard.standard')">${{ formatCost(stats.today_cost) }}</span>
              </p>
              <div v-if="hasSparkData(costSeries)" class="kpi-card-spark">
                <SparklineMini :data="costSeries" color="#10b981" :height="40" />
              </div>
            </div>

            <!-- 新增用户：无时间序列；有今日新增时才显示"今日 vs 累计"占比条 -->
            <div class="kpi-card">
              <div class="kpi-card-header">
                <div class="metric-icon metric-icon-sky">
                  <Icon name="userPlus" size="sm" :stroke-width="1.75" />
                </div>
                <span v-if="stats.today_new_users > 0" class="trend-chip trend-up">
                  <Icon name="arrowUp" size="xs" />
                  {{ t('admin.dashboard.todayDelta') }}
                </span>
              </div>
              <p class="kpi-card-label">{{ t('admin.dashboard.users') }}</p>
              <p class="kpi-card-value">+{{ formatNumber(stats.today_new_users) }}</p>
              <p class="kpi-card-hint">
                {{ t('common.total') }}
                <span class="font-medium text-gray-700 dark:text-gray-300">{{ formatNumber(stats.total_users) }}</span>
              </p>
              <div v-if="stats.today_new_users > 0" class="kpi-card-spark flex items-end pb-2">
                <div class="h-1.5 w-full overflow-hidden rounded-full bg-sky-100 dark:bg-sky-500/15">
                  <div
                    class="h-full rounded-full bg-sky-500 transition-all"
                    :style="{ width: `${userGrowthPercent}%` }"
                  ></div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 系统健康度 -->
        <section class="space-y-4">
          <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">
            {{ t('admin.dashboard.systemHealth') }}
          </h2>
          <!-- 移动端 2 列：4 张卡叠成 2×2 比单列长长一条更紧凑 -->
          <div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
            <!-- 账号 -->
            <div class="metric-card">
              <div class="metric-icon metric-icon-amber">
                <Icon name="server" size="md" :stroke-width="1.75" />
              </div>
              <div class="metric-body">
                <p class="metric-label">{{ t('admin.dashboard.accounts') }}</p>
                <p class="metric-value">{{ stats.total_accounts }}</p>
                <p class="metric-hint">
                  <span class="inline-flex items-center gap-1">
                    <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                    <span class="font-medium text-emerald-600 dark:text-emerald-400">{{ stats.normal_accounts }}</span>
                    {{ t('common.active') }}
                  </span>
                  <template v-if="stats.error_accounts > 0">
                    <span class="mx-1.5 text-gray-300 dark:text-dark-600">·</span>
                    <span class="inline-flex items-center gap-1">
                      <span class="h-1.5 w-1.5 rounded-full bg-red-500"></span>
                      <span class="font-medium text-red-600 dark:text-red-400">{{ stats.error_accounts }}</span>
                      {{ t('common.error') }}
                    </span>
                  </template>
                </p>
              </div>
            </div>

            <!-- API Keys -->
            <div class="metric-card">
              <div class="metric-icon metric-icon-rose">
                <Icon name="key" size="md" :stroke-width="1.75" />
              </div>
              <div class="metric-body">
                <p class="metric-label">{{ t('admin.dashboard.apiKeys') }}</p>
                <p class="metric-value">{{ stats.total_api_keys }}</p>
                <p class="metric-hint">
                  <span class="font-medium text-emerald-600 dark:text-emerald-400">{{ stats.active_api_keys }}</span>
                  {{ t('common.active') }}
                </p>
              </div>
            </div>

            <!-- 性能 RPM/TPM -->
            <div class="metric-card">
              <div class="metric-icon metric-icon-teal">
                <Icon name="bolt" size="md" :stroke-width="1.75" />
              </div>
              <div class="metric-body">
                <p class="metric-label">{{ t('admin.dashboard.performance') }}</p>
                <p class="metric-value">
                  {{ formatTokens(stats.rpm) }}
                  <span class="text-xs font-normal text-gray-500 dark:text-dark-400">RPM</span>
                </p>
                <p class="metric-hint tabular-nums">
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{ formatTokens(stats.tpm) }}</span>
                  TPM
                </p>
              </div>
            </div>

            <!-- 平均响应 -->
            <div class="metric-card">
              <div class="metric-icon metric-icon-indigo">
                <Icon name="clock" size="md" :stroke-width="1.75" />
              </div>
              <div class="metric-body">
                <p class="metric-label">{{ t('admin.dashboard.avgResponse') }}</p>
                <p class="metric-value">{{ formatDuration(stats.average_duration_ms) }}</p>
                <p class="metric-hint">
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{ stats.active_users }}</span>
                  {{ t('admin.dashboard.activeUsers') }}
                </p>
              </div>
            </div>
          </div>
        </section>

        <!-- 累计成本 + Token 大数据条带 -->
        <section class="surface-card overflow-hidden">
          <div class="grid grid-cols-1 divide-y divide-gray-200/60 dark:divide-dark-700/60 md:grid-cols-2 md:divide-x md:divide-y-0">
            <div class="p-6">
              <p class="text-[13px] font-medium text-gray-500 dark:text-dark-400">
                {{ t('admin.dashboard.totalCost') }}
              </p>
              <p class="mt-2 flex items-baseline gap-2 tabular-nums">
                <span class="text-[28px] font-semibold leading-none text-gray-900 dark:text-white">${{ formatCost(stats.total_actual_cost) }}</span>
                <span class="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
                  {{ t('admin.dashboard.actual') }}
                </span>
              </p>
              <div class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
                <span class="inline-flex items-center gap-1">
                  <span class="h-1 w-1 rounded-full bg-gray-400"></span>
                  ${{ formatCost(stats.total_account_cost) }} {{ t('admin.dashboard.accountCost') }}
                </span>
                <span class="inline-flex items-center gap-1">
                  <span class="h-1 w-1 rounded-full bg-gray-300"></span>
                  ${{ formatCost(stats.total_cost) }} {{ t('admin.dashboard.standard') }}
                </span>
              </div>
            </div>
            <div class="p-6">
              <p class="text-[13px] font-medium text-gray-500 dark:text-dark-400">
                {{ t('admin.dashboard.totalTokens') }}
              </p>
              <p class="mt-2 flex items-baseline gap-2 tabular-nums">
                <span class="text-[28px] font-semibold leading-none text-gray-900 dark:text-white">{{ formatTokens(stats.total_tokens) }}</span>
                <span class="rounded-full bg-brand-50 px-2 py-0.5 text-[11px] font-medium text-brand-700 dark:bg-brand-500/10 dark:text-brand-300">
                  {{ t('admin.dashboard.todayDelta') }} +{{ formatTokens(stats.today_tokens) }}
                </span>
              </p>
              <!-- 累计 Token 分解：input + output + cache 加起来等于总数 -->
              <div class="mt-3 grid grid-cols-2 gap-x-4 gap-y-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
                <span class="inline-flex items-center gap-1">
                  <Icon name="arrowDown" size="xs" class="text-emerald-500" />
                  <span>{{ formatTokens(stats.total_input_tokens) }}</span>
                  <span class="opacity-70">{{ t('dashboard.input') }}</span>
                </span>
                <span class="inline-flex items-center gap-1">
                  <Icon name="arrowUp" size="xs" class="text-violet-500" />
                  <span>{{ formatTokens(stats.total_output_tokens) }}</span>
                  <span class="opacity-70">{{ t('dashboard.output') }}</span>
                </span>
                <span v-if="hasTotalCache" class="inline-flex items-center gap-1">
                  <Icon name="inbox" size="xs" class="text-sky-500" />
                  <span>{{ formatTokens(stats.total_cache_read_tokens) }}</span>
                  <span class="opacity-70">{{ t('pricing.cacheRead') }}</span>
                </span>
                <span v-if="hasTotalCache" class="inline-flex items-center gap-1">
                  <Icon name="edit" size="xs" class="text-amber-500" />
                  <span>{{ formatTokens(stats.total_cache_creation_tokens) }}</span>
                  <span class="opacity-70">{{ t('pricing.cacheWrite') }}</span>
                </span>
              </div>
            </div>
          </div>
        </section>

        <!-- 图表区域 -->
        <section class="space-y-4">
          <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">
            {{ t('admin.dashboard.tokenUsageTrend') }}
          </h2>
          <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
            <ModelDistributionChart
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-actual-cost="rankingTotalActualCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToUserUsage"
            />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <!-- 用户使用趋势 -->
          <div class="surface-card overflow-hidden">
            <div class="flex items-center justify-between border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
              <div class="flex items-center gap-2.5">
                <h3 class="text-[15px] font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.dashboard.userTrendTitle') }}
                </h3>
                <span class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300">
                  Top 12
                </span>
              </div>
              <span class="text-xs text-gray-400 dark:text-dark-500">{{ startDate }} ~ {{ endDate }}</span>
            </div>
            <div class="p-6">
              <div class="h-72">
                <div v-if="userTrendLoading" class="flex h-full items-center justify-center">
                  <LoadingSpinner size="md" />
                </div>
                <Line v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
                <div
                  v-else
                  class="flex h-full flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-dark-400"
                >
                  <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500">
                    <Icon name="chart" size="md" />
                  </div>
                  {{ t('admin.dashboard.noDataAvailable') }}
                </div>
              </div>
            </div>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { formatCompactNumber } from '@/utils/format'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  UserUsageTrendPoint,
  UserSpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import SparklineMini from '@/components/charts/SparklineMini.vue'
import TrendChip from '@/components/charts/TrendChip.vue'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const userTrendLoading = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)

const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
let chartLoadSeq = 0
let usersTrendLoadSeq = 0
let rankingLoadSeq = 0
const rankingLimit = 12

const lastUpdated = ref<Date | null>(null)
const nowTick = ref<number>(Date.now())
let nowTimer: ReturnType<typeof setInterval> | null = null

const lastUpdatedLabel = computed(() => {
  if (!lastUpdated.value) return t('common.time.never')
  const diffSec = Math.max(0, Math.round((nowTick.value - lastUpdated.value.getTime()) / 1000))
  if (diffSec < 5) return t('common.time.justNow')
  if (diffSec < 60) return t('admin.dashboard.secondsAgo', { n: diffSec })
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return t('common.time.minutesAgo', { n: diffMin })
  const diffHour = Math.floor(diffMin / 60)
  return t('common.time.hoursAgo', { n: diffHour })
})

const todayLabel = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
})

// ============ KPI sparkline 数据派生 ============
// 从 trendData（默认近 24h 按小时）派生 sparkline 序列，至少 4 个点；不足时给空数组让 sparkline 不渲染
const requestsSeries = computed(() => trendData.value.map((d) => d.requests))
const tokensSeries = computed(() => trendData.value.map((d) => d.total_tokens))
const costSeries = computed(() => trendData.value.map((d) => d.actual_cost))

// 趋势百分比：对比"前半段平均"vs"后半段平均"，反映近期增长/下降走势
// 没有"昨日"数据时这是次优近似，但视觉上能直观传达"在增长 / 在下降"
const calcTrend = (series: number[]): number | null => {
  if (series.length < 4) return null
  const mid = Math.floor(series.length / 2)
  const firstHalf = series.slice(0, mid)
  const secondHalf = series.slice(mid)
  const sum = (arr: number[]) => arr.reduce((s, v) => s + v, 0)
  const a = sum(firstHalf) / firstHalf.length
  const b = sum(secondHalf) / secondHalf.length
  if (a === 0) return b > 0 ? 100 : null
  return ((b - a) / a) * 100
}

const requestsTrend = computed(() => calcTrend(requestsSeries.value))
const tokensTrend = computed(() => calcTrend(tokensSeries.value))
const costTrend = computed(() => calcTrend(costSeries.value))

// sparkline 数据有效性：至少 2 个点 + 不全为 0，否则隐藏整块 spark 区，避免出现"空白条"
const hasSparkData = (series: number[]): boolean => {
  return series.length >= 2 && series.some((v) => v > 0)
}

// 缓存 token 是否有数据：任一字段 > 0 才展示 cache 行，避免显示 "0 · 0"
const hasTodayCache = computed(() => {
  const s = stats.value
  if (!s) return false
  return (s.today_cache_read_tokens || 0) > 0 || (s.today_cache_creation_tokens || 0) > 0
})
const hasTotalCache = computed(() => {
  const s = stats.value
  if (!s) return false
  return (s.total_cache_read_tokens || 0) > 0 || (s.total_cache_creation_tokens || 0) > 0
})

// 新增用户占总用户的百分比（给"用户" KPI 卡的占比条做数据）
const userGrowthPercent = computed(() => {
  if (!stats.value || !stats.value.total_users) return 0
  const pct = (stats.value.today_new_users / stats.value.total_users) * 100
  // 至少显示一点宽度让条状可见，最高 100
  return Math.min(100, Math.max(stats.value.today_new_users > 0 ? 3 : 0, pct))
})

const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const getLast24HoursRangeDates = (): { start: string; end: string } => {
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end)
  }
}

const granularity = ref<'day' | 'hour'>('hour')
const defaultRange = getLast24HoursRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb'
}))

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: { size: 11 }
      }
    },
    tooltip: {
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => `${context.dataset.label}: ${formatTokens(context.raw)}`
      }
    }
  },
  scales: {
    x: {
      grid: { color: chartColors.value.grid },
      ticks: { color: chartColors.value.text, font: { size: 10 } }
    },
    y: {
      grid: { color: chartColors.value.grid },
      ticks: {
        color: chartColors.value.text,
        font: { size: 10 },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    }
  }
}))

const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null

  const getDisplayName = (point: UserUsageTrendPoint): string => {
    const username = point.username?.trim()
    if (username) return username
    const email = point.email?.trim()
    if (email) return email
    return t('admin.redeem.userPrefix', { id: point.user_id })
  }

  const userGroups = new Map<number, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = point.user_id
    if (!userGroups.has(key)) {
      userGroups.set(key, { name: getDisplayName(point), data: new Map() })
    }
    userGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#6366f1', '#84cc16', '#06b6d4', '#a855f7']

  const datasets = Array.from(userGroups.values()).map((group, idx) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: colors[idx % colors.length],
    backgroundColor: `${colors[idx % colors.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return { labels: sortedDates, datasets }
})

// 走全局 formatCompactNumber，统一支持到 T/P 单位
const formatTokens = (value: number | undefined): string =>
  formatCompactNumber(value, { decimals: 2 })

const formatNumber = (value: number): string => value.toLocaleString()

const formatCost = (value: number | null | undefined): string => {
  // null / undefined / NaN 兜底为 0：API 返回字段缺失时不让 .toFixed() 抛 TypeError 让面板崩
  if (value == null || Number.isNaN(value)) return '0.00'
  if (value >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (value >= 1) return value.toFixed(2)
  if (value >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${Math.round(ms)}ms`
}

const goToUserUsage = (item: UserSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      user_id: String(item.user_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))
  granularity.value = daysDiff <= 1 ? 'hour' : 'day'
  loadChartData()
}

const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) loading.value = true
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) stats.value = response.stats
    trendData.value = response.trend || []
    modelStats.value = response.models || []
    lastUpdated.value = new Date()
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUsersTrend = async () => {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: 12
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === usersTrendLoadSeq) userTrendLoading.value = false
  }
}

const loadUserSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getUserSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalActualCost.value = response.total_actual_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading user spending ranking:', error)
    rankingItems.value = []
    rankingTotalActualCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) rankingLoading.value = false
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadUsersTrend(),
    loadUserSpendingRanking()
  ])
}

onMounted(() => {
  loadDashboardStats()
  nowTimer = setInterval(() => { nowTick.value = Date.now() }, 5000)
})

onBeforeUnmount(() => {
  if (nowTimer) { clearInterval(nowTimer); nowTimer = null }
})
</script>

<style scoped>
/* ============ Surface 通用卡片容器：中性白底，让 metric-icon 主题色做区分 ============ */
.surface-card {
  @apply rounded-2xl border border-gray-200/70 bg-white shadow-card transition-all;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}

/* ============ KPI 主卡：白底 + 不同 metric-icon 区分 ============ */
.kpi-card {
  @apply relative flex flex-col gap-1 overflow-hidden rounded-2xl border border-gray-200/70 bg-white p-4 shadow-card transition-all duration-200 sm:p-5;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
.kpi-card:hover {
  border-color: rgb(209 213 219);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04), 0 8px 24px -12px rgb(15 23 42 / 0.18);
}
:root.dark .kpi-card:hover {
  border-color: rgb(75 85 99);
}

.kpi-card-header {
  @apply flex items-start justify-between gap-2;
}

.kpi-card-label {
  @apply mt-3 text-[13px] font-medium text-gray-500 dark:text-dark-400;
}

.kpi-card-value {
  @apply mt-1 flex items-baseline gap-1.5 text-2xl font-semibold leading-tight tabular-nums text-gray-900 sm:text-[30px] dark:text-white;
}

.kpi-card-hint {
  @apply mt-1 flex flex-wrap items-center gap-x-1 text-xs text-gray-500 dark:text-dark-400;
}

/* sparkline 贴底显示，撑满卡片下半，让数据"有图形回应" */
.kpi-card-spark {
  @apply -mx-4 -mb-4 mt-4 h-10 sm:-mx-5 sm:-mb-5;
}
.kpi-card-spark > svg {
  @apply h-full w-full;
}

/* ============ 系统健康度副卡：白底，metric-icon 已区分 ============ */
.metric-card {
  @apply flex items-start gap-3 rounded-2xl border border-gray-200/70 bg-white p-4 shadow-card transition-all duration-200 sm:p-5;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
.metric-card:hover {
  border-color: rgb(209 213 219);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04), 0 4px 16px -8px rgb(15 23 42 / 0.12);
}
:root.dark .metric-card:hover {
  border-color: rgb(75 85 99);
}

.metric-icon {
  @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl;
}

/* colored icon backgrounds：柔和的 50 浅底 + 600 描边图标 */
.metric-icon-brand   { @apply bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300; }
.metric-icon-emerald { @apply bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300; }
.metric-icon-sky     { @apply bg-sky-50 text-sky-600 dark:bg-sky-500/15 dark:text-sky-300; }
.metric-icon-violet  { @apply bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300; }
.metric-icon-amber   { @apply bg-amber-50 text-amber-600 dark:bg-amber-500/15 dark:text-amber-300; }
.metric-icon-rose    { @apply bg-rose-50 text-rose-600 dark:bg-rose-500/15 dark:text-rose-300; }
.metric-icon-teal    { @apply bg-teal-50 text-teal-600 dark:bg-teal-500/15 dark:text-teal-300; }
.metric-icon-indigo  { @apply bg-indigo-50 text-indigo-600 dark:bg-indigo-500/15 dark:text-indigo-300; }

.metric-body { @apply min-w-0 flex-1; }
.metric-label { @apply text-[13px] font-medium text-gray-500 dark:text-dark-400; }
.metric-value { @apply mt-1 flex items-baseline gap-1.5 text-[26px] font-semibold leading-tight tabular-nums text-gray-900 dark:text-white; }
.metric-hint { @apply mt-1.5 flex flex-wrap items-center gap-x-1 text-xs text-gray-500 dark:text-dark-400; }

/* ============ trend-chip 兜底（用户卡片直接用了类名，未走 TrendChip 组件） ============ */
.trend-chip {
  @apply inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-[11px] font-semibold tabular-nums;
}
.trend-up { @apply bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300; }
.trend-down { @apply bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300; }
</style>

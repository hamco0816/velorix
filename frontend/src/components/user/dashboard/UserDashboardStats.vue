<template>
  <div class="space-y-8">
    <!-- 核心数据：余额 + 今日运营，带 sparkline -->
    <section class="space-y-4">
      <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">
        {{ t('dashboard.coreStats') }}
      </h2>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Balance：余额作为账户首要指标，brand 主色 + 占比条 -->
        <div v-if="!isSimple" class="kpi-card">
          <div class="kpi-card-header">
            <div class="metric-icon metric-icon-brand">
              <Icon name="dollar" size="sm" :stroke-width="1.75" />
            </div>
          </div>
          <p class="kpi-card-label">{{ t('dashboard.balance') }}</p>
          <p class="kpi-card-value">${{ formatBalance(balance) }}</p>
          <p class="kpi-card-hint">{{ t('common.available') }}</p>
        </div>

        <!-- Today Requests：带 sparkline + trend -->
        <div class="kpi-card">
          <div class="kpi-card-header">
            <div class="metric-icon metric-icon-violet">
              <Icon name="chart" size="sm" :stroke-width="1.75" />
            </div>
            <TrendChip :value="requestsTrend" />
          </div>
          <p class="kpi-card-label">{{ t('dashboard.todayRequests') }}</p>
          <p class="kpi-card-value">{{ formatNumber(stats?.today_requests || 0) }}</p>
          <p class="kpi-card-hint">
            {{ t('common.total') }}
            <span class="font-medium text-gray-700 dark:text-gray-300">{{ formatNumber(stats?.total_requests || 0) }}</span>
          </p>
          <div v-if="hasSpark(requestsSeries)" class="kpi-card-spark">
            <SparklineMini :data="requestsSeries" color="#8b5cf6" :height="40" />
          </div>
        </div>

        <!-- Today Cost：带 sparkline + trend -->
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
            <SparklineMini :data="costSeries" color="#10b981" :height="40" />
          </div>
        </div>

        <!-- API Keys：当前快照值，无 sparkline -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-sky">
            <Icon name="key" size="sm" :stroke-width="1.75" />
          </div>
          <div class="metric-body">
            <p class="metric-label">{{ t('dashboard.apiKeys') }}</p>
            <p class="metric-value">{{ stats?.total_api_keys || 0 }}</p>
            <p class="metric-hint">
              <span class="inline-flex items-center gap-1">
                <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                <span class="font-medium text-emerald-600 dark:text-emerald-400">{{ stats?.active_api_keys || 0 }}</span>
                {{ t('common.active') }}
              </span>
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Token / 性能数据 -->
    <section class="space-y-4">
      <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">
        {{ t('dashboard.tokenStats') }}
      </h2>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Today Tokens：带 sparkline -->
        <div class="kpi-card">
          <div class="kpi-card-header">
            <div class="metric-icon metric-icon-amber">
              <Icon name="cube" size="sm" :stroke-width="1.75" />
            </div>
            <TrendChip :value="tokensTrend" />
          </div>
          <p class="kpi-card-label">{{ t('dashboard.todayTokens') }}</p>
          <p class="kpi-card-value">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <!-- 输入/输出 + 缓存读取/创建 分两行展示，保证四项加起来等于 total_tokens -->
          <div class="kpi-card-hint space-y-0.5 tabular-nums">
            <p>
              <Icon name="arrowDown" size="xs" class="mr-0.5 inline-block text-emerald-500" />
              <span>{{ formatTokens(stats?.today_input_tokens || 0) }}</span>
              <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
              <Icon name="arrowUp" size="xs" class="mr-0.5 inline-block text-violet-500" />
              <span>{{ formatTokens(stats?.today_output_tokens || 0) }}</span>
            </p>
            <p v-if="hasCacheToday">
              <Icon name="inbox" size="xs" class="mr-0.5 inline-block text-sky-500" />
              <span>{{ formatTokens(stats?.today_cache_read_tokens || 0) }}</span>
              <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
              <Icon name="edit" size="xs" class="mr-0.5 inline-block text-amber-500" />
              <span>{{ formatTokens(stats?.today_cache_creation_tokens || 0) }}</span>
            </p>
          </div>
          <div v-if="hasSpark(tokensSeries)" class="kpi-card-spark">
            <SparklineMini :data="tokensSeries" color="#f59e0b" :height="40" />
          </div>
        </div>

        <!-- Total Tokens -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-rose">
            <Icon name="database" size="sm" :stroke-width="1.75" />
          </div>
          <div class="metric-body">
            <p class="metric-label">{{ t('dashboard.totalTokens') }}</p>
            <p class="metric-value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
            <div class="metric-hint space-y-0.5 tabular-nums">
              <p>
                <Icon name="arrowDown" size="xs" class="mr-0.5 inline-block text-emerald-500" />
                <span>{{ formatTokens(stats?.total_input_tokens || 0) }}</span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <Icon name="arrowUp" size="xs" class="mr-0.5 inline-block text-violet-500" />
                <span>{{ formatTokens(stats?.total_output_tokens || 0) }}</span>
              </p>
              <p v-if="hasCacheTotal">
                <Icon name="inbox" size="xs" class="mr-0.5 inline-block text-sky-500" />
                <span>{{ formatTokens(stats?.total_cache_read_tokens || 0) }}</span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <Icon name="edit" size="xs" class="mr-0.5 inline-block text-amber-500" />
                <span>{{ formatTokens(stats?.total_cache_creation_tokens || 0) }}</span>
              </p>
            </div>
          </div>
        </div>

        <!-- Performance -->
        <div class="metric-card">
          <div class="metric-icon metric-icon-teal">
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
          <div class="metric-icon metric-icon-indigo">
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

// 至少 2 个点 + 不全为 0 才显示 sparkline，避免无数据时出现空白条
const hasSpark = (series?: number[]) => Array.isArray(series) && series.length >= 2 && series.some((v) => v > 0)

// 是否有缓存数据（任一字段 > 0）— 没有就不渲染 cache 行，避免出现一整行 0
const hasCacheToday = computed(() =>
  (props.stats?.today_cache_read_tokens || 0) > 0 || (props.stats?.today_cache_creation_tokens || 0) > 0,
)
const hasCacheTotal = computed(() =>
  (props.stats?.total_cache_read_tokens || 0) > 0 || (props.stats?.total_cache_creation_tokens || 0) > 0,
)

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>

<style scoped>
/* 用户仪表盘 KPI 卡：violet 染色（核心数据氛围） */
.kpi-card {
  @apply relative flex flex-col gap-1 overflow-hidden rounded-2xl p-5 transition-all duration-200;
  border: 1px solid rgb(221 214 254 / 0.7);
  background:
    radial-gradient(circle at 100% 0%, rgb(139 92 246 / 0.10), transparent 32%),
    radial-gradient(circle at 0% 100%, rgb(168 85 247 / 0.06), transparent 38%),
    linear-gradient(135deg, rgb(245 243 255), rgb(255 255 255) 55%, rgb(243 232 255));
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}
:root.dark .kpi-card {
  border-color: rgb(76 29 149 / 0.4);
  background:
    radial-gradient(circle at 100% 0%, rgb(139 92 246 / 0.10), transparent 32%),
    radial-gradient(circle at 0% 100%, rgb(168 85 247 / 0.06), transparent 38%),
    linear-gradient(135deg, rgb(30 41 59 / 0.6), rgb(17 24 39 / 0.5));
}
.kpi-card:hover {
  border-color: rgb(196 181 253 / 0.9);
  box-shadow: 0 8px 24px rgb(139 92 246 / 0.10);
}

.kpi-card-header { @apply flex items-start justify-between gap-2; }
.kpi-card-label { @apply mt-3 text-[13px] font-medium text-gray-500 dark:text-dark-400; }
.kpi-card-value { @apply mt-1 flex items-baseline gap-1.5 text-[30px] font-semibold leading-tight tabular-nums text-gray-900 dark:text-white; }
.kpi-card-hint { @apply mt-1 flex flex-wrap items-center gap-x-1 text-xs text-gray-500 dark:text-dark-400; }
.kpi-card-spark { @apply -mx-5 -mb-5 mt-4 h-10; }
.kpi-card-spark > svg { @apply h-full w-full; }

/* 用户仪表盘次级 metric 卡：sky 染色（信息辅助色） */
.metric-card {
  @apply flex items-start gap-3 rounded-2xl p-5 transition-all duration-200;
  border: 1px solid rgb(186 230 253 / 0.7);
  background:
    radial-gradient(circle at 100% 0%, rgb(14 165 233 / 0.10), transparent 32%),
    radial-gradient(circle at 0% 100%, rgb(59 130 246 / 0.06), transparent 38%),
    linear-gradient(135deg, rgb(240 249 255), rgb(255 255 255) 55%, rgb(224 242 254));
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}
:root.dark .metric-card {
  border-color: rgb(7 89 133 / 0.4);
  background:
    radial-gradient(circle at 100% 0%, rgb(14 165 233 / 0.10), transparent 32%),
    radial-gradient(circle at 0% 100%, rgb(59 130 246 / 0.06), transparent 38%),
    linear-gradient(135deg, rgb(30 41 59 / 0.6), rgb(17 24 39 / 0.5));
}
.metric-card:hover {
  border-color: rgb(125 211 252 / 0.9);
  box-shadow: 0 4px 16px rgb(14 165 233 / 0.10);
}

.metric-icon { @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl; }
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
</style>

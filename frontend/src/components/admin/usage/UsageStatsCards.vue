<template>
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <!-- Total Requests -->
    <div class="metric-card">
      <span class="metric-icon bg-info-soft text-info dark:bg-info-deep/30 dark:text-info">
        <Icon name="document" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('usage.totalRequests') }}</p>
        <p class="mt-1 text-[26px] font-semibold leading-tight tabular-nums text-gray-900 dark:text-white">
          {{ stats?.total_requests?.toLocaleString() || '0' }}
        </p>
        <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">{{ t('usage.inSelectedRange') }}</p>
      </div>
    </div>

    <!-- Total Tokens -->
    <div class="metric-card">
      <span class="metric-icon bg-warning-soft text-warning dark:bg-warning-deep/30 dark:text-brand-300">
        <Icon name="cube" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('usage.totalTokens') }}</p>
        <p class="mt-1 text-[26px] font-semibold leading-tight tabular-nums text-gray-900 dark:text-white">
          {{ formatTokens(stats?.total_tokens || 0) }}
        </p>
        <p class="mt-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
          <span :title="t('usage.tokenIconHint.input')" class="cursor-help">{{ t('usage.in') }} {{ formatTokens(stats?.total_input_tokens || 0) }}</span>
          <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
          <span :title="t('usage.tokenIconHint.output')" class="cursor-help">{{ t('usage.out') }} {{ formatTokens(stats?.total_output_tokens || 0) }}</span>
          <template v-if="(stats?.total_cache_tokens || 0) > 0">
            <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
            <span :title="t('usage.tokenIconHint.cacheRead')" class="cursor-help">{{ t('usage.cache') }} {{ formatTokens(stats?.total_cache_tokens || 0) }}</span>
          </template>
        </p>
      </div>
    </div>

    <!-- Total Cost -->
    <div class="metric-card">
      <span class="metric-icon bg-success-soft text-success dark:bg-success-deep/30 dark:text-tea-300">
        <Icon name="dollar" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('usage.totalCost') }}</p>
        <p class="mt-1 text-[26px] font-semibold leading-tight tabular-nums text-success dark:text-tea-300">
          ${{ (stats?.total_actual_cost || 0).toFixed(4) }}
        </p>
        <p class="mt-1 flex flex-wrap items-center gap-x-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
          <span>{{ t('usage.accountCost') }} ${{ (stats?.total_account_cost || 0).toFixed(4) }}</span>
          <span class="text-gray-300 dark:text-dark-600">·</span>
          <span>{{ t('usage.standardCost') }} ${{ (stats?.total_cost || 0).toFixed(4) }}</span>
        </p>
      </div>
    </div>

    <!-- Avg Duration -->
    <div class="metric-card">
      <span class="metric-icon bg-brand-50 text-brand-600 dark:bg-brand-900/30 dark:text-brand-300">
        <Icon name="clock" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('usage.avgDuration') }}</p>
        <p class="mt-1 text-[26px] font-semibold leading-tight tabular-nums text-gray-900 dark:text-white">
          {{ formatDuration(stats?.average_duration_ms || 0) }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import Icon from '@/components/icons/Icon.vue'
import { formatCompactNumber } from '@/utils/format'

defineProps<{ stats: AdminUsageStatsResponse | null }>()

const { t } = useI18n()

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

// 走全局 formatCompactNumber，统一支持到 T/P 单位；保留 2 位小数与原有显示风格一致
const formatTokens = (value: number) => formatCompactNumber(value, { decimals: 2 })
</script>

<style scoped>
.metric-card {
  @apply flex items-start gap-3 rounded-2xl border border-gray-200/70 bg-white p-4 shadow-card transition-all duration-200;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
  @apply hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-card-hover dark:hover:border-dark-600;
}

.metric-icon {
  @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl;
}
</style>

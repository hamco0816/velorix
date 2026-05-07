<template>
  <div class="space-y-8">
    <!-- 核心数据：余额作为用户首要关注，用 brand 强调；其余中性灰统一节奏 -->
    <section class="space-y-3">
      <h2 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ t('dashboard.coreStats') }}
      </h2>
      <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <!-- Balance：用户首要关注，brand 点睛 -->
        <div v-if="!isSimple" class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-brand-50 text-brand-700 dark:bg-brand-500/10 dark:text-brand-300">
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.balance') }}</p>
              <p class="mt-0.5 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">${{ formatBalance(balance) }}</p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('common.available') }}</p>
            </div>
          </div>
        </div>

        <!-- API Keys -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              <Icon name="key" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.apiKeys') }}</p>
              <p class="mt-0.5 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ stats?.total_api_keys || 0 }}</p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                <span class="font-medium text-emerald-600 dark:text-emerald-400">{{ stats?.active_api_keys || 0 }}</span>
                {{ t('common.active') }}
              </p>
            </div>
          </div>
        </div>

        <!-- Today Requests -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              <Icon name="chart" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.todayRequests') }}</p>
              <p class="mt-0.5 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatNumber(stats?.today_requests || 0) }}</p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('common.total') }} {{ formatNumber(stats?.total_requests || 0) }}</p>
            </div>
          </div>
        </div>

        <!-- Today Cost：直接关联到余额扣减，brand 强调 -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-brand-50 text-brand-700 dark:bg-brand-500/10 dark:text-brand-300">
              <Icon name="dollar" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.todayCost') }}</p>
              <p class="mt-0.5 flex items-baseline gap-1.5">
                <span class="text-2xl font-semibold tabular-nums text-gray-900 dark:text-white" :title="t('dashboard.actual')">${{ formatCost(stats?.today_actual_cost || 0) }}</span>
                <span class="text-xs tabular-nums text-gray-400 dark:text-dark-500" :title="t('dashboard.standard')">/ ${{ formatCost(stats?.today_cost || 0) }}</span>
              </p>
              <p class="mt-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
                {{ t('common.total') }}
                <span class="font-medium text-gray-700 dark:text-gray-300">${{ formatCost(stats?.total_actual_cost || 0) }}</span>
                <span class="text-gray-400 dark:text-dark-500"> / ${{ formatCost(stats?.total_cost || 0) }}</span>
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Token / 性能数据 -->
    <section class="space-y-3">
      <h2 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-dark-400">
        {{ t('dashboard.tokenStats') }}
      </h2>
      <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <!-- Today Tokens -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              <Icon name="cube" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.todayTokens') }}</p>
              <p class="mt-0.5 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatTokens(stats?.today_tokens || 0) }}</p>
              <p class="mt-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
                <span>{{ t('dashboard.input') }} {{ formatTokens(stats?.today_input_tokens || 0) }}</span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <span>{{ t('dashboard.output') }} {{ formatTokens(stats?.today_output_tokens || 0) }}</span>
              </p>
            </div>
          </div>
        </div>

        <!-- Total Tokens -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              <Icon name="database" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.totalTokens') }}</p>
              <p class="mt-0.5 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatTokens(stats?.total_tokens || 0) }}</p>
              <p class="mt-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
                <span>{{ t('dashboard.input') }} {{ formatTokens(stats?.total_input_tokens || 0) }}</span>
                <span class="mx-1 text-gray-300 dark:text-dark-600">·</span>
                <span>{{ t('dashboard.output') }} {{ formatTokens(stats?.total_output_tokens || 0) }}</span>
              </p>
            </div>
          </div>
        </div>

        <!-- Performance (RPM/TPM) -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              <Icon name="bolt" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.performance') }}</p>
              <div class="mt-0.5 flex items-baseline gap-1.5">
                <p class="text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatTokens(stats?.rpm || 0) }}</p>
                <span class="text-xs font-medium text-gray-500 dark:text-dark-400">RPM</span>
              </div>
              <div class="mt-0.5 flex items-baseline gap-1.5 text-xs text-gray-500 dark:text-dark-400">
                <span class="font-semibold tabular-nums text-gray-700 dark:text-gray-300">{{ formatTokens(stats?.tpm || 0) }}</span>
                <span>TPM</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Avg Response Time -->
        <div class="card p-5">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-200">
              <Icon name="clock" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('dashboard.avgResponse') }}</p>
              <p class="mt-0.5 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.averageTime') }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'

defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
}>()
const { t } = useI18n()

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>

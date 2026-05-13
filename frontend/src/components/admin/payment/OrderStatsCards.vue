<template>
  <div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
    <!-- Today Revenue -->
    <div class="metric-card">
      <span class="metric-icon bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300">
        <Icon name="dollar" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.todayRevenue') }}</p>
        <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-gray-900 dark:text-white">¥{{ formatMoney(stats.today_amount) }}</p>
        <p class="mt-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ stats.today_count }}</span>
          {{ t('payment.admin.orders') }}
        </p>
      </div>
    </div>

    <!-- Total Revenue -->
    <div class="metric-card">
      <span class="metric-icon bg-sky-50 text-sky-600 dark:bg-sky-500/15 dark:text-sky-300">
        <Icon name="creditCard" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.totalRevenue') }}</p>
        <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-gray-900 dark:text-white">¥{{ formatMoney(stats.total_amount) }}</p>
        <p class="mt-1 text-xs tabular-nums text-gray-500 dark:text-dark-400">
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ stats.total_count }}</span>
          {{ t('payment.admin.orders') }}
        </p>
      </div>
    </div>

    <!-- Today Orders -->
    <div class="metric-card">
      <span class="metric-icon bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300">
        <Icon name="chart" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.todayOrders') }}</p>
        <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-gray-900 dark:text-white">{{ stats.today_count }}</p>
      </div>
    </div>

    <!-- Average Amount -->
    <div class="metric-card">
      <span class="metric-icon bg-amber-50 text-amber-600 dark:bg-amber-500/15 dark:text-amber-300">
        <Icon name="chart" size="sm" :stroke-width="1.75" />
      </span>
      <div class="min-w-0 flex-1">
        <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('payment.admin.avgAmount') }}</p>
        <p class="mt-1 text-xl font-semibold leading-tight tabular-nums sm:text-[24px] text-gray-900 dark:text-white">¥{{ formatMoney(stats.avg_amount) }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DashboardStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  stats: DashboardStats
}>()

function formatMoney(value: number): string {
  return value.toFixed(2)
}
</script>

<style scoped>
.metric-card {
  @apply flex items-start gap-2.5 rounded-2xl border border-gray-200/70 bg-white p-3 shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-all duration-200 sm:gap-3 sm:p-4;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
  @apply hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-[0_4px_16px_rgba(15,23,42,0.06)] dark:hover:border-dark-600;
}

.metric-icon {
  @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl;
}
</style>

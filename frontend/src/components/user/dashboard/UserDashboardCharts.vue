<template>
  <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
    <!-- Model Distribution -->
    <div class="surface-card relative overflow-hidden">
      <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-dark-800/50">
        <LoadingSpinner size="md" />
      </div>
      <div class="border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
        <h3 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('dashboard.modelDistribution') }}</h3>
      </div>
      <div class="p-6">
        <div class="flex flex-wrap items-center gap-6">
          <div class="relative h-44 w-44 flex-shrink-0">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('dashboard.noDataAvailable') }}</div>
          </div>
          <div class="max-h-44 min-w-0 flex-1 overflow-y-auto">
            <table class="w-full text-xs">
              <thead class="sticky top-0 bg-white dark:bg-dark-800">
                <tr class="text-gray-500 dark:text-gray-400">
                  <th class="pb-2 text-left font-medium">{{ t('dashboard.model') }}</th>
                  <th class="pb-2 text-right font-medium">{{ t('dashboard.requests') }}</th>
                  <th class="pb-2 text-right font-medium">{{ t('dashboard.tokens') }}</th>
                  <th class="pb-2 text-right font-medium">{{ t('dashboard.actual') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in models" :key="model.model" class="border-t border-gray-100 dark:border-gray-700/60">
                  <td class="max-w-[140px] truncate py-2 font-medium text-gray-900 dark:text-white" :title="model.model">{{ model.model }}</td>
                  <td class="py-2 text-right tabular-nums text-gray-600 dark:text-gray-400">{{ formatNumber(model.requests) }}</td>
                  <td class="py-2 text-right tabular-nums text-gray-600 dark:text-gray-400">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="py-2 text-right tabular-nums font-medium text-emerald-600 dark:text-emerald-400">${{ formatCost(model.actual_cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Token Usage Trend Chart -->
    <TokenUsageTrend :trend-data="trend" :loading="loading" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, trend: TrendDataPoint[], models: ModelStat[] }>()
const { t } = useI18n()

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: ['#f97316', '#10b981', '#3b82f6', '#8b5cf6', '#f59e0b', '#ec4899', '#06b6d4', '#84cc16'],
    borderWidth: 0
  }]
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: { label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens` }
    }
  }
}
</script>

<style scoped>
.surface-card {
  @apply rounded-2xl border border-gray-200/70 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)];
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
</style>

<!--
  仪表盘图表区：模型分布环图（品牌化配色 + 中心总量）+ Token 使用趋势折线图。
-->
<template>
  <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
    <!-- Model Distribution -->
    <div class="surface-card relative overflow-hidden">
      <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-dark-800/50">
        <LoadingSpinner size="md" />
      </div>
      <div class="border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
        <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('dashboard.modelDistribution') }}</h3>
      </div>
      <div class="p-6">
        <div class="flex flex-wrap items-center gap-6">
          <div class="relative h-44 w-44 flex-shrink-0">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <!-- 环图中心：Token 总量摘要 -->
            <div v-if="modelData" class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
              <p class="text-lg font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">{{ totalTokensLabel }}</p>
              <p class="text-2xs text-gray-500 dark:text-dark-400">{{ t('dashboard.tokens') }}</p>
            </div>
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
                <tr v-for="(model, index) in models" :key="model.model" class="border-t border-gray-100 dark:border-gray-700/60">
                  <td class="max-w-[140px] py-2" :title="model.model">
                    <span class="flex items-center gap-1.5">
                      <!-- 图例色点：与环图分段同色，方便对照 -->
                      <span class="h-2 w-2 flex-shrink-0 rounded-full" :style="{ backgroundColor: segmentColor(index) }"></span>
                      <span class="truncate font-mono font-medium text-gray-900 dark:text-white">{{ model.model }}</span>
                    </span>
                  </td>
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

    <!-- Token Usage Trend Chart（品牌化折线图） -->
    <UserDashboardTrendChart :trend-data="trend" :loading="loading" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { Doughnut } from 'vue-chartjs'
import UserDashboardTrendChart from '@/components/user/dashboard/UserDashboardTrendChart.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{ loading: boolean, trend: TrendDataPoint[], models: ModelStat[] }>()
const { t } = useI18n()

// 环图品牌化色板：brand 橙领衔，emerald/sky/amber 语义色 + 同族浅档 + 中性灰兜底（全部取自 Tailwind 令牌）
const DOUGHNUT_PALETTE = ['#f97316', '#10b981', '#0ea5e9', '#f59e0b', '#fb923c', '#34d399', '#38bdf8', '#a1a1aa'] as const

const segmentColor = (index: number) => DOUGHNUT_PALETTE[index % DOUGHNUT_PALETTE.length]

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: props.models.map((_: ModelStat, i: number) => segmentColor(i)),
    borderWidth: 0,
    borderRadius: 4,
    spacing: 2,
    hoverOffset: 6
  }]
})

// 环图中心展示的 Token 总量
const totalTokensLabel = computed(() => {
  const total = (props.models || []).reduce((sum, m) => sum + (m.total_tokens || 0), 0)
  return formatTokens(total)
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '70%',
  plugins: {
    legend: { display: false },
    tooltip: {
      usePointStyle: true,
      backgroundColor: 'rgba(9,9,11,0.92)',
      titleColor: '#fafafa',
      bodyColor: '#e4e4e7',
      borderColor: 'rgba(255,255,255,0.08)',
      borderWidth: 1,
      padding: 12,
      cornerRadius: 10,
      boxPadding: 6,
      callbacks: { label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens` }
    }
  }
}
</script>

<style scoped>
.surface-card {
  @apply rounded-2xl border border-gray-200/70 bg-white shadow-card;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
</style>

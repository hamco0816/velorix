<!--
  用户仪表盘 Token 使用趋势图：品牌化配色的折线图。
  数据系列：输入=brand 橙、输出=emerald、缓存创建=amber、缓存读取=sky、缓存命中率=中性灰虚线（右轴百分比）。
  网格线中性灰、tooltip 深色，明暗模式均适配（跟随全局 useTheme 响应式切换）。
-->
<template>
  <div class="surface-card relative overflow-hidden">
    <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-dark-800/50">
      <LoadingSpinner size="md" />
    </div>
    <div class="border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
      <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('dashboard.tokenUsageTrend') }}</h3>
    </div>
    <div class="p-6">
      <div v-if="chartData" class="h-56">
        <Line :data="chartData" :options="lineOptions" />
      </div>
      <div v-else class="flex h-56 flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-gray-400">
        <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500">
          <Icon name="chart" size="md" />
        </div>
        {{ t('dashboard.noDataAvailable') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler } from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { useTheme } from '@/composables/useTheme'
import { formatCompactNumber } from '@/utils/format'
import type { TrendDataPoint } from '@/types'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const props = defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
}>()

const { t } = useI18n()
const { isDark } = useTheme()

// 品牌化系列色（对应 Tailwind 令牌：brand-500 / emerald-500 / amber-500 / sky-500 / zinc-500）
const SERIES_COLORS = {
  input: '#f97316',
  output: '#10b981',
  cacheCreation: '#f59e0b',
  cacheRead: '#0ea5e9',
  cacheHitRate: '#71717a'
} as const

const axisColors = computed(() => ({
  text: isDark.value ? '#d4d4d8' : '#3f3f46',
  muted: '#a1a1aa',
  grid: isDark.value ? 'rgba(255,255,255,0.06)' : 'rgba(9,9,11,0.06)'
}))

// 单条数据系列的公共外观：细线、隐点、柔和同色填充
const makeDataset = (label: string, data: number[], color: string) => ({
  label,
  data,
  borderColor: color,
  backgroundColor: `${color}1f`,
  fill: true,
  tension: 0.3
})

const chartData = computed(() => {
  if (!props.trendData?.length) return null
  return {
    labels: props.trendData.map((d) => d.date),
    datasets: [
      makeDataset(t('dashboard.input'), props.trendData.map((d) => d.input_tokens), SERIES_COLORS.input),
      makeDataset(t('dashboard.output'), props.trendData.map((d) => d.output_tokens), SERIES_COLORS.output),
      makeDataset(t('dashboard.cacheCreation'), props.trendData.map((d) => d.cache_creation_tokens), SERIES_COLORS.cacheCreation),
      makeDataset(t('dashboard.cacheRead'), props.trendData.map((d) => d.cache_read_tokens), SERIES_COLORS.cacheRead),
      {
        // 缓存命中率：辅助参考线，用中性灰虚线挂右轴，不与 token 量级混轴
        label: t('dashboard.cacheHitRate'),
        data: props.trendData.map((d) => {
          const total = d.cache_read_tokens + d.cache_creation_tokens
          return total > 0 ? (d.cache_read_tokens / total) * 100 : 0
        }),
        borderColor: SERIES_COLORS.cacheHitRate,
        backgroundColor: 'transparent',
        borderDash: [5, 5],
        fill: false,
        tension: 0.3,
        yAxisID: 'yPercent'
      }
    ]
  }
})

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  // 默认隐藏数据点、hover 时浮现；线条更细，曲线克制
  elements: {
    point: { radius: 0, hoverRadius: 4, hitRadius: 12, borderWidth: 2 },
    line: { borderWidth: 2 }
  },
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: axisColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: { size: 11 }
      }
    },
    tooltip: {
      usePointStyle: true,
      backgroundColor: 'rgba(9,9,11,0.92)',
      titleColor: '#fafafa',
      bodyColor: '#e4e4e7',
      footerColor: '#a1a1aa',
      borderColor: 'rgba(255,255,255,0.08)',
      borderWidth: 1,
      padding: 12,
      cornerRadius: 10,
      boxPadding: 6,
      callbacks: {
        label: (context: any) => {
          if (context.dataset.yAxisID === 'yPercent') {
            return `${context.dataset.label}: ${context.raw.toFixed(1)}%`
          }
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        },
        footer: (tooltipItems: any) => {
          const dataIndex = tooltipItems[0]?.dataIndex
          if (dataIndex !== undefined && props.trendData[dataIndex]) {
            const data = props.trendData[dataIndex]
            return `${t('dashboard.actual')}: $${formatCost(data.actual_cost)} | ${t('dashboard.standard')}: $${formatCost(data.cost)}`
          }
          return ''
        }
      }
    }
  },
  scales: {
    x: {
      grid: { display: false },
      border: { display: false },
      ticks: { color: axisColors.value.muted, font: { size: 10 } }
    },
    y: {
      grid: { color: axisColors.value.grid },
      border: { display: false },
      ticks: {
        color: axisColors.value.muted,
        font: { size: 10 },
        padding: 8,
        callback: (value: string | number) => formatTokens(Number(value))
      }
    },
    yPercent: {
      position: 'right' as const,
      min: 0,
      max: 100,
      grid: { drawOnChartArea: false },
      border: { display: false },
      ticks: {
        color: axisColors.value.muted,
        font: { size: 10 },
        callback: (value: string | number) => `${value}%`
      }
    }
  }
}))

// 走全局 formatCompactNumber，统一支持到 T/P 单位
const formatTokens = (value: number): string => formatCompactNumber(value, { decimals: 2 })

// 费用按量级自适应小数位：金额越小保留越多位，方便核对
const formatCost = (value: number): string => {
  if (value >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (value >= 1) return value.toFixed(2)
  if (value >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}
</script>

<style scoped>
.surface-card {
  @apply rounded-2xl border border-gray-200/70 bg-white shadow-card;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
</style>

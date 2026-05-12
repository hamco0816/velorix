<template>
  <section class="mb-4 md:mb-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <!-- 左侧概览：三段紧凑数字 — 整体状态 / 正常渠道 / 平均可用率 -->
      <div class="flex flex-wrap items-center gap-2">
        <span
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold ring-1 ring-inset"
          :class="overallChipClass"
        >
          <span class="w-1.5 h-1.5 rounded-full" :class="overallDotClass"></span>
          {{ overallLabel }}
        </span>
        <span
          v-if="totalCount > 0"
          class="inline-flex items-center gap-1 rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
        >
          <Icon name="server" size="xs" class="text-gray-400" />
          <span class="tabular-nums">{{ operationalCount }} / {{ totalCount }}</span>
          <span class="text-gray-400 dark:text-dark-400">{{ t('channelStatus.heroOperational') }}</span>
        </span>
        <span
          v-if="avgAvailability != null"
          class="inline-flex items-center gap-1 rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:ring-dark-700/60"
          :class="availabilityTextClass"
        >
          <Icon name="trendingUp" size="xs" class="opacity-70" />
          <span class="tabular-nums">{{ avgAvailability.toFixed(2) }}%</span>
          <span class="text-gray-400 dark:text-dark-400">{{ t('channelStatus.heroAvgAvailability') }}</span>
        </span>
      </div>

      <!-- 右侧 toolbar：时间窗口切换 + 刷新 + 自动刷新 -->
      <div class="flex items-center gap-2 flex-wrap">
        <div
          role="tablist"
          class="inline-flex p-0.5 rounded-xl bg-gray-100 dark:bg-dark-800 border border-gray-200/60 dark:border-dark-700/60 text-xs"
        >
          <button
            v-for="opt in windowOptions"
            :key="opt.value"
            type="button"
            role="tab"
            :aria-selected="window === opt.value"
            class="px-3 py-1 rounded-lg transition-colors"
            :class="window === opt.value
              ? 'bg-white dark:bg-dark-700 shadow-sm text-gray-900 dark:text-white font-semibold'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="emit('update:window', opt.value)"
          >
            {{ opt.label }}
          </button>
        </div>

        <button
          type="button"
          class="h-8 w-8 rounded-lg flex items-center justify-center text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-dark-700 transition-colors disabled:opacity-50"
          :disabled="loading"
          :title="t('common.refresh')"
          @click="emit('refresh')"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>

        <AutoRefreshButton
          v-if="autoRefresh"
          :enabled="autoRefresh.enabled.value"
          :interval-seconds="autoRefresh.intervalSeconds.value"
          :countdown="autoRefresh.countdown.value"
          :intervals="autoRefresh.intervals"
          @update:enabled="autoRefresh.setEnabled"
          @update:interval="autoRefresh.setInterval"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import AutoRefreshButton from '@/components/common/AutoRefreshButton.vue'
export type MonitorWindow = '7d' | '15d' | '30d'
export type OverallStatus = 'operational' | 'degraded'

const props = withDefaults(defineProps<{
  overallStatus: OverallStatus
  intervalSeconds: number
  window: MonitorWindow
  loading: boolean
  totalCount?: number
  operationalCount?: number
  avgAvailability?: number | null
  autoRefresh?: {
    enabled: { value: boolean }
    intervalSeconds: { value: number }
    countdown: { value: number }
    intervals: readonly number[]
    setEnabled: (v: boolean) => void
    setInterval: (v: number) => void
  }
}>(), {
  totalCount: 0,
  operationalCount: 0,
  avgAvailability: null,
})

const emit = defineEmits<{
  (e: 'update:window', value: MonitorWindow): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])

const overallLabel = computed(() => t(`channelStatus.overall.${props.overallStatus}`))

const overallChipClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-50 text-emerald-700 ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30'
    case 'degraded':
    default:
      return 'bg-amber-50 text-amber-700 ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30'
  }
})

const overallDotClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-500 animate-pulse'
    case 'degraded':
    default:
      return 'bg-amber-500 animate-pulse'
  }
})

// 平均可用率色阶：≥99 emerald / 95-99 sky / 90-95 amber / <90 rose
const availabilityTextClass = computed(() => {
  const v = props.avgAvailability
  if (v == null) return 'text-gray-700 dark:text-dark-200'
  if (v >= 99) return 'text-emerald-700 dark:text-emerald-300'
  if (v >= 95) return 'text-sky-700 dark:text-sky-300'
  if (v >= 90) return 'text-amber-700 dark:text-amber-300'
  return 'text-rose-700 dark:text-rose-300'
})
</script>

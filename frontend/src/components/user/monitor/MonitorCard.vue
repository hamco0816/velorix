<template>
  <button
    type="button"
    class="monitor-card group flex min-h-[280px] w-full flex-col p-5 text-left transition-all duration-300 ease-out hover:-translate-y-1"
    @click="emit('click')"
  >
    <!-- Header: icon + name/model + status chip -->
    <div class="flex items-start gap-3">
      <span
        class="w-9 h-9 rounded-xl ring-1 ring-black/5 dark:ring-white/10 grid place-items-center flex-shrink-0"
        :class="[providerGradient(item.provider), providerTintClass]"
      >
        <ProviderIcon :provider="item.provider" :size="20" />
      </span>
      <div class="flex-1 min-w-0">
        <div class="text-base font-semibold truncate text-gray-900 dark:text-gray-100">
          {{ item.name }}
        </div>
        <!-- 一行 chip 组：provider + 模型名 + group。把模型名也做成 chip（之前是裸文字），
             跟其他两个 chip 视觉层级一致，整行有规整韵律 -->
        <div class="mt-1 flex flex-wrap items-center gap-1.5 min-w-0">
          <span
            class="inline-flex items-center rounded-md px-1.5 py-0.5 text-2xs font-medium flex-shrink-0"
            :class="providerBadgeClass(item.provider)"
          >
            {{ providerLabel(item.provider) }}
          </span>
          <span
            class="inline-flex max-w-[16ch] items-center truncate rounded-md bg-gray-100 px-1.5 py-0.5 font-mono text-2xs font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-200"
            :title="item.primary_model"
          >
            {{ item.primary_model }}
          </span>
          <span
            v-if="item.group_name"
            class="inline-flex items-center rounded-md px-1.5 py-0.5 text-2xs font-medium bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300 flex-shrink-0"
          >
            {{ item.group_name }}
          </span>
        </div>
      </div>
      <span
        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold flex-shrink-0"
        :class="statusBadgeClass(item.primary_status)"
      >
        <span class="h-1.5 w-1.5 rounded-full" :class="statusDotClass(item.primary_status)"></span>
        {{ statusLabel(item.primary_status) }}
      </span>
    </div>

    <!-- Metrics -->
    <MonitorMetricPair
      primary-icon="bolt"
      :primary-label="t('monitorCommon.dialogLatency')"
      :primary-value="formatLatency(item.primary_latency_ms)"
      primary-unit="ms"
      secondary-icon="globe"
      :secondary-label="t('monitorCommon.endpointPing')"
      :secondary-value="formatLatency(item.primary_ping_latency_ms)"
      secondary-unit="ms"
    />

    <!-- Divider -->
    <div class="mt-4 border-t border-gray-100 dark:border-dark-700/60"></div>

    <!-- Availability row -->
    <MonitorAvailabilityRow
      :window-label="availabilityLabel"
      :value="availabilityValue"
      :samples-label="extraModelsCountLabel"
    />

    <!-- Timeline -->
    <MonitorTimeline
      :buckets="item.timeline"
      :countdown-seconds="cardCountdownSeconds"
    />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import {
  useChannelMonitorFormat,
  providerGradient,
} from '@/composables/useChannelMonitorFormat'
import { useNowTick } from '@/composables/useNowTick'
import ProviderIcon from './ProviderIcon.vue'
import MonitorMetricPair from './MonitorMetricPair.vue'
import MonitorAvailabilityRow from './MonitorAvailabilityRow.vue'
import MonitorTimeline from './MonitorTimeline.vue'

const PROVIDER_TINT: Record<string, string> = {
  openai: 'text-emerald-600 dark:text-emerald-300',
  anthropic: 'text-orange-600 dark:text-orange-300',
  gemini: 'text-sky-600 dark:text-sky-300',
}

const props = defineProps<{
  item: UserMonitorView
  window: '7d' | '15d' | '30d'
  availabilityValue: number | null
  /** 兜底倒计时（来自页面级 autoRefresh）。仅在监控数据缺少 interval/last_checked_at 时使用 */
  countdownSeconds: number
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const { t } = useI18n()
const {
  statusLabel,
  statusBadgeClass,
  statusDotClass,
  providerLabel,
  providerBadgeClass,
  formatLatency,
} = useChannelMonitorFormat()

// 每秒驱动倒计时重算：last_checked_at + interval_seconds - now
const now = useNowTick()
const cardCountdownSeconds = computed<number>(() => {
  const interval = props.item.interval_seconds
  const lastCheckedAt = props.item.last_checked_at
  // 没有探测过 / 没有间隔配置 → 回退到页面级倒计时
  if (!interval || interval <= 0 || !lastCheckedAt) return props.countdownSeconds
  const lastMs = Date.parse(lastCheckedAt)
  if (Number.isNaN(lastMs)) return props.countdownSeconds
  const nextProbeAt = lastMs + interval * 1000
  const remainingMs = nextProbeAt - now.value
  return Math.max(0, Math.ceil(remainingMs / 1000))
})

const providerTintClass = computed(() =>
  PROVIDER_TINT[props.item.provider] ?? 'text-gray-500 dark:text-gray-300'
)

const availabilityLabel = computed(() => {
  const win = t(`channelStatus.windowTab.${props.window}`)
  return `${t('monitorCommon.availabilityPrefix')} · ${win}`
})

const extraModelsCountLabel = computed(() => {
  const count = props.item.extra_models?.length ?? 0
  if (count === 0) return undefined
  return t('monitorCommon.extraModelsCount', { n: count })
})
</script>

<style scoped>
/* 渠道状态卡：中性白底（列表中 N 张卡，染色会让整页同色，状态色已通过 chip 表达） */
.monitor-card {
  border-radius: 1rem;
  border: 1px solid rgb(229 231 235 / 0.8);
  background: rgb(255 255 255 / 0.85);
  box-shadow: 0 1px 1px rgb(16 24 40 / 0.03), 0 2px 5px -1px rgb(16 24 40 / 0.05);
  backdrop-filter: blur(14px);
}
:root.dark .monitor-card {
  border-color: rgb(55 65 81 / 0.7);
  background: rgb(31 41 55 / 0.6);
}
.monitor-card:hover {
  border-color: rgb(209 213 219);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04), 0 12px 28px -16px rgb(15 23 42 / 0.18);
}
:root.dark .monitor-card:hover {
  border-color: rgb(99 102 241 / 0.4);
}
</style>

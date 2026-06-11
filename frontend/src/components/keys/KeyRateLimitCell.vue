<!-- 密钥列表「速率限制」单元格：5h/1d/7d 时间窗用量进度 + 重置倒计时 + 重置按钮 -->
<template>
  <div v-if="windows.length" class="min-w-[150px] space-y-1.5">
    <div v-for="w in windows" :key="w.label">
      <QuotaProgressBar :used="w.used" :limit="w.limit" :label="w.label" dense />
      <div
        v-if="w.resetAt && countdown(w.resetAt)"
        class="mt-0.5 flex items-center gap-1 text-2xs tabular-nums text-gray-400 dark:text-dark-500"
      >
        <Icon name="refresh" size="xs" />
        {{ countdown(w.resetAt) }}
      </div>
    </div>
    <button
      v-if="hasUsage"
      type="button"
      @click.stop="$emit('reset')"
      class="mt-0.5 inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-xs text-gray-500 transition-colors duration-150 hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-dark-700 dark:hover:text-white"
      :title="t('keys.resetRateLimitUsage')"
    >
      <Icon name="refresh" size="xs" />
      {{ t('keys.resetUsage') }}
    </button>
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import QuotaProgressBar from './QuotaProgressBar.vue'
import type { ApiKey } from '@/types'

const props = defineProps<{
  apiKey: ApiKey
  /** 当前时间：由父组件统一驱动（每分钟刷新），避免每行各开一个定时器 */
  now: Date
}>()

defineEmits<{ (e: 'reset'): void }>()

const { t } = useI18n()

// 三个时间窗配置：只渲染设置了限额的窗口
const windows = computed(() =>
  [
    { label: '5h', limit: props.apiKey.rate_limit_5h, used: props.apiKey.usage_5h || 0, resetAt: props.apiKey.reset_5h_at },
    { label: '1d', limit: props.apiKey.rate_limit_1d, used: props.apiKey.usage_1d || 0, resetAt: props.apiKey.reset_1d_at },
    { label: '7d', limit: props.apiKey.rate_limit_7d, used: props.apiKey.usage_7d || 0, resetAt: props.apiKey.reset_7d_at }
  ].filter((w) => w.limit > 0)
)

const hasUsage = computed(() => windows.value.some((w) => w.used > 0))

// 距离时间窗重置的剩余时间（如 2h 30m）
function countdown(resetAt: string | null): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - props.now.getTime()
  if (diff <= 0) return t('keys.resetNow')
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}
</script>

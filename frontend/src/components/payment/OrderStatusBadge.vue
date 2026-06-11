<template>
  <!-- 订单状态徽章：状态点 + 文字，语义色统一（成功 emerald / 处理中 sky / 待支付 amber / 失败 red） -->
  <span
    class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium"
    :class="entry.pill"
  >
    <span class="h-1.5 w-1.5 shrink-0 rounded-full" :class="entry.dot" aria-hidden="true"></span>
    {{ statusLabel }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { OrderStatus } from '@/types/payment'

const props = defineProps<{
  status: OrderStatus
}>()

const { t } = useI18n()

// 语义色五档：amber=待支付/待处理、sky=处理中、emerald=成功、red=失败、gray=终止；
// 退款完成保留 purple 作为"已退款"终态的功能性区分
const TONE = {
  amber: { pill: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300', dot: 'bg-amber-500' },
  sky: { pill: 'bg-sky-100 text-sky-800 dark:bg-sky-900/30 dark:text-sky-300', dot: 'bg-sky-500' },
  emerald: { pill: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
  red: { pill: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300', dot: 'bg-red-500' },
  gray: { pill: 'bg-gray-100 text-gray-700 dark:bg-dark-700/60 dark:text-dark-300', dot: 'bg-gray-400' },
  purple: { pill: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300', dot: 'bg-purple-500' },
} as const

type Tone = keyof typeof TONE

const statusMap: Record<OrderStatus, { key: string; tone: Tone }> = {
  PENDING: { key: 'payment.status.pending', tone: 'amber' },
  PAID: { key: 'payment.status.paid', tone: 'sky' },
  RECHARGING: { key: 'payment.status.recharging', tone: 'sky' },
  COMPLETED: { key: 'payment.status.completed', tone: 'emerald' },
  EXPIRED: { key: 'payment.status.expired', tone: 'gray' },
  CANCELLED: { key: 'payment.status.cancelled', tone: 'gray' },
  FAILED: { key: 'payment.status.failed', tone: 'red' },
  REFUND_REQUESTED: { key: 'payment.status.refund_requested', tone: 'amber' },
  REFUNDING: { key: 'payment.status.refunding', tone: 'sky' },
  REFUNDED: { key: 'payment.status.refunded', tone: 'purple' },
  PARTIALLY_REFUNDED: { key: 'payment.status.partially_refunded', tone: 'purple' },
  REFUND_FAILED: { key: 'payment.status.refund_failed', tone: 'red' },
}

const statusLabel = computed(() => {
  const item = statusMap[props.status]
  return item ? t(item.key) : props.status
})

const entry = computed(() => TONE[statusMap[props.status]?.tone ?? 'gray'])
</script>

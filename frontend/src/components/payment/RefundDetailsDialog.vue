<template>
  <BaseDialog :show="!!order" :title="t('payment.orders.refundDetailsTitle')" width="narrow" @close="$emit('close')">
    <div v-if="order" class="space-y-4">
      <!-- 顶部：金额大字 + 状态徽章；克制的 rose 软底替代彩虹渐变 -->
      <div class="rounded-xl border border-rose-200/70 bg-rose-50/60 px-5 py-4 dark:border-rose-500/30 dark:bg-rose-500/10">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-xs font-medium text-rose-600/80 dark:text-rose-300/80">
              {{ t('payment.orders.refundAmountLabel') }}
            </p>
            <div class="mt-1 flex items-end gap-1">
              <span class="text-sm font-semibold text-rose-600 dark:text-rose-300">¥</span>
              <span class="text-3xl font-semibold leading-none tabular-nums tracking-tight text-rose-600 dark:text-rose-300">
                {{ (order.refund_amount || 0).toFixed(2) }}
              </span>
            </div>
            <p v-if="isPartial" class="mt-1.5 text-2xs text-rose-500/80 dark:text-rose-300/70">
              {{ t('payment.orders.refundPartialNote', { paid: order.pay_amount.toFixed(2) }) }}
            </p>
          </div>
          <OrderStatusBadge :status="order.status" />
        </div>
      </div>

      <!-- 详情列表 -->
      <div class="rounded-xl border border-gray-200/70 bg-gray-50/40 px-4 py-3 dark:border-dark-700/60 dark:bg-dark-800/30">
        <dl class="divide-y divide-gray-100 text-sm dark:divide-dark-700/60">
          <div class="flex justify-between py-2.5">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</dt>
            <dd class="font-mono font-medium tabular-nums text-gray-900 dark:text-white">#{{ order.id }}</dd>
          </div>
          <div v-if="order.out_trade_no" class="flex justify-between gap-3 py-2.5">
            <dt class="shrink-0 text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderNo') }}</dt>
            <dd class="truncate font-mono text-xs text-gray-700 dark:text-gray-300">{{ order.out_trade_no }}</dd>
          </div>
          <div class="flex justify-between py-2.5">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.paymentMethod') }}</dt>
            <dd class="inline-flex items-center gap-1.5 text-gray-900 dark:text-white">
              <PaymentBrandIcon :type="order.payment_type" size="16px" />
              {{ paymentMethodLabel }}
              <span class="ml-1 text-2xs text-gray-400">{{ t('payment.orders.refundOriginalRoute') }}</span>
            </dd>
          </div>
          <div v-if="refundedAtDisplay" class="flex justify-between py-2.5">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.refundedAt') }}</dt>
            <dd class="tabular-nums text-gray-900 dark:text-white">{{ refundedAtDisplay }}</dd>
          </div>
          <div v-if="requestedAtDisplay" class="flex justify-between py-2.5">
            <dt class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.refundRequestedAt') }}</dt>
            <dd class="tabular-nums text-gray-900 dark:text-white">{{ requestedAtDisplay }}</dd>
          </div>
        </dl>
      </div>

      <!-- 退款原因 -->
      <div v-if="reasonText">
        <p class="mb-1.5 text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('payment.orders.refundReason') }}
        </p>
        <div class="rounded-xl border border-amber-200/70 bg-amber-50/60 px-4 py-3 text-sm leading-relaxed text-amber-900 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200">
          <div class="flex items-start gap-2.5">
            <Icon name="infoCircle" size="sm" class="mt-0.5 shrink-0 text-amber-500 dark:text-amber-400" :stroke-width="2.2" />
            <div>
              <span class="font-medium">{{ reasonHumanLabel }}</span>
              <p v-if="reasonHumanLabel !== reasonText" class="mt-1 text-xs text-amber-700/80 dark:text-amber-300/70">
                {{ reasonText }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 提示 -->
      <p class="text-center text-2xs text-gray-400 dark:text-gray-500">
        {{ t('payment.orders.refundProcessHint') }}
      </p>
    </div>
    <template #footer>
      <button class="btn btn-secondary" @click="$emit('close')">
        {{ t('common.close') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import BaseDialog from '@/components/common/BaseDialog.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import PaymentBrandIcon from '@/components/payment/PaymentBrandIcon.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ order: PaymentOrder | null }>()
defineEmits<{ close: [] }>()

const { t, te } = useI18n()

const isPartial = computed(() => props.order?.status === 'PARTIALLY_REFUNDED')

const paymentMethodLabel = computed(() => {
  if (!props.order) return ''
  return t(`payment.methods.${props.order.payment_type}`, props.order.payment_type)
})

// 系统码 → 用户友好文案；找不到对应 i18n 时直接展示原始原因
const reasonText = computed<string>(() => {
  return (props.order?.refund_reason || props.order?.refund_request_reason || '').trim()
})

const reasonHumanLabel = computed<string>(() => {
  const raw = reasonText.value
  if (!raw) return ''
  const key = `payment.orders.refundReasonCodes.${raw}`
  return te(key) ? t(key) : raw
})

function formatDateTime(s?: string | null): string {
  if (!s) return ''
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  return d.toLocaleString()
}

const refundedAtDisplay = computed(() => formatDateTime(props.order?.refund_at as string | undefined))
const requestedAtDisplay = computed(() => formatDateTime(props.order?.refund_requested_at))
</script>

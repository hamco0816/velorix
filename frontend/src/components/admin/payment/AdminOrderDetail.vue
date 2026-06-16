<template>
  <BaseDialog
    :show="show"
    :title="t('payment.admin.orderDetail')"
    width="wide"
    @close="emit('close')"
  >
    <div v-if="order" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</p>
          <p class="font-mono text-sm font-medium text-gray-900 dark:text-white">#{{ order.id }}</p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.status') }}</p>
          <span :class="['badge', statusBadgeClass(order.status)]">
            {{ t('payment.status.' + order.status.toLowerCase(), order.status) }}
          </span>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.baseAmount') }}</p>
          <p class="text-sm font-medium text-gray-900 dark:text-white">¥{{ baseAmount.toFixed(2) }}</p>
        </div>
        <div v-if="order.fee_rate > 0">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.fee') }} ({{ order.fee_rate }}%)</p>
          <p class="text-sm font-medium text-gray-900 dark:text-white">¥{{ feeAmount.toFixed(2) }}</p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.payAmount') }}</p>
          <p class="text-sm font-medium text-gray-900 dark:text-white">¥{{ order.pay_amount.toFixed(2) }}</p>
        </div>
        <div v-if="order.amount !== order.pay_amount">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.creditedAmount') }}</p>
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ order.order_type === 'balance' ? '$' : '¥' }}{{ order.amount.toFixed(2) }}</p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.paymentMethod') }}</p>
          <p class="inline-flex items-center gap-1.5 text-sm text-gray-700 dark:text-gray-300">
            <PaymentBrandIcon :type="order.payment_type" size="18px" />
            {{ t('payment.methods.' + order.payment_type, order.payment_type) }}
          </p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.orderType') }}</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">
            {{ t('payment.admin.' + order.order_type + 'Order', order.order_type) }}
          </p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.userId') }}</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">#{{ order.user_id }}</p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.createdAt') }}</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(order.created_at) }}</p>
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.expiresAt') }}</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(order.expires_at) }}</p>
        </div>
        <div v-if="order.paid_at">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.paidAt') }}</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(order.paid_at) }}</p>
        </div>
        <div v-if="order.completed_at">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.completedAt') }}</p>
          <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(order.completed_at) }}</p>
        </div>
      </div>

      <div
        v-if="order.refund_amount"
        class="rounded-lg border border-danger/30 bg-danger-soft p-3 dark:border-danger-deep/60 dark:bg-danger-deep/20"
      >
        <h4 class="mb-2 text-sm font-semibold text-danger-deep dark:text-danger">
          {{ t('payment.admin.refundInfo') }}
        </h4>
        <div class="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span class="text-danger dark:text-danger">{{ t('payment.admin.refundAmount') }}:</span>
            <span class="ml-1 font-medium text-danger-deep dark:text-danger-soft">{{ order.order_type === 'balance' ? '$' : '¥' }}{{ order.refund_amount.toFixed(2) }}</span>
          </div>
          <div v-if="order.refund_reason" class="col-span-2">
            <span class="text-danger dark:text-danger">{{ t('payment.admin.refundReason') }}:</span>
            <span class="ml-1 text-danger-deep dark:text-danger-soft">{{ order.refund_reason }}</span>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-end gap-2 border-t border-gray-200 pt-4 dark:border-dark-700">
        <button
          v-if="order.status === 'PENDING'"
          @click="emit('cancel', order)"
          class="btn btn-sm rounded-md bg-warning-soft px-3 py-1.5 text-sm text-warning hover:bg-warning-soft/70 dark:bg-warning-deep/20 dark:text-brand-300 dark:hover:bg-warning-deep/30"
        >
          {{ t('payment.orders.cancel') }}
        </button>
        <button
          v-if="order.status === 'FAILED'"
          @click="emit('retry', order)"
          class="btn btn-sm btn-secondary"
        >
          {{ t('payment.admin.retry') }}
        </button>
        <button
          v-if="canRefund(order)"
          @click="emit('refund', order)"
          class="btn btn-sm rounded-md bg-danger-soft px-3 py-1.5 text-sm text-danger hover:bg-danger-soft/70 dark:bg-danger-deep/20 dark:text-danger dark:hover:bg-danger-deep/30"
        >
          {{ t('payment.admin.refund') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PaymentOrder } from '@/types/payment'
import { statusBadgeClass, canRefund as canRefundStatus, formatOrderDateTime } from '@/components/payment/orderUtils'
import PaymentBrandIcon from '@/components/payment/PaymentBrandIcon.vue'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  order: PaymentOrder | null
}>()

/** 充值金额 (base amount before fee) = pay_amount - fee = pay_amount / (1 + fee_rate/100) */
const baseAmount = computed(() => {
  if (!props.order) return 0
  if (props.order.fee_rate <= 0) return props.order.pay_amount
  return props.order.pay_amount / (1 + props.order.fee_rate / 100)
})

/** 手续费 = pay_amount - baseAmount */
const feeAmount = computed(() => {
  if (!props.order || props.order.fee_rate <= 0) return 0
  return props.order.pay_amount - baseAmount.value
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'cancel', order: PaymentOrder): void
  (e: 'retry', order: PaymentOrder): void
  (e: 'refund', order: PaymentOrder): void
}>()

function canRefund(order: PaymentOrder): boolean {
  return canRefundStatus(order.status)
}

function formatDateTime(dateStr: string): string {
  return formatOrderDateTime(dateStr)
}
</script>

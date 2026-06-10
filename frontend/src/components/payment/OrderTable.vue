<template>
  <DataTable :columns="columns" :data="orders" :loading="loading" :error="error" @retry="$emit('retry')">
    <template #cell-id="{ value }">
      <span class="font-mono text-sm">#{{ value }}</span>
    </template>
    <template #cell-out_trade_no="{ value }">
      <span class="text-sm text-gray-900 dark:text-white">{{ value }}</span>
    </template>
    <template v-if="showUser" #cell-user_email="{ value, row }">
      <div class="text-sm">
        <span class="text-gray-900 dark:text-white">{{ value || row.user_name || '#' + row.user_id }}</span>
        <span v-if="row.user_notes" class="ml-1 text-xs text-gray-400">({{ row.user_notes }})</span>
      </div>
    </template>
    <template #cell-pay_amount="{ value, row }">
      <div class="text-sm">
        <span class="font-medium text-gray-900 dark:text-white">¥{{ value.toFixed(2) }}</span>
        <span v-if="row.fee_rate > 0" class="ml-1 text-xs text-gray-400" :title="t('payment.orders.fee') + ': ' + row.fee_rate + '%'">
          ({{ t('payment.orders.fee') }} {{ row.fee_rate }}%)
        </span>
        <div v-if="row.amount !== row.pay_amount" class="text-xs text-gray-500">
          {{ t('payment.orders.creditedAmount') }}: {{ row.order_type === 'balance' ? '$' : '¥' }}{{ row.amount.toFixed(2) }}
        </div>
      </div>
    </template>
    <template #cell-payment_type="{ value }">
      <span class="inline-flex items-center gap-1.5 text-sm text-gray-700 dark:text-gray-300">
        <PaymentBrandIcon :type="value" size="18px" />
        {{ t('payment.methods.' + value, value) }}
      </span>
    </template>
    <template #cell-status="{ value, row }">
      <div class="inline-flex items-center gap-1.5">
        <OrderStatusBadge :status="value" />
        <button
          v-if="isRefundStatus(value) && row.refund_amount > 0"
          type="button"
          class="flex h-5 w-5 items-center justify-center rounded-full bg-rose-50 text-rose-500 transition-colors hover:bg-rose-100 hover:text-rose-600 dark:bg-rose-900/30 dark:text-rose-300 dark:hover:bg-rose-900/50"
          :title="t('payment.orders.viewRefund')"
          @click="$emit('inspect-refund', row)"
        >
          <Icon name="infoCircle" size="xs" :stroke-width="2.5" />
        </button>
      </div>
    </template>
    <template #cell-created_at="{ value }">
      <span class="text-xs text-gray-500 dark:text-gray-400">{{ formatDate(value) }}</span>
    </template>
    <template #cell-actions="{ row }">
      <slot name="actions" :row="row" />
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PaymentOrder } from '@/types/payment'
import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import PaymentBrandIcon from '@/components/payment/PaymentBrandIcon.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  orders: PaymentOrder[]
  loading: boolean
  showUser?: boolean
  /** 父页面加载订单失败时为 true，表格显示错误态 */
  error?: boolean
}>(), {
  showUser: false,
  error: false,
})

// retry：错误态点击重试时向父页面转发
defineEmits<{ 'inspect-refund': [order: PaymentOrder]; retry: [] }>()

function isRefundStatus(status: string): boolean {
  return status === 'REFUNDED' || status === 'PARTIALLY_REFUNDED' || status === 'REFUND_REQUESTED' || status === 'REFUNDING' || status === 'REFUND_FAILED'
}

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString() }

const columns = computed((): Column[] => {
  const cols: Column[] = [
    { key: 'id', label: t('payment.orders.orderId') },
    { key: 'out_trade_no', label: t('payment.orders.orderNo') },
  ]
  if (props.showUser) {
    cols.push({ key: 'user_email', label: t('payment.admin.colUser') })
  }
  cols.push(
    { key: 'pay_amount', label: t('payment.orders.payAmount'), numeric: true },
    { key: 'payment_type', label: t('payment.orders.paymentMethod'), align: 'center' },
    { key: 'status', label: t('payment.orders.status'), align: 'center' },
    { key: 'created_at', label: t('payment.orders.createdAt') },
    { key: 'actions', label: t('common.actions'), align: 'center' },
  )
  return cols
})
</script>

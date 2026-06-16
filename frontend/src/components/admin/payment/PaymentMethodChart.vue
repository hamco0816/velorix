<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
      {{ t('payment.admin.paymentDistribution') }}
    </h3>
    <div
      v-if="!methods?.length"
      class="flex h-32 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else class="space-y-3">
      <div v-for="method in methods" :key="method.type" class="space-y-1">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <PaymentBrandIcon :type="method.type" size="18px" />
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ t('payment.methods.' + method.type, method.type) }}
            </span>
          </div>
          <div class="text-right">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              ¥{{ method.amount.toFixed(2) }}
            </span>
            <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">
              ({{ method.count }})
            </span>
          </div>
        </div>
        <div class="h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
          <div
            :class="['h-full rounded-full transition-all', barColorMap[method.type] || 'bg-gray-400']"
            :style="{ width: barWidth(method.amount) + '%' }"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PaymentBrandIcon from '@/components/payment/PaymentBrandIcon.vue'

const { t } = useI18n()

const props = defineProps<{
  methods: { type: string; amount: number; count: number }[]
}>()

// 进度条按支付渠道使用各品牌识别真彩（支付宝蓝 / 微信绿 / Stripe 紫），与全站收银台保持一致
const barColorMap: Record<string, string> = {
  alipay: 'bg-[#00AEEF]',
  wxpay: 'bg-[#2BB741]',
  alipay_direct: 'bg-[#00AEEF]/70',
  wxpay_direct: 'bg-[#2BB741]/70',
  stripe: 'bg-[#635bff]',
}

const maxAmount = computed(() => {
  if (!props.methods?.length) return 1
  return Math.max(...props.methods.map(m => m.amount), 1)
})

function barWidth(amount: number): number {
  return Math.min((amount / maxAmount.value) * 100, 100)
}
</script>

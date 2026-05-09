<template>
  <div>
    <label class="mb-3 block text-sm font-semibold text-gray-800 dark:text-gray-100">
      {{ t('payment.paymentMethod') }}
    </label>
    <div :class="['grid gap-3', layout === 'list' ? 'grid-cols-1' : 'grid-cols-1 sm:grid-cols-2']">
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :disabled="!method.available"
        :class="[
          'relative flex min-h-[64px] items-center gap-3 rounded-xl border px-3.5 py-2.5 text-left transition-colors',
          !method.available
            ? 'cursor-not-allowed border-gray-200 bg-gray-50 opacity-50 dark:border-dark-700 dark:bg-dark-800/50'
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'border-gray-300 bg-white text-gray-700 hover:border-gray-400 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-500',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span class="flex h-9 w-9 shrink-0 items-center justify-center">
          <PaymentBrandIcon
            :type="method.type"
            :alt="t(`payment.methods.${method.type}`)"
            size="32px"
          />
        </span>
        <span class="min-w-0 flex-1">
          <span class="block truncate text-sm font-semibold leading-tight">{{ t(`payment.methods.${method.type}`) }}</span>
          <span class="mt-0.5 block truncate text-[11px] leading-tight text-gray-500 dark:text-dark-400">
            <template v-if="method.fee_rate > 0">{{ t('payment.fee') }} {{ method.fee_rate }}%</template>
            <template v-else>{{ method.available ? t('common.available') : t('common.notAvailable') }}</template>
          </span>
        </span>
        <span
          v-if="selected === method.type"
          class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-gray-900 text-white dark:bg-white dark:text-gray-950"
        >
          <Icon name="check" size="xs" :stroke-width="3" />
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { METHOD_ORDER } from './providerConfig'
import PaymentBrandIcon from './PaymentBrandIcon.vue'

export interface PaymentMethodOption {
  type: string
  fee_rate: number
  available: boolean
}

const props = withDefaults(defineProps<{
  methods: PaymentMethodOption[]
  selected: string
  layout?: 'grid' | 'list'
}>(), {
  layout: 'grid',
})

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

function methodSelectedClass(type: string): string {
  if (type.includes('alipay')) return 'border-[#02A9F1] bg-blue-50 text-gray-900 shadow-sm ring-1 ring-[#02A9F1]/20 dark:bg-blue-950/50 dark:text-gray-100'
  if (type.includes('wxpay')) return 'border-[#09BB07] bg-green-50 text-gray-900 shadow-sm ring-1 ring-[#09BB07]/20 dark:bg-green-950/50 dark:text-gray-100'
  if (type === 'stripe') return 'border-[#676BE5] bg-indigo-50 text-gray-900 shadow-sm ring-1 ring-[#676BE5]/20 dark:bg-indigo-950/50 dark:text-gray-100'
  return 'border-primary-500 bg-primary-50 text-gray-900 shadow-sm ring-1 ring-primary-500/20 dark:bg-primary-950/50 dark:text-gray-100'
}
</script>

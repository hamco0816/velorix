<template>
  <div
    :class="[
      'group relative rounded-lg border transition-all',
      enabled ? 'border-gray-200 dark:border-dark-600' : 'border-gray-200 bg-gray-50 opacity-50 dark:border-dark-700 dark:bg-dark-800/50',
    ]"
    :title="!enabled ? t('admin.settings.payment.typeDisabled') + ' — ' + t('admin.settings.payment.enableTypesFirst') : undefined"
  >
    <div :class="[
      'flex items-center justify-between gap-4 px-4 py-3',
      !enabled && 'pointer-events-none',
    ]">
      <!-- Left: icon + name + key badge + type badges -->
      <div class="flex min-w-0 flex-1 flex-wrap items-center gap-3">
        <div :class="[
          'flex h-8 w-8 shrink-0 items-center justify-center rounded-md',
          provider.enabled && enabled ? 'bg-green-100 dark:bg-green-900/30' : 'bg-gray-100 dark:bg-dark-700',
        ]">
          <PaymentBrandIcon
            v-if="providerBrandIcon"
            :type="provider.provider_key"
            size="18px"
          />
          <Icon
            v-else
            name="server"
            size="sm"
            :class="provider.enabled && enabled ? 'text-green-600 dark:text-green-400' : 'text-gray-400'"
          />
        </div>
        <span class="text-sm font-medium text-gray-900 dark:text-white">{{ provider.name }}</span>
        <span class="text-xs text-gray-400 dark:text-gray-500">{{ keyLabel }}</span>
        <span v-if="provider.payment_mode" class="text-xs text-gray-400 dark:text-gray-500">· {{ modeLabel }}</span>
        <span v-if="enabled && availableTypes.length" class="text-xs text-gray-300 dark:text-gray-600">|</span>
        <div v-if="enabled" class="flex flex-wrap items-center gap-1.5">
          <button
            v-for="pt in availableTypes"
            :key="pt.value"
            type="button"
            @click="emit('toggleType', pt.value)"
            :class="typeBadgeClass(pt.value)"
          >
            <PaymentBrandIcon :type="pt.value" size="14px" />
            <span>{{ pt.label }}</span>
          </button>
        </div>
      </div>

      <!-- Right: toggles + actions -->
      <div class="flex shrink-0 items-center justify-end gap-4">
        <ToggleSwitch :label="t('common.enabled')" :checked="provider.enabled" @toggle="emit('toggleField', 'enabled')" />
        <ToggleSwitch :label="t('admin.settings.payment.refundEnabled')" :checked="provider.refund_enabled" @toggle="emit('toggleField', 'refund_enabled')" />
        <ToggleSwitch v-if="provider.refund_enabled" :label="t('admin.settings.payment.allowUserRefund')" :checked="provider.allow_user_refund" @toggle="emit('toggleField', 'allow_user_refund')" />
        <div class="flex items-center gap-1.5 border-l border-gray-200 pl-3 dark:border-dark-600">
          <button
            type="button"
            @click="emit('edit')"
            class="inline-flex h-9 w-9 items-center justify-center rounded-md text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-dark-700 dark:hover:text-white"
            :title="t('common.edit')"
          >
            <Icon name="edit" size="sm" />
          </button>
          <button
            type="button"
            @click="emit('delete')"
            class="inline-flex h-9 w-9 items-center justify-center rounded-md text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
            :title="t('common.delete')"
          >
            <Icon name="trash" size="sm" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import ToggleSwitch from './ToggleSwitch.vue'
import PaymentBrandIcon from './PaymentBrandIcon.vue'
import type { ProviderInstance } from '@/types/payment'
import type { TypeOption } from './providerConfig'
import { PAYMENT_MODE_QRCODE, PAYMENT_MODE_POPUP } from './providerConfig'
import { hasPaymentBrandIcon } from './paymentBrand'

const PROVIDER_KEY_LABELS: Record<string, string> = {
  easypay: 'admin.settings.payment.providerEasypay',
  xunhupay: 'admin.settings.payment.providerXunhupay',
  alipay: 'admin.settings.payment.providerAlipay',
  wxpay: 'admin.settings.payment.providerWxpay',
  stripe: 'admin.settings.payment.providerStripe',
}

const props = defineProps<{
  provider: ProviderInstance
  enabled: boolean
  availableTypes: TypeOption[]
}>()

const emit = defineEmits<{
  toggleField: [field: 'enabled' | 'refund_enabled' | 'allow_user_refund']
  toggleType: [type: string]
  edit: []
  delete: []
}>()

const { t } = useI18n()

const keyLabel = computed(() => t(PROVIDER_KEY_LABELS[props.provider.provider_key] || props.provider.provider_key))
const providerBrandIcon = computed(() => hasPaymentBrandIcon(props.provider.provider_key))

const modeLabel = computed(() => {
  if (props.provider.payment_mode === PAYMENT_MODE_QRCODE) return t('admin.settings.payment.modeQRCode')
  if (props.provider.payment_mode === PAYMENT_MODE_POPUP) return t('admin.settings.payment.modePopup')
  return ''
})

function isSelected(type: string): boolean {
  return props.provider.supported_types.includes(type)
}

// 支付类型 chip：选中态统一品牌橙描边 + 浅橙底（品牌色只出现在 chip 内的品牌图标上）
function typeBadgeClass(type: string): string {
  const base = 'inline-flex h-6 items-center gap-1 rounded-full border px-2 text-xs font-semibold transition-colors duration-150 ease-out'
  if (!isSelected(type)) {
    return `${base} border-gray-200 bg-gray-50 text-gray-400 hover:border-gray-300 hover:bg-white dark:border-dark-700 dark:bg-dark-800 dark:text-gray-500 dark:hover:border-dark-600`
  }
  return `${base} border-brand-500 bg-brand-50 text-brand-800 dark:border-brand-400 dark:bg-brand-500/10 dark:text-brand-300`
}
</script>

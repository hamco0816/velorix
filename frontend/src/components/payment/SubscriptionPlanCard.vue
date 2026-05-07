<template>
  <div
    :class="[
      'group relative flex min-h-[250px] flex-col overflow-hidden rounded-xl border transition-colors',
      'bg-white shadow-sm hover:border-gray-300 hover:shadow-md dark:bg-dark-900 dark:shadow-none dark:hover:border-dark-500',
      borderClass,
    ]"
  >
    <div :class="['h-1.5', accentClass]" />

    <div class="flex flex-1 flex-col p-5">
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div class="min-w-0 flex-1 space-y-3">
          <div class="flex items-center gap-2">
            <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-50 ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-600">
              <BrandIcon
                v-if="platformBrand"
                :brand="platformBrand"
                size="22px"
              />
              <span v-else class="text-sm font-bold text-gray-500 dark:text-gray-300">{{ platformInitial }}</span>
            </span>
            <div class="min-w-0">
              <h3 class="truncate text-lg font-bold text-gray-900 dark:text-white">{{ plan.name }}</h3>
              <span :class="['mt-1 inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium', badgeLightClass]">
                {{ pLabel }}
              </span>
            </div>
          </div>
          <p v-if="plan.description" class="line-clamp-2 text-sm leading-relaxed text-gray-500 dark:text-dark-400">
            {{ plan.description }}
          </p>
        </div>

        <div class="flex shrink-0 items-center justify-between gap-4 rounded-xl bg-slate-50 px-4 py-3 text-right ring-1 ring-gray-100 dark:bg-dark-800/70 dark:ring-dark-700 sm:min-w-[176px] sm:flex-col sm:items-stretch sm:justify-center sm:gap-2">
          <span class="text-xs font-medium text-gray-400 dark:text-dark-500 sm:hidden">{{ t('payment.admin.price') }}</span>
          <div class="flex flex-col items-end sm:items-stretch">
            <div class="flex items-end justify-end gap-1.5 whitespace-nowrap sm:justify-center">
              <span :class="['mb-1 text-xl font-black leading-none', textClass]">¥</span>
              <span :class="['text-4xl font-black leading-none tracking-tight', textClass]">{{ plan.price }}</span>
            </div>
            <div class="mt-2 flex flex-wrap items-center justify-end gap-2 sm:justify-center">
              <span class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-full bg-white px-2.5 py-1 text-xs font-semibold text-gray-600 ring-1 ring-slate-200 dark:bg-dark-900 dark:text-dark-200 dark:ring-dark-600">
                <Icon name="calendar" size="xs" :stroke-width="2" />
                {{ validitySuffix }}
              </span>
              <span v-if="plan.original_price" class="inline-flex items-center gap-1.5">
                <span class="text-sm text-gray-400 line-through dark:text-dark-500">¥{{ plan.original_price }}</span>
                <span v-if="discountText" :class="['rounded-full px-2 py-0.5 text-[11px] font-bold', discountClass]">{{ discountText }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="mb-4 space-y-3 rounded-xl border border-gray-100 bg-slate-50/70 p-3 dark:border-dark-700 dark:bg-dark-800/45">
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <div class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.rate') }}</span>
          <span :class="['mt-1 block text-base font-bold', textClass]">{{ rateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="mt-1 block text-base font-bold text-gray-900 dark:text-white">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="mt-1 block text-base font-bold text-gray-900 dark:text-white">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="mt-1 block text-base font-bold text-gray-900 dark:text-white">${{ plan.monthly_limit_usd }}</span>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.quota') }}</span>
          <span class="mt-1 block text-base font-bold text-gray-900 dark:text-white">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        </div>

        <div v-if="modelScopeItems.length > 0" class="rounded-lg bg-white px-3 py-3 dark:bg-dark-900/70">
          <div class="mb-2 flex items-center justify-between gap-3">
            <span class="text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.models') }}</span>
            <span class="h-px flex-1 bg-slate-200 dark:bg-dark-600" />
          </div>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="scope in modelScopeItems"
              :key="scope.key"
              class="inline-flex items-center gap-1.5 rounded-full bg-white px-2.5 py-1.5 text-xs font-semibold text-gray-700 shadow-sm ring-1 ring-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:ring-dark-600"
            >
              <ModelIcon :model="scope.iconModel" size="14px" />
              {{ scope.label }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="plan.features.length > 0" class="mb-3 space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-start gap-1.5">
          <Icon name="check" size="xs" :stroke-width="2.5" :class="['mt-0.5 flex-shrink-0', iconClass]" />
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <div class="mt-4">
        <button
          type="button"
          :class="['inline-flex w-full items-center justify-center rounded-xl px-6 py-2.5 text-sm font-semibold transition-colors active:scale-[0.99]', btnClass]"
          @click="emit('select', plan)"
        >
          {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import BrandIcon from '@/components/common/BrandIcon.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformBorderClass,
  platformTextClass,
  platformIconClass,
  platformButtonClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

// Derived color classes from central config
const accentClass = computed(() => platformAccentBarClass(platform.value))
const borderClass = computed(() => platformBorderClass(platform.value))
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const btnClass = computed(() => platformButtonClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))
const platformBrand = computed<'claude' | 'openai' | 'gemini' | null>(() => {
  if (platform.value === 'anthropic') return 'claude'
  if (platform.value === 'openai') return 'openai'
  if (platform.value === 'gemini') return 'gemini'
  return null
})
const platformInitial = computed(() => pLabel.value.charAt(0).toUpperCase())

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const MODEL_SCOPE_META: Record<string, { label: string; iconModel: string }> = {
  claude: { label: 'Claude', iconModel: 'claude-3-5-sonnet' },
  gemini_text: { label: 'Gemini', iconModel: 'gemini-2.5-pro' },
  gemini_image: { label: 'Imagen', iconModel: 'imagen-3' },
}

const modelScopeItems = computed(() => {
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => ({
    key: s,
    label: MODEL_SCOPE_META[s]?.label || s,
    iconModel: MODEL_SCOPE_META[s]?.iconModel || s,
  }))
})

const validitySuffix = computed(() => {
  const u = props.plan.validity_unit || 'day'
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${props.plan.validity_days}${t('payment.days')}`
})
</script>

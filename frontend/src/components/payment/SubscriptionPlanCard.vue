<template>
  <div
    :class="[
      'group relative flex min-h-[280px] flex-col overflow-hidden rounded-2xl border transition-all',
      'bg-white shadow-sm shadow-gray-200/60 hover:-translate-y-0.5 hover:shadow-xl dark:bg-dark-800 dark:shadow-none',
      borderClass,
    ]"
  >
    <div :class="['h-2', accentClass]" />

    <div class="flex flex-1 flex-col bg-gradient-to-br from-white via-white to-slate-50/80 p-5 dark:from-dark-800 dark:via-dark-800 dark:to-dark-900">
      <div class="mb-5 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div class="min-w-0 flex-1 space-y-3">
          <div class="flex items-center gap-2">
            <span class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-white shadow-sm ring-1 ring-gray-200 dark:bg-dark-700 dark:ring-dark-600">
              <BrandIcon
                v-if="platformBrand"
                :brand="platformBrand"
                size="23px"
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

        <div class="flex shrink-0 items-end justify-between rounded-2xl bg-white/90 px-4 py-3 text-right shadow-sm ring-1 ring-gray-100 dark:bg-dark-700/70 dark:ring-dark-600 sm:block sm:min-w-[126px]">
          <span class="text-xs font-medium text-gray-400 dark:text-dark-500 sm:hidden">{{ t('payment.admin.price') }}</span>
          <div class="flex items-baseline justify-end gap-1">
            <span class="text-xs font-semibold text-gray-400 dark:text-dark-500">$</span>
            <span :class="['text-3xl font-extrabold tracking-tight', textClass]">{{ plan.price }}</span>
          </div>
          <span class="block text-xs text-gray-400 dark:text-dark-500">/ {{ validitySuffix }}</span>
          <div v-if="plan.original_price" class="mt-0.5 flex items-center justify-end gap-1.5">
            <span class="text-xs text-gray-400 line-through dark:text-dark-500">${{ plan.original_price }}</span>
            <span :class="['rounded px-1 py-0.5 text-[10px] font-semibold', discountClass]">{{ discountText }}</span>
          </div>
        </div>
      </div>

      <div class="mb-4 grid grid-cols-2 gap-2 rounded-2xl bg-white/80 p-3 text-xs ring-1 ring-gray-100 dark:bg-dark-700/45 dark:ring-dark-600">
        <div class="rounded-xl bg-slate-50 px-3 py-2 dark:bg-dark-800/80">
          <span class="block text-gray-400 dark:text-dark-500">{{ t('payment.planCard.rate') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ rateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="rounded-xl bg-slate-50 px-3 py-2 dark:bg-dark-800/80">
          <span class="block text-gray-400 dark:text-dark-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="rounded-xl bg-slate-50 px-3 py-2 dark:bg-dark-800/80">
          <span class="block text-gray-400 dark:text-dark-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="rounded-xl bg-slate-50 px-3 py-2 dark:bg-dark-800/80">
          <span class="block text-gray-400 dark:text-dark-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.monthly_limit_usd }}</span>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="rounded-xl bg-slate-50 px-3 py-2 dark:bg-dark-800/80">
          <span class="block text-gray-400 dark:text-dark-500">{{ t('payment.planCard.quota') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        <div v-if="modelScopeItems.length > 0" class="col-span-2 rounded-xl bg-slate-50 px-3 py-2 dark:bg-dark-800/80">
          <span class="mb-2 block text-gray-400 dark:text-dark-500">{{ t('payment.planCard.models') }}</span>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="scope in modelScopeItems"
              :key="scope.key"
              class="inline-flex items-center gap-1.5 rounded-full bg-gray-100 px-2 py-1 text-[11px] font-semibold text-gray-700 ring-1 ring-gray-200 dark:bg-dark-700 dark:text-gray-200 dark:ring-dark-600"
            >
              <ModelIcon :model="scope.iconModel" size="14px" />
              {{ scope.label }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="plan.features.length > 0" class="mb-3 space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-start gap-1.5">
          <svg :class="['mt-0.5 h-3.5 w-3.5 flex-shrink-0', iconClass]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
          </svg>
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <button
        type="button"
        :class="['mt-2 w-full rounded-xl py-3 text-sm font-semibold transition-all active:scale-[0.98]', btnClass]"
        @click="emit('select', plan)"
      >
        {{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}
      </button>
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

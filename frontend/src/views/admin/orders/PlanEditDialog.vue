<template>
  <BaseDialog :show="show" :title="plan ? t('payment.admin.editPlan') : t('payment.admin.createPlan')" width="wide" @close="emit('close')">
    <form id="plan-form" @submit.prevent="handleSavePlan" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('payment.admin.planName') }} <span class="text-danger">*</span></label>
          <input v-model="planForm.name" type="text" class="input" required />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.group') }} <span class="text-danger">*</span></label>
          <Select v-model="planForm.group_id" :options="groupOptions" :placeholder="t('payment.admin.selectGroup')" class="w-full">
            <template #selected="{ option }">
              <span v-if="option?.platform" :class="platformTextClass(String(option.platform))">{{ option.label }}</span>
              <span v-else>{{ option?.label || t('payment.admin.selectGroup') }}</span>
            </template>
            <template #option="{ option, selected }">
              <span class="flex-1 truncate text-left" :class="option.platform ? platformTextClass(String(option.platform)) : ''">{{ option.label }}</span>
              <Icon v-if="selected" name="check" size="sm" class="text-primary-500" :stroke-width="2" />
            </template>
          </Select>
        </div>
      </div>

      <!-- Group Info Preview -->
      <div v-if="selectedGroupInfo" class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-800">
        <div class="mb-2 flex items-center gap-2">
          <GroupBadge :name="selectedGroupInfo.name" :platform="selectedGroupInfo.platform" :rate-multiplier="selectedGroupInfo.rate_multiplier" />
        </div>
        <div class="grid grid-cols-2 gap-2 text-xs">
          <div><span class="text-gray-500">{{ t('payment.admin.dailyLimit') }}:</span> <span class="ml-1 font-medium text-gray-700 dark:text-gray-300">{{ selectedGroupInfo.daily_limit_usd != null ? '$' + selectedGroupInfo.daily_limit_usd : t('payment.admin.unlimited') }}</span></div>
          <div><span class="text-gray-500">{{ t('payment.admin.weeklyLimit') }}:</span> <span class="ml-1 font-medium text-gray-700 dark:text-gray-300">{{ selectedGroupInfo.weekly_limit_usd != null ? '$' + selectedGroupInfo.weekly_limit_usd : t('payment.admin.unlimited') }}</span></div>
          <div><span class="text-gray-500">{{ t('payment.admin.monthlyLimit') }}:</span> <span class="ml-1 font-medium text-gray-700 dark:text-gray-300">{{ selectedGroupInfo.monthly_limit_usd != null ? '$' + selectedGroupInfo.monthly_limit_usd : t('payment.admin.unlimited') }}</span></div>
        </div>
      </div>

      <div><label class="input-label">{{ t('payment.admin.planDescription') }} <span class="text-danger">*</span></label><textarea v-model="planForm.description" rows="2" class="input" required></textarea></div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="input-label">{{ t('payment.admin.price') }} <span class="text-danger">*</span></label><input v-model.number="planForm.price" type="number" step="0.01" min="0.01" class="input" required @input="handleCostPriceInput" /></div>
        <div><label class="input-label">{{ t('payment.admin.originalPrice') }}</label><input v-model.number="planForm.original_price" type="number" step="0.01" min="0" class="input" /></div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="input-label">{{ t('payment.admin.validityDays') }} <span class="text-danger">*</span></label><input v-model.number="planForm.validity_days" type="number" min="1" class="input" required /></div>
        <div><label class="input-label">{{ t('payment.admin.validityUnit') }} <span class="text-danger">*</span></label><Select v-model="planForm.validity_unit" :options="validityUnitOptions" /></div>
      </div>
      <!-- 档位名 + 档位样式：决定对比表的列头标识（图标 + 配色，越高越豪华）-->
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('payment.admin.planLabel') }}</label>
          <input v-model.trim="planForm.plan_label" type="text" maxlength="24" class="input" :placeholder="t('payment.admin.planLabelPlaceholder')" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.planLabelHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.tierStyle.label') }}</label>
          <Select v-model="planForm.tier_style" :options="tierStyleOptions" />
          <div class="mt-1 flex items-center gap-2">
            <span :class="['h-3.5 w-3.5 rounded-full', tierSwatchClass(planForm.tier_style)]" />
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.tierStyle.hint') }}</span>
          </div>
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('payment.admin.planBadgeText') }}</label>
        <input v-model.trim="planForm.badge_text" type="text" maxlength="12" class="input" :placeholder="t('payment.admin.planBadgePlaceholder')" />
        <div class="mt-1 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
          <span>{{ t('payment.admin.planBadgeHint') }}</span>
          <span
            v-if="planForm.badge_text"
            :class="['inline-flex items-center rounded-full px-2 py-0.5 text-2xs font-bold tracking-wider shadow-sm', badgeToneClass(planForm.badge_color)]"
          >
            {{ planForm.badge_text }}
          </span>
        </div>
        <!-- 角标配色：尊贵预设色板，仅在填了角标文字时才需要选 -->
        <div v-if="planForm.badge_text" class="mt-2.5 flex flex-wrap items-center gap-2.5">
          <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.badgeColor.label') }}</span>
          <button
            v-for="tone in BADGE_TONE_KEYS"
            :key="tone"
            type="button"
            :title="t(`payment.admin.badgeColor.${tone}`)"
            :class="[
              'h-6 w-6 rounded-full transition-transform hover:scale-110',
              badgeToneSwatchClass(tone),
              planForm.badge_color === tone
                ? 'ring-2 ring-offset-2 ring-gray-900 dark:ring-white dark:ring-offset-dark-800'
                : 'ring-1 ring-black/10 dark:ring-white/15',
            ]"
            @click="planForm.badge_color = tone"
          />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="input-label">{{ t('payment.admin.sortOrder') }}</label><input v-model.number="planForm.sort_order" type="number" min="0" class="input" /></div>
        <div>
          <label class="input-label">{{ t('payment.admin.planKind') }} <span class="text-danger">*</span></label>
          <Select v-model="planForm.kind" :options="planKindOptions" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ planForm.kind === 'exclusive' ? t('payment.admin.planKindExclusiveHint') : t('payment.admin.planKindSharedHint') }}
          </p>
        </div>
      </div>
      <div>
        <label class="input-label">{{ t('payment.admin.features') }}</label>
        <textarea v-model="planFeaturesText" rows="3" class="input" :placeholder="t('payment.admin.featuresPlaceholder')"></textarea>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.featuresHint') }}</p>
      </div>

      <!-- 套餐级限额/倍率覆盖：留空 = 沿用 group 默认；填值 = 覆盖（同 group 下做差异化档位） -->
      <div class="rounded-lg border border-gray-200 bg-gray-50/60 p-4 dark:border-dark-700 dark:bg-dark-800/40">
        <div class="mb-3 flex items-start justify-between gap-3">
          <div class="flex-1">
            <p class="text-sm font-semibold text-gray-900 dark:text-gray-50">
              {{ t('payment.admin.limitOverrideTitle') }}
            </p>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-300">
              {{ t('payment.admin.limitOverrideHint') }}
            </p>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="input-label">{{ t('payment.admin.planDailyLimitUSD') }}</label>
            <input v-model.number="planForm.daily_limit_usd" type="number" step="0.01" min="0" class="input" :placeholder="t('payment.admin.limitInheritGroup')" @input="handlePeriodLimitInput" />
            <p class="mt-1 text-2xs leading-relaxed text-gray-400 dark:text-dark-500">
              {{ t('payment.admin.limitHint') }}
            </p>
          </div>
          <div>
            <label class="input-label">{{ t('payment.admin.planWeeklyLimitUSD') }}</label>
            <input v-model.number="planForm.weekly_limit_usd" type="number" step="0.01" min="0" class="input"
                   :class="limitWarnings.weekly ? 'border-warning focus:border-warning focus:ring-warning/40 dark:border-warning/60' : ''"
                   :placeholder="t('payment.admin.limitInheritGroup')" @input="handlePeriodLimitInput" />
            <p v-if="limitWarnings.weekly" class="mt-1 flex items-start gap-1 text-2xs leading-relaxed text-warning dark:text-warning">
              <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
              <span>{{ t('payment.admin.warnWeeklyRedundant', { max: Math.floor(limitWarnings.weekly.effectiveMax) }) }}</span>
            </p>
          </div>
          <div>
            <label class="input-label">{{ t('payment.admin.planMonthlyLimitUSD') }}</label>
            <input v-model.number="planForm.monthly_limit_usd" type="number" step="0.01" min="0" class="input"
                   :class="limitWarnings.monthly ? 'border-warning focus:border-warning focus:ring-warning/40 dark:border-warning/60' : ''"
                   :placeholder="t('payment.admin.limitInheritGroup')" @input="handlePeriodLimitInput" />
            <p v-if="limitWarnings.monthly" class="mt-1 flex items-start gap-1 text-2xs leading-relaxed text-warning dark:text-warning">
              <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
              <span>{{ t('payment.admin.warnMonthlyRedundant_' + limitWarnings.monthly.cappedBy, { max: Math.floor(limitWarnings.monthly.effectiveMax) }) }}</span>
            </p>
          </div>
          <div>
            <label class="input-label">{{ t('payment.admin.planRateMultiplier') }}</label>
            <input v-model.number="planForm.rate_multiplier" type="number" step="0.01" min="0" class="input" :placeholder="t('payment.admin.limitInheritGroup')" />
          </div>
        </div>
        <div class="col-span-2 mt-4 rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-xs dark:border-dark-700 dark:bg-dark-900/60">
          <div class="mb-2 flex items-center justify-between gap-3">
            <span class="font-semibold text-gray-800 dark:text-gray-100">{{ t('payment.admin.costMultiplierPreview') }}</span>
            <span class="rounded-full bg-gray-100 px-2 py-0.5 text-2xs font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300">
              {{ t('payment.admin.adminOnly') }}
            </span>
          </div>
          <div v-if="costEstimate.effectiveCostMultiplier !== null" class="grid grid-cols-3 gap-2">
            <div>
              <span class="block text-gray-400">{{ t('payment.admin.quotaPriceRate') }}</span>
              <span class="mt-0.5 block font-semibold text-gray-800 dark:text-gray-100">{{ formatCostMultiplier(costEstimate.priceQuotaMultiplier) }}</span>
            </div>
            <div>
              <span class="block text-gray-400">{{ t('payment.admin.billingRate') }}</span>
              <span class="mt-0.5 block font-semibold text-gray-800 dark:text-gray-100">{{ formatCostMultiplier(costEstimate.rateMultiplier) }}</span>
            </div>
            <div>
              <span class="block text-gray-400">{{ t('payment.admin.effectiveCostRate') }}</span>
              <span class="mt-0.5 block font-semibold text-success dark:text-success">{{ formatCostMultiplier(costEstimate.effectiveCostMultiplier) }}</span>
            </div>
          </div>
          <p v-else class="text-gray-500 dark:text-dark-400">{{ t('payment.admin.costMultiplierUnavailable') }}</p>
          <p v-if="costEstimate.periodLimitUSD !== null" class="mt-2 text-2xs text-gray-400 dark:text-dark-500">
            {{ t('payment.admin.periodQuotaUsed', { quota: Number(costEstimate.periodLimitUSD.toFixed(4)) }) }}
          </p>
          <div class="mt-3 grid gap-2 border-t border-gray-100 pt-3 dark:border-dark-700 sm:grid-cols-2">
            <div>
              <label class="input-label">{{ t('payment.admin.targetCostMultiplier') }}</label>
              <input
                v-model.number="targetCostMultiplier"
                type="number"
                step="0.0001"
                min="0"
                class="input"
                :placeholder="t('payment.admin.targetCostMultiplierPlaceholder')"
                @input="handleCostTargetInput"
              />
            </div>
            <div>
              <label class="input-label">{{ t('payment.admin.periodQuotaInput') }}</label>
              <input
                v-model.number="periodQuotaInput"
                type="number"
                step="0.01"
                min="0"
                class="input"
                :placeholder="t('payment.admin.periodQuotaInputPlaceholder')"
                @input="handleCostQuotaInput"
              />
            </div>
            <p class="sm:col-span-2 text-2xs leading-relaxed text-gray-400 dark:text-dark-500">
              {{ calculatorStatusText }}
            </p>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <label class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.admin.forSale') }}</label>
        <button
          type="button"
          :class="[
            'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
            planForm.for_sale ? 'bg-success' : 'bg-gray-300 dark:bg-dark-600'
          ]"
          @click="planForm.for_sale = !planForm.for_sale"
        >
          <span :class="[
            'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
            planForm.for_sale ? 'translate-x-5' : 'translate-x-0'
          ]" />
        </button>
        <!-- 显式状态文字 + 颜色，让"已开启 / 已关闭"一眼可辨 -->
        <span
          v-if="planForm.for_sale"
          class="inline-flex items-center gap-1 rounded-full bg-success-soft px-2 py-0.5 text-xs font-medium text-success-deep dark:bg-success/20 dark:text-success"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-success"></span>
          {{ t('payment.admin.forSaleOn') }}
        </span>
        <span
          v-else
          class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500 dark:bg-dark-700 dark:text-gray-400"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-gray-400"></span>
          {{ t('payment.admin.forSaleOff') }}
        </span>
      </div>

    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" @click="emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="plan-form" :disabled="saving" class="btn btn-primary">{{ saving ? t('common.saving') : t('common.save') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import { platformTextClass } from '@/utils/platformColors'
import { validatePlanLimits } from '@/utils/planLimits'
import { normalizeToDays } from '@/utils/planCardType'
import { calculatePlanCostEstimate, formatCostMultiplier } from '@/utils/planCost'
import { BADGE_TONE_KEYS, DEFAULT_BADGE_TONE, badgeToneClass, badgeToneSwatchClass, type BadgeTone } from '@/utils/badgeTone'
import { TIER_STYLE_KEYS, DEFAULT_TIER_STYLE, tierSwatchClass, type TierStyle } from '@/utils/tierStyle'

const props = defineProps<{
  show: boolean
  plan: SubscriptionPlan | null
  groups: AdminGroup[]
  /**
   * 创建新套餐时的预填值（来自订阅定价助手"用此建议创建套餐"）。
   * 仅在 plan === null（创建模式）下生效；plan != null（编辑模式）下忽略。
   */
  prefill?: Partial<{
    name: string
    price: number
    daily_limit_usd: number
    weekly_limit_usd: number
    monthly_limit_usd: number
    rate_multiplier: number
  }> | null
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const saving = ref(false)
type CostCalcSource = 'price' | 'quota'
const targetCostMultiplier = ref<number | null>(null)
const periodQuotaInput = ref<number | null>(null)
const costCalcSource = ref<CostCalcSource>('price')
let syncingCostCalculator = false
const planForm = reactive({
  name: '',
  group_id: null as number | null,
  description: '',
  price: 0,
  original_price: 0,
  validity_days: 30,
  validity_unit: 'days',
  sort_order: 0,
  for_sale: true,
  is_popular: false,
  badge_text: '',
  badge_color: DEFAULT_BADGE_TONE as BadgeTone,
  plan_label: '',
  tier_style: DEFAULT_TIER_STYLE as TierStyle,
  kind: 'shared' as 'shared' | 'exclusive',
  // 套餐级覆盖字段：null/0 表示沿用 group 默认值
  daily_limit_usd: null as number | null,
  weekly_limit_usd: null as number | null,
  monthly_limit_usd: null as number | null,
  rate_multiplier: null as number | null,
})
const planFeaturesText = ref('')

// 跨字段限额校验：weekly >= daily × 7 或 monthly >= 更紧限额折算月上限 时给 admin 提示
// 不阻断提交，让 admin 自主决定（万一是故意配的）
const limitWarnings = computed(() =>
  validatePlanLimits(
    planForm.daily_limit_usd,
    planForm.weekly_limit_usd,
    planForm.monthly_limit_usd,
  )
)

const validityUnitOptions = computed(() => [
  { value: 'days', label: t('payment.admin.days') },
  { value: 'weeks', label: t('payment.admin.weeks') },
  { value: 'months', label: t('payment.admin.months') },
])

const planKindOptions = computed(() => [
  { value: 'shared', label: t('payment.admin.planKindShared') },
  { value: 'exclusive', label: t('payment.admin.planKindExclusive') },
])

const tierStyleOptions = computed(() =>
  TIER_STYLE_KEYS.map(k => ({ value: k, label: t(`payment.admin.tierStyle.${k}`) })),
)

const groupOptions = computed(() =>
  props.groups
    .filter(g => g.subscription_type === 'subscription')
    .map(g => ({
      value: g.id,
      label: `${g.name} — ${g.platform} (${g.rate_multiplier}x)`,
      platform: g.platform,
    })),
)

const selectedGroupInfo = computed(() => {
  if (!planForm.group_id) return null
  return props.groups.find(g => g.id === planForm.group_id) || null
})

function displayBadgeText(plan: SubscriptionPlan | null): string {
  if (!plan) return ''
  // 角标文字与"推荐档"(is_popular) 解耦，这里只回填角标本身
  return (plan.badge_text || '').trim()
}

// 把后端可能为空/旧值的 badge_color 收敛到合法色调 key
function normalizeBadgeTone(color: string | null | undefined): BadgeTone {
  const v = (color || '').trim().toLowerCase()
  return (BADGE_TONE_KEYS as string[]).includes(v) ? (v as BadgeTone) : DEFAULT_BADGE_TONE
}

// 把后端可能为空/旧值的 tier_style 收敛到合法 key
function normalizeTierStyle(style: string | null | undefined): TierStyle {
  const v = (style || '').trim().toLowerCase()
  return (TIER_STYLE_KEYS as string[]).includes(v) ? (v as TierStyle) : DEFAULT_TIER_STYLE
}

function positiveNumber(value: number | null | undefined): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : null
}

function roundCurrency(value: number): number {
  return Number(value.toFixed(2))
}

function roundMultiplierValue(value: number): number {
  return Number(value.toFixed(4))
}

function activePeriodLimitTarget(): { field: 'daily_limit_usd' | 'weekly_limit_usd' | 'monthly_limit_usd'; label: string; factor: number } | null {
  const totalDays = normalizeToDays(planForm.validity_days, planForm.validity_unit)
  if (totalDays <= 0) return null
  if (totalDays % 30 === 0) {
    return { field: 'monthly_limit_usd', label: t('payment.admin.planMonthlyLimitUSD'), factor: totalDays / 30 }
  }
  if (totalDays % 7 === 0) {
    return { field: 'weekly_limit_usd', label: t('payment.admin.planWeeklyLimitUSD'), factor: totalDays / 7 }
  }
  return { field: 'daily_limit_usd', label: t('payment.admin.planDailyLimitUSD'), factor: totalDays }
}

function setPeriodQuota(periodQuota: number | null) {
  const target = activePeriodLimitTarget()
  if (!target || periodQuota === null) return
  planForm[target.field] = roundCurrency(periodQuota / target.factor)
  periodQuotaInput.value = roundCurrency(periodQuota)
}

function syncPeriodQuotaFromEstimate() {
  const quota = positiveNumber(costEstimate.value.periodLimitUSD)
  periodQuotaInput.value = quota === null ? null : roundCurrency(quota)
}

function syncTargetFromEstimate() {
  const multiplier = positiveNumber(costEstimate.value.effectiveCostMultiplier)
  targetCostMultiplier.value = multiplier === null ? null : roundMultiplierValue(multiplier)
}

function recalculateQuotaFromPrice() {
  const price = positiveNumber(planForm.price)
  const target = positiveNumber(targetCostMultiplier.value)
  const billingRate = positiveNumber(costEstimate.value.rateMultiplier) ?? 1
  if (price === null || target === null || billingRate <= 0) return
  setPeriodQuota((price * billingRate) / target)
}

function recalculatePriceFromQuota() {
  const quota = positiveNumber(periodQuotaInput.value)
  const target = positiveNumber(targetCostMultiplier.value)
  const billingRate = positiveNumber(costEstimate.value.rateMultiplier) ?? 1
  if (quota === null || target === null || billingRate <= 0) return
  planForm.price = roundCurrency((target * quota) / billingRate)
}

function handleCostPriceInput() {
  costCalcSource.value = 'price'
  recalculateQuotaFromPrice()
}

function handleCostTargetInput() {
  if (costCalcSource.value === 'quota') {
    recalculatePriceFromQuota()
    return
  }
  recalculateQuotaFromPrice()
}

function handleCostQuotaInput() {
  costCalcSource.value = 'quota'
  const quota = positiveNumber(periodQuotaInput.value)
  setPeriodQuota(quota)
  recalculatePriceFromQuota()
}

async function handlePeriodLimitInput() {
  costCalcSource.value = 'quota'
  await nextTick()
  syncPeriodQuotaFromEstimate()
  recalculatePriceFromQuota()
}

const calculatorStatusText = computed(() => {
  const target = activePeriodLimitTarget()
  if (!target) return t('payment.admin.costCalculatorUnavailable')
  if (targetCostMultiplier.value && periodQuotaInput.value) {
    return t('payment.admin.costCalculatorTarget', {
      field: target.label,
      value: roundCurrency(periodQuotaInput.value / target.factor),
    })
  }
  return t('payment.admin.costCalculatorHint')
})

const costEstimate = computed(() =>
  calculatePlanCostEstimate(
    {
      price: planForm.price,
      validity_days: planForm.validity_days,
      validity_unit: planForm.validity_unit,
      rate_multiplier: planForm.rate_multiplier,
      daily_limit_usd: planForm.daily_limit_usd,
      weekly_limit_usd: planForm.weekly_limit_usd,
      monthly_limit_usd: planForm.monthly_limit_usd,
    },
    selectedGroupInfo.value,
  )
)

// Reset form when dialog opens
watch(() => props.show, (visible) => {
  if (!visible) return
  if (props.plan) {
    Object.assign(planForm, {
      name: props.plan.name,
      group_id: props.plan.group_id,
      description: props.plan.description,
      price: props.plan.price,
      original_price: props.plan.original_price || 0,
      validity_days: props.plan.validity_days,
      validity_unit: props.plan.validity_unit || 'days',
      sort_order: props.plan.sort_order || 0,
      for_sale: props.plan.for_sale,
      is_popular: props.plan.is_popular === true,
      badge_text: displayBadgeText(props.plan),
      badge_color: normalizeBadgeTone(props.plan.badge_color),
      plan_label: props.plan.plan_label || '',
      tier_style: normalizeTierStyle(props.plan.tier_style),
      kind: props.plan.kind || 'shared',
      daily_limit_usd: (props.plan as any).daily_limit_usd ?? null,
      weekly_limit_usd: (props.plan as any).weekly_limit_usd ?? null,
      monthly_limit_usd: (props.plan as any).monthly_limit_usd ?? null,
      rate_multiplier: (props.plan as any).rate_multiplier ?? null,
    })
    planFeaturesText.value = (props.plan.features || []).join('\n')
  } else {
    // 创建模式：默认值 + 可选的 prefill 覆盖（来自定价助手"用此建议创建"链路）
    const p = props.prefill || {}
    Object.assign(planForm, {
      name: p.name ?? '',
      group_id: null,
      description: '',
      price: p.price ?? 0,
      original_price: 0,
      validity_days: 30,
      validity_unit: 'days',
      sort_order: 0,
      for_sale: true,
      is_popular: false,
      badge_text: '',
      badge_color: DEFAULT_BADGE_TONE,
      plan_label: '',
      tier_style: DEFAULT_TIER_STYLE,
      kind: 'shared',
      daily_limit_usd: p.daily_limit_usd ?? null,
      weekly_limit_usd: p.weekly_limit_usd ?? null,
      monthly_limit_usd: p.monthly_limit_usd ?? null,
      rate_multiplier: p.rate_multiplier ?? null,
    })
    planFeaturesText.value = ''
  }
  syncPeriodQuotaFromEstimate()
  syncTargetFromEstimate()
})

watch(
  () => costEstimate.value.periodLimitUSD,
  () => {
    if (!syncingCostCalculator) syncPeriodQuotaFromEstimate()
  },
)

watch(
  () => [planForm.validity_days, planForm.validity_unit, planForm.rate_multiplier, selectedGroupInfo.value?.rate_multiplier],
  () => {
    if (!targetCostMultiplier.value) return
    syncingCostCalculator = true
    try {
      if (costCalcSource.value === 'quota') recalculatePriceFromQuota()
      else recalculateQuotaFromPrice()
    } finally {
      syncingCostCalculator = false
    }
  },
)

/** Build request payload with snake_case keys matching backend JSON tags */
function buildPlanPayload() {
  const features = planFeaturesText.value.split('\n').map(f => f.trim()).filter(Boolean).join('\n')
  // 限额覆盖字段语义：
  //   后端 *float64 nil = "不修改"（保持 DB 现状）；0 = "清空覆盖"（→ 调度回落到 group）；> 0 = "设置覆盖"
  // 前端永远显式传值（0 或正数），不传 null/undefined，避免 admin 清空字段后 DB 仍保留旧值
  const optionalLimit = (v: number | null): number => {
    if (v === null || v === undefined || Number.isNaN(v) || v <= 0) return 0
    return v
  }
  return {
    name: planForm.name,
    group_id: planForm.group_id,
    description: planForm.description,
    price: planForm.price,
    original_price: planForm.original_price || 0,
    validity_days: planForm.validity_days,
    validity_unit: planForm.validity_unit,
    sort_order: planForm.sort_order,
    for_sale: planForm.for_sale,
    is_popular: planForm.is_popular,
    badge_text: planForm.badge_text.trim(),
    badge_color: planForm.badge_color,
    plan_label: planForm.plan_label.trim(),
    tier_style: planForm.tier_style,
    features,
    kind: planForm.kind,
    daily_limit_usd: optionalLimit(planForm.daily_limit_usd),
    weekly_limit_usd: optionalLimit(planForm.weekly_limit_usd),
    monthly_limit_usd: optionalLimit(planForm.monthly_limit_usd),
    rate_multiplier: optionalLimit(planForm.rate_multiplier),
  }
}

async function handleSavePlan() {
  if (!planForm.group_id) {
    appStore.showError(t('payment.admin.groupRequired'))
    return
  }
  if (!planForm.price || planForm.price <= 0) {
    appStore.showError(t('payment.admin.priceRequired'))
    return
  }
  if (!planForm.validity_days || planForm.validity_days < 1) {
    appStore.showError(t('payment.admin.validityDaysRequired'))
    return
  }
  saving.value = true
  try {
    const data = buildPlanPayload()
    if (props.plan) { await adminPaymentAPI.updatePlan(props.plan.id, data) }
    else { await adminPaymentAPI.createPlan(data) }
    appStore.showSuccess(t('common.saved'))
    emit('close')
    emit('saved')
  } catch (err: unknown) { appStore.showError(extractApiErrorMessage(err, t('common.error'))) }
  finally { saving.value = false }
}
</script>

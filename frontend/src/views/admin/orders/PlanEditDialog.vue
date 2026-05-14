<template>
  <BaseDialog :show="show" :title="plan ? t('payment.admin.editPlan') : t('payment.admin.createPlan')" width="wide" @close="emit('close')">
    <form id="plan-form" @submit.prevent="handleSavePlan" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="input-label">{{ t('payment.admin.planName') }} <span class="text-red-500">*</span></label>
          <input v-model="planForm.name" type="text" class="input" required />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.group') }} <span class="text-red-500">*</span></label>
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

      <div><label class="input-label">{{ t('payment.admin.planDescription') }} <span class="text-red-500">*</span></label><textarea v-model="planForm.description" rows="2" class="input" required></textarea></div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="input-label">{{ t('payment.admin.price') }} <span class="text-red-500">*</span></label><input v-model.number="planForm.price" type="number" step="0.01" min="0.01" class="input" required /></div>
        <div><label class="input-label">{{ t('payment.admin.originalPrice') }}</label><input v-model.number="planForm.original_price" type="number" step="0.01" min="0" class="input" /></div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="input-label">{{ t('payment.admin.validityDays') }} <span class="text-red-500">*</span></label><input v-model.number="planForm.validity_days" type="number" min="1" class="input" required /></div>
        <div><label class="input-label">{{ t('payment.admin.validityUnit') }} <span class="text-red-500">*</span></label><Select v-model="planForm.validity_unit" :options="validityUnitOptions" /></div>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="input-label">{{ t('payment.admin.sortOrder') }}</label><input v-model.number="planForm.sort_order" type="number" min="0" class="input" /></div>
        <div>
          <label class="input-label">{{ t('payment.admin.planKind') }} <span class="text-red-500">*</span></label>
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
            <input v-model.number="planForm.daily_limit_usd" type="number" step="0.01" min="0" class="input" :placeholder="t('payment.admin.limitInheritGroup')" />
            <p class="mt-1 text-[11px] leading-relaxed text-gray-400 dark:text-dark-500">
              {{ t('payment.admin.limitHint') }}
            </p>
          </div>
          <div>
            <label class="input-label">{{ t('payment.admin.planWeeklyLimitUSD') }}</label>
            <input v-model.number="planForm.weekly_limit_usd" type="number" step="0.01" min="0" class="input"
                   :class="limitWarnings.weekly ? 'border-amber-400 focus:border-amber-500 focus:ring-amber-300/40 dark:border-amber-500/60' : ''"
                   :placeholder="t('payment.admin.limitInheritGroup')" />
            <p v-if="limitWarnings.weekly" class="mt-1 flex items-start gap-1 text-[11px] leading-relaxed text-amber-600 dark:text-amber-400">
              <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
              <span>{{ t('payment.admin.warnWeeklyRedundant', { max: Math.floor(limitWarnings.weekly.effectiveMax) }) }}</span>
            </p>
          </div>
          <div>
            <label class="input-label">{{ t('payment.admin.planMonthlyLimitUSD') }}</label>
            <input v-model.number="planForm.monthly_limit_usd" type="number" step="0.01" min="0" class="input"
                   :class="limitWarnings.monthly ? 'border-amber-400 focus:border-amber-500 focus:ring-amber-300/40 dark:border-amber-500/60' : ''"
                   :placeholder="t('payment.admin.limitInheritGroup')" />
            <p v-if="limitWarnings.monthly" class="mt-1 flex items-start gap-1 text-[11px] leading-relaxed text-amber-600 dark:text-amber-400">
              <Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" />
              <span>{{ t('payment.admin.warnMonthlyRedundant_' + limitWarnings.monthly.cappedBy, { max: Math.floor(limitWarnings.monthly.effectiveMax) }) }}</span>
            </p>
          </div>
          <div>
            <label class="input-label">{{ t('payment.admin.planRateMultiplier') }}</label>
            <input v-model.number="planForm.rate_multiplier" type="number" step="0.01" min="0" class="input" :placeholder="t('payment.admin.limitInheritGroup')" />
          </div>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <label class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.admin.forSale') }}</label>
        <button
          type="button"
          :class="[
            'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
            planForm.for_sale ? 'bg-green-500' : 'bg-gray-300 dark:bg-dark-600'
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
          class="inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-300"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-green-500"></span>
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
import { ref, reactive, computed, watch } from 'vue'
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
      kind: 'shared',
      daily_limit_usd: p.daily_limit_usd ?? null,
      weekly_limit_usd: p.weekly_limit_usd ?? null,
      monthly_limit_usd: p.monthly_limit_usd ?? null,
      rate_multiplier: p.rate_multiplier ?? null,
    })
    planFeaturesText.value = ''
  }
})

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

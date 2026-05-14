<template>
  <!-- 卡片：用中性边框 + 兑换码同款"远距离浮起阴影"，平台色仅在 badge / 价格 / 按钮处点缀，
       避免整张卡片被绿/橙色边框包围造成视觉冲击 -->
  <div
    :class="[
      'plan-card group relative flex min-h-[250px] flex-col overflow-hidden rounded-lg border-2 bg-white transition-colors dark:bg-dark-900',
      soldOut
        ? 'border-gray-200 opacity-65 grayscale dark:border-dark-700'
        : isPopular
          ? 'border-amber-400/80 shadow-amber-100 hover:border-amber-500 dark:border-amber-500/60 dark:hover:border-amber-400/80'
          : 'border-gray-200 hover:border-gray-300 dark:border-dark-700 dark:hover:border-dark-500',
    ]"
  >
    <!-- 主推角标：右上角斜带，渐变 + 内层白线高光 + 暖色光晕，比平铺 badge 更显眼。
         purely SVG（sparkles 图标），无 emoji，符合系统设计规范。 -->
    <div
      v-if="isPopular && !soldOut"
      class="popular-ribbon pointer-events-none absolute right-0 top-0 z-10 overflow-hidden"
    >
      <div class="relative h-[96px] w-[96px]">
        <!-- 外圈：渐变带 + 双层阴影（深色投影 + 暖色光晕） -->
        <div class="absolute -right-[28px] top-[20px] flex w-[136px] rotate-45 flex-col items-center justify-center bg-gradient-to-r from-amber-500 via-orange-500 to-amber-600 py-[5px] text-white shadow-[0_2px_10px_rgba(245,158,11,0.55),0_1px_3px_rgba(0,0,0,0.25)] ring-1 ring-inset ring-white/35">
          <span class="flex items-center gap-[3px] text-[11px] font-extrabold uppercase tracking-[0.12em] drop-shadow-[0_1px_1px_rgba(0,0,0,0.3)]">
            <Icon name="sparkles" size="xs" :stroke-width="2.75" />
            {{ t('payment.planCard.popularBadge') }}
          </span>
        </div>
      </div>
    </div>

    <!-- 售罄角标：右上角红色斜带；与主推角标互斥（售罄优先） -->
    <div
      v-if="soldOut"
      class="pointer-events-none absolute right-0 top-0 z-10 overflow-hidden"
    >
      <div class="relative h-[88px] w-[88px]">
        <div class="absolute -right-[26px] top-[18px] flex w-[120px] rotate-45 items-center justify-center bg-rose-500 py-1 text-[11px] font-bold uppercase tracking-wider text-white shadow-md">
          {{ t('payment.admin.stockSoldOut') }}
        </div>
      </div>
    </div>

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
              <div class="mt-1 flex flex-wrap items-center gap-1.5">
                <span :class="['inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium', badgeLightClass]">
                  {{ pLabel }}
                </span>
                <span v-if="plan.kind === 'exclusive'"
                  class="inline-flex items-center gap-1 rounded-full bg-violet-50 px-2 py-0.5 text-[11px] font-semibold text-violet-700 ring-1 ring-violet-200 dark:bg-violet-900/30 dark:text-violet-300 dark:ring-violet-900/50">
                  <Icon name="badge" size="xs" :stroke-width="2.5" />
                  {{ t('payment.admin.kindBadgeExclusive') }}
                </span>
                <span v-if="cardType !== 'custom'"
                  :class="['inline-flex rounded-full px-2 py-0.5 text-[11px] font-semibold', cardTypeBadgeClass(cardType)]">
                  {{ t(`payment.admin.cardType.${cardType}`) }}
                </span>
                <!-- 套餐自带限额覆盖：标识"独立档位"，让买家看到这档跟同 group 其他档限额不同 -->
                <span v-if="plan.has_plan_limit_override"
                  class="inline-flex items-center gap-1 rounded-full bg-indigo-50 px-2 py-0.5 text-[11px] font-semibold text-indigo-700 ring-1 ring-indigo-200 dark:bg-indigo-900/30 dark:text-indigo-300 dark:ring-indigo-900/50"
                  :title="t('payment.planCard.tierBadgeHint')">
                  <Icon name="badge" size="xs" :stroke-width="2.5" />
                  {{ t('payment.planCard.tierBadge') }}
                </span>
                <span v-if="plan.kind === 'exclusive' && stockInfo"
                  :class="['inline-flex rounded-full px-2 py-0.5 text-[11px] font-semibold', stockInfo.cls]">
                  {{ stockInfo.text }}
                </span>
                <!-- 已订阅状态：同一 group 已有活跃订阅时显示，3 个色档让用户看到紧迫感 -->
                <span v-if="subscriptionInfo"
                  :class="['inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold ring-1', subscriptionChipClass]">
                  <Icon name="checkCircle" size="xs" :stroke-width="2.5" />
                  {{ t('payment.planCard.subscribed', { days: subscriptionInfo.daysRemaining }) }}
                </span>
              </div>
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
        <!-- 限额：value > 0 显示具体额度；value <= 0 显示"无限制"。
             被更紧限额覆盖的"废限额"由 limitVisibility 自动隐藏，避免用户看到 ×0.6 倍率 + 周$280 + 月$1200 这种没意义的并列。 -->
        <div v-if="limitVisibility.showDaily" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <span v-if="(plan.daily_limit_usd ?? 0) > 0" class="mt-1 block text-base font-bold text-gray-900 dark:text-white">${{ plan.daily_limit_usd }}</span>
          <span v-else class="mt-1 block text-base font-bold text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        <div v-if="limitVisibility.showWeekly" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span v-if="(plan.weekly_limit_usd ?? 0) > 0" class="mt-1 block text-base font-bold text-gray-900 dark:text-white">${{ plan.weekly_limit_usd }}</span>
          <span v-else class="mt-1 block text-base font-bold text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        <div v-if="limitVisibility.showMonthly" class="rounded-lg bg-white px-3 py-2.5 dark:bg-dark-900/70">
          <span class="block text-[11px] font-medium text-gray-400 dark:text-dark-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span v-if="(plan.monthly_limit_usd ?? 0) > 0" class="mt-1 block text-base font-bold text-gray-900 dark:text-white">${{ plan.monthly_limit_usd }}</span>
          <span v-else class="mt-1 block text-base font-bold text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</span>
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

      <!-- 按钮：统一黑色 CTA（platform 色已在 badge/价格/headline 处体现，避免再用饱和按钮造成视觉重复） -->
      <div class="mt-4 flex justify-center">
        <button
          type="button"
          :disabled="soldOut"
          :class="[
            'inline-flex items-center justify-center rounded-md px-10 py-2.5 text-sm font-semibold transition-colors active:scale-[0.99]',
            soldOut
              ? 'cursor-not-allowed bg-gray-200 text-gray-400 dark:bg-dark-700 dark:text-dark-500'
              : 'bg-gray-900 text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100',
          ]"
          @click="!soldOut && emit('select', plan)"
        >
          {{ soldOut ? t('payment.admin.stockSoldOut') : (isRenewal ? t('payment.renewNow') : t('payment.subscribeNow')) }}
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
  platformBadgeLightClass,
  platformTextClass,
  platformIconClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'
import { derivePlanCardType, cardTypeBadgeClass } from '@/utils/planCardType'
import { getEffectiveLimitVisibility } from '@/utils/planLimits'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
// 找到与当前 plan 同 group 的活跃订阅（最多一个）；用于"已订阅"状态提示和续费 CTA
const matchingActiveSubscription = computed(() =>
  props.activeSubscriptions?.find(s => s.group_id === props.plan.group_id && s.status === 'active') ?? null
)
const isRenewal = computed(() => matchingActiveSubscription.value !== null)

// 剩余天数 + 颜色档：> 7 天 = 中性 emerald；3-7 天 = amber 预警；≤ 3 天 = rose 强提醒
// 让重度续费用户在卡片角上直接看到"还剩 N 天"，不用切到"我的订阅"页核对
const subscriptionInfo = computed(() => {
  const sub = matchingActiveSubscription.value
  if (!sub || !sub.expires_at) return null
  const expiresAt = new Date(sub.expires_at)
  if (Number.isNaN(expiresAt.getTime())) return null
  const msRemaining = expiresAt.getTime() - Date.now()
  const daysRemaining = Math.max(0, Math.ceil(msRemaining / (1000 * 60 * 60 * 24)))
  let tone: 'neutral' | 'warning' | 'urgent' = 'neutral'
  if (daysRemaining <= 3) tone = 'urgent'
  else if (daysRemaining <= 7) tone = 'warning'
  return { daysRemaining, tone }
})

const subscriptionChipClass = computed(() => {
  if (!subscriptionInfo.value) return ''
  switch (subscriptionInfo.value.tone) {
    case 'urgent':
      return 'bg-rose-50 text-rose-700 ring-rose-200 dark:bg-rose-900/20 dark:text-rose-300 dark:ring-rose-900/50'
    case 'warning':
      return 'bg-amber-50 text-amber-700 ring-amber-200 dark:bg-amber-900/20 dark:text-amber-300 dark:ring-amber-900/50'
    default:
      return 'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-300 dark:ring-emerald-900/50'
  }
})

// Derived color classes from central config（accent bar / border / 主按钮已统一为中性色，platform 色仅用在 badge / 价格 / 折扣 / 模型对勾）
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const textClass = computed(() => platformTextClass(platform.value))
const iconClass = computed(() => platformIconClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))
const platformBrand = computed<'claude' | 'openai' | 'gemini' | null>(() => {
  if (platform.value === 'anthropic') return 'claude'
  if (platform.value === 'openai') return 'openai'
  if (platform.value === 'gemini') return 'gemini'
  return null
})
const platformInitial = computed(() => pLabel.value.charAt(0).toUpperCase())

// 卡类型（日/周/月/季/年/自定义）—— 纯前端推导，不影响业务逻辑
const cardType = computed(() => derivePlanCardType(props.plan.validity_days, props.plan.validity_unit))

// 独享池库存：sold_out 时禁用购买按钮
const stock = computed<number | null>(() => {
  if (props.plan.kind !== 'exclusive') return null
  const v = (props.plan as SubscriptionPlan & { stock_available?: number }).stock_available
  return typeof v === 'number' ? v : null
})
const soldOut = computed(() => stock.value !== null && stock.value <= 0)
const stockInfo = computed<{ text: string; cls: string } | null>(() => {
  if (stock.value === null) return null
  if (stock.value <= 0) {
    return {
      text: t('payment.admin.stockSoldOut'),
      cls: 'bg-red-50 text-red-700 ring-1 ring-red-200 dark:bg-red-900/20 dark:text-red-300 dark:ring-red-900/50',
    }
  }
  return {
    text: t('payment.admin.stockAvailable', { n: stock.value }),
    cls: 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-300 dark:ring-emerald-900/50',
  }
})

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

// 主推标记：admin 在 plan 上勾选 is_popular 后，卡片显示琥珀角标 + 描边强化
const isPopular = computed(() => props.plan.is_popular === true)

// 限额可视性：废限额自动隐藏（被更紧的限额覆盖时不展示，避免用户困惑）
const limitVisibility = computed(() =>
  getEffectiveLimitVisibility({
    daily_limit_usd: props.plan.daily_limit_usd,
    weekly_limit_usd: props.plan.weekly_limit_usd,
    monthly_limit_usd: props.plan.monthly_limit_usd,
  })
)

const MODEL_SCOPE_META: Record<string, { label: string; iconModel: string }> = {
  claude: { label: 'Claude', iconModel: 'claude-3-5-sonnet' },
  gemini_text: { label: 'Gemini', iconModel: 'gemini-2.5-pro' },
  gemini_image: { label: 'Imagen', iconModel: 'imagen-3' },
}

// supported_model_scopes 是 antigravity 平台专属字段（用来选 claude / gemini_text / gemini_image 子能力）。
// 非 antigravity 平台（openai / anthropic / gemini）该字段无意义——但 GroupsView createForm 默认会塞一组
// antigravity scope 进去，结果 GPT 套餐卡片误显示 "Claude · Gemini · Imagen"。
// 这里强制只在 antigravity 平台才渲染 scope chips，根治视觉错乱。
const modelScopeItems = computed(() => {
  if (platform.value !== 'antigravity') return []
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

<style scoped>
/* 卡片浮起：与兑换码 redeem-panel / 文档 docs-panel 一致的远距离淡阴影 */
.plan-card {
  box-shadow: 0 18px 44px -34px rgb(15 23 42 / 0.55);
}
.plan-card:hover {
  box-shadow: 0 22px 48px -28px rgb(15 23 42 / 0.6);
}
:global(:root.dark) .plan-card,
:global(:root.dark) .plan-card:hover {
  box-shadow: none;
}

/* 主推徽章：缓慢"扫光"动效，强化高级感（5s 一次，淡淡的不抢戏）。
   用 ::after 叠一条 45° 白色半透明斜光带，从徽章左侧扫到右侧。
   prefers-reduced-motion 用户禁用动画。 */
.popular-ribbon > div > div {
  position: absolute; /* 保留 inline 的定位继承 */
  overflow: hidden;
}
.popular-ribbon > div > div::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(
    100deg,
    transparent 30%,
    rgba(255, 255, 255, 0.45) 50%,
    transparent 70%
  );
  transform: translateX(-100%);
  animation: popular-shine 5s ease-in-out infinite;
  pointer-events: none;
}
@keyframes popular-shine {
  0%, 60% { transform: translateX(-100%); }
  80% { transform: translateX(100%); }
  100% { transform: translateX(100%); }
}
@media (prefers-reduced-motion: reduce) {
  .popular-ribbon > div > div::after {
    animation: none;
  }
}
</style>

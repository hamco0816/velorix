<template>
  <!-- 同一「平台 × 周期」下的所有档位，以卡片并排呈现。
       表头标识厂商；每张卡片带档位专属图标/配色、价格、限额、CTA；豪华/至尊档抬高突出。 -->
  <section class="overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-[0_18px_44px_-34px_rgba(15,23,42,0.4)] dark:border-dark-700/60 dark:bg-dark-900 dark:shadow-none">
    <header class="flex flex-wrap items-center gap-2.5 border-b border-gray-100 bg-gray-50/40 px-5 py-3.5 dark:border-dark-700/60 dark:bg-dark-800/30">
      <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-600">
        <BrandIcon v-if="brand" :brand="brand" size="18px" />
        <span v-else class="text-sm font-bold text-gray-500 dark:text-gray-300">{{ (label || 'API').charAt(0) }}</span>
      </span>
      <h3 class="text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ label }}</h3>
      <span v-if="periodLabel" class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-200">{{ periodLabel }}</span>
      <span v-if="allExclusive" class="rounded-full bg-violet-50 px-2 py-0.5 text-[11px] font-semibold text-violet-700 ring-1 ring-violet-200 dark:bg-violet-900/30 dark:text-violet-300 dark:ring-violet-900/50">{{ t('payment.admin.kindBadgeExclusive') }}</span>
      <span class="ml-auto text-[11px] tabular-nums text-gray-400 dark:text-dark-500">{{ t('payment.tierCount', { count: plans.length }) }}</span>
    </header>

    <div class="grid grid-cols-[repeat(auto-fit,minmax(14rem,1fr))] gap-5 p-6">
      <div
        v-for="(plan, i) in plans"
        :key="plan.id"
        :class="['relative flex flex-col overflow-hidden rounded-2xl border p-6 transition-all', cardClass(plan), isSoldOut(plan) ? 'opacity-70 grayscale' : '']"
      >
        <!-- 右上角真角标：售罄 > 促销角标 -->
        <div v-if="isSoldOut(plan) || badgeOf(plan)" class="absolute right-0 top-0 z-10">
          <span v-if="isSoldOut(plan)" class="inline-flex items-center rounded-bl-xl bg-rose-500 px-2.5 py-1.5 text-xs font-bold text-white shadow-sm">{{ t('payment.admin.stockSoldOut') }}</span>
          <span v-else :class="['inline-flex items-center rounded-bl-xl px-2.5 py-1.5 text-xs font-bold shadow-sm', badgeToneClass(plan.badge_color)]">{{ badgeOf(plan) }}</span>
        </div>

        <!-- 档位：图标 + 名 -->
        <div class="flex items-center gap-2.5">
          <Icon v-if="themes[i].icon" :name="themes[i].icon!" size="xl" :stroke-width="2.2" :class="themes[i].iconClass" />
          <span :class="['text-2xl font-bold tracking-tight', themes[i].nameClass]">{{ tierLabels[i] }}</span>
        </div>

        <!-- 价格（左）+ 原价/省%（右），左右撑开更平衡 -->
        <div class="mt-5 flex items-end justify-between gap-2">
          <div class="flex items-baseline gap-1">
            <span class="text-xl font-semibold text-gray-400 dark:text-dark-400">¥</span>
            <span class="text-[40px] font-bold tabular-nums leading-none tracking-tight text-gray-900 dark:text-white">{{ plan.price }}</span>
            <span class="ml-0.5 text-xs text-gray-400 dark:text-dark-500">/ {{ validitySuffix }}</span>
          </div>
          <div v-if="plan.original_price && plan.original_price > 0" class="flex shrink-0 flex-col items-end gap-1 pb-1 leading-none">
            <span class="text-sm text-gray-400 line-through decoration-gray-300 dark:text-dark-500">¥{{ plan.original_price }}</span>
            <span v-if="discountOf(plan)" class="rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-bold text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-400">{{ t('payment.savePercent', { pct: discountOf(plan) }) }}</span>
          </div>
        </div>

        <!-- 限额清单：分隔线 + 小图标，更精致 -->
        <dl class="mt-5 divide-y divide-gray-100 border-t border-gray-100 text-sm dark:divide-dark-700/50 dark:border-dark-700/50">
          <div v-if="showRate" class="flex items-center justify-between py-2.5">
            <dt class="flex items-center gap-1.5 text-gray-500 dark:text-dark-400"><Icon name="chartBar" size="xs" class="text-gray-400 dark:text-dark-500" />{{ t('payment.planCard.rate') }}</dt>
            <dd class="font-semibold tabular-nums text-gray-900 dark:text-white">{{ rateOf(plan) }}</dd>
          </div>
          <div v-if="showDaily" class="flex items-center justify-between py-2.5">
            <dt class="flex items-center gap-1.5 text-gray-500 dark:text-dark-400"><Icon name="clock" size="xs" class="text-gray-400 dark:text-dark-500" />{{ t('payment.planCard.dailyLimit') }}</dt>
            <dd><LimitValue :value="plan.daily_limit_usd" /></dd>
          </div>
          <div v-if="showWeekly" class="flex items-center justify-between py-2.5">
            <dt class="flex items-center gap-1.5 text-gray-500 dark:text-dark-400"><Icon name="calendar" size="xs" class="text-gray-400 dark:text-dark-500" />{{ t('payment.planCard.weeklyLimit') }}</dt>
            <dd><LimitValue :value="plan.weekly_limit_usd" /></dd>
          </div>
          <div v-if="showMonthly" class="flex items-center justify-between py-2.5">
            <dt class="flex items-center gap-1.5 text-gray-500 dark:text-dark-400"><Icon name="calendar" size="xs" class="text-gray-400 dark:text-dark-500" />{{ t('payment.planCard.monthlyLimit') }}</dt>
            <dd><LimitValue :value="plan.monthly_limit_usd" /></dd>
          </div>
          <div v-if="!showDaily && !showWeekly && !showMonthly" class="flex items-center justify-between py-2.5">
            <dt class="flex items-center gap-1.5 text-gray-500 dark:text-dark-400"><Icon name="sparkles" size="xs" class="text-gray-400 dark:text-dark-500" />{{ t('payment.planCard.quota') }}</dt>
            <dd class="font-semibold text-emerald-600 dark:text-emerald-400">{{ t('payment.planCard.unlimited') }}</dd>
          </div>
        </dl>

        <div class="flex-1" />

        <button
          type="button"
          :disabled="isSoldOut(plan)"
          :class="[
            'mt-6 w-full rounded-xl px-3 py-2.5 text-[13px] font-semibold transition-all active:scale-[0.99]',
            isSoldOut(plan)
              ? 'cursor-not-allowed bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500'
              : ctaClassOf(plan) || 'bg-gray-900 text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100',
          ]"
          @click="!isSoldOut(plan) && emit('select', plan)"
        >
          {{ ctaText(plan) }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, h, type FunctionalComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import BrandIcon from '@/components/common/BrandIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { badgeToneClass } from '@/utils/badgeTone'
import { tierTheme } from '@/utils/tierStyle'
import { normalizeToDays } from '@/utils/planCardType'

const props = defineProps<{
  plans: SubscriptionPlan[]
  platform: string
  brand: 'claude' | 'openai' | 'gemini' | null
  label: string
  periodLabel: string
  activeSubscriptions?: UserSubscription[]
}>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

// 限额值单元：>0 显示 $额度，否则"无限制"
const LimitValue: FunctionalComponent<{ value: number | null | undefined }> = (cellProps) => {
  const v = cellProps.value
  if (v != null && v > 0) {
    return h('span', { class: 'font-semibold tabular-nums text-gray-900 dark:text-white' }, `$${v}`)
  }
  return h('span', { class: 'font-semibold text-emerald-600 dark:text-emerald-400' }, t('payment.planCard.unlimited'))
}

const allExclusive = computed(() => props.plans.length > 0 && props.plans.every(p => p.kind === 'exclusive'))

// 同组所有卡片展示同一组限额行（取并集），保证卡片高度/结构一致
const showRate = computed(() => props.plans.some(p => (p.rate_multiplier ?? 1) !== 1))
const showDaily = computed(() => props.plans.some(p => p.daily_limit_usd != null))
const showWeekly = computed(() => props.plans.some(p => p.weekly_limit_usd != null))
const showMonthly = computed(() => props.plans.some(p => p.monthly_limit_usd != null))

// 有效期标签（如 30天 / 7天），同组通常一致，取第一个
const validitySuffix = computed(() => {
  const first = props.plans[0]
  if (!first) return ''
  const days = normalizeToDays(first.validity_days, first.validity_unit)
  return `${days}${t('payment.days')}`
})

// 每个档位的样式主题（图标 / 配色 / CTA），按列索引取用
const themes = computed(() => props.plans.map(p => tierTheme(p.tier_style)))

// 档位名：admin 填了 plan_label 优先用；否则自动从套餐名推导
const tierLabels = computed(() => {
  const derived = deriveTierLabels(props.plans.map(p => p.name), props.periodLabel)
  return props.plans.map((p, i) => (p.plan_label || '').trim() || derived[i])
})

// 自动推导：剥同组套餐名公共前缀（GPT周卡-Lite → Lite），再去掉开头漏出的周期词；任一档剥空则回退完整名
function deriveTierLabels(names: string[], periodLabel: string): string[] {
  const base = stripCommonPrefix(names)
  const cleaned = base.map(s => stripLeadingPeriod(s, periodLabel))
  return cleaned.every(s => s.length > 0) ? cleaned : names
}

function stripCommonPrefix(names: string[]): string[] {
  if (names.length <= 1) return [...names]
  let prefix = names[0]
  for (const name of names) {
    while (prefix && !name.startsWith(prefix)) {
      prefix = prefix.slice(0, -1)
    }
    if (!prefix) break
  }
  return prefix.length < 2 ? [...names] : names.map(name => name.slice(prefix.length))
}

const SEP = /^[\s\-·/|:：]+/
function stripLeadingPeriod(raw: string, periodLabel: string): string {
  let s = raw.replace(SEP, '')
  if (periodLabel && s.startsWith(periodLabel)) s = s.slice(periodLabel.length)
  return s.replace(SEP, '').trim()
}

function rateOf(plan: SubscriptionPlan): string {
  const rate = plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
}

function discountOf(plan: SubscriptionPlan): string {
  if (!plan.original_price || plan.original_price <= 0) return ''
  const pct = Math.round((1 - plan.price / plan.original_price) * 100)
  return pct > 0 ? `${pct}%` : ''
}

function badgeOf(plan: SubscriptionPlan): string {
  return (plan.badge_text || '').trim()
}

// 卡片整体样式：豪华/至尊档抬高突出（描边 + 渐变底 + 阴影）
function cardClass(plan: SubscriptionPlan): string {
  if (plan.tier_style === 'luxury') {
    // 豪华：金色描边 + 双环 + 鎏金光晕，明显抬高突出
    return 'border-amber-300 ring-2 ring-amber-200 bg-gradient-to-b from-amber-50/60 to-white shadow-[0_30px_60px_-24px_rgba(245,158,11,0.65)] dark:border-amber-500/50 dark:ring-amber-500/25 dark:from-amber-500/[0.08] dark:to-dark-900'
  }
  if (plan.tier_style === 'supreme') {
    // 至尊：黑金描边 + 深色光晕
    return 'border-gray-800 ring-2 ring-gray-300 bg-gradient-to-b from-gray-50 to-white shadow-[0_30px_60px_-24px_rgba(15,23,42,0.5)] dark:border-amber-500/30 dark:ring-white/10 dark:from-white/[0.05] dark:to-dark-900'
  }
  return 'border-gray-200 hover:border-gray-300 hover:shadow-lg hover:shadow-gray-200/60 dark:border-dark-700/60 dark:hover:border-dark-600'
}

// 该档 CTA 专属配色（豪华金 / 至尊黑金）；空则用默认黑按钮
function ctaClassOf(plan: SubscriptionPlan): string {
  return tierTheme(plan.tier_style).ctaClass
}

function stockOf(plan: SubscriptionPlan): number | null {
  if (plan.kind !== 'exclusive') return null
  const v = plan.stock_available
  return typeof v === 'number' ? v : null
}

function isSoldOut(plan: SubscriptionPlan): boolean {
  const s = stockOf(plan)
  return s !== null && s <= 0
}

// 与该档位同 group 的活跃订阅剩余天数；无则返回 null（用于 CTA 文案）
function subscribedDays(plan: SubscriptionPlan): number | null {
  const sub = props.activeSubscriptions?.find(s => s.group_id === plan.group_id && s.status === 'active')
  if (!sub || !sub.expires_at) return null
  const expiresAt = new Date(sub.expires_at)
  if (Number.isNaN(expiresAt.getTime())) return null
  return Math.max(0, Math.ceil((expiresAt.getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
}

function ctaText(plan: SubscriptionPlan): string {
  if (isSoldOut(plan)) return t('payment.admin.stockSoldOut')
  return subscribedDays(plan) !== null ? t('payment.renewNow') : t('payment.subscribeNow')
}
</script>

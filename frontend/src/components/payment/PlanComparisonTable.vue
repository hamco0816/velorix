<template>
  <!-- 同一「平台 × 周期」下所有档位的横向对比表：行 = 价格 / 倍率 / 日周月限额，列 = 档位。
       表头带品牌图标 + 平台名 + 周期，明确标识厂商；推荐档整列高亮。 -->
  <section class="overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-[0_18px_44px_-34px_rgba(15,23,42,0.4)] dark:border-dark-700/60 dark:bg-dark-900 dark:shadow-none">
    <header class="flex flex-wrap items-center gap-2.5 border-b border-gray-100 bg-gray-50/40 px-4 py-3 dark:border-dark-700/60 dark:bg-dark-800/30">
      <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-white ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-600">
        <BrandIcon v-if="brand" :brand="brand" size="18px" />
        <span v-else class="text-sm font-bold text-gray-500 dark:text-gray-300">{{ (label || 'API').charAt(0) }}</span>
      </span>
      <h3 class="text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ label }}</h3>
      <span v-if="periodLabel" class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-200">{{ periodLabel }}</span>
      <span v-if="allExclusive" class="rounded-full bg-violet-50 px-2 py-0.5 text-[11px] font-semibold text-violet-700 ring-1 ring-violet-200 dark:bg-violet-900/30 dark:text-violet-300 dark:ring-violet-900/50">{{ t('payment.admin.kindBadgeExclusive') }}</span>
      <span class="ml-auto text-[11px] tabular-nums text-gray-400 dark:text-dark-500">{{ t('payment.tierCount', { count: plans.length }) }}</span>
    </header>

    <div class="overflow-x-auto">
      <table class="w-full border-collapse text-sm">
        <thead>
          <tr>
            <th class="sticky left-0 z-10 w-24 bg-white dark:bg-dark-900"></th>
            <th
              v-for="(plan, i) in plans"
              :key="plan.id"
              :class="['relative min-w-[10.5rem] px-4 text-center align-top', colClass(plan)]"
            >
              <!-- 右上角真角标：售罄 > 促销角标 -->
              <div v-if="isSoldOut(plan) || badgeOf(plan)" class="absolute right-2.5 top-2.5 z-10">
                <span v-if="isSoldOut(plan)" class="inline-flex items-center rounded-md bg-rose-500 px-1.5 py-0.5 text-[10px] font-bold text-white shadow-sm">{{ t('payment.admin.stockSoldOut') }}</span>
                <span v-else :class="['inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-bold shadow-sm', badgeToneClass(plan.badge_color)]">{{ badgeOf(plan) }}</span>
              </div>
              <!-- 档位名（专属图标 + 配色）+ 价格（原价/折扣横放右侧）-->
              <div class="pb-5 pt-6">
                <div class="flex items-center justify-center gap-1.5">
                  <Icon v-if="themes[i].icon" :name="themes[i].icon!" size="sm" :stroke-width="2" :class="themes[i].iconClass" />
                  <span :class="['text-sm font-bold', themes[i].nameClass]">{{ tierLabels[i] }}</span>
                </div>
                <div class="mt-2 flex items-center justify-center gap-2">
                  <div class="flex items-baseline gap-0.5">
                    <span class="text-base font-semibold text-gray-400 dark:text-dark-400">¥</span>
                    <span class="text-[30px] font-bold tabular-nums leading-none tracking-tight text-gray-900 dark:text-white">{{ plan.price }}</span>
                  </div>
                  <div v-if="plan.original_price && plan.original_price > 0" class="flex flex-col items-start gap-0.5 leading-none">
                    <span class="text-[11px] text-gray-400 line-through decoration-gray-300 dark:text-dark-500">¥{{ plan.original_price }}</span>
                    <span v-if="discountOf(plan)" class="text-[10px] font-bold text-emerald-600 dark:text-emerald-400">{{ discountOf(plan) }}</span>
                  </div>
                </div>
              </div>
            </th>
          </tr>
        </thead>

        <tbody>
          <!-- 限额区：行间用极淡分隔，整体留白为主 -->
          <tr v-if="showRate" class="border-t border-gray-100/70 dark:border-dark-700/40">
            <th class="sticky left-0 z-10 bg-white px-4 py-3 text-left text-[13px] font-medium text-gray-400 dark:bg-dark-900 dark:text-dark-400">{{ t('payment.planCard.rate') }}</th>
            <td v-for="plan in plans" :key="plan.id" :class="['px-4 py-3 text-center font-semibold tabular-nums text-gray-900 dark:text-white', colClass(plan)]">{{ rateOf(plan) }}</td>
          </tr>
          <tr v-if="showDaily" class="border-t border-gray-100/70 dark:border-dark-700/40">
            <th class="sticky left-0 z-10 bg-white px-4 py-3 text-left text-[13px] font-medium text-gray-400 dark:bg-dark-900 dark:text-dark-400">{{ t('payment.planCard.dailyLimit') }}</th>
            <td v-for="plan in plans" :key="plan.id" :class="['px-4 py-3 text-center', colClass(plan)]"><LimitCell :value="plan.daily_limit_usd" /></td>
          </tr>
          <tr v-if="showWeekly" class="border-t border-gray-100/70 dark:border-dark-700/40">
            <th class="sticky left-0 z-10 bg-white px-4 py-3 text-left text-[13px] font-medium text-gray-400 dark:bg-dark-900 dark:text-dark-400">{{ t('payment.planCard.weeklyLimit') }}</th>
            <td v-for="plan in plans" :key="plan.id" :class="['px-4 py-3 text-center', colClass(plan)]"><LimitCell :value="plan.weekly_limit_usd" /></td>
          </tr>
          <tr v-if="showMonthly" class="border-t border-gray-100/70 dark:border-dark-700/40">
            <th class="sticky left-0 z-10 bg-white px-4 py-3 text-left text-[13px] font-medium text-gray-400 dark:bg-dark-900 dark:text-dark-400">{{ t('payment.planCard.monthlyLimit') }}</th>
            <td v-for="plan in plans" :key="plan.id" :class="['px-4 py-3 text-center', colClass(plan)]"><LimitCell :value="plan.monthly_limit_usd" /></td>
          </tr>
          <tr v-if="!showDaily && !showWeekly && !showMonthly" class="border-t border-gray-100/70 dark:border-dark-700/40">
            <th class="sticky left-0 z-10 bg-white px-4 py-3 text-left text-[13px] font-medium text-gray-400 dark:bg-dark-900 dark:text-dark-400">{{ t('payment.planCard.quota') }}</th>
            <td v-for="plan in plans" :key="plan.id" :class="['px-4 py-3 text-center text-base font-bold text-emerald-600 dark:text-emerald-400', colClass(plan)]">{{ t('payment.planCard.unlimited') }}</td>
          </tr>
        </tbody>

        <tfoot>
          <tr>
            <th class="sticky left-0 z-10 bg-white px-4 pb-5 pt-4 dark:bg-dark-900"></th>
            <td v-for="plan in plans" :key="plan.id" :class="['px-4 pb-5 pt-4 align-top', colClass(plan)]">
              <button
                type="button"
                :disabled="isSoldOut(plan)"
                :class="[
                  'w-full rounded-xl px-3 py-2.5 text-[13px] font-semibold transition-all active:scale-[0.99]',
                  isSoldOut(plan)
                    ? 'cursor-not-allowed bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500'
                    : ctaClassOf(plan) || 'bg-gray-900 text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100',
                ]"
                @click="!isSoldOut(plan) && emit('select', plan)"
              >
                {{ ctaText(plan) }}
              </button>
            </td>
          </tr>
        </tfoot>
      </table>
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

// 限额单元格：value 为 null/<=0 视为无限制，否则展示 $额度。内联函数式组件避免重复模板。
const LimitCell: FunctionalComponent<{ value: number | null | undefined }> = (cellProps) => {
  const v = cellProps.value
  if (v != null && v > 0) {
    return h('span', { class: 'text-base font-bold tabular-nums text-gray-900 dark:text-white' }, `$${v}`)
  }
  return h('span', { class: 'text-base font-bold text-emerald-600 dark:text-emerald-400' }, t('payment.planCard.unlimited'))
}

const allExclusive = computed(() => props.plans.length > 0 && props.plans.every(p => p.kind === 'exclusive'))

// 倍率行：任一档位倍率 ≠ 1 才显示，全 ×1 时这行是噪音
const showRate = computed(() => props.plans.some(p => (p.rate_multiplier ?? 1) !== 1))
const showDaily = computed(() => props.plans.some(p => p.daily_limit_usd != null))
const showWeekly = computed(() => props.plans.some(p => p.weekly_limit_usd != null))
const showMonthly = computed(() => props.plans.some(p => p.monthly_limit_usd != null))

// 档位名：admin 填了 plan_label 优先用；否则自动从套餐名推导
const tierLabels = computed(() => {
  const derived = deriveTierLabels(props.plans.map(p => p.name), props.periodLabel)
  return props.plans.map((p, i) => (p.plan_label || '').trim() || derived[i])
})

// 自动推导：先剥同组套餐名的公共前缀（GPT周卡-Lite → Lite），
// 再去掉开头漏出来的周期词（如命名不统一导致的 "周卡-Lite" → "Lite"）；任一档剥空则整体回退完整名。
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
  return pct > 0 ? `-${pct}%` : ''
}

function badgeOf(plan: SubscriptionPlan): string {
  return (plan.badge_text || '').trim()
}

// 每个档位的样式主题（图标 / 配色 / 列底 / CTA），按列索引取用
const themes = computed(() => props.plans.map(p => tierTheme(p.tier_style)))

// 高档（豪华/至尊）整列点缀底色，做视觉突出（替代原"推荐"高亮）
function colClass(plan: SubscriptionPlan): string {
  return tierTheme(plan.tier_style).columnClass
}

// 该档 CTA 专属配色（豪华金 / 至尊黑金）；空则组件用默认黑按钮
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

// 与该档位同 group 的活跃订阅剩余天数；无则返回 null
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

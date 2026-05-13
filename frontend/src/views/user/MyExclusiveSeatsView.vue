<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 工具栏：左侧概览 chip / 右侧操作 -->
      <div v-if="!loading && seats.length > 0" class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
          >
            <Icon name="badge" size="xs" class="text-gray-400" />
            <span class="tabular-nums">{{ seats.length }}</span>
            <span class="text-gray-400 dark:text-dark-400">{{ t('exclusiveSeats.statTotal') }}</span>
          </span>
          <span
            v-if="activeCount > 0"
            class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30"
          >
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span class="tabular-nums">{{ activeCount }}</span>
            <span class="opacity-70">{{ t('exclusiveSeats.statActive') }}</span>
          </span>
          <span
            v-if="expiringSoonCount > 0"
            class="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30"
          >
            <Icon name="clock" size="xs" />
            <span class="tabular-nums">{{ expiringSoonCount }}</span>
            <span class="opacity-70">{{ t('exclusiveSeats.statExpiringSoon') }}</span>
          </span>
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
            @click="loadSeats"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button
            type="button"
            class="btn btn-primary btn-sm"
            @click="router.push('/purchase')"
          >
            <Icon name="plus" size="sm" class="mr-1.5" />
            {{ t('exclusiveSeats.browsePlans') }}
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <LoadingSpinner size="md" />
      </div>

      <!-- Empty -->
      <EmptyState
        v-else-if="seats.length === 0"
        variant="violet"
        :description="t('exclusiveSeats.empty')"
        :action-text="t('exclusiveSeats.browsePlans')"
        @action="router.push('/purchase')"
      >
        <template #icon>
          <Icon name="badge" class="empty-state-icon" />
        </template>
      </EmptyState>

      <!-- 使用引导：用户有 active seat 时提醒需要 ApiKey；Notion 风克制 -->
      <div
        v-else-if="hasActiveSeats"
        class="flex items-start gap-3 rounded-2xl border border-sky-200/70 bg-sky-50/60 px-5 py-3.5 dark:border-sky-500/30 dark:bg-sky-500/10"
      >
        <div class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-white/80 text-sky-600 ring-1 ring-inset ring-sky-200/70 dark:bg-sky-500/15 dark:text-sky-300 dark:ring-sky-500/30">
          <Icon name="infoCircle" size="sm" :stroke-width="2.2" />
        </div>
        <div class="flex-1 text-sm text-sky-900 dark:text-sky-100">
          <p class="font-semibold">{{ t('exclusiveSeats.usageGuideTitle') }}</p>
          <p class="mt-0.5 text-xs text-sky-700/80 dark:text-sky-200/80">
            {{ t('exclusiveSeats.usageGuideHint') }}
          </p>
        </div>
        <button
          class="shrink-0 rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-sky-700"
          @click="router.push('/keys')">
          {{ t('exclusiveSeats.goToKeys') }}
        </button>
      </div>

      <!-- Seats list -->
      <div v-if="seats.length > 0" class="grid gap-4 lg:grid-cols-2">
        <div v-for="seat in seats" :key="seat.id"
          class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-shadow hover:shadow-[0_1px_2px_rgba(15,23,42,0.04),0_8px_24px_-18px_rgba(15,23,42,0.22)] dark:border-dark-700/60 dark:bg-dark-800/40">
          <!-- Header -->
          <div class="flex items-center gap-3 border-b border-gray-100 px-4 py-3 dark:border-dark-700/60">
            <div :class="['flex h-10 w-10 items-center justify-center rounded-xl', statusIconClass(seat.status)]">
              <Icon name="badge" size="md" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-sm font-semibold text-gray-900 dark:text-white truncate">{{ seat.plan_name || `Plan #${seat.plan_id}` }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ seat.account_label }}</p>
            </div>
            <span :class="['inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 ring-inset', statusPillClass(seat.status)]">
              <span class="h-1.5 w-1.5 rounded-full" :class="statusDotClass(seat.status)"></span>
              {{ t(`exclusiveSeats.status.${seat.status}`) }}
            </span>
          </div>

          <!-- Body -->
          <div class="space-y-2 px-4 py-3 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ t('exclusiveSeats.assignedAt') }}</span>
              <span class="text-gray-900 dark:text-white">{{ formatDate(seat.assigned_at) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ t('exclusiveSeats.expiresAt') }}</span>
              <span :class="['font-medium', expiryColor(seat)]">
                {{ formatDate(seat.expires_at) }}
                <span v-if="seat.status === 'active'" class="ml-1 text-xs">
                  ({{ daysRemainingLabel(seat.expires_at) }})
                </span>
              </span>
            </div>
            <div v-if="seat.last_renewal_at" class="flex justify-between text-xs">
              <span class="text-gray-400 dark:text-gray-500">{{ t('exclusiveSeats.lastRenewalAt') }}</span>
              <span class="text-gray-500 dark:text-gray-400">{{ formatDate(seat.last_renewal_at) }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-500 dark:text-gray-400">{{ t('exclusiveSeats.usage') }}</span>
              <span class="text-gray-900 dark:text-white">${{ seat.usage_usd.toFixed(4) }}</span>
            </div>

            <!-- 日/周/月窗口用量进度条：有 limit 才画条，没有 limit 则跳过 -->
            <div
              v-if="usageWindows(seat).length > 0"
              class="mt-2 space-y-1.5 rounded-lg bg-gray-50 p-2.5 dark:bg-dark-800/60"
            >
              <div v-for="w in usageWindows(seat)" :key="w.key">
                <div class="flex items-baseline justify-between text-xs">
                  <span class="text-gray-500 dark:text-gray-400">{{ w.label }}</span>
                  <span class="font-mono tabular-nums" :class="usageTextColor(w.ratio)">
                    ${{ w.used.toFixed(4) }} / ${{ w.limit.toFixed(2) }}
                  </span>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                  <div
                    class="h-full rounded-full transition-all"
                    :class="usageBarColor(w.ratio)"
                    :style="{ width: Math.min(100, Math.max(0, w.ratio * 100)).toFixed(1) + '%' }"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Footer actions -->
          <div v-if="canRenew(seat)" class="flex border-t border-gray-100 dark:border-dark-700/60">
            <button class="flex-1 py-2.5 text-sm font-medium text-emerald-600 transition-colors hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-500/10"
              @click="openRenew(seat)">
              <Icon name="refresh" size="sm" class="mr-1 inline-block" />
              {{ t('exclusiveSeats.renew') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Renew preview dialog: 显示价格与续费时长，点击"前往支付"跳转购买页继续走支付流程 -->
    <BaseDialog :show="!!renewTarget" :title="t('exclusiveSeats.renewTitle')" width="narrow" @close="renewTarget = null">
      <div v-if="renewTarget && renewPreview" class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ t('exclusiveSeats.renewHint', { name: renewPreview.plan_name }) }}
        </p>

        <!-- 价格变动提示块（三态都展示，保持高度一致）：涨价 amber、降价 emerald、不变中性灰 -->
        <div v-if="renewPreview.last_paid_price > 0"
          class="flex items-start gap-2.5 rounded-lg border-l-4 px-3 py-2.5"
          :class="renewDialogTrend === 'up'
            ? 'border-amber-500 bg-amber-50 dark:bg-amber-950/30'
            : renewDialogTrend === 'down'
              ? 'border-emerald-500 bg-emerald-50 dark:bg-emerald-950/30'
              : 'border-gray-300 bg-gray-50 dark:border-dark-600 dark:bg-dark-800/60'">
          <Icon
            :name="renewDialogTrend === 'up' ? 'exclamationTriangle' : 'check'"
            size="sm"
            class="mt-0.5 shrink-0"
            :class="renewDialogTrend === 'up'
              ? 'text-amber-600 dark:text-amber-300'
              : renewDialogTrend === 'down'
                ? 'text-emerald-600 dark:text-emerald-300'
                : 'text-gray-500 dark:text-dark-300'" />
          <div class="flex-1 text-xs leading-5"
            :class="renewDialogTrend === 'up'
              ? 'text-amber-800 dark:text-amber-200'
              : renewDialogTrend === 'down'
                ? 'text-emerald-800 dark:text-emerald-200'
                : 'text-gray-700 dark:text-dark-200'">
            <span v-if="renewDialogTrend === 'up'">{{ t('exclusiveSeats.priceIncreasedWarning', { delta: (renewPreview.price - renewPreview.last_paid_price).toFixed(2) }) }}</span>
            <span v-else-if="renewDialogTrend === 'down'">{{ t('exclusiveSeats.priceDecreasedNote', { delta: (renewPreview.last_paid_price - renewPreview.price).toFixed(2) }) }}</span>
            <span v-else>{{ t('exclusiveSeats.priceUnchangedNote') }}</span>
          </div>
        </div>

        <!-- 续费明细卡：行间距加大、字号 hierarchy 更明显 -->
        <div class="rounded-xl border border-gray-200 bg-gray-50/60 p-5 dark:border-dark-700 dark:bg-dark-800/40">
          <!-- 有效期 -->
          <div class="flex items-center justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('exclusiveSeats.validityDays') }}</span>
            <span class="font-medium tabular-nums">{{ renewPreview.validity_days }} {{ renewPreview.validity_unit }}</span>
          </div>

          <!-- 上次价行：行间距 mt-3，删除线加粗到 decoration-2 让"过时"语义更明确 -->
          <div v-if="renewPreview.last_paid_price > 0" class="mt-3 flex items-center justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.renewalBanner.lastPaid') }}</span>
            <span class="font-mono tabular-nums"
              :class="renewDialogTrend === 'same'
                ? 'text-gray-700 dark:text-dark-100'
                : 'text-gray-400 line-through decoration-2 decoration-gray-300/80 dark:text-dark-400 dark:decoration-dark-600/80'">
              ¥{{ renewPreview.last_paid_price.toFixed(2) }}
            </span>
          </div>

          <!-- 本次价：跟其他行一致结构（label 左、金额右），chip 内联到金额左侧 -->
          <div class="mt-3 flex items-baseline justify-between">
            <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.actualPay') }}</span>
            <div class="flex items-baseline gap-2">
              <span v-if="renewDialogTrend === 'up'"
                class="inline-flex items-center rounded-md bg-amber-100/80 px-1.5 py-0.5 text-[10px] font-bold tabular-nums text-amber-700 dark:bg-amber-900/40 dark:text-amber-300">
                +¥{{ (renewPreview.price - renewPreview.last_paid_price).toFixed(2) }}
              </span>
              <span v-else-if="renewDialogTrend === 'down'"
                class="inline-flex items-center rounded-md bg-emerald-100/80 px-1.5 py-0.5 text-[10px] font-bold tabular-nums text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
                −¥{{ (renewPreview.last_paid_price - renewPreview.price).toFixed(2) }}
              </span>
              <span v-else
                class="inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300">
                {{ t('payment.renewalBanner.priceSame') }}
              </span>
              <span class="font-mono text-2xl font-bold tabular-nums leading-none"
                :class="renewDialogTrend === 'up' ? 'text-amber-600 dark:text-amber-300'
                  : renewDialogTrend === 'down' ? 'text-emerald-600 dark:text-emerald-300'
                  : 'text-primary-600 dark:text-primary-400'">
                ¥{{ renewPreview.price.toFixed(2) }}
              </span>
            </div>
          </div>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('exclusiveSeats.renewPaymentHint') }}</p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="renewTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="!renewPreview" @click="goToRenewPayment">
            {{ t('exclusiveSeats.goToPay') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { ExclusiveSeat } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const seats = ref<ExclusiveSeat[]>([])
const renewTarget = ref<ExclusiveSeat | null>(null)
const renewPreview = ref<{
  seat_id: number
  plan_id: number
  plan_name: string
  price: number
  validity_days: number
  validity_unit: string
  current_expires_at: string
  last_paid_price: number
} | null>(null)

const hasActiveSeats = computed(() => seats.value.some((s) => s.status === 'active'))

// 续费 dialog 内的价格变动方向（与 PaymentView 一致，差额 < 0.005 视为不变）
const renewDialogTrend = computed<'up' | 'down' | 'same'>(() => {
  const p = renewPreview.value
  if (!p || p.last_paid_price <= 0) return 'same'
  const delta = p.price - p.last_paid_price
  if (delta > 0.005) return 'up'
  if (delta < -0.005) return 'down'
  return 'same'
})

async function loadSeats() {
  loading.value = true
  try {
    const res = await paymentAPI.getMyExclusiveSeats()
    seats.value = res.data.items || []
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function canRenew(seat: ExclusiveSeat): boolean {
  if (seat.status === 'active') return true
  // 过期 7 天宽限期内可补缴
  if (seat.status === 'expired') {
    const expiredMs = Date.now() - Date.parse(seat.expires_at)
    return expiredMs > 0 && expiredMs <= 7 * 24 * 60 * 60 * 1000
  }
  return false
}

async function openRenew(seat: ExclusiveSeat) {
  renewTarget.value = seat
  renewPreview.value = null
  try {
    const res = await paymentAPI.previewRenewal(seat.id)
    renewPreview.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
    renewTarget.value = null
  }
}

// 续费走完整支付链路：跳转到购买页，PaymentView 检测到 query.renew_seat 后
// 自动选中对应 plan 并把 renewal_seat_id 透传到 CreateOrder。
function goToRenewPayment() {
  if (!renewTarget.value || !renewPreview.value) return
  router.push({
    path: '/purchase',
    query: {
      tab: 'subscription',
      plan_id: String(renewPreview.value.plan_id),
      renew_seat: String(renewTarget.value.id),
    },
  })
  renewTarget.value = null
}

// 状态左侧大图标筐配色：克制软底 + 主题色文字
function statusIconClass(status: string): string {
  switch (status) {
    case 'active': return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'expired': return 'bg-amber-50 text-amber-600 dark:bg-amber-500/15 dark:text-amber-300'
    case 'refunded': return 'bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300'
    default: return 'bg-gray-50 text-gray-500 dark:bg-dark-700/40 dark:text-gray-400'
  }
}

// 状态 chip：与全站 ring inset 软色调统一
function statusPillClass(status: string): string {
  switch (status) {
    case 'active': return 'bg-emerald-50 text-emerald-700 ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30'
    case 'expired': return 'bg-amber-50 text-amber-700 ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30'
    case 'refunded': return 'bg-violet-50 text-violet-700 ring-violet-200/70 dark:bg-violet-500/15 dark:text-violet-300 dark:ring-violet-500/30'
    default: return 'bg-gray-50 text-gray-600 ring-gray-200/70 dark:bg-dark-700/40 dark:text-gray-400 dark:ring-dark-600/60'
  }
}

function statusDotClass(status: string): string {
  switch (status) {
    case 'active': return 'bg-emerald-500 animate-pulse'
    case 'expired': return 'bg-amber-500'
    case 'refunded': return 'bg-violet-500'
    default: return 'bg-gray-400'
  }
}

function expiryColor(seat: ExclusiveSeat): string {
  if (seat.status !== 'active') return 'text-gray-500 dark:text-gray-400'
  const remainingMs = Date.parse(seat.expires_at) - Date.now()
  if (remainingMs <= 24 * 60 * 60 * 1000) return 'text-rose-600 dark:text-rose-400'
  if (remainingMs <= 7 * 24 * 60 * 60 * 1000) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-900 dark:text-white'
}

function daysRemainingLabel(expiresAt: string): string {
  const ms = Date.parse(expiresAt) - Date.now()
  const days = Math.ceil(ms / (24 * 60 * 60 * 1000))
  if (days <= 0) return t('exclusiveSeats.expiringNow')
  return t('exclusiveSeats.daysLeft', { days })
}

function formatDate(s: string): string {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  return d.toLocaleString()
}

// 用量窗口聚合：只输出"有 limit 且 > 0"的窗口，避免无限额池上画一条 0% 条
type UsageWindow = { key: 'daily' | 'weekly' | 'monthly'; label: string; used: number; limit: number; ratio: number }
function usageWindows(seat: ExclusiveSeat): UsageWindow[] {
  const out: UsageWindow[] = []
  const push = (key: UsageWindow['key'], label: string, used: number | undefined, limit: number | null | undefined) => {
    if (!limit || limit <= 0) return
    const u = used || 0
    out.push({ key, label, used: u, limit, ratio: u / limit })
  }
  push('daily', t('exclusiveSeats.usageDaily'), seat.daily_usage_usd, seat.daily_limit_usd)
  push('weekly', t('exclusiveSeats.usageWeekly'), seat.weekly_usage_usd, seat.weekly_limit_usd)
  push('monthly', t('exclusiveSeats.usageMonthly'), seat.monthly_usage_usd, seat.monthly_limit_usd)
  return out
}

// 用量进度色阶（与全站统一）：<80% emerald / 80–100% amber / ≥100% rose
function usageBarColor(ratio: number): string {
  if (ratio >= 1) return 'bg-rose-500'
  if (ratio >= 0.8) return 'bg-amber-500'
  return 'bg-emerald-500'
}
function usageTextColor(ratio: number): string {
  if (ratio >= 1) return 'text-rose-600 dark:text-rose-400'
  if (ratio >= 0.8) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-700 dark:text-gray-200'
}

// 概览统计：active 数量 / 即将过期 (≤7 天) 数量 — 用于工具栏概览 chip
const activeCount = computed(
  () => seats.value.filter(s => s.status === 'active').length,
)

const expiringSoonCount = computed(() => {
  const now = Date.now()
  const sevenDaysMs = 7 * 24 * 60 * 60 * 1000
  return seats.value.filter(s => {
    if (s.status !== 'active' || !s.expires_at) return false
    const diff = Date.parse(s.expires_at) - now
    return diff > 0 && diff <= sevenDaysMs
  }).length
})

onMounted(loadSeats)
</script>

<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部操作栏 + 概览 chip 行：标题由全局 AppHeader 提供，这里只放操作按钮 -->
      <div class="flex flex-wrap items-center justify-end gap-2">
        <button
          type="button"
          class="btn btn-secondary btn-sm"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
          @click="loadSubscriptions"
        >
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button
          type="button"
          class="btn btn-primary btn-sm"
          @click="router.push({ path: '/purchase', query: { tab: 'subscription' } })"
        >
          <Icon name="plus" size="sm" class="mr-1.5" />
          {{ t('payment.subscribeNow') }}
        </button>
      </div>

      <div v-if="!loading" class="flex flex-wrap items-center gap-2">
        <span
          class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
        >
          <Icon name="badge" size="xs" class="text-gray-400" />
          <span class="tabular-nums">{{ subscriptions.length }}</span>
          <span class="text-gray-400 dark:text-dark-400">{{ t('userSubscriptions.statTotal') }}</span>
        </span>
        <span
          v-if="activeCount > 0"
          class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 ring-1 ring-inset ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          <span class="tabular-nums">{{ activeCount }}</span>
          <span class="opacity-70">{{ t('userSubscriptions.statActive') }}</span>
        </span>
        <span
          v-if="expiringSoonCount > 0"
          class="inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30"
        >
          <Icon name="clock" size="xs" />
          <span class="tabular-nums">{{ expiringSoonCount }}</span>
          <span class="opacity-70">{{ t('userSubscriptions.statExpiringSoon') }}</span>
        </span>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <LoadingSpinner size="md" />
      </div>

      <!-- Empty State -->
      <EmptyState
        v-else-if="subscriptions.length === 0"
        variant="emerald"
        :title="t('userSubscriptions.noActiveSubscriptions')"
        :description="t('userSubscriptions.noActiveSubscriptionsDesc')"
        :action-text="t('payment.subscribeNow')"
        @action="router.push({ path: '/purchase', query: { tab: 'subscription' } })"
      >
        <template #icon>
          <Icon name="creditCard" class="empty-state-icon" />
        </template>
      </EmptyState>

      <!-- Subscriptions Grid -->
      <div v-else class="grid gap-5 lg:grid-cols-2">
        <div
          v-for="subscription in subscriptions"
          :key="subscription.id"
          class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-card transition-all duration-200 hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-dark-600"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between gap-3 border-b border-gray-100 p-4 dark:border-dark-700/60"
          >
            <div class="flex items-start gap-3 min-w-0">
              <div :class="['mt-2 h-1.5 w-1.5 shrink-0 rounded-full', platformAccentDotClass(subscription.group?.platform || '')]" />
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <h3 class="font-semibold text-gray-900 dark:text-white truncate">
                    {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                  </h3>
                  <span :class="['inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-2xs font-medium', platformBadgeClass(subscription.group?.platform || '')]">
                    {{ platformLabel(subscription.group?.platform || '') }}
                  </span>
                </div>
                <p v-if="subscription.group?.description" class="mt-0.5 text-xs leading-relaxed text-gray-500 dark:text-dark-400 line-clamp-2">
                  {{ subscription.group.description }}
                </p>
              </div>
            </div>
            <div class="flex flex-col items-end gap-2 shrink-0">
              <span
                :class="[
                  'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset',
                  statusChipClass(subscription.status),
                ]"
              >
                <span class="h-1.5 w-1.5 rounded-full" :class="statusDotClass(subscription.status)"></span>
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
              <button
                v-if="subscription.status === 'active'"
                :class="['rounded-lg px-3 py-1 text-xs font-semibold text-white transition-colors', platformButtonClass(subscription.group?.platform || '')]"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
          </div>

          <!-- Usage Progress -->
          <div class="space-y-4 p-4">
            <!-- Expiration Info -->
            <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(subscription.expires_at)">
                {{ formatExpirationDate(subscription.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <!-- Daily Usage -->
            <div v-if="getDailyLimit(subscription)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  {{ formatUsage(subscription.daily_usage_usd, getDailyLimit(subscription)) }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      getDailyLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      getDailyLimit(subscription)
                    )
                  }"
                ></div>
              </div>
              <p class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('common.remainingQuota') }}:
                {{ formatUSD(getRemaining(subscription.daily_usage_usd, getDailyLimit(subscription))) }}
                <span v-if="subscription.daily_window_start">
                  |
                  {{
                    t('userSubscriptions.resetIn', {
                      time: formatResetTime(subscription.daily_window_start, 24)
                    })
                  }}
                </span>
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="getWeeklyLimit(subscription)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  {{ formatUsage(subscription.weekly_usage_usd, getWeeklyLimit(subscription)) }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.weekly_usage_usd,
                      getWeeklyLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.weekly_usage_usd,
                      getWeeklyLimit(subscription)
                    )
                  }"
                ></div>
              </div>
              <p class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('common.remainingQuota') }}:
                {{ formatUSD(getRemaining(subscription.weekly_usage_usd, getWeeklyLimit(subscription))) }}
                <span v-if="subscription.weekly_window_start">
                  |
                  {{
                    t('userSubscriptions.resetIn', {
                      time: formatResetTime(subscription.weekly_window_start, 168)
                    })
                  }}
                </span>
              </p>
            </div>

            <!-- Monthly Usage -->
            <div v-if="getMonthlyLimit(subscription)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  {{ formatUsage(subscription.monthly_usage_usd, getMonthlyLimit(subscription)) }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.monthly_usage_usd,
                      getMonthlyLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.monthly_usage_usd,
                      getMonthlyLimit(subscription)
                    )
                  }"
                ></div>
              </div>
              <p class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('common.remainingQuota') }}:
                {{ formatUSD(getRemaining(subscription.monthly_usage_usd, getMonthlyLimit(subscription))) }}
                <span v-if="subscription.monthly_window_start">
                  |
                  {{
                    t('userSubscriptions.resetIn', {
                      time: formatResetTime(subscription.monthly_window_start, 720)
                    })
                  }}
                </span>
              </p>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="!hasUsageLimit(subscription)"
              class="flex items-center justify-center gap-3 rounded-2xl border border-emerald-200/70 bg-emerald-50/60 py-5 dark:border-emerald-500/30 dark:bg-emerald-500/10"
            >
              <span class="text-3xl font-light leading-none text-emerald-600 dark:text-emerald-400">∞</span>
              <div>
                <p class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">
                  {{ t('userSubscriptions.unlimited') }}
                </p>
                <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                  {{ t('userSubscriptions.unlimitedDesc') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatDateOnly } from '@/utils/format'
import { platformBadgeClass, platformButtonClass, platformLabel } from '@/utils/platformColors'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-orange-500'
    case 'openai': return 'bg-emerald-500'
    case 'antigravity': return 'bg-purple-500'
    case 'gemini': return 'bg-blue-500'
    default: return 'bg-gray-400'
  }
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function validLimit(limit: number | null | undefined): number | null {
  return typeof limit === 'number' && limit > 0 ? limit : null
}

function getEffectiveLimit(
  subscriptionLimit: number | null | undefined,
  groupLimit: number | null | undefined
): number | null {
  return validLimit(subscriptionLimit) ?? validLimit(groupLimit)
}

function getDailyLimit(subscription: UserSubscription): number | null {
  return getEffectiveLimit(
    subscription.daily_limit_usd,
    subscription.group?.daily_limit_usd
  )
}

function getWeeklyLimit(subscription: UserSubscription): number | null {
  return getEffectiveLimit(
    subscription.weekly_limit_usd,
    subscription.group?.weekly_limit_usd
  )
}

function getMonthlyLimit(subscription: UserSubscription): number | null {
  return getEffectiveLimit(
    subscription.monthly_limit_usd,
    subscription.group?.monthly_limit_usd
  )
}

function hasUsageLimit(subscription: UserSubscription): boolean {
  return Boolean(getDailyLimit(subscription) || getWeeklyLimit(subscription) || getMonthlyLimit(subscription))
}

function getRemaining(used: number | undefined, limit: number | null | undefined): number {
  if (!limit || limit <= 0) return 0
  return Math.max(limit - (used || 0), 0)
}

function formatUSD(value: number): string {
  return `$${value.toFixed(2)}`
}

function formatUsage(used: number | undefined, limit: number | null | undefined): string {
  return `${formatUSD(used || 0)} / ${limit ? formatUSD(limit) : '∞'}`
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

// 用量进度色阶（与全站统一）：<80% emerald 健康 / 80–100% amber 接近上限 / ≥100% rose 已耗尽
function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 100) return 'bg-rose-500'
  if (percentage >= 80) return 'bg-amber-500'
  return 'bg-emerald-500'
}

// 订阅状态 chip：active 绿 / expired 琥珀（提示续费） / 其他 (cancelled 等) 玫红
function statusChipClass(status: string): string {
  switch (status) {
    case 'active':
      return 'bg-emerald-50 text-emerald-700 ring-emerald-200/70 dark:bg-emerald-500/15 dark:text-emerald-300 dark:ring-emerald-500/30'
    case 'expired':
      return 'bg-amber-50 text-amber-700 ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30'
    default:
      return 'bg-rose-50 text-rose-700 ring-rose-200/70 dark:bg-rose-500/15 dark:text-rose-300 dark:ring-rose-500/30'
  }
}

function statusDotClass(status: string): string {
  switch (status) {
    case 'active':
      return 'bg-emerald-500 animate-pulse'
    case 'expired':
      return 'bg-amber-500'
    default:
      return 'bg-rose-500'
  }
}

// 概览统计：active 数量 / 即将过期 (≤7 天) 数量 — 用于工具栏概览 chip
const activeCount = computed(
  () => subscriptions.value.filter(s => s.status === 'active').length,
)

const expiringSoonCount = computed(() => {
  const now = Date.now()
  const sevenDaysMs = 7 * 24 * 60 * 60 * 1000
  return subscriptions.value.filter(s => {
    if (s.status !== 'active' || !s.expires_at) return false
    const diff = new Date(s.expires_at).getTime() - now
    return diff > 0 && diff <= sevenDaysMs
  }).length
})

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateOnly(expires)

  if (days === 0) {
    return `${dateStr} (${t('common.today')})`
  }
  if (days === 1) {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const now = new Date()
  const diff = end.getTime() - now.getTime()

  if (diff <= 0) return t('userSubscriptions.windowNotActive')

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

  if (hours > 24) {
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24
    return `${days}d ${remainingHours}h`
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m`
}

onMounted(() => {
  loadSubscriptions()
})
</script>

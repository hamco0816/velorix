<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 移动端 2 列防止 4 张卡纵向堆成长长一条；md+ 仍可 2 列；xl 起 4 列 -->
      <section class="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
            <div class="redeem-stat-card">
              <div class="mb-4 flex items-start justify-between gap-3 sm:mb-6 sm:gap-4">
                <div class="flex h-10 w-10 items-center justify-center sm:h-12 sm:w-12 rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-900/25 dark:text-sky-300">
                  <Icon name="creditCard" size="lg" />
                </div>
                <span class="inline-flex items-center rounded-md bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300">
                  <Icon name="shield" size="xs" class="mr-1" />
                  Velorix
                </span>
              </div>
              <p class="text-sm font-medium text-gray-500 dark:text-dark-400">
                {{ t('redeem.currentBalance') }}
              </p>
              <p class="mt-2 text-2xl font-semibold text-gray-950 sm:text-3xl dark:text-white">
                ${{ user?.balance?.toFixed(2) || '0.00' }}
              </p>
            </div>

            <div class="redeem-stat-card">
              <div class="mb-4 flex items-start justify-between gap-3 sm:mb-6 sm:gap-4">
                <div class="flex h-10 w-10 items-center justify-center sm:h-12 sm:w-12 rounded-lg bg-emerald-50 text-emerald-600 dark:bg-emerald-900/25 dark:text-emerald-300">
                  <Icon name="bolt" size="lg" />
                </div>
                <span class="rounded-md bg-sky-50 px-2.5 py-1 text-xs font-medium text-sky-700 dark:bg-sky-900/25 dark:text-sky-300">
                  {{ t('redeem.requests') }}
                </span>
              </div>
              <p class="text-sm font-medium text-gray-500 dark:text-dark-400">
                {{ t('redeem.concurrency') }}
              </p>
              <p class="mt-2 text-2xl font-semibold text-gray-950 sm:text-3xl dark:text-white">
                {{ user?.concurrency || 0 }}
              </p>
            </div>

            <div class="redeem-stat-card">
              <div class="mb-4 flex items-start justify-between gap-3 sm:mb-6 sm:gap-4">
                <div class="flex h-10 w-10 items-center justify-center sm:h-12 sm:w-12 rounded-lg bg-amber-50 text-amber-600 dark:bg-amber-900/25 dark:text-amber-300">
                  <Icon name="badge" size="lg" />
                </div>
                <span class="rounded-md bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700 dark:bg-amber-900/25 dark:text-amber-300">
                  {{ t('redeem.planLabel') }}
                </span>
              </div>
              <p class="text-sm font-medium text-gray-500 dark:text-dark-400">
                {{ t('redeem.currentPlan') }}
              </p>
              <p class="mt-2 truncate text-4xl font-semibold text-gray-950 dark:text-white">
                {{ activePlanLabel }}
              </p>
              <p class="mt-2 truncate text-xs text-gray-400 dark:text-dark-500">
                {{ activePlanMeta }}
              </p>
            </div>

            <div class="redeem-stat-card">
              <div class="mb-4 flex items-start justify-between gap-3 sm:mb-6 sm:gap-4">
                <div class="flex h-10 w-10 items-center justify-center sm:h-12 sm:w-12 rounded-lg bg-teal-50 text-teal-600 dark:bg-teal-900/25 dark:text-teal-300">
                  <Icon name="shield" size="lg" />
                </div>
                <span :class="['rounded-md px-2.5 py-1 text-xs font-medium', accountStatusBadgeClass]">
                  {{ accountStatusLabel }}
                </span>
              </div>
              <p class="text-sm font-medium text-gray-500 dark:text-dark-400">
                {{ t('redeem.accountStatus') }}
              </p>
              <p class="mt-2 text-2xl font-semibold text-gray-950 sm:text-3xl dark:text-white">
                {{ accountStatusLabel }}
              </p>
              <p :class="['mt-2 text-xs font-medium', accountStatusClass]">
                {{ accountStatusMeta }}
              </p>
            </div>
          </section>

      <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_430px]">
        <main class="space-y-5">

          <!-- Redeem Form -->
          <section class="redeem-panel overflow-hidden">
            <div class="border-b border-sky-100 bg-sky-50/70 px-6 py-5 dark:border-dark-700 dark:bg-sky-900/10">
              <div class="flex items-center gap-3">
                <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-white text-sky-600 shadow-sm dark:bg-dark-700 dark:text-sky-300">
                  <Icon name="gift" size="md" />
                </div>
                <div>
                  <h2 class="text-base font-semibold text-gray-950 dark:text-white">
                    {{ t('redeem.redeemCodeLabel') }}
                  </h2>
                  <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">
                    {{ t('redeem.redeemCodeHint') }}
                  </p>
                </div>
              </div>
            </div>

            <div class="p-6">
              <form @submit.prevent="handleRedeem" class="space-y-5">
                <div>
                  <label for="code" class="input-label">
                    {{ t('redeem.redeemCodeLabel') }}
                  </label>
                  <div class="relative mt-2">
                    <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                      <Icon name="gift" size="md" class="text-sky-500 dark:text-sky-300" />
                    </div>
                    <input
                      id="code"
                      v-model="redeemCode"
                      type="text"
                      required
                      :placeholder="t('redeem.redeemCodePlaceholder')"
                      :disabled="submitting"
                      class="input h-14 rounded-xl border-gray-200 py-3 pl-12 pr-24 text-base font-medium shadow-sm shadow-gray-100/60 dark:border-dark-600 dark:shadow-none"
                    />
                    <button
                      type="button"
                      class="absolute inset-y-2 right-2 rounded-lg border border-gray-200 px-3 text-sm font-semibold text-sky-600 transition-colors hover:border-sky-300 hover:bg-sky-50 dark:border-dark-600 dark:text-sky-300 dark:hover:bg-sky-900/20"
                      @click="pasteRedeemCode"
                    >
                      {{ t('redeem.paste') }}
                    </button>
                  </div>
                  <p class="input-hint mt-2">
                    {{ t('redeem.redeemCodeHint') }}
                  </p>
                </div>

                <button
                  type="submit"
                  :disabled="!redeemCode || submitting"
                  class="inline-flex w-full items-center justify-center rounded-xl bg-gradient-to-r from-sky-600 via-blue-600 to-violet-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-blue-500/20 transition hover:brightness-105 disabled:cursor-not-allowed disabled:bg-none disabled:bg-gray-100 disabled:text-gray-400 disabled:shadow-none dark:disabled:bg-dark-700 dark:disabled:text-dark-400"
                >
                  <svg
                    v-if="submitting"
                    class="-ml-1 mr-2 h-5 w-5 animate-spin"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      class="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      stroke-width="4"
                    ></circle>
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <Icon v-else name="checkCircle" size="md" class="mr-2" />
                  {{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}
                </button>
              </form>
            </div>
          </section>

      <!-- Success Message -->
      <transition name="fade">
        <div
          v-if="redeemResult"
          class="redeem-panel border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30"
              >
                <Icon name="checkCircle" size="md" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">
                  {{ t('redeem.redeemSuccess') }}
                </h3>
                <div class="mt-2 text-sm text-emerald-700 dark:text-emerald-400">
                  <p>{{ redeemResult.message }}</p>
                  <div class="mt-3 space-y-1">
                    <p v-if="redeemResult.type === 'balance'" class="font-medium">
                      {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
                    </p>
                    <p v-else-if="redeemResult.type === 'concurrency'" class="font-medium">
                      {{ t('redeem.added') }}: {{ redeemResult.value }}
                      {{ t('redeem.concurrentRequests') }}
                    </p>
                    <p v-else-if="redeemResult.type === 'subscription'" class="font-medium">
                      {{ t('redeem.subscriptionAssigned') }}
                      <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                      <span v-if="redeemResult.validity_days">
                        ({{
                          t('redeem.subscriptionDays', { days: redeemResult.validity_days })
                        }})</span
                      >
                    </p>
                    <p v-if="redeemResult.new_balance !== undefined">
                      {{ t('redeem.newBalance') }}:
                      <span class="font-semibold">${{ redeemResult.new_balance.toFixed(2) }}</span>
                    </p>
                    <p v-if="redeemResult.new_concurrency !== undefined">
                      {{ t('redeem.newConcurrency') }}:
                      <span class="font-semibold"
                        >{{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}</span
                      >
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Error Message -->
      <transition name="fade">
        <div
          v-if="errorMessage"
          class="redeem-panel border-red-200 bg-red-50 dark:border-red-800/50 dark:bg-red-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-red-100 dark:bg-red-900/30"
              >
                <Icon
                  name="exclamationCircle"
                  size="md"
                  class="text-red-600 dark:text-red-400"
                />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-red-800 dark:text-red-300">
                  {{ t('redeem.redeemFailed') }}
                </h3>
                <p class="mt-2 text-sm text-red-700 dark:text-red-400">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Recent Activity -->
      <section class="redeem-panel overflow-hidden">
        <div class="border-b border-gray-100 bg-gray-50/70 px-6 py-4 dark:border-dark-700 dark:bg-dark-800/60">
          <div class="flex items-center gap-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-amber-50 text-amber-600 dark:bg-amber-900/25 dark:text-amber-300">
              <Icon name="clock" size="md" />
            </div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('redeem.recentActivity') }}
            </h2>
          </div>
        </div>
        <div class="p-6">
          <!-- Loading State -->
          <div v-if="loadingHistory" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          </div>

          <!-- History List -->
          <div v-else-if="history.length > 0" class="space-y-3">
            <div
              v-for="item in history"
              :key="item.id"
              class="flex items-center justify-between gap-4 rounded-lg border border-gray-100 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-800"
            >
              <div class="flex items-center gap-4">
                <div
                  :class="[
                    'flex h-10 w-10 items-center justify-center rounded-xl',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : 'bg-red-100 dark:bg-red-900/30'
                      : isSubscriptionType(item.type)
                        ? 'bg-purple-100 dark:bg-purple-900/30'
                        : item.value >= 0
                          ? 'bg-blue-100 dark:bg-blue-900/30'
                          : 'bg-orange-100 dark:bg-orange-900/30'
                  ]"
                >
                  <!-- 余额类型图标 -->
                  <Icon
                    v-if="isBalanceType(item.type)"
                    name="dollar"
                    size="md"
                    :class="
                      item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                    "
                  />
                  <!-- 订阅类型图标 -->
                  <Icon
                    v-else-if="isSubscriptionType(item.type)"
                    name="badge"
                    size="md"
                    class="text-purple-600 dark:text-purple-400"
                  />
                  <!-- 并发类型图标 -->
                  <Icon
                    v-else
                    name="bolt"
                    size="md"
                    :class="
                      item.value >= 0
                        ? 'text-blue-600 dark:text-blue-400'
                        : 'text-orange-600 dark:text-orange-400'
                    "
                  />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ getHistoryItemTitle(item) }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{ formatDateTime(item.used_at) }}
                  </p>
                </div>
              </div>
              <div class="text-right">
                <p
                  :class="[
                    'text-sm font-semibold',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                      : isSubscriptionType(item.type)
                        ? 'text-purple-600 dark:text-purple-400'
                        : item.value >= 0
                          ? 'text-blue-600 dark:text-blue-400'
                          : 'text-orange-600 dark:text-orange-400'
                  ]"
                >
                  {{ formatHistoryValue(item) }}
                </p>
                <p
                  v-if="!isAdminAdjustment(item.type)"
                  class="font-mono text-xs text-gray-400 dark:text-dark-500"
                >
                  {{ item.code.slice(0, 8) }}...
                </p>
                <p v-else class="text-xs text-gray-400 dark:text-dark-500">
                  {{ t('redeem.adminAdjustment') }}
                </p>
                <!-- Display notes for admin adjustments -->
                <p
                  v-if="item.notes"
                  class="mt-1 text-xs text-gray-500 dark:text-dark-400 italic max-w-[200px] truncate"
                  :title="item.notes"
                >
                  {{ item.notes }}
                </p>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="empty-state py-8">
            <div
              class="mb-4 flex h-16 w-16 items-center justify-center rounded-lg bg-amber-50 dark:bg-amber-900/20"
            >
              <Icon name="clock" size="xl" class="text-amber-500 dark:text-amber-300" />
            </div>
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('redeem.historyWillAppear') }}
            </p>
          </div>
        </div>
      </section>
        </main>

        <!-- Information Card -->
        <aside class="space-y-5 xl:sticky xl:top-6 xl:self-start">
          <section class="redeem-panel p-6">
            <div class="mb-5 flex items-center gap-3">
              <div class="flex h-11 w-11 items-center justify-center rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-900/25 dark:text-sky-300">
                <Icon name="infoCircle" size="md" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-950 dark:text-white">
                  {{ t('redeem.aboutCodes') }}
                </h3>
                <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('redeem.syncOnRedeem') }}
                </p>
              </div>
            </div>
            <ul class="space-y-3 text-sm text-gray-700 dark:text-dark-200">
              <li class="flex gap-2">
                <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-sky-500" />
                <span>{{ t('redeem.codeRule1') }}</span>
              </li>
              <li class="flex gap-2">
                <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
                <span>{{ t('redeem.codeRule2') }}</span>
              </li>
              <li class="flex gap-2">
                <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-amber-500" />
                <span>{{ t('redeem.codeRule3') }}</span>
              </li>
              <li class="flex gap-2">
                <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-violet-500" />
                <span>{{ t('redeem.codeRule4') }}</span>
              </li>
            </ul>

            <div
              v-if="hasContactMethods"
              class="mt-5 rounded-xl border border-gray-100 bg-gray-50/80 p-3 dark:border-dark-700 dark:bg-dark-900/40"
            >
              <span class="mb-3 block text-sm font-semibold text-gray-800 dark:text-dark-100">
                {{ t('common.contactSupport') }}
              </span>
              <ContactMethodsDisplay
                :methods="contactMethods"
                :legacy-info="contactInfo"
                compact
                :show-label="false"
              />
            </div>
          </section>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { redeemAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import ContactMethodsDisplay from '@/components/common/ContactMethodsDisplay.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { ContactMethod } from '@/types'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const user = computed(() => authStore.user)

const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
} | null>(null)
const errorMessage = ref('')

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const contactInfo = ref('')
const contactMethods = ref<ContactMethod[]>([])
const hasContactMethods = computed(() => contactMethods.value.length > 0 || !!contactInfo.value)

const activeSubscription = computed(() =>
  subscriptionStore.activeSubscriptions.find((item) => item.status === 'active') || null,
)

const activePlanLabel = computed(() =>
  activeSubscription.value?.group?.name || (activeSubscription.value ? t('redeem.subscriptionAssigned') : t('redeem.notSubscribed')),
)

const activePlanMeta = computed(() => {
  if (!activeSubscription.value) return t('redeem.noSubscription')
  if (!activeSubscription.value.expires_at) return t('redeem.permanentValidity')
  return t('redeem.validUntil', { date: formatDateTime(activeSubscription.value.expires_at) })
})

const accountStatusLabel = computed(() => (user.value?.status === 'active' ? t('redeem.statusNormal') : t('redeem.statusBanned')))
const accountStatusMeta = computed(() =>
  user.value?.status === 'active' ? t('redeem.statusNormalDesc') : t('redeem.statusBannedDesc'),
)
const accountStatusClass = computed(() =>
  user.value?.status === 'active'
    ? 'text-emerald-600 dark:text-emerald-400'
    : 'text-red-600 dark:text-red-400',
)
const accountStatusBadgeClass = computed(() =>
  user.value?.status === 'active'
    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300'
    : 'bg-red-50 text-red-700 dark:bg-red-900/25 dark:text-red-300',
)

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance' || type === 'admin_balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

const isAdminAdjustment = (type: string) => {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const pasteRedeemCode = async () => {
  try {
    const text = await navigator.clipboard?.readText()
    if (text?.trim()) {
      redeemCode.value = text.trim()
    }
  } catch {
    appStore.showWarning(t('redeem.clipboardReadFailed'))
  }
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    redeemResult.value = result

    // Refresh user data to get updated balance/concurrency
    await authStore.refreshUser()

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || t('redeem.failedToRedeem')

    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  fetchHistory()
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load active subscriptions:', error)
  })
  try {
    const settings = await appStore.fetchPublicSettings()
    contactInfo.value = settings?.contact_info || ''
    contactMethods.value = Array.isArray(settings?.contact_methods) ? settings.contact_methods : []
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
/* redeem 统计卡 + 主面板：白底（4 张卡 / 多个面板，染色会让整页同色） */
.redeem-panel,
.redeem-stat-card {
  border-radius: 1rem;
  border: 1px solid rgb(229 231 235);
  background: rgb(255 255 255);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.redeem-stat-card {
  padding: 1rem;
}
@media (min-width: 640px) {
  .redeem-stat-card {
    padding: 1.5rem;
  }
}

.dark .redeem-panel,
.dark .redeem-stat-card {
  border-color: rgb(55 65 81 / 0.6);
  background: rgb(31 41 55 / 0.4);
  box-shadow: none;
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>

<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部工具栏：days switcher + 刷新 -->
      <div class="flex flex-wrap items-center justify-end gap-2">
        <div class="flex items-center gap-1 rounded-full bg-gray-100 p-1 dark:bg-dark-800">
          <button
            v-for="d in DAYS_OPTIONS"
            :key="d"
            type="button"
            class="rounded-full px-3 py-1 text-xs font-medium transition-colors"
            :class="days === d
              ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
              : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
            @click="days = d"
          >
            {{ d }}{{ t('payment.admin.daySuffix') }}
          </button>
        </div>
        <button @click="refreshAll" :disabled="loading || financeLoading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
          <Icon name="refresh" size="sm" :class="loading || financeLoading ? 'animate-spin' : ''" />
        </button>
      </div>

      <!-- Dashboard Content -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <section class="surface-card overflow-hidden">
          <header class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
            <div class="flex items-start gap-3">
              <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300">
                <Icon name="chart" size="sm" :stroke-width="1.75" />
              </span>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.financeTitle') }}</h3>
                <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.financeDesc') }}</p>
              </div>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <div class="flex items-center gap-1 rounded-full bg-gray-100 p-1 dark:bg-dark-800">
                <button
                  type="button"
                  class="rounded-full px-3 py-1 text-xs font-medium transition-colors"
                  :class="financePeriod === 'month'
                    ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                    : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
                  @click="financePeriod = 'month'"
                >
                  {{ t('payment.admin.financePeriodMonth') }}
                </button>
                <button
                  type="button"
                  class="rounded-full px-3 py-1 text-xs font-medium transition-colors"
                  :class="financePeriod === 'day'
                    ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                    : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
                  @click="financePeriod = 'day'"
                >
                  {{ t('payment.admin.financePeriodDay') }}
                </button>
              </div>
              <input
                v-if="financePeriod === 'month'"
                v-model="financeMonth"
                type="month"
                class="input h-9 w-36 text-sm"
                :aria-label="t('payment.admin.financeMonth')"
              />
              <input
                v-else
                v-model="financeDate"
                type="date"
                class="input h-9 w-40 text-sm"
                :aria-label="t('payment.admin.financeDate')"
              />
            </div>
          </header>

          <div v-if="financeLoading && !financeStats" class="flex items-center justify-center py-10">
            <LoadingSpinner />
          </div>
          <div v-else-if="financeStats" class="space-y-4 p-6">
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-5">
              <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.financeNetIncome') }}</p>
                <p class="mt-1 text-xl font-semibold tabular-nums text-gray-900 dark:text-white">&yen;{{ financeStats.total_amount.toFixed(2) }}</p>
              </div>
              <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.financeGrossIncome') }}</p>
                <p class="mt-1 text-xl font-semibold tabular-nums text-gray-900 dark:text-white">&yen;{{ financeStats.gross_amount.toFixed(2) }}</p>
              </div>
              <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.financeRefunds') }}</p>
                <p class="mt-1 text-xl font-semibold tabular-nums text-red-600 dark:text-red-300">&yen;{{ financeStats.refund_amount.toFixed(2) }}</p>
              </div>
              <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.orderCount') }}</p>
                <p class="mt-1 text-xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ financeStats.total_count }}</p>
              </div>
              <div class="rounded-lg border border-gray-100 p-4 dark:border-dark-700">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.financeAvgOrder') }}</p>
                <p class="mt-1 text-xl font-semibold tabular-nums text-gray-900 dark:text-white">&yen;{{ financeStats.avg_amount.toFixed(2) }}</p>
              </div>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.admin.financeLedgerNote') }}</p>
            <div class="overflow-hidden rounded-lg border border-gray-100 dark:border-dark-700">
              <div class="max-h-80 overflow-auto">
                <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
                  <thead class="sticky top-0 bg-gray-50/95 text-left text-xs font-medium uppercase tracking-wide text-gray-500 backdrop-blur dark:bg-dark-800/95 dark:text-gray-400">
                    <tr>
                      <th class="px-4 py-3">{{ t('payment.admin.colDate') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('payment.admin.netIncome') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('payment.admin.grossIncome') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('payment.admin.refundDeduction') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('payment.admin.orderCount') }}</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                    <tr v-for="bucket in financeStats.series" :key="bucket.date" class="hover:bg-gray-50/60 dark:hover:bg-dark-800/40">
                      <td class="px-4 py-3 font-mono text-gray-600 dark:text-gray-300">{{ bucket.date }}</td>
                      <td class="px-4 py-3 text-right font-medium tabular-nums text-gray-900 dark:text-white">&yen;{{ bucket.amount.toFixed(2) }}</td>
                      <td class="px-4 py-3 text-right tabular-nums text-gray-600 dark:text-gray-300">&yen;{{ bucket.gross_amount.toFixed(2) }}</td>
                      <td class="px-4 py-3 text-right tabular-nums text-red-600 dark:text-red-300">&yen;{{ bucket.refund_amount.toFixed(2) }}</td>
                      <td class="px-4 py-3 text-right tabular-nums text-gray-600 dark:text-gray-300">{{ bucket.count }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <!-- 支付方式分布 -->
          <div class="surface-card overflow-hidden">
            <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
              <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-500/15 dark:text-sky-300">
                <Icon name="creditCard" size="sm" :stroke-width="1.75" />
              </span>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.paymentDistribution') }}</h3>
            </header>
            <div class="p-6">
              <div v-if="!stats.payment_methods?.length" class="flex h-32 flex-col items-center justify-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500">
                  <Icon name="inbox" size="md" />
                </div>
                {{ t('payment.admin.noData') }}
              </div>
              <div v-else class="space-y-3">
                <div v-for="method in stats.payment_methods" :key="method.type" class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <PaymentBrandIcon :type="method.type" size="18px" />
                    <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('payment.methods.' + method.type, method.type) }}</span>
                  </div>
                  <div class="text-right tabular-nums">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">&yen;{{ method.amount.toFixed(2) }}</span>
                    <span class="ml-2 text-xs text-gray-500 dark:text-gray-400">({{ method.count }})</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 消费排行 -->
          <div class="surface-card overflow-hidden">
            <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
              <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300">
                <Icon name="chart" size="sm" :stroke-width="1.75" />
              </span>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.topUsers') }}</h3>
            </header>
            <div class="p-6">
              <div v-if="!stats.top_users?.length" class="flex h-32 flex-col items-center justify-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-dark-500">
                  <Icon name="inbox" size="md" />
                </div>
                {{ t('payment.admin.noData') }}
              </div>
              <div v-else class="space-y-1">
                <div v-for="(user, idx) in stats.top_users" :key="user.user_id" class="flex items-center justify-between rounded-lg px-3 py-2 transition-colors hover:bg-gray-50 dark:hover:bg-dark-800/60">
                  <div class="flex min-w-0 items-center gap-3">
                    <span class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full text-xs font-semibold" :class="rankClass(idx)">{{ idx + 1 }}</span>
                    <span class="truncate text-sm text-gray-700 dark:text-gray-300">{{ user.email }}</span>
                  </div>
                  <span class="text-sm font-semibold tabular-nums text-gray-900 dark:text-white">&yen;{{ user.amount.toFixed(2) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { DashboardStats, FinanceRevenueStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'
import PaymentBrandIcon from '@/components/payment/PaymentBrandIcon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const DAYS_OPTIONS = [7, 30, 90] as const
const days = ref<number>(30)
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)
type FinancePeriod = 'month' | 'day'
const financePeriod = ref<FinancePeriod>('month')
const financeMonth = ref(formatLocalDate().slice(0, 7))
const financeDate = ref(formatLocalDate())
const financeLoading = ref(false)
const financeStats = ref<FinanceRevenueStats | null>(null)

function formatLocalDate(value = new Date()): string {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function rankClass(idx: number): string {
  if (idx === 0) return 'bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
  if (idx === 1) return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
  if (idx === 2) return 'bg-orange-50 text-orange-700 dark:bg-orange-500/15 dark:text-orange-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}

async function loadDashboard() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getDashboard(days.value)
    stats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function loadFinanceStats() {
  financeLoading.value = true
  try {
    const params = financePeriod.value === 'month'
      ? { period: 'month' as const, month: financeMonth.value }
      : { period: 'day' as const, date: financeDate.value }
    const res = await adminPaymentAPI.getFinanceStats(params)
    financeStats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    financeLoading.value = false
  }
}

async function refreshAll() {
  await Promise.all([loadDashboard(), loadFinanceStats()])
}

watch(days, () => loadDashboard())
watch([financePeriod, financeMonth, financeDate], () => loadFinanceStats())
onMounted(() => refreshAll())
</script>

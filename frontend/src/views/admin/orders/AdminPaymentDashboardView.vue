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
        <button @click="loadDashboard" :disabled="loading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <!-- Dashboard Content -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
      <template v-else-if="stats">
        <OrderStatsCards :stats="stats" />
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" />
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <!-- 支付方式分布 -->
          <div class="surface-card overflow-hidden">
            <header class="flex items-start gap-3 border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
              <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-500/15 dark:text-sky-300">
                <Icon name="creditCard" size="sm" :stroke-width="1.75" />
              </span>
              <h3 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.paymentDistribution') }}</h3>
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
              <h3 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('payment.admin.topUsers') }}</h3>
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
import type { DashboardStats } from '@/types/payment'
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

watch(days, () => loadDashboard())
onMounted(() => loadDashboard())
</script>

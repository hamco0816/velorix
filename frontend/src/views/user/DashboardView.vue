<template>
  <AppLayout wide>
    <div class="space-y-8">
      <!-- 工具栏：sm 起两端对齐，小屏 stack 避免溢出 -->
      <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-sm text-gray-600 dark:text-dark-300">
            {{ t('dashboard.welcomeTitle', { name: greetingName }) }}
          </span>
          <span class="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
            {{ t('admin.dashboard.liveUpdated') }} {{ lastUpdatedLabel }}
          </span>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <DateRangePicker
            v-model:start-date="startDate"
            v-model:end-date="endDate"
            @change="onDateRangeChange"
          />
          <div class="w-28">
            <Select
              v-model="granularity"
              :options="granularityOptions"
              @change="loadCharts"
            />
          </div>
          <button @click="refreshAll" :disabled="loading || loadingCharts" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
            <Icon name="refresh" size="sm" :class="(loading || loadingCharts) ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-20"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <UserDashboardStats
          :stats="stats"
          :balance="user?.balance || 0"
          :is-simple="authStore.isSimpleMode"
          :requests-series="requestsSeries"
          :tokens-series="tokensSeries"
          :cost-series="costSeries"
          :requests-trend="requestsTrend"
          :tokens-trend="tokensTrend"
          :cost-trend="costTrend"
        />
        <UserDashboardCharts :loading="loadingCharts" :trend="trendData" :models="modelStats" />
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'

const { t } = useI18n()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const greetingName = computed(() => {
  return user.value?.username || user.value?.email?.split('@')[0] || t('common.user')
})

const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatLD(new Date()))
const granularity = ref<'day' | 'hour'>('day')

const granularityOptions = computed(() => [
  { value: 'day', label: t('dashboard.day') },
  { value: 'hour', label: t('dashboard.hour') }
])

// ============ KPI sparkline 数据派生 ============
const requestsSeries = computed(() => trendData.value.map((d) => d.requests))
const tokensSeries = computed(() => trendData.value.map((d) => d.total_tokens))
const costSeries = computed(() => trendData.value.map((d) => d.actual_cost))

const calcTrend = (series: number[]): number | null => {
  if (series.length < 4) return null
  const mid = Math.floor(series.length / 2)
  const a = series.slice(0, mid).reduce((s, v) => s + v, 0) / mid
  const b = series.slice(mid).reduce((s, v) => s + v, 0) / (series.length - mid)
  if (a === 0) return b > 0 ? 100 : null
  return ((b - a) / a) * 100
}

const requestsTrend = computed(() => calcTrend(requestsSeries.value))
const tokensTrend = computed(() => calcTrend(tokensSeries.value))
const costTrend = computed(() => calcTrend(costSeries.value))

const lastUpdated = ref<Date | null>(null)
const nowTick = ref<number>(Date.now())
let nowTimer: ReturnType<typeof setInterval> | null = null

const lastUpdatedLabel = computed(() => {
  if (!lastUpdated.value) return t('common.time.never')
  const diffSec = Math.max(0, Math.round((nowTick.value - lastUpdated.value.getTime()) / 1000))
  if (diffSec < 5) return t('common.time.justNow')
  if (diffSec < 60) return t('admin.dashboard.secondsAgo', { n: diffSec })
  const diffMin = Math.floor(diffSec / 60)
  if (diffMin < 60) return t('common.time.minutesAgo', { n: diffMin })
  const diffHour = Math.floor(diffMin / 60)
  return t('common.time.hoursAgo', { n: diffHour })
})

const onDateRangeChange = (range: { startDate: string; endDate: string; preset: string | null }) => {
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))
  granularity.value = daysDiff <= 1 ? 'hour' : 'day'
  loadCharts()
  loadRecent()
}

const loadStats = async () => {
  loading.value = true
  try {
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
    lastUpdated.value = new Date()
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  } finally {
    loading.value = false
  }
}

const loadCharts = async () => {
  loadingCharts.value = true
  try {
    const res = await Promise.all([
      usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }),
      usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })
    ])
    trendData.value = res[0].trend || []
    modelStats.value = res[1].models || []
    lastUpdated.value = new Date()
  } catch (error) {
    console.error('Failed to load charts:', error)
  } finally {
    loadingCharts.value = false
  }
}

const loadRecent = async () => {
  loadingUsage.value = true
  try {
    const res = await usageAPI.getByDateRange(startDate.value, endDate.value)
    recentUsage.value = res.items.slice(0, 5)
  } catch (error) {
    console.error('Failed to load recent usage:', error)
  } finally {
    loadingUsage.value = false
  }
}

const refreshAll = () => {
  loadStats()
  loadCharts()
  loadRecent()
}

onMounted(() => {
  refreshAll()
  nowTimer = setInterval(() => { nowTick.value = Date.now() }, 5000)
})

onBeforeUnmount(() => {
  if (nowTimer) { clearInterval(nowTimer); nowTimer = null }
})
</script>

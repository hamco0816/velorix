<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="risk-info-banner">
        <span class="risk-info-icon">
          <Icon name="infoCircle" size="sm" />
        </span>
        <span>
          风控拦截会在请求发送到上游前执行。这里记录命中规则、是否经过 AI 审核以及管理员复核状态，便于排查风险和清空用户管控记录。
        </span>
      </div>

      <section class="card p-5">
        <div class="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(190px,1fr)_minmax(190px,1fr)_minmax(190px,1fr)_minmax(320px,1.25fr)_auto]">
          <label class="space-y-2">
            <span class="filter-label">状态</span>
            <Select
              v-model="selectedStatus"
              :options="statusFilterOptions"
              class="risk-select"
            />
          </label>

          <label class="space-y-2">
            <span class="filter-label">时间</span>
            <Select
              v-model="selectedTimeRange"
              :options="timeRangeOptions"
              class="risk-select"
            />
          </label>

          <label class="space-y-2">
            <span class="filter-label">用户 ID 可选</span>
            <input
              v-model.trim="userIdInput"
              type="number"
              min="1"
              class="risk-filter-control"
              placeholder="请输入用户 ID"
            />
          </label>

          <label class="space-y-2">
            <span class="filter-label">搜索用户/邮箱、API Key、请求</span>
            <span class="relative block">
              <input
                v-model.trim="filters.q"
                class="risk-filter-control pr-10"
                placeholder="请输入关键词"
                @keyup.enter="applyFilters"
              />
              <Icon name="search" size="sm" class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400" />
            </span>
          </label>

          <div class="flex items-end">
            <button class="risk-icon-button h-10 w-full justify-center xl:w-auto" :disabled="loading" @click="loadEvents">
              <Icon name="refresh" size="sm" />
              <span>刷新</span>
            </button>
          </div>
        </div>

        <div class="mt-5 flex flex-wrap items-center gap-3">
          <button class="risk-action-button risk-action-primary min-w-[88px]" :disabled="loading" @click="applyFilters">查询</button>
          <button class="risk-action-button risk-action-secondary min-w-[88px]" :disabled="loading" @click="resetFilters">重置</button>
        </div>
      </section>

      <section class="card overflow-hidden p-0">
        <div class="grid grid-cols-1 divide-y divide-gray-100 dark:divide-dark-700 md:grid-cols-5 md:divide-x md:divide-y-0">
          <div v-for="stat in riskStats" :key="stat.key" class="risk-stat">
            <span class="risk-stat-icon" :class="stat.iconClass">
              <Icon :name="stat.icon" size="md" />
            </span>
            <span class="min-w-0">
              <span class="block text-sm text-gray-500 dark:text-gray-400">{{ stat.label }}</span>
              <span class="mt-1 block text-2xl font-semibold leading-none text-gray-900 dark:text-gray-100">{{ stat.value }}</span>
              <span class="mt-2 block text-xs text-gray-400 dark:text-gray-500">{{ stat.hint }}</span>
            </span>
          </div>
        </div>
      </section>

      <section class="card overflow-hidden p-0">
        <div class="overflow-x-auto px-4 pt-4">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="table-th">时间</th>
                <th class="table-th">用户 / Key</th>
                <th class="table-th">路径</th>
                <th class="table-th">命中规则</th>
                <th class="table-th">AI 审核</th>
                <th class="table-th">状态</th>
                <th class="table-th">请求预览</th>
                <th class="table-th text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-if="loading">
                <td colspan="8" class="px-4 py-16 text-center text-sm text-gray-500 dark:text-gray-400">加载中...</td>
              </tr>
              <tr v-else-if="events.length === 0">
                <td colspan="8">
                  <div class="risk-empty-state">
                    <div class="risk-empty-icon">
                      <Icon name="inbox" size="xl" />
                    </div>
                    <div class="text-base font-semibold text-gray-800 dark:text-gray-100">暂无风控日志</div>
                    <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">当前没有符合条件的日志记录</div>
                  </div>
                </td>
              </tr>
              <template v-else>
                <tr v-for="item in events" :key="item.id" class="transition hover:bg-gray-50 dark:hover:bg-dark-800/70">
                  <td class="table-td whitespace-nowrap">
                    <div class="text-sm text-gray-900 dark:text-gray-100">{{ formatDate(item.created_at) }}</div>
                    <div class="text-xs text-gray-500">#{{ item.id }}</div>
                  </td>
                  <td class="table-td min-w-[190px]">
                    <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {{ item.user_email || formatID('用户', item.user_id) }}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">
                      {{ item.api_key_name || formatID('Key', item.api_key_id) }}
                    </div>
                  </td>
                  <td class="table-td min-w-[190px]">
                    <div class="text-sm text-gray-900 dark:text-gray-100">{{ item.method }} {{ item.path }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.client_ip || '-' }}</div>
                  </td>
                  <td class="table-td min-w-[180px]">
                    <div class="flex flex-wrap items-center gap-1.5">
                      <span class="badge" :class="item.rule_source === 'custom' ? 'badge-blue' : 'badge-gray'">
                        {{ sourceText(item.rule_source) }}
                      </span>
                      <span class="badge badge-amber">{{ item.action || 'blocked' }}</span>
                    </div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ item.rule_word || '-' }}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.rule_path || '-' }}</div>
                  </td>
                  <td class="table-td whitespace-nowrap">
                    <span class="badge" :class="item.ai_reviewed ? 'badge-green' : 'badge-gray'">
                      {{ item.ai_reviewed ? '已经过 AI' : '未经过 AI' }}
                    </span>
                    <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ item.ai_review_result || 'not_used' }}</div>
                  </td>
                  <td class="table-td whitespace-nowrap">
                    <span class="badge" :class="statusClass(item.status)">{{ statusText(item.status) }}</span>
                    <div v-if="item.reviewed_at" class="mt-1 text-xs text-gray-500">{{ formatDate(item.reviewed_at) }}</div>
                  </td>
                  <td class="table-td max-w-[380px]">
                    <div class="line-clamp-3 text-sm text-gray-700 dark:text-gray-300">{{ item.prompt_preview || '-' }}</div>
                    <div v-if="item.request_id || item.client_request_id" class="mt-1 text-xs text-gray-500">
                      {{ item.request_id || item.client_request_id }}
                    </div>
                  </td>
                  <td class="table-td whitespace-nowrap text-right">
                    <button
                      v-if="item.status !== 'reviewed'"
                      class="btn btn-secondary btn-sm mr-2"
                      :disabled="loading"
                      @click="markReviewed(item)"
                    >
                      标记复核
                    </button>
                    <button
                      v-if="item.user_id && item.status !== 'cleared'"
                      class="btn btn-danger btn-sm"
                      :disabled="loading"
                      @click="openClearUser(item)"
                    >
                      清空用户
                    </button>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div v-if="pagination.total > 0" class="border-t border-gray-100 px-4 py-3 dark:border-dark-700">
          <Pagination
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </section>
    </div>
  </AppLayout>

  <ConfirmDialog
    :show="clearDialogVisible"
    title="清空用户警告"
    :message="clearDialogMessage"
    confirm-text="确认清空"
    cancel-text="取消"
    danger
    @confirm="clearSelectedUser"
    @cancel="clearDialogVisible = false"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import {
  clearSafetyRiskEventsForUser,
  listSafetyRiskEvents,
  reviewSafetyRiskEvent,
  type SafetyRiskEvent,
  type SafetyRiskQueryParams,
} from '@/api/admin/safetyRisk'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

type StatIcon = 'document' | 'shield' | 'brain' | 'checkCircle' | 'xCircle'

interface RiskStat {
  key: string
  label: string
  value: number
  hint: string
  icon: StatIcon
  iconClass: string
}

const appStore = useAppStore()
const loading = ref(false)
const events = ref<SafetyRiskEvent[]>([])
const selectedClearEvent = ref<SafetyRiskEvent | null>(null)
const clearDialogVisible = ref(false)
const userIdInput = ref('')

const filters = reactive<SafetyRiskQueryParams>({
  status: 'pending',
  time_range: '24h',
  q: '',
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})

const statusFilterOptions: SelectOption[] = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待复核' },
  { value: 'reviewed', label: '已复核' },
  { value: 'cleared', label: '已清空' },
]

const timeRangeOptions: SelectOption[] = [
  { value: '5m', label: '最近 5 分钟' },
  { value: '30m', label: '最近 30 分钟' },
  { value: '1h', label: '最近 1 小时' },
  { value: '6h', label: '最近 6 小时' },
  { value: '24h', label: '最近 24 小时' },
  { value: '7d', label: '最近 7 天' },
  { value: '30d', label: '最近 30 天' },
]

const selectedStatus = computed<string | number | boolean | null>({
  get: () => filters.status ?? '',
  set: (value) => {
    filters.status = typeof value === 'string' ? value : ''
  },
})

const selectedTimeRange = computed<string | number | boolean | null>({
  get: () => filters.time_range ?? '24h',
  set: (value) => {
    const nextValue = typeof value === 'string' ? value : '24h'
    filters.time_range = nextValue as SafetyRiskQueryParams['time_range']
  },
})

const riskStats = computed<RiskStat[]>(() => [
  {
    key: 'total',
    label: '总拦截数',
    value: pagination.total,
    hint: timeRangeText(filters.time_range),
    icon: 'document',
    iconClass: 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-300',
  },
  {
    key: 'pending',
    label: '待复核',
    value: countByStatus('pending'),
    hint: '当前页',
    icon: 'shield',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-900/30 dark:text-violet-300',
  },
  {
    key: 'ai',
    label: 'AI 审核中',
    value: events.value.filter(isAIReviewing).length,
    hint: '当前页',
    icon: 'brain',
    iconClass: 'bg-orange-50 text-orange-600 dark:bg-orange-900/30 dark:text-orange-300',
  },
  {
    key: 'approved',
    label: '已通过',
    value: events.value.filter(isApproved).length,
    hint: '当前页',
    icon: 'checkCircle',
    iconClass: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-300',
  },
  {
    key: 'rejected',
    label: '已拒绝',
    value: events.value.filter(isRejected).length,
    hint: '当前页',
    icon: 'xCircle',
    iconClass: 'bg-rose-50 text-rose-600 dark:bg-rose-900/30 dark:text-rose-300',
  },
])

const clearDialogMessage = computed(() => {
  const item = selectedClearEvent.value
  if (!item) return ''
  const user = item.user_email || formatID('用户', item.user_id)
  return `将把 ${user} 的未清空风控警告全部标记为已清空。该操作只清空警告记录，不会删除日志。`
})

async function loadEvents() {
  loading.value = true
  try {
    const params: SafetyRiskQueryParams = {
      ...filters,
      page: pagination.page,
      page_size: pagination.page_size,
    }
    const userId = Number(userIdInput.value)
    if (Number.isFinite(userId) && userId > 0) {
      params.user_id = userId
    }
    if (!params.q) delete params.q
    const data = await listSafetyRiskEvents(params)
    events.value = data.items
    pagination.total = data.total
    pagination.page = data.page
    pagination.page_size = data.page_size
  } catch (error: any) {
    appStore.showError(error?.message || '加载风控日志失败')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  loadEvents()
}

function resetFilters() {
  filters.status = 'pending'
  filters.time_range = '24h'
  filters.q = ''
  userIdInput.value = ''
  applyFilters()
}

async function markReviewed(item: SafetyRiskEvent) {
  try {
    await reviewSafetyRiskEvent(item.id, {
      status: 'reviewed',
      review_note: 'manual reviewed',
    })
    appStore.showSuccess('已标记为复核')
    await loadEvents()
  } catch (error: any) {
    appStore.showError(error?.message || '复核失败')
  }
}

function openClearUser(item: SafetyRiskEvent) {
  selectedClearEvent.value = item
  clearDialogVisible.value = true
}

async function clearSelectedUser() {
  const item = selectedClearEvent.value
  if (!item?.user_id) {
    clearDialogVisible.value = false
    return
  }
  try {
    const result = await clearSafetyRiskEventsForUser(item.user_id, 'manual clear after review')
    appStore.showSuccess(`已清空 ${result.cleared} 条警告记录`)
    clearDialogVisible.value = false
    selectedClearEvent.value = null
    await loadEvents()
  } catch (error: any) {
    appStore.showError(error?.message || '清空用户警告失败')
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadEvents()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page = 1
  pagination.page_size = pageSize
  loadEvents()
}

function countByStatus(status: string): number {
  return events.value.filter((item) => item.status === status).length
}

function normalizeReviewResult(value?: string | null): string {
  return String(value || '').trim().toLowerCase()
}

function isAIReviewing(item: SafetyRiskEvent): boolean {
  const result = normalizeReviewResult(item.ai_review_result)
  return item.ai_reviewed && item.status === 'pending' && (!result || result === 'pending' || result === 'reviewing')
}

function isApproved(item: SafetyRiskEvent): boolean {
  const status = String(item.status || '').toLowerCase()
  const result = normalizeReviewResult(item.ai_review_result)
  return ['approved', 'passed', 'allowed'].includes(status) || ['approve', 'approved', 'pass', 'passed', 'allow', 'allowed'].includes(result)
}

function isRejected(item: SafetyRiskEvent): boolean {
  const status = String(item.status || '').toLowerCase()
  const result = normalizeReviewResult(item.ai_review_result)
  return ['rejected', 'blocked', 'denied'].includes(status) || ['reject', 'rejected', 'block', 'blocked', 'deny', 'denied'].includes(result)
}

function timeRangeText(range?: string): string {
  switch (range) {
    case '5m':
      return '最近 5 分钟'
    case '30m':
      return '最近 30 分钟'
    case '1h':
      return '最近 1 小时'
    case '6h':
      return '最近 6 小时'
    case '7d':
      return '最近 7 天'
    case '30d':
      return '最近 30 天'
    case '24h':
    default:
      return '最近 24 小时'
  }
}

function formatDate(value?: string | null): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatID(label: string, value?: number | null): string {
  return value ? `${label} #${value}` : '-'
}

function sourceText(source: string): string {
  if (source === 'custom') return '自定义'
  if (source === 'builtin') return '内置'
  if (source === 'ai') return 'AI'
  return source || '本地'
}

function statusText(status: string): string {
  switch (status) {
    case 'pending':
      return '待复核'
    case 'reviewed':
      return '已复核'
    case 'cleared':
      return '已清空'
    case 'rejected':
      return '已拒绝'
    case 'approved':
    case 'passed':
      return '已通过'
    default:
      return status || '-'
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'pending':
      return 'badge-amber'
    case 'reviewed':
      return 'badge-blue'
    case 'cleared':
    case 'approved':
    case 'passed':
      return 'badge-green'
    case 'rejected':
      return 'badge-red'
    default:
      return 'badge-gray'
  }
}

onMounted(loadEvents)
</script>

<style scoped>
.risk-info-banner {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  border: 1px solid rgb(147 197 253);
  border-radius: 0.5rem;
  background: rgb(239 246 255 / 0.72);
  padding: 0.875rem 1rem;
  color: rgb(37 99 235);
  font-size: 0.875rem;
  line-height: 1.6;
}

.dark .risk-info-banner {
  border-color: rgb(30 64 175 / 0.75);
  background: rgb(30 58 138 / 0.18);
  color: rgb(147 197 253);
}

.risk-info-icon {
  display: inline-flex;
  flex: 0 0 auto;
}

.filter-label {
  display: block;
  color: rgb(55 65 81);
  font-size: 0.8125rem;
  font-weight: 500;
}

.dark .filter-label {
  color: rgb(209 213 219);
}

.risk-select :deep(.select-trigger) {
  height: 2.75rem;
  border-color: rgb(209 213 219);
  border-radius: 0.5rem;
  padding: 0.625rem 0.875rem;
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.risk-select :deep(.select-trigger-open) {
  border-color: rgb(59 130 246);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.14);
}

.dark .risk-select :deep(.select-trigger) {
  border-color: rgb(75 85 99);
  box-shadow: none;
}

.risk-filter-control {
  display: block;
  width: 100%;
  height: 2.75rem;
  border: 1px solid rgb(209 213 219);
  border-radius: 0.5rem;
  background-color: rgb(255 255 255);
  padding: 0.625rem 0.875rem;
  color: rgb(17 24 39);
  font-size: 0.875rem;
  line-height: 1.25rem;
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
  transition: border-color 0.15s ease, box-shadow 0.15s ease, background-color 0.15s ease;
}

.risk-filter-control::placeholder {
  color: rgb(156 163 175);
}

.risk-filter-control:focus {
  border-color: rgb(59 130 246);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.14);
  outline: none;
}

.dark .risk-filter-control {
  border-color: rgb(75 85 99);
  background-color: rgb(17 24 39);
  color: rgb(243 244 246);
  box-shadow: none;
}

.dark .risk-filter-control::placeholder {
  color: rgb(107 114 128);
}

.risk-action-button,
.risk-icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.25rem;
  transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
}

.risk-action-button {
  height: 2.5rem;
  padding: 0 1.25rem;
}

.risk-action-primary {
  border: 1px solid rgb(37 99 235);
  background: linear-gradient(180deg, rgb(37 99 235), rgb(29 78 216));
  color: white;
  box-shadow: 0 8px 18px rgb(37 99 235 / 0.18);
}

.risk-action-primary:hover:not(:disabled) {
  border-color: rgb(29 78 216);
  background: linear-gradient(180deg, rgb(29 78 216), rgb(30 64 175));
}

.risk-action-secondary,
.risk-icon-button {
  border: 1px solid rgb(209 213 219);
  background: rgb(255 255 255);
  color: rgb(55 65 81);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.risk-icon-button {
  padding: 0 1rem;
}

.risk-action-secondary:hover:not(:disabled),
.risk-icon-button:hover:not(:disabled) {
  border-color: rgb(156 163 175);
  background: rgb(249 250 251);
  color: rgb(17 24 39);
}

.risk-action-button:disabled,
.risk-icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.dark .risk-action-secondary,
.dark .risk-icon-button {
  border-color: rgb(75 85 99);
  background: rgb(17 24 39);
  color: rgb(229 231 235);
  box-shadow: none;
}

.dark .risk-action-secondary:hover:not(:disabled),
.dark .risk-icon-button:hover:not(:disabled) {
  border-color: rgb(107 114 128);
  background: rgb(31 41 55);
  color: white;
}

.risk-stat {
  display: flex;
  min-height: 6.75rem;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem 1.25rem;
}

.risk-stat-icon {
  display: inline-flex;
  height: 3rem;
  width: 3rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
}

.risk-empty-state {
  display: flex;
  min-height: 18rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  text-align: center;
}

.risk-empty-icon {
  display: inline-flex;
  height: 5rem;
  width: 5rem;
  align-items: center;
  justify-content: center;
  border-radius: 1.5rem;
  background: linear-gradient(180deg, rgb(219 234 254), rgb(238 242 255));
  color: rgb(79 70 229);
  box-shadow: 0 18px 40px rgb(99 102 241 / 0.18);
  margin-bottom: 1.25rem;
}

.dark .risk-empty-icon {
  background: linear-gradient(180deg, rgb(30 64 175 / 0.4), rgb(49 46 129 / 0.45));
  color: rgb(191 219 254);
}

.table-th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-size: 0.8125rem;
  font-weight: 700;
  letter-spacing: 0;
  color: rgb(17 24 39);
  white-space: nowrap;
}

.dark .table-th {
  color: rgb(229 231 235);
}

.table-td {
  padding: 0.875rem 1rem;
  vertical-align: top;
}

.badge {
  display: inline-flex;
  align-items: center;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.25rem;
}

.badge-gray {
  background: rgb(243 244 246);
  color: rgb(55 65 81);
}

.badge-blue {
  background: rgb(219 234 254);
  color: rgb(30 64 175);
}

.badge-amber {
  background: rgb(254 243 199);
  color: rgb(146 64 14);
}

.badge-green {
  background: rgb(220 252 231);
  color: rgb(22 101 52);
}

.badge-red {
  background: rgb(254 226 226);
  color: rgb(153 27 27);
}

.dark .badge-gray {
  background: rgb(55 65 81);
  color: rgb(229 231 235);
}

.dark .badge-blue {
  background: rgb(30 58 138 / 0.35);
  color: rgb(147 197 253);
}

.dark .badge-amber {
  background: rgb(120 53 15 / 0.35);
  color: rgb(252 211 77);
}

.dark .badge-green {
  background: rgb(20 83 45 / 0.35);
  color: rgb(134 239 172);
}

.dark .badge-red {
  background: rgb(127 29 29 / 0.35);
  color: rgb(252 165 165);
}
</style>

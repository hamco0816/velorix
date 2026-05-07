<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-800/70 dark:bg-amber-950/30 dark:text-amber-200">
        风控拦截会在请求转发到上游前执行。这里记录命中的规则、是否经过 AI 审核以及管理员复核状态，便于排查误伤和清空用户警告记录。
      </div>

      <div class="card p-4">
        <div class="grid grid-cols-1 gap-3 md:grid-cols-5">
          <label class="space-y-1">
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">状态</span>
            <select v-model="filters.status" class="form-input">
              <option value="">全部</option>
              <option value="pending">待复核</option>
              <option value="reviewed">已复核</option>
              <option value="cleared">已清空</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">时间</span>
            <select v-model="filters.time_range" class="form-input">
              <option value="1h">最近 1 小时</option>
              <option value="6h">最近 6 小时</option>
              <option value="24h">最近 24 小时</option>
              <option value="7d">最近 7 天</option>
              <option value="30d">最近 30 天</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">用户 ID</span>
            <input v-model.trim="userIdInput" type="number" min="1" class="form-input" placeholder="可选" />
          </label>
          <label class="space-y-1 md:col-span-2">
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">搜索</span>
            <input v-model.trim="filters.q" class="form-input" placeholder="用户邮箱、API Key、请求 ID、规则、预览内容" @keyup.enter="applyFilters" />
          </label>
        </div>
        <div class="mt-3 flex flex-wrap items-center gap-2">
          <button class="btn btn-primary" :disabled="loading" @click="applyFilters">查询</button>
          <button class="btn btn-secondary" :disabled="loading" @click="resetFilters">重置</button>
          <button class="btn btn-secondary ml-auto" :disabled="loading" @click="loadEvents">刷新</button>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
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
                <td colspan="8" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">加载中...</td>
              </tr>
              <tr v-else-if="events.length === 0">
                <td colspan="8" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">暂无风控日志</td>
              </tr>
              <tr v-for="item in events" v-else :key="item.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/70">
                <td class="table-td whitespace-nowrap">
                  <div class="text-sm text-gray-900 dark:text-gray-100">{{ formatDate(item.created_at) }}</div>
                  <div class="text-xs text-gray-500">#{{ item.id }}</div>
                </td>
                <td class="table-td min-w-[180px]">
                  <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {{ item.user_email || formatID('用户', item.user_id) }}
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">
                    {{ item.api_key_name || formatID('Key', item.api_key_id) }}
                  </div>
                </td>
                <td class="table-td min-w-[180px]">
                  <div class="text-sm text-gray-900 dark:text-gray-100">{{ item.method }} {{ item.path }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.client_ip || '-' }}</div>
                </td>
                <td class="table-td min-w-[170px]">
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
                <td class="table-td max-w-[360px]">
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
      </div>
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
import { useAppStore } from '@/stores/app'
import {
  clearSafetyRiskEventsForUser,
  listSafetyRiskEvents,
  reviewSafetyRiskEvent,
  type SafetyRiskEvent,
  type SafetyRiskQueryParams,
} from '@/api/admin/safetyRisk'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

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
      return 'badge-green'
    default:
      return 'badge-gray'
  }
}

onMounted(loadEvents)
</script>

<style scoped>
.table-th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0;
  color: rgb(75 85 99);
  white-space: nowrap;
}

.dark .table-th {
  color: rgb(209 213 219);
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
</style>

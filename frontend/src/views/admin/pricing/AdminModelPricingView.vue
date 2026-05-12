<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 3 张元信息 metric-card：数据源 / 上次更新 / 模型数 -->
      <div class="grid gap-4 md:grid-cols-3">
        <div class="surface-card p-5">
          <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('admin.pricing.source') }}</p>
          <p v-if="metadata" class="mt-2 break-all font-mono text-xs text-gray-700 dark:text-gray-300">
            {{ metadata.remote_url || t('admin.pricing.sourceUnknown') }}
          </p>
          <p v-else class="mt-2 text-gray-400">—</p>
        </div>
        <div class="surface-card p-5">
          <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('admin.pricing.lastUpdated') }}</p>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-300">
            {{ metadata?.last_updated ? formatRelativeWithDateTime(metadata.last_updated) : '—' }}
          </p>
        </div>
        <div class="surface-card flex items-start gap-3 p-5">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300">
            <Icon name="cube" size="sm" :stroke-width="1.75" />
          </span>
          <div>
            <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('admin.pricing.modelCount') }}</p>
            <p class="mt-1 text-[26px] font-semibold leading-tight tabular-nums text-gray-900 dark:text-white">
              {{ metadata?.model_count ?? '—' }}
            </p>
          </div>
        </div>
      </div>

      <!-- 倍率视角选择器 + 操作 -->
      <div class="surface-card p-5">
        <div class="flex flex-wrap items-end gap-4">
          <div class="min-w-[12rem]">
            <label class="input-label">{{ t('admin.pricing.viewMode') }}</label>
            <Select v-model="viewMode" :options="viewModeOptions" class="mt-1 w-full" />
          </div>
          <div v-if="viewMode === 'group'" class="min-w-[16rem] flex-1">
            <label class="input-label">{{ t('admin.pricing.selectGroup') }}</label>
            <Select v-model="selectedGroupId" :options="groupOptions" class="mt-1 w-full" />
          </div>
          <div v-if="viewMode === 'account'" class="min-w-[16rem] flex-1">
            <label class="input-label">{{ t('admin.pricing.selectAccount') }}</label>
            <Select v-model="selectedAccountId" :options="accountOptions" class="mt-1 w-full" />
          </div>
          <div class="min-w-[14rem] flex-1">
            <label class="input-label">{{ t('admin.pricing.searchModel') }}</label>
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="t('admin.pricing.searchPlaceholder')"
              class="input mt-1 w-full"
            />
          </div>
          <div class="ml-auto flex items-end gap-3">
            <div class="text-sm">
              <p class="text-[11px] font-medium text-gray-500 dark:text-dark-400">{{ t('admin.pricing.effectiveMultiplier') }}</p>
              <p class="mt-0.5 text-xl font-semibold tabular-nums" :class="multiplierColor">
                {{ effectiveMultiplier.toFixed(2) }}x
              </p>
            </div>
            <button class="btn btn-secondary" :disabled="loading" @click="reload">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
        <p class="mt-3 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.pricing.viewModeHint') }}
        </p>
      </div>

      <!-- Provider tabs：black active（Linear 风），与全站时间窗口按钮一致 -->
      <div class="flex flex-wrap items-center gap-2">
        <button
          v-for="tab in providerTabs"
          :key="tab.key"
          type="button"
          class="inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium transition-colors"
          :class="tab.key === activeProvider
            ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
          @click="activeProvider = tab.key"
        >
          {{ tab.label }}
          <span
            class="rounded-full px-1.5 py-0.5 text-[10px] font-semibold tabular-nums"
            :class="tab.key === activeProvider
              ? 'bg-white/20 text-white dark:bg-gray-900/20 dark:text-gray-900'
              : 'bg-gray-200 text-gray-600 dark:bg-dark-600 dark:text-gray-300'"
          >
            {{ tab.count }}
          </span>
        </button>
      </div>

      <!-- 表格 -->
      <div class="surface-card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200/60 dark:divide-dark-700/60">
            <thead class="sticky top-0 z-10 bg-gray-50/60 backdrop-blur-sm dark:bg-dark-800/60">
              <tr>
                <th
                  v-for="col in tableColumns"
                  :key="col.key"
                  :class="[
                    'px-4 py-3 text-[13px] font-medium text-gray-500 dark:text-dark-400 border-b border-gray-200/60 dark:border-dark-700/60',
                    col.align === 'right' ? 'text-right' : col.align === 'center' ? 'text-center' : 'text-left',
                    col.sortable ? 'cursor-pointer select-none hover:bg-gray-100/80 dark:hover:bg-dark-700/40' : '',
                  ]"
                  @click="col.sortable ? toggleSort(col.key as SortKey) : undefined"
                >
                  <span class="inline-flex items-center gap-1">
                    {{ col.label }}
                    <template v-if="col.sortable && sortKey === col.key">
                      <Icon :name="sortDir === 'asc' ? 'chevronUp' : 'chevronDown'" size="sm" />
                    </template>
                    <template v-else-if="col.sortable">
                      <Icon name="arrowsUpDown" size="xs" class="text-gray-300 dark:text-gray-600" />
                    </template>
                  </span>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <!-- align-middle 保证每行内容垂直居中（chips/数字/badge 混排高度不一时不再上下错位） -->
              <tr v-for="row in sortedRows" :key="row.model" class="align-middle text-sm hover:bg-gray-50 dark:hover:bg-dark-800">
                <td class="px-4 py-2.5 text-left font-mono text-xs text-gray-900 dark:text-white">{{ row.model }}</td>
                <td class="px-4 py-2.5 text-left">
                  <span :class="['rounded-full px-2 py-0.5 text-[11px] font-semibold', providerPillClass(row.provider)]">
                    {{ row.provider || '—' }}
                  </span>
                </td>
                <td class="px-4 py-2.5 text-right font-mono tabular-nums text-gray-900 dark:text-gray-100">{{ formatPriceMTok(row.input_cost_per_token) }}</td>
                <td class="px-4 py-2.5 text-right font-mono tabular-nums text-gray-900 dark:text-gray-100">{{ formatPriceMTok(row.output_cost_per_token) }}</td>
                <td class="px-4 py-2.5 text-right font-mono tabular-nums text-gray-700 dark:text-gray-300">{{ formatPriceMTok(row.cache_read_input_token_cost) }}</td>
                <td class="px-4 py-2.5 text-right font-mono tabular-nums text-gray-700 dark:text-gray-300">{{ formatPriceMTok(row.cache_creation_input_token_cost) }}</td>
                <td class="px-4 py-2.5 text-center text-xs">
                  <span v-if="row.supports_prompt_caching" class="mr-1 inline-block rounded bg-emerald-100 px-1.5 py-0.5 text-[10px] font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300" :title="t('admin.pricing.flagCachingTip')">cache</span>
                  <span v-if="row.supports_service_tier" class="mr-1 inline-block rounded bg-violet-100 px-1.5 py-0.5 text-[10px] font-medium text-violet-700 dark:bg-violet-900/30 dark:text-violet-300" :title="t('admin.pricing.flagPriorityTip')">priority</span>
                  <span v-if="row.mode" class="inline-block rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ row.mode }}</span>
                </td>
              </tr>
              <tr v-if="!loading && sortedRows.length === 0">
                <td colspan="7" class="py-12 text-center text-sm text-gray-400">
                  {{ t('admin.pricing.empty') }}
                </td>
              </tr>
              <tr v-if="loading">
                <td colspan="7" class="py-12 text-center text-sm text-gray-400">
                  <Icon name="refresh" size="md" class="mr-2 inline-block animate-spin" />
                  {{ t('common.loading') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="sortedRows.length > 0" class="border-t border-gray-200/60 bg-gray-50/30 px-4 py-2.5 text-xs text-gray-500 dark:border-dark-700/60 dark:bg-dark-800/30 dark:text-gray-400">
          {{ t('admin.pricing.tableFooter', { shown: sortedRows.length, total: pricingRows.length }) }}
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import adminAPI from '@/api/admin'
import type { PricingModelEntry, PricingMetadata } from '@/api/admin/pricing'
import type { AdminGroup, Account } from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatRelativeWithDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const pricingRows = ref<PricingModelEntry[]>([])
const metadata = ref<PricingMetadata | null>(null)

type ViewMode = 'default' | 'group' | 'account'
const viewMode = ref<ViewMode>('default')

const groups = ref<AdminGroup[]>([])
const accounts = ref<Account[]>([])
const selectedGroupId = ref<number | null>(null)
const selectedAccountId = ref<number | null>(null)
const searchQuery = ref('')

// Provider tabs：187 个模型分桶，按 LiteLLM provider 字段聚类
type ProviderTabKey = 'all' | 'anthropic' | 'openai' | 'google' | 'others'
const activeProvider = ref<ProviderTabKey>('all')

// 表格排序
type SortKey = 'model' | 'provider' | 'input' | 'output' | 'cacheRead' | 'cacheWrite'
const sortKey = ref<SortKey>('model')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
}

interface TableColumn {
  key: 'model' | 'provider' | 'input' | 'output' | 'cacheRead' | 'cacheWrite' | 'flags'
  label: string
  align: 'left' | 'right' | 'center'
  sortable: boolean
}

const tableColumns = computed<TableColumn[]>(() => [
  { key: 'model', label: t('admin.pricing.col.model'), align: 'left', sortable: true },
  { key: 'provider', label: t('admin.pricing.col.provider'), align: 'left', sortable: true },
  { key: 'input', label: t('admin.pricing.col.input'), align: 'right', sortable: true },
  { key: 'output', label: t('admin.pricing.col.output'), align: 'right', sortable: true },
  { key: 'cacheRead', label: t('admin.pricing.col.cacheRead'), align: 'right', sortable: true },
  { key: 'cacheWrite', label: t('admin.pricing.col.cacheWrite'), align: 'right', sortable: true },
  { key: 'flags', label: t('admin.pricing.col.flags'), align: 'center', sortable: false },
])

// 把 LiteLLM provider 字段聚类到 4 个 tab：anthropic / openai（含 codex）/ google（含 gemini/vertex）/ others
function providerBucket(provider: string): Exclude<ProviderTabKey, 'all'> {
  const p = (provider || '').toLowerCase()
  if (p.includes('anthropic')) return 'anthropic'
  if (p.includes('openai') || p.includes('codex')) return 'openai'
  if (p.includes('google') || p.includes('gemini') || p.includes('vertex')) return 'google'
  return 'others'
}

const providerTabs = computed(() => {
  const counts: Record<Exclude<ProviderTabKey, 'all'>, number> = {
    anthropic: 0, openai: 0, google: 0, others: 0,
  }
  for (const r of pricingRows.value) counts[providerBucket(r.provider)]++
  return [
    { key: 'all' as ProviderTabKey, label: t('admin.pricing.tabAll'), count: pricingRows.value.length },
    { key: 'anthropic' as ProviderTabKey, label: 'Anthropic', count: counts.anthropic },
    { key: 'openai' as ProviderTabKey, label: 'OpenAI', count: counts.openai },
    { key: 'google' as ProviderTabKey, label: 'Google', count: counts.google },
    { key: 'others' as ProviderTabKey, label: t('admin.pricing.tabOthers'), count: counts.others },
  ]
})

const viewModeOptions = computed(() => [
  { value: 'default', label: t('admin.pricing.viewMode_default') },
  { value: 'group', label: t('admin.pricing.viewMode_group') },
  { value: 'account', label: t('admin.pricing.viewMode_account') },
])

const groupOptions = computed(() =>
  groups.value.map((g) => ({
    value: g.id,
    label: `${g.name} (${g.rate_multiplier.toFixed(2)}x)`,
  })),
)

// 账号列表可能很大，最多展示 200 条；如果倍率为 null/undefined 默认按 1.0 显示
const accountOptions = computed(() =>
  accounts.value.map((a) => ({
    value: a.id,
    label: `${a.name} (${(a.rate_multiplier ?? 1).toFixed(2)}x)`,
  })),
)

// 当前生效的倍率：default → 1.0；group → 所选 group 的 rate_multiplier；account → 所选 account 的 rate_multiplier
const effectiveMultiplier = computed(() => {
  if (viewMode.value === 'group' && selectedGroupId.value) {
    const g = groups.value.find((x) => x.id === selectedGroupId.value)
    return g?.rate_multiplier ?? 1
  }
  if (viewMode.value === 'account' && selectedAccountId.value) {
    const a = accounts.value.find((x) => x.id === selectedAccountId.value)
    return a?.rate_multiplier ?? 1
  }
  return 1
})

const multiplierColor = computed(() => {
  const m = effectiveMultiplier.value
  if (m > 1.001) return 'text-red-600 dark:text-red-400'
  if (m < 0.999) return 'text-emerald-600 dark:text-emerald-400'
  return 'text-gray-700 dark:text-gray-300'
})

// 把 $/单 token × 1e6 换算为 $/MTok；应用倍率；保留 2-4 位小数
function formatPriceMTok(perToken: number): string {
  if (!perToken || perToken <= 0) return '—'
  const perMTok = perToken * 1_000_000 * effectiveMultiplier.value
  if (perMTok >= 100) return `$${perMTok.toFixed(2)}`
  if (perMTok >= 1) return `$${perMTok.toFixed(3)}`
  return `$${perMTok.toFixed(4)}`
}

// provider 颜色徽章
function providerPillClass(provider: string): string {
  const key = (provider || '').toLowerCase()
  if (key.includes('anthropic')) return 'bg-orange-50 text-orange-700 ring-1 ring-orange-200 dark:bg-orange-900/20 dark:text-orange-300 dark:ring-orange-900/50'
  if (key.includes('openai')) return 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-300 dark:ring-emerald-900/50'
  if (key.includes('google') || key.includes('gemini') || key.includes('vertex')) return 'bg-blue-50 text-blue-700 ring-1 ring-blue-200 dark:bg-blue-900/20 dark:text-blue-300 dark:ring-blue-900/50'
  if (key.includes('xai') || key.includes('grok')) return 'bg-gray-50 text-gray-700 ring-1 ring-gray-200 dark:bg-dark-800 dark:text-gray-300 dark:ring-dark-600'
  if (key.includes('mistral') || key.includes('cohere') || key.includes('groq')) return 'bg-violet-50 text-violet-700 ring-1 ring-violet-200 dark:bg-violet-900/20 dark:text-violet-300 dark:ring-violet-900/50'
  return 'bg-gray-50 text-gray-600 ring-1 ring-gray-200 dark:bg-dark-800 dark:text-gray-400 dark:ring-dark-600'
}

// 先按 provider tab 过滤、再按搜索过滤；最后由 sortedRows 应用排序
const filteredRows = computed(() => {
  let rows = pricingRows.value
  if (activeProvider.value !== 'all') {
    rows = rows.filter((r) => providerBucket(r.provider) === activeProvider.value)
  }
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    rows = rows.filter((r) => r.model.toLowerCase().includes(q) || (r.provider || '').toLowerCase().includes(q))
  }
  return rows
})

// 排序：model/provider 按字典序；价格列按数值（不应用倍率，因为倍率全表统一）
const sortedRows = computed(() => {
  const list = [...filteredRows.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  list.sort((a, b) => {
    let av: number | string = ''
    let bv: number | string = ''
    switch (sortKey.value) {
      case 'model': av = a.model; bv = b.model; break
      case 'provider': av = a.provider || ''; bv = b.provider || ''; break
      case 'input': av = a.input_cost_per_token; bv = b.input_cost_per_token; break
      case 'output': av = a.output_cost_per_token; bv = b.output_cost_per_token; break
      case 'cacheRead': av = a.cache_read_input_token_cost; bv = b.cache_read_input_token_cost; break
      case 'cacheWrite': av = a.cache_creation_input_token_cost; bv = b.cache_creation_input_token_cost; break
    }
    if (typeof av === 'number' && typeof bv === 'number') return (av - bv) * dir
    return String(av).localeCompare(String(bv)) * dir
  })
  return list
})

async function loadPricing() {
  loading.value = true
  try {
    const res = await adminAPI.pricing.listAllModelPricing()
    pricingRows.value = res.models || []
    metadata.value = res.metadata
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'admin.pricing.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch {
    /* 倍率视角是可选功能，groups 拉失败不阻塞主表格 */
  }
}

async function loadAccounts() {
  try {
    const res = await adminAPI.accounts.list(1, 200)
    accounts.value = res?.items ?? []
  } catch {
    /* same as groups */
  }
}

async function reload() {
  await Promise.all([loadPricing(), loadGroups(), loadAccounts()])
}

onMounted(reload)
</script>

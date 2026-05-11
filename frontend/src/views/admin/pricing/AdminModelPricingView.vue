<template>
  <AppLayout>
    <div class="space-y-5">
      <!-- Hero -->
      <header class="page-hero page-hero-amber">
        <div class="relative z-10 max-w-3xl">
          <span class="page-hero-tag page-hero-tag-amber">
            <Icon name="badge" size="sm" />
            {{ t('admin.pricing.title') }}
          </span>
          <h1 class="mt-3 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-[28px]">
            {{ t('admin.pricing.title') }}
          </h1>
          <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-600 dark:text-dark-200">
            {{ t('admin.pricing.subtitle') }}
          </p>
        </div>
      </header>

      <!-- 数据源元信息 + 视图控制 -->
      <div class="grid gap-3 md:grid-cols-3">
        <!-- 数据源 -->
        <div class="card p-4 text-sm">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.pricing.source') }}</p>
          <p v-if="metadata" class="mt-1 break-all font-mono text-xs text-gray-700 dark:text-gray-300">
            {{ metadata.remote_url || t('admin.pricing.sourceUnknown') }}
          </p>
          <p v-else class="mt-1 text-gray-400">—</p>
        </div>
        <div class="card p-4 text-sm">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.pricing.lastUpdated') }}</p>
          <p class="mt-1 text-gray-700 dark:text-gray-300">
            {{ metadata?.last_updated ? formatRelativeWithDateTime(metadata.last_updated) : '—' }}
          </p>
        </div>
        <div class="card p-4 text-sm">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.pricing.modelCount') }}</p>
          <p class="mt-1 text-2xl font-bold text-amber-600 dark:text-amber-300">
            {{ metadata?.model_count ?? '—' }}
          </p>
        </div>
      </div>

      <!-- 倍率视角选择器 + 操作 -->
      <div class="card p-4">
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
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.pricing.effectiveMultiplier') }}</p>
              <p class="mt-0.5 text-xl font-bold tabular-nums" :class="multiplierColor">
                {{ effectiveMultiplier.toFixed(2) }}x
              </p>
            </div>
            <button class="btn btn-secondary" :disabled="loading" @click="reload">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
        <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.pricing.viewModeHint') }}
        </p>
      </div>

      <!-- 表格 -->
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.model') }}</th>
                <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.provider') }}</th>
                <th class="px-4 py-2 text-right text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.input') }}</th>
                <th class="px-4 py-2 text-right text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.output') }}</th>
                <th class="px-4 py-2 text-right text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.cacheRead') }}</th>
                <th class="px-4 py-2 text-right text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.cacheWrite') }}</th>
                <th class="px-4 py-2 text-center text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.pricing.col.flags') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in filteredRows" :key="row.model" class="text-sm hover:bg-gray-50 dark:hover:bg-dark-800">
                <td class="px-4 py-2 font-mono text-xs text-gray-900 dark:text-white">{{ row.model }}</td>
                <td class="px-4 py-2">
                  <span :class="['rounded-full px-2 py-0.5 text-[11px] font-semibold', providerPillClass(row.provider)]">
                    {{ row.provider || '—' }}
                  </span>
                </td>
                <td class="px-4 py-2 text-right font-mono tabular-nums text-gray-900 dark:text-gray-100">{{ formatPriceMTok(row.input_cost_per_token) }}</td>
                <td class="px-4 py-2 text-right font-mono tabular-nums text-gray-900 dark:text-gray-100">{{ formatPriceMTok(row.output_cost_per_token) }}</td>
                <td class="px-4 py-2 text-right font-mono tabular-nums text-gray-700 dark:text-gray-300">{{ formatPriceMTok(row.cache_read_input_token_cost) }}</td>
                <td class="px-4 py-2 text-right font-mono tabular-nums text-gray-700 dark:text-gray-300">{{ formatPriceMTok(row.cache_creation_input_token_cost) }}</td>
                <td class="px-4 py-2 text-center text-xs">
                  <span v-if="row.supports_prompt_caching" class="mr-1 inline-block rounded bg-emerald-100 px-1.5 py-0.5 text-[10px] font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300" :title="t('admin.pricing.flagCachingTip')">cache</span>
                  <span v-if="row.supports_service_tier" class="mr-1 inline-block rounded bg-violet-100 px-1.5 py-0.5 text-[10px] font-medium text-violet-700 dark:bg-violet-900/30 dark:text-violet-300" :title="t('admin.pricing.flagPriorityTip')">priority</span>
                  <span v-if="row.mode" class="inline-block rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ row.mode }}</span>
                </td>
              </tr>
              <tr v-if="!loading && filteredRows.length === 0">
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
        <div v-if="filteredRows.length > 0" class="border-t border-gray-100 px-4 py-2 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
          {{ t('admin.pricing.tableFooter', { shown: filteredRows.length, total: pricingRows.length }) }}
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

const filteredRows = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return pricingRows.value
  return pricingRows.value.filter(
    (r) => r.model.toLowerCase().includes(q) || (r.provider || '').toLowerCase().includes(q),
  )
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

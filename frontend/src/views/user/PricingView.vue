<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部 toolbar：概览 chip + 搜索 + 刷新 -->
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-wrap items-center gap-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
          >
            <Icon name="cube" size="xs" class="text-gray-400" />
            <span class="tabular-nums">{{ filteredModels.length }}</span>
            <span class="text-gray-400 dark:text-dark-400">{{ t('pricing.statModels') }}</span>
          </span>
          <span
            v-if="availableGroups.length > 0"
            class="inline-flex items-center gap-1.5 rounded-full bg-violet-50 px-2.5 py-1 text-xs font-medium text-violet-700 ring-1 ring-inset ring-violet-200/70 dark:bg-violet-500/15 dark:text-violet-300 dark:ring-violet-500/30"
          >
            <Icon name="shield" size="xs" />
            <span class="tabular-nums">{{ availableGroups.length }}</span>
            <span class="opacity-70">{{ t('pricing.statGroups') }}</span>
          </span>
          <span
            v-if="selectedGroup"
            class="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-2.5 py-1 text-xs font-medium text-brand-700 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30"
          >
            <Icon name="bolt" size="xs" />
            <span class="opacity-70">{{ t('pricing.appliedRate') }}</span>
            <span class="tabular-nums">×{{ effectiveRate(selectedGroup).toFixed(2) }}</span>
          </span>
        </div>
        <div class="flex items-center gap-2">
          <div class="relative w-full sm:w-72">
            <Icon
              name="search"
              size="sm"
              class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
            />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('pricing.searchPlaceholder')"
              class="input pl-9 text-sm"
            />
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="loading"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
            @click="loadAll"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <!-- 主体：左侧筛选 + 右侧网格 -->
      <div class="grid gap-5 lg:grid-cols-[280px_minmax(0,1fr)]">
        <!-- 左侧筛选 panel：sticky 跟随滚动 -->
        <aside class="lg:sticky lg:top-20 lg:self-start">
          <div class="space-y-5 rounded-2xl border border-gray-200/70 bg-white p-4 shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
            <!-- 我可用的分组 -->
            <section>
              <div class="mb-2 flex items-center justify-between">
                <h3 class="text-[12px] font-semibold tracking-tight text-gray-900 dark:text-white">
                  {{ t('pricing.filterGroups') }}
                </h3>
                <button
                  v-if="selectedGroupId !== null"
                  type="button"
                  class="text-[11px] font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400"
                  @click="selectedGroupId = null"
                >
                  {{ t('common.reset') }}
                </button>
              </div>
              <div class="flex flex-col gap-1.5">
                <!-- "全部分组"：选中时显示原价（×1） -->
                <button
                  type="button"
                  class="group flex items-center justify-between rounded-lg border px-2.5 py-1.5 text-left text-[12px] transition-colors"
                  :class="
                    selectedGroupId === null
                      ? 'border-brand-300 bg-brand-50 text-brand-700 dark:border-brand-500/40 dark:bg-brand-500/10 dark:text-brand-300'
                      : 'border-gray-200/70 bg-white text-gray-700 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-dark-500'
                  "
                  @click="selectedGroupId = null"
                >
                  <span class="font-medium">{{ t('pricing.allGroups') }}</span>
                  <span class="text-[11px] opacity-70">{{ availableGroups.length }}</span>
                </button>

                <button
                  v-for="group in availableGroups"
                  :key="group.id"
                  type="button"
                  class="group flex items-center justify-between rounded-lg border px-2.5 py-1.5 text-left text-[12px] transition-colors"
                  :class="
                    selectedGroupId === group.id
                      ? 'border-brand-300 bg-brand-50 text-brand-700 dark:border-brand-500/40 dark:bg-brand-500/10 dark:text-brand-300'
                      : 'border-gray-200/70 bg-white text-gray-700 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-dark-500'
                  "
                  @click="selectedGroupId = group.id"
                >
                  <span class="flex min-w-0 items-center gap-1.5">
                    <Icon
                      v-if="group.is_exclusive"
                      name="shield"
                      size="xs"
                      class="shrink-0 text-violet-500"
                      :title="t('pricing.exclusiveGroup')"
                    />
                    <PlatformIcon :platform="group.platform as GroupPlatform" size="xs" class="shrink-0" />
                    <span class="truncate font-medium">{{ group.name }}</span>
                  </span>
                  <span
                    class="ml-2 inline-flex shrink-0 items-center rounded-full px-1.5 py-0.5 text-[10px] font-semibold tabular-nums ring-1 ring-inset"
                    :class="
                      selectedGroupId === group.id
                        ? 'bg-brand-100 text-brand-700 ring-brand-200/70 dark:bg-brand-500/20 dark:text-brand-200 dark:ring-brand-500/30'
                        : 'bg-gray-100 text-gray-600 ring-gray-200/70 dark:bg-dark-700/40 dark:text-dark-200 dark:ring-dark-600/60'
                    "
                  >
                    ×{{ effectiveRate(group).toFixed(2) }}
                  </span>
                </button>

                <p
                  v-if="!loading && availableGroups.length === 0"
                  class="text-[12px] text-gray-500 dark:text-dark-400"
                >
                  {{ t('pricing.noGroups') }}
                </p>
              </div>
            </section>

            <!-- 平台筛选 -->
            <section v-if="availablePlatforms.length > 1">
              <h3 class="mb-2 text-[12px] font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('pricing.filterPlatform') }}
              </h3>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-colors"
                  :class="
                    selectedPlatform === null
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedPlatform = null"
                >
                  {{ t('common.all') }}
                </button>
                <button
                  v-for="p in availablePlatforms"
                  :key="p"
                  type="button"
                  class="inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-colors"
                  :class="
                    selectedPlatform === p
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedPlatform = p"
                >
                  <PlatformIcon :platform="p as GroupPlatform" size="xs" />
                  {{ p }}
                </button>
              </div>
            </section>

            <!-- 计费类型筛选 -->
            <section v-if="availableBillingModes.length > 1">
              <h3 class="mb-2 text-[12px] font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('pricing.filterBillingMode') }}
              </h3>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-colors"
                  :class="
                    selectedBillingMode === null
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedBillingMode = null"
                >
                  {{ t('common.all') }}
                </button>
                <button
                  v-for="m in availableBillingModes"
                  :key="m"
                  type="button"
                  class="rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-colors"
                  :class="
                    selectedBillingMode === m
                      ? 'border-gray-900 bg-gray-900 text-white dark:border-white dark:bg-white dark:text-gray-900'
                      : 'border-gray-200/70 bg-white text-gray-600 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200'
                  "
                  @click="selectedBillingMode = m"
                >
                  {{ billingModeLabel(m) }}
                </button>
              </div>
            </section>
          </div>
        </aside>

        <!-- 右侧：模型卡片网格 -->
        <main>
          <!-- Loading -->
          <div v-if="loading" class="flex items-center justify-center py-16">
            <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>

          <!-- Empty -->
          <div
            v-else-if="filteredModels.length === 0"
            class="rounded-2xl border border-gray-200/70 bg-white p-12 text-center dark:border-dark-700/60 dark:bg-dark-800/40"
          >
            <div class="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-2xl bg-gray-50 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-700/40 dark:ring-dark-600/60">
              <Icon name="search" size="lg" class="text-gray-400" />
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ searchQuery || selectedPlatform || selectedBillingMode ? t('pricing.emptyFiltered') : t('pricing.empty') }}
            </p>
          </div>

          <!-- Grid -->
          <div v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <div
              v-for="model in filteredModels"
              :key="`${model.platform}-${model.name}`"
              class="rounded-2xl border border-gray-200/70 bg-white p-4 shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-colors hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-dark-500"
            >
              <!-- Header：平台 icon + 模型名 + 计费 chip -->
              <div class="flex items-start gap-2.5">
                <div
                  :class="[
                    'flex h-9 w-9 shrink-0 items-center justify-center rounded-xl ring-1 ring-inset',
                    platformIconBg(model.platform),
                  ]"
                >
                  <PlatformIcon :platform="model.platform as GroupPlatform" size="sm" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-semibold tracking-tight text-gray-900 dark:text-white" :title="model.name">
                    {{ model.name }}
                  </p>
                  <div class="mt-0.5 flex flex-wrap items-center gap-1.5">
                    <span
                      :class="[
                        'inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
                        billingModeChipClass(model.pricing?.billing_mode),
                      ]"
                    >
                      {{ billingModeLabel(model.pricing?.billing_mode || 'token') }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- 价格主体：根据计费模式分支 -->
              <div class="mt-3 space-y-1.5 text-[12px]">
                <template v-if="!model.pricing">
                  <p class="text-gray-400 dark:text-dark-500">{{ t('pricing.noPricing') }}</p>
                </template>

                <!-- Per-request 计费 -->
                <template v-else-if="model.pricing.billing_mode === 'per_request'">
                  <div class="flex items-center justify-between">
                    <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.perRequest') }}</span>
                    <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPrice(applyRate(model.pricing.per_request_price)) }}
                    </span>
                  </div>
                </template>

                <!-- Image 计费 -->
                <template v-else-if="model.pricing.billing_mode === 'image'">
                  <div class="flex items-center justify-between">
                    <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.imageOutput') }}</span>
                    <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPrice(applyRate(model.pricing.image_output_price)) }}
                    </span>
                  </div>
                </template>

                <!-- Token 计费（默认） -->
                <template v-else>
                  <div class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1 text-gray-500 dark:text-dark-400">
                      <Icon name="arrowDown" size="xs" class="text-emerald-500" />
                      {{ t('pricing.input') }}
                    </span>
                    <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPricePerM(applyRate(model.pricing.input_price)) }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1 text-gray-500 dark:text-dark-400">
                      <Icon name="arrowUp" size="xs" class="text-violet-500" />
                      {{ t('pricing.output') }}
                    </span>
                    <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPricePerM(applyRate(model.pricing.output_price)) }}
                    </span>
                  </div>
                  <div v-if="model.pricing.cache_read_price != null" class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1 text-gray-500 dark:text-dark-400">
                      <Icon name="inbox" size="xs" class="text-sky-500" />
                      {{ t('pricing.cacheRead') }}
                    </span>
                    <span class="font-medium tabular-nums text-sky-700 dark:text-sky-300">
                      {{ formatPricePerM(applyRate(model.pricing.cache_read_price)) }}
                    </span>
                  </div>
                  <div v-if="model.pricing.cache_write_price != null" class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1 text-gray-500 dark:text-dark-400">
                      <Icon name="edit" size="xs" class="text-amber-500" />
                      {{ t('pricing.cacheWrite') }}
                    </span>
                    <span class="font-medium tabular-nums text-amber-700 dark:text-amber-300">
                      {{ formatPricePerM(applyRate(model.pricing.cache_write_price)) }}
                    </span>
                  </div>
                </template>
              </div>

              <!-- 阶梯定价 hint：仅当有 intervals 且 token 计费时提示 -->
              <p
                v-if="model.pricing?.intervals && model.pricing.intervals.length > 0"
                class="mt-2 inline-flex items-center gap-1 rounded-full bg-indigo-50 px-2 py-0.5 text-[10px] font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200/70 dark:bg-indigo-500/15 dark:text-indigo-300 dark:ring-indigo-500/30"
                :title="t('pricing.tieredHint')"
              >
                <Icon name="chart" size="xs" />
                {{ t('pricing.tiered') }} ×{{ model.pricing.intervals.length }}
              </p>

              <!-- 模型可访问分组 hint：让用户知道这个模型属于哪些分组 -->
              <div
                v-if="!selectedGroupId && model.accessibleGroups.length > 0"
                class="mt-3 flex flex-wrap items-center gap-1 border-t border-gray-100 pt-2.5 dark:border-dark-700/60"
              >
                <span class="text-[10px] text-gray-400 dark:text-dark-500">{{ t('pricing.availableIn') }}</span>
                <span
                  v-for="g in model.accessibleGroups.slice(0, 3)"
                  :key="g.id"
                  class="inline-flex items-center gap-0.5 rounded bg-gray-50 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
                  :title="`×${effectiveRate(g).toFixed(2)}`"
                >
                  <Icon v-if="g.is_exclusive" name="shield" size="xs" class="text-violet-500" />
                  <span class="truncate max-w-[10ch]">{{ g.name }}</span>
                  <span class="opacity-60">×{{ effectiveRate(g).toFixed(1) }}</span>
                </span>
                <span
                  v-if="model.accessibleGroups.length > 3"
                  class="text-[10px] text-gray-500 dark:text-dark-400"
                >
                  +{{ model.accessibleGroups.length - 3 }}
                </span>
              </div>
            </div>
          </div>

          <!-- 倍率提示：选中"全部分组"时说明显示原价 -->
          <p
            v-if="!loading && filteredModels.length > 0 && !selectedGroup"
            class="mt-4 inline-flex items-center gap-1.5 text-[12px] text-gray-500 dark:text-dark-400"
          >
            <Icon name="infoCircle" size="xs" />
            {{ t('pricing.standardPriceHint') }}
          </p>
        </main>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserAvailableGroup,
  type UserSupportedModel,
} from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { GroupPlatform } from '@/types'
import type { BillingMode } from '@/constants/channel'

const { t } = useI18n()
const appStore = useAppStore()

// ============ 原始数据 ============
const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)

// ============ 筛选状态 ============
const searchQuery = ref('')
const selectedGroupId = ref<number | null>(null)
const selectedPlatform = ref<string | null>(null)
const selectedBillingMode = ref<BillingMode | null>(null)

// ============ 数据加载 ============
async function loadAll() {
  loading.value = true
  try {
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch(() => ({}) as Record<number, number>),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

// 实际倍率：用户专属 > 分组默认
function effectiveRate(group: UserAvailableGroup): number {
  const custom = userGroupRates.value[group.id]
  return typeof custom === 'number' ? custom : group.rate_multiplier
}

// ============ 派生：所有可访问分组（按 platform 内去重） ============
const availableGroups = computed<UserAvailableGroup[]>(() => {
  const seen = new Set<number>()
  const out: UserAvailableGroup[] = []
  for (const ch of channels.value) {
    for (const sec of ch.platforms) {
      for (const g of sec.groups) {
        if (seen.has(g.id)) continue
        seen.add(g.id)
        out.push(g)
      }
    }
  }
  // 排序：专属在前 + 倍率从小到大（更便宜的分组放前面）
  return out.sort((a, b) => {
    if (a.is_exclusive !== b.is_exclusive) return a.is_exclusive ? -1 : 1
    return effectiveRate(a) - effectiveRate(b)
  })
})

const selectedGroup = computed<UserAvailableGroup | null>(() => {
  if (selectedGroupId.value === null) return null
  return availableGroups.value.find(g => g.id === selectedGroupId.value) ?? null
})

// ============ 派生：所有可用平台 ============
const availablePlatforms = computed<string[]>(() => {
  const set = new Set<string>()
  for (const ch of channels.value) {
    for (const sec of ch.platforms) set.add(sec.platform)
  }
  return Array.from(set).sort()
})

// ============ 派生：模型聚合 ============
interface FlatModel extends UserSupportedModel {
  accessibleGroups: UserAvailableGroup[]
}

// 聚合所有 unique 模型（按 platform+name），关联可访问分组
const allModels = computed<FlatModel[]>(() => {
  const map = new Map<string, FlatModel>()
  for (const ch of channels.value) {
    for (const sec of ch.platforms) {
      for (const m of sec.supported_models) {
        const key = `${m.platform}|${m.name}`
        let entry = map.get(key)
        if (!entry) {
          entry = { ...m, accessibleGroups: [] }
          map.set(key, entry)
        }
        // 合并该 section 的分组进入此模型的可访问列表（去重）
        for (const g of sec.groups) {
          if (!entry.accessibleGroups.some(eg => eg.id === g.id)) {
            entry.accessibleGroups.push(g)
          }
        }
      }
    }
  }
  // 排序：先按平台、再按名字（让卡片排列稳定）
  return Array.from(map.values()).sort((a, b) => {
    if (a.platform !== b.platform) return a.platform.localeCompare(b.platform)
    return a.name.localeCompare(b.name)
  })
})

// ============ 派生：所有计费类型 ============
const availableBillingModes = computed<BillingMode[]>(() => {
  const set = new Set<BillingMode>()
  for (const m of allModels.value) {
    if (m.pricing?.billing_mode) set.add(m.pricing.billing_mode)
  }
  return Array.from(set).sort()
})

// ============ 派生：过滤后的模型 ============
const filteredModels = computed<FlatModel[]>(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return allModels.value.filter(m => {
    // 选中分组时：模型必须属于该分组
    if (selectedGroupId.value !== null) {
      if (!m.accessibleGroups.some(g => g.id === selectedGroupId.value)) return false
    }
    // 平台
    if (selectedPlatform.value !== null && m.platform !== selectedPlatform.value) return false
    // 计费类型
    if (selectedBillingMode.value !== null && m.pricing?.billing_mode !== selectedBillingMode.value) return false
    // 搜索
    if (q && !m.name.toLowerCase().includes(q) && !m.platform.toLowerCase().includes(q)) return false
    return true
  })
})

// ============ 价格换算 ============
// 当前生效的倍率（无选中分组时为 1，表示展示原价）
const activeRate = computed(() => (selectedGroup.value ? effectiveRate(selectedGroup.value) : 1))

function applyRate(price: number | null | undefined): number | null {
  if (price == null) return null
  return price * activeRate.value
}

// 单价（per_request / image）：4 位小数
function formatPrice(price: number | null | undefined): string {
  if (price == null) return '-'
  return `¥${price.toFixed(4)}`
}

// 每百万 token 价格：4 位小数 + /M 后缀
function formatPricePerM(price: number | null | undefined): string {
  if (price == null) return '-'
  return `¥${price.toFixed(4)}/M`
}

// ============ 计费类型 chip 配色 ============
function billingModeLabel(mode: BillingMode | string | undefined): string {
  if (!mode) return t('pricing.billingToken')
  if (mode === 'per_request') return t('pricing.billingPerRequest')
  if (mode === 'image') return t('pricing.billingImage')
  return t('pricing.billingToken')
}

function billingModeChipClass(mode: BillingMode | string | undefined): string {
  if (mode === 'per_request')
    return 'bg-violet-50 text-violet-700 ring-violet-200/70 dark:bg-violet-500/15 dark:text-violet-300 dark:ring-violet-500/30'
  if (mode === 'image')
    return 'bg-pink-50 text-pink-700 ring-pink-200/70 dark:bg-pink-500/15 dark:text-pink-300 dark:ring-pink-500/30'
  return 'bg-sky-50 text-sky-700 ring-sky-200/70 dark:bg-sky-500/15 dark:text-sky-300 dark:ring-sky-500/30'
}

// ============ 平台图标筐配色 ============
function platformIconBg(platform: string): string {
  switch (platform) {
    case 'anthropic':
      return 'bg-orange-50 ring-orange-200/70 dark:bg-orange-500/15 dark:ring-orange-500/30'
    case 'openai':
      return 'bg-emerald-50 ring-emerald-200/70 dark:bg-emerald-500/15 dark:ring-emerald-500/30'
    case 'gemini':
      return 'bg-sky-50 ring-sky-200/70 dark:bg-sky-500/15 dark:ring-sky-500/30'
    case 'antigravity':
      return 'bg-violet-50 ring-violet-200/70 dark:bg-violet-500/15 dark:ring-violet-500/30'
    default:
      return 'bg-gray-50 ring-gray-200/70 dark:bg-dark-700/40 dark:ring-dark-600/60'
  }
}

onMounted(loadAll)
</script>

<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部 toolbar：概览 chip + 搜索 + 刷新 + 汇率切换 -->
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
          <!-- 货币切换：USD 标准价 / CNY 换算后价 -->
          <div class="inline-flex rounded-lg border border-gray-200/70 bg-white p-0.5 text-xs dark:border-dark-700/60 dark:bg-dark-800/40">
            <button
              type="button"
              class="rounded-md px-2.5 py-1 font-medium transition-colors"
              :class="
                currency === 'USD'
                  ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                  : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'
              "
              @click="currency = 'USD'"
            >
              USD
            </button>
            <button
              type="button"
              class="rounded-md px-2.5 py-1 font-medium transition-colors"
              :class="
                currency === 'CNY'
                  ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900'
                  : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'
              "
              @click="currency = 'CNY'"
              :title="t('pricing.cnyHint', { rate: usdToCny.toFixed(2) })"
            >
              CNY
            </button>
          </div>
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
      <div class="grid gap-5 lg:grid-cols-[300px_minmax(0,1fr)]">
        <!-- 左侧筛选 panel：sticky 跟随滚动；字号 + 间距增大 -->
        <aside class="lg:sticky lg:top-20 lg:self-start">
          <div class="space-y-5 rounded-2xl border border-gray-200/70 bg-white p-5 shadow-[0_1px_2px_rgba(15,23,42,0.04)] dark:border-dark-700/60 dark:bg-dark-800/40">
            <!-- 我可用的分组 -->
            <section>
              <div class="mb-2.5 flex items-center justify-between">
                <h3 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
                  {{ t('pricing.filterGroups') }}
                </h3>
                <button
                  v-if="selectedGroupId !== null"
                  type="button"
                  class="text-xs font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400"
                  @click="selectedGroupId = null"
                >
                  {{ t('common.reset') }}
                </button>
              </div>
              <div class="flex flex-col gap-1.5">
                <!-- "全部分组"：选中时显示原价（×1） -->
                <button
                  type="button"
                  class="group flex items-center justify-between rounded-lg border px-3 py-2 text-left text-sm transition-colors"
                  :class="
                    selectedGroupId === null
                      ? 'border-brand-300 bg-brand-50 text-brand-700 dark:border-brand-500/40 dark:bg-brand-500/10 dark:text-brand-300'
                      : 'border-gray-200/70 bg-white text-gray-700 hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-dark-500'
                  "
                  @click="selectedGroupId = null"
                >
                  <span class="font-medium">{{ t('pricing.allGroups') }}</span>
                  <span class="text-xs opacity-70 tabular-nums">{{ availableGroups.length }}</span>
                </button>

                <button
                  v-for="group in availableGroups"
                  :key="group.id"
                  type="button"
                  class="group flex items-center justify-between gap-2 rounded-lg border px-3 py-2 text-left text-sm transition-colors"
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
                    class="ml-1 inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-[11px] font-semibold tabular-nums ring-1 ring-inset"
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
                  class="text-xs text-gray-500 dark:text-dark-400"
                >
                  {{ t('pricing.noGroups') }}
                </p>
              </div>
            </section>

            <!-- 平台筛选 -->
            <section v-if="availablePlatforms.length > 1">
              <h3 class="mb-2.5 text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('pricing.filterPlatform') }}
              </h3>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
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
                  class="inline-flex items-center gap-1 rounded-full border px-3 py-1 text-xs font-medium transition-colors"
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
              <h3 class="mb-2.5 text-sm font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ t('pricing.filterBillingMode') }}
              </h3>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
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
                  class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
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
          <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <div
              v-for="model in filteredModels"
              :key="`${model.platform}-${model.name}`"
              class="flex flex-col rounded-2xl border border-gray-200/70 bg-white p-5 shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-colors hover:border-gray-300 dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-dark-500"
            >
              <!-- Header：平台 icon + 模型名 + 计费 chip -->
              <div class="flex items-start gap-3">
                <div
                  :class="[
                    'flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ring-1 ring-inset',
                    platformIconBg(model.platform),
                  ]"
                >
                  <PlatformIcon :platform="model.platform as GroupPlatform" size="md" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-[15px] font-semibold tracking-tight text-gray-900 dark:text-white" :title="model.name">
                    {{ model.name }}
                  </p>
                  <div class="mt-1 flex flex-wrap items-center gap-1.5">
                    <span
                      :class="[
                        'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 ring-inset',
                        billingModeChipClass(model.billingMode),
                      ]"
                    >
                      {{ billingModeLabel(model.billingMode) }}
                    </span>
                    <span
                      v-if="!model.fromChannel"
                      class="inline-flex items-center gap-1 rounded-full bg-gray-50 px-2 py-0.5 text-[11px] font-medium text-gray-500 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60"
                      :title="t('pricing.standardModelHint')"
                    >
                      <Icon name="database" size="xs" />
                      {{ t('pricing.standardModel') }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- 价格主体：按计费模式分支 -->
              <div class="mt-4 flex-1 space-y-2 text-sm">
                <template v-if="!model.hasPricing">
                  <p class="text-gray-400 dark:text-dark-500">{{ t('pricing.noPricing') }}</p>
                </template>

                <!-- Per-request 计费 -->
                <template v-else-if="model.billingMode === 'per_request'">
                  <div class="flex items-center justify-between">
                    <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.perRequest') }}</span>
                    <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPrice(applyRate(model.perRequestPrice)) }}
                    </span>
                  </div>
                </template>

                <!-- Image 计费 -->
                <template v-else-if="model.billingMode === 'image'">
                  <div class="flex items-center justify-between">
                    <span class="text-gray-500 dark:text-dark-400">{{ t('pricing.imageOutput') }}</span>
                    <span class="text-base font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPrice(applyRate(model.imageOutputPrice)) }}
                    </span>
                  </div>
                </template>

                <!-- Token 计费（默认） -->
                <template v-else>
                  <div class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="arrowDown" size="sm" class="text-emerald-500" />
                      {{ t('pricing.input') }}
                    </span>
                    <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPricePerM(applyRate(model.inputPrice)) }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="arrowUp" size="sm" class="text-violet-500" />
                      {{ t('pricing.output') }}
                    </span>
                    <span class="font-semibold tabular-nums text-gray-900 dark:text-white">
                      {{ formatPricePerM(applyRate(model.outputPrice)) }}
                    </span>
                  </div>
                  <div v-if="model.cacheReadPrice != null && model.cacheReadPrice > 0" class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="inbox" size="sm" class="text-sky-500" />
                      {{ t('pricing.cacheRead') }}
                    </span>
                    <span class="font-medium tabular-nums text-sky-700 dark:text-sky-300">
                      {{ formatPricePerM(applyRate(model.cacheReadPrice)) }}
                    </span>
                  </div>
                  <div v-if="model.cacheWritePrice != null && model.cacheWritePrice > 0" class="flex items-center justify-between">
                    <span class="inline-flex items-center gap-1.5 text-gray-500 dark:text-dark-400">
                      <Icon name="edit" size="sm" class="text-amber-500" />
                      {{ t('pricing.cacheWrite') }}
                    </span>
                    <span class="font-medium tabular-nums text-amber-700 dark:text-amber-300">
                      {{ formatPricePerM(applyRate(model.cacheWritePrice)) }}
                    </span>
                  </div>
                </template>
              </div>

              <!-- 可用分组：完整显示 + 可点击切换 -->
              <div
                v-if="model.accessibleGroups.length > 0"
                class="mt-4 border-t border-gray-100 pt-3 dark:border-dark-700/60"
              >
                <p class="mb-1.5 text-xs font-medium text-gray-500 dark:text-dark-400">
                  {{ t('pricing.availableIn') }}
                </p>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="g in sortGroupsByRate(model.accessibleGroups)"
                    :key="g.id"
                    type="button"
                    class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium ring-1 ring-inset transition-colors"
                    :class="
                      selectedGroupId === g.id
                        ? 'bg-brand-50 text-brand-700 ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30'
                        : 'bg-gray-50 text-gray-600 ring-gray-200/70 hover:bg-gray-100 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60 dark:hover:bg-dark-700/60'
                    "
                    :title="t('pricing.switchToGroup', { name: g.name })"
                    @click="selectedGroupId = g.id"
                  >
                    <Icon v-if="g.is_exclusive" name="shield" size="xs" class="text-violet-500" />
                    <span>{{ g.name }}</span>
                    <span class="tabular-nums opacity-70">×{{ effectiveRate(g).toFixed(2) }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 倍率提示：选中"全部分组"时说明显示原价 -->
          <p
            v-if="!loading && filteredModels.length > 0 && !selectedGroup"
            class="mt-4 inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-3 py-1 text-[12px] font-medium text-amber-700 ring-1 ring-inset ring-amber-200/70 dark:bg-amber-500/15 dark:text-amber-300 dark:ring-amber-500/30"
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
  type PricingListEntry,
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
const allPricingEntries = ref<PricingListEntry[]>([])
const loading = ref(false)

// ============ 筛选状态 ============
const searchQuery = ref('')
const selectedGroupId = ref<number | null>(null)
const selectedPlatform = ref<string | null>(null)
const selectedBillingMode = ref<BillingMode | null>(null)

// ============ 货币 ============
// 默认显示 CNY（中国用户更习惯）；USD 按钮切回美元原价
const currency = ref<'USD' | 'CNY'>('CNY')

// USD → CNY 汇率：默认 7.2；后续可从 public settings 加 `usd_to_cny_rate` 字段覆盖
const usdToCny = computed(() => {
  const fromSettings = appStore.cachedPublicSettings as { usd_to_cny_rate?: number } | null
  const rate = fromSettings?.usd_to_cny_rate
  return typeof rate === 'number' && rate > 0 ? rate : 7.2
})

// ============ 数据加载 ============
async function loadAll() {
  loading.value = true
  try {
    const [list, rates, pricingResp] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch(() => ({}) as Record<number, number>),
      userChannelsAPI.listAllPricing().catch(() => ({ models: [], metadata: {} as never })),
    ])
    channels.value = list
    userGroupRates.value = rates
    allPricingEntries.value = pricingResp.models || []
    // 默认选中最便宜的分组（按 effectiveRate 升序后第一个）
    if (selectedGroupId.value === null && availableGroups.value.length > 0) {
      selectedGroupId.value = availableGroups.value[0].id
    }
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

// 按 effectiveRate 升序排列分组
function sortGroupsByRate(groups: UserAvailableGroup[]): UserAvailableGroup[] {
  return [...groups].sort((a, b) => effectiveRate(a) - effectiveRate(b))
}

// ============ 派生：所有可访问分组 ============
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
  return out.sort((a, b) => {
    if (a.is_exclusive !== b.is_exclusive) return a.is_exclusive ? -1 : 1
    return effectiveRate(a) - effectiveRate(b)
  })
})

const selectedGroup = computed<UserAvailableGroup | null>(() => {
  if (selectedGroupId.value === null) return null
  return availableGroups.value.find(g => g.id === selectedGroupId.value) ?? null
})

// ============ 平台 mapping ============
// LiteLLM 的 provider 字段（如 "anthropic" / "openai" / "vertex_ai-anthropic_models"）
// 与系统平台（anthropic / openai / gemini / antigravity）的映射
function providerToPlatform(provider: string, modelName: string): string | null {
  const p = (provider || '').toLowerCase()
  const m = (modelName || '').toLowerCase()
  if (p.includes('anthropic') || m.startsWith('claude')) return 'anthropic'
  if (p === 'openai' || p === 'azure' || p.startsWith('text-completion-openai') || m.startsWith('gpt') || /^o[1-9]/.test(m)) return 'openai'
  if (p.includes('gemini') || p.includes('vertex_ai-language-models') || p === 'google' || m.startsWith('gemini')) return 'gemini'
  return null
}

const availablePlatforms = computed<string[]>(() => {
  const set = new Set<string>()
  // 从用户可访问分组得到
  for (const g of availableGroups.value) {
    if (g.platform) set.add(g.platform)
  }
  return Array.from(set).sort()
})

// 可访问平台集合（用户能用的）
const accessiblePlatformSet = computed<Set<string>>(() => new Set(availablePlatforms.value))

// platform → 可访问分组列表（用于反向关联模型 → 分组）
const platformToGroups = computed<Map<string, UserAvailableGroup[]>>(() => {
  const map = new Map<string, UserAvailableGroup[]>()
  for (const g of availableGroups.value) {
    if (!g.platform) continue
    const arr = map.get(g.platform) ?? []
    arr.push(g)
    map.set(g.platform, arr)
  }
  return map
})

// ============ 派生：模型聚合 ============
// 合并两个来源：
//   1) admin 渠道里手动加的支持模型（fromChannel=true，可能有自定义定价 / intervals）
//   2) LiteLLM 全量定价（fromChannel=false，使用标准定价；这是用户期望的"自动列全"）
//
// 同名同平台优先使用 1（admin 自定义）。
interface FlatModel {
  name: string
  platform: string
  fromChannel: boolean
  hasPricing: boolean
  billingMode: BillingMode
  inputPrice: number | null
  outputPrice: number | null
  cacheReadPrice: number | null
  cacheWritePrice: number | null
  perRequestPrice: number | null
  imageOutputPrice: number | null
  accessibleGroups: UserAvailableGroup[]
}

const allModels = computed<FlatModel[]>(() => {
  const map = new Map<string, FlatModel>()

  // ── Step 1: 先放 admin 渠道自定义的模型（优先级最高）
  for (const ch of channels.value) {
    for (const sec of ch.platforms) {
      for (const m of sec.supported_models) {
        const key = `${m.platform}|${m.name}`
        let entry = map.get(key)
        if (!entry) {
          const p = m.pricing
          entry = {
            name: m.name,
            platform: m.platform,
            fromChannel: true,
            hasPricing: !!p,
            billingMode: (p?.billing_mode as BillingMode) || 'token',
            inputPrice: p?.input_price ?? null,
            outputPrice: p?.output_price ?? null,
            cacheReadPrice: p?.cache_read_price ?? null,
            cacheWritePrice: p?.cache_write_price ?? null,
            perRequestPrice: p?.per_request_price ?? null,
            imageOutputPrice: p?.image_output_price ?? null,
            accessibleGroups: [],
          }
          map.set(key, entry)
        }
        for (const g of sec.groups) {
          if (!entry.accessibleGroups.some(eg => eg.id === g.id)) {
            entry.accessibleGroups.push(g)
          }
        }
      }
    }
  }

  // ── Step 2: 再加 LiteLLM 全量模型（仅限用户可访问平台，不覆盖 Step 1 已存在的 key）
  for (const entry of allPricingEntries.value) {
    const platform = providerToPlatform(entry.provider, entry.model)
    if (!platform) continue
    if (!accessiblePlatformSet.value.has(platform)) continue
    const key = `${platform}|${entry.model}`
    if (map.has(key)) continue

    // 推断 billing mode：mode === 'image_generation' → image，其它都按 token
    const mode: BillingMode = entry.mode === 'image_generation' ? 'image' : 'token'

    map.set(key, {
      name: entry.model,
      platform,
      fromChannel: false,
      hasPricing: entry.input_cost_per_token > 0 || entry.output_cost_per_token > 0 || entry.output_cost_per_image > 0,
      billingMode: mode,
      inputPrice: entry.input_cost_per_token || null,
      outputPrice: entry.output_cost_per_token || null,
      cacheReadPrice: entry.cache_read_input_token_cost || null,
      cacheWritePrice: entry.cache_creation_input_token_cost || null,
      perRequestPrice: null,
      imageOutputPrice: entry.output_cost_per_image || null,
      // 全量模型自动关联该平台的所有可访问分组
      accessibleGroups: platformToGroups.value.get(platform) ?? [],
    })
  }

  return Array.from(map.values()).sort((a, b) => {
    if (a.platform !== b.platform) return a.platform.localeCompare(b.platform)
    return a.name.localeCompare(b.name)
  })
})

const availableBillingModes = computed<BillingMode[]>(() => {
  const set = new Set<BillingMode>()
  for (const m of allModels.value) {
    if (m.billingMode) set.add(m.billingMode)
  }
  return Array.from(set).sort()
})

const filteredModels = computed<FlatModel[]>(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return allModels.value.filter(m => {
    if (selectedGroupId.value !== null) {
      if (!m.accessibleGroups.some(g => g.id === selectedGroupId.value)) return false
    }
    if (selectedPlatform.value !== null && m.platform !== selectedPlatform.value) return false
    if (selectedBillingMode.value !== null && m.billingMode !== selectedBillingMode.value) return false
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

// 单位说明：API 返回的是 USD per token；CNY 模式下再 × usdToCny 汇率
const PER_MILLION = 1_000_000

function smartFixed(n: number): string {
  if (n === 0) return '0'
  if (n >= 100) return n.toFixed(2)
  if (n >= 1) return n.toFixed(3)
  return n.toFixed(4)
}

function applyCurrency(usdPrice: number): { value: number; symbol: string } {
  if (currency.value === 'CNY') {
    return { value: usdPrice * usdToCny.value, symbol: '¥' }
  }
  return { value: usdPrice, symbol: '$' }
}

// 单次价格（per_request / image）
function formatPrice(price: number | null | undefined): string {
  if (price == null) return '-'
  const { value, symbol } = applyCurrency(price)
  return `${symbol}${smartFixed(value)}`
}

// 每百万 token 价格
function formatPricePerM(price: number | null | undefined): string {
  if (price == null) return '-'
  const { value, symbol } = applyCurrency(price * PER_MILLION)
  return `${symbol}${smartFixed(value)}/M`
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

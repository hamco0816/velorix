<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 工具栏：左侧搜索 + 右侧刷新 + 创建按钮 -->
      <div class="flex items-center gap-2">
        <div class="relative max-w-sm flex-1">
          <Icon name="search" size="sm" class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            v-model="searchQuery"
            type="text"
            class="input pl-9"
            :placeholder="t('payment.admin.searchPlaceholder')"
          />
          <button
            v-if="searchQuery"
            type="button"
            class="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700"
            @click="searchQuery = ''"
            :title="t('common.clear')"
          >
            <Icon name="x" size="xs" />
          </button>
        </div>
        <span v-if="searchQuery" class="text-xs text-gray-500 dark:text-dark-400">
          {{ t('payment.admin.searchMatched', { n: filteredPlans.length, total: plans.length }) }}
        </span>
        <div class="ml-auto flex items-center gap-2">
          <button @click="loadPlans" :disabled="plansLoading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
            <Icon name="refresh" size="sm" :class="plansLoading ? 'animate-spin' : ''" />
          </button>
          <button @click="openPlanEdit(null)" class="btn btn-primary">
            <Icon name="plus" size="sm" class="mr-1.5" />
            {{ t('payment.admin.createPlan') }}
          </button>
        </div>
      </div>

      <!-- 批量操作工具栏：仅当有选中时显示，避免占视觉空间 -->
      <div v-if="selectedIds.size > 0" class="flex items-center gap-3 rounded-lg border border-primary-200/60 bg-primary-50/40 px-4 py-2 text-sm dark:border-primary-500/20 dark:bg-primary-500/5">
        <span class="font-medium text-primary-700 dark:text-primary-300">
          {{ t('payment.admin.bulkSelected', { n: selectedIds.size }) }}
        </span>
        <button
          type="button"
          :disabled="bulkActionLoading"
          class="inline-flex items-center gap-1 rounded-md border border-emerald-200 bg-white px-2.5 py-1 text-xs font-medium text-emerald-700 transition-colors hover:bg-emerald-50 disabled:opacity-50 dark:border-emerald-900/50 dark:bg-dark-800 dark:text-emerald-300 dark:hover:bg-emerald-950/20"
          @click="bulkSetForSale(true)"
        >
          <Icon name="checkCircle" size="xs" />
          {{ t('payment.admin.bulkActionEnable') }}
        </button>
        <button
          type="button"
          :disabled="bulkActionLoading"
          class="inline-flex items-center gap-1 rounded-md border border-gray-200 bg-white px-2.5 py-1 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-50 disabled:opacity-50 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-300 dark:hover:bg-dark-700"
          @click="bulkSetForSale(false)"
        >
          <Icon name="ban" size="xs" />
          {{ t('payment.admin.bulkActionDisable') }}
        </button>
        <button
          type="button"
          class="ml-auto text-xs text-gray-500 hover:text-gray-700 dark:text-dark-400 dark:hover:text-dark-200"
          @click="clearSelection"
        >
          {{ t('common.clear') }}
        </button>
      </div>

      <!-- Plans Table -->
      <DataTable :columns="planColumns" :data="filteredPlans" :loading="plansLoading">
        <!-- 复选框列：表头主复选 + 行内复选；indeterminate 状态表示"部分已选" -->
        <template #header-__select>
          <input
            type="checkbox"
            class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700"
            :checked="allVisibleSelected"
            :indeterminate="someVisibleSelected"
            :title="t('payment.admin.bulkSelectAll')"
            @change="toggleSelectAllVisible"
          />
        </template>
        <template #cell-__select="{ row }">
          <input
            type="checkbox"
            class="h-4 w-4 cursor-pointer rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700"
            :checked="selectedIds.has(row.id)"
            @change="toggleSelect(row.id)"
            @click.stop
          />
        </template>
        <template #cell-name="{ value, row }">
          <div class="space-y-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium" :class="getPlanNameClass(row.group_id)">{{ value }}</span>
              <!-- 主推标识：admin 一眼分清哪些档是主推（同 group 多档时尤其有用）。
                   渐变填充 + sparkles 图标，跟用户端徽章保持视觉一致。 -->
              <span
                v-if="planBadgeText(row)"
                :class="['inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold tracking-wider ring-1 shadow-sm', badgeToneClass(row.badge_color)]"
                :title="t('payment.admin.planBadgeHint')"
              >
                <Icon name="sparkles" size="xs" :stroke-width="2.5" />
                {{ planBadgeText(row) }}
              </span>
              <span
                v-if="row.kind === 'exclusive'"
                class="inline-flex items-center gap-1 rounded-full bg-violet-50 px-2 py-0.5 text-[10px] font-semibold text-violet-700 ring-1 ring-violet-200 dark:bg-violet-900/30 dark:text-violet-300 dark:ring-violet-900/50"
                :title="t('payment.admin.planKindExclusiveHint')"
              >
                <Icon name="badge" size="xs" :stroke-width="2.5" />
                {{ t('payment.admin.kindBadgeExclusive') }}
              </span>
              <span
                v-for="ct in [derivePlanCardType(row.validity_days, row.validity_unit)]"
                :key="ct"
                v-show="ct !== 'custom'"
                :class="['inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold', cardTypeBadgeClass(ct)]"
              >
                {{ t(`payment.admin.cardType.${ct}`) }}
              </span>
              <!-- 独立档位徽章：plan 有自定义限额覆盖时展示，方便管理员一眼识别 -->
              <span
                v-if="hasPlanLimitOverride(row)"
                class="inline-flex items-center gap-1 rounded-full bg-indigo-50 px-2 py-0.5 text-[10px] font-semibold text-indigo-700 ring-1 ring-indigo-200 dark:bg-indigo-900/30 dark:text-indigo-300 dark:ring-indigo-900/50"
                :title="t('payment.admin.limitOverrideHint')"
              >
                <Icon name="badge" size="xs" :stroke-width="2.5" />
                {{ t('payment.planCard.tierBadge') }}
              </span>
            </div>
            <!-- 套餐描述：admin 列表里直接看到说明文字（多行 truncate 到 2 行），hover 看完整 -->
            <p
              v-if="row.description"
              class="line-clamp-2 max-w-md whitespace-normal text-xs text-gray-500 dark:text-dark-400"
              :title="row.description"
            >
              {{ row.description }}
            </p>
            <!-- 套餐特性 chips：紧凑展示前 3 条，更多用 +N 隐藏 -->
            <div v-if="row.features && row.features.length > 0" class="flex flex-wrap gap-1">
              <span
                v-for="(f, idx) in row.features.slice(0, 3)"
                :key="idx"
                class="inline-flex max-w-[16rem] truncate rounded bg-gray-50 px-1.5 py-0.5 text-[10px] text-gray-600 dark:bg-dark-800 dark:text-dark-300"
                :title="f"
              >
                {{ f }}
              </span>
              <span
                v-if="row.features.length > 3"
                class="inline-flex rounded bg-gray-50 px-1.5 py-0.5 text-[10px] text-gray-500 dark:bg-dark-800 dark:text-dark-400"
                :title="row.features.slice(3).join('\n')"
              >
                +{{ row.features.length - 3 }}
              </span>
            </div>
          </div>
        </template>
        <template #cell-group_id="{ value }">
          <span v-if="isGroupMissing(value)" class="inline-flex items-center gap-1 text-sm">
            <span class="text-gray-400">#{{ value }}</span>
            <span class="ml-1 inline-flex items-center gap-1 rounded-md bg-rose-50 px-2 py-0.5 text-xs font-medium text-rose-700 dark:bg-rose-500/15 dark:text-rose-300">
              <span class="h-1.5 w-1.5 rounded-full bg-rose-500"></span>
              {{ t('payment.admin.groupMissing') }}
            </span>
          </span>
          <GroupBadge
            v-else-if="getGroup(value)"
            :name="getGroup(value)!.name"
            :platform="getGroup(value)!.platform"
            :rate-multiplier="getGroup(value)!.rate_multiplier"
            :promo-rate-multiplier="getGroup(value)!.promo_rate_multiplier ?? null"
            :promo-starts-at="getGroup(value)!.promo_starts_at ?? null"
            :promo-ends-at="getGroup(value)!.promo_ends_at ?? null"
          />
          <span v-else class="text-sm text-gray-400">-</span>
        </template>
        <template #cell-price="{ value, row }">
          <div class="text-sm">
            <span class="font-medium text-gray-900 dark:text-white">¥{{ (value ?? 0).toFixed(2) }}</span>
            <span v-if="row.original_price" class="ml-1 text-xs text-gray-400 line-through">¥{{ row.original_price.toFixed(2) }}</span>
          </div>
        </template>
        <template #cell-cost_multiplier="{ row }">
          <div class="text-sm" :title="costMultiplierTitle(row)">
            <span v-if="planCostEstimate(row).effectiveCostMultiplier !== null" class="font-semibold text-emerald-700 dark:text-emerald-300">
              {{ formatCostMultiplier(planCostEstimate(row).effectiveCostMultiplier) }}
            </span>
            <span v-else class="text-gray-400">-</span>
            <div v-if="planCostEstimate(row).effectiveCostMultiplier !== null" class="mt-0.5 text-[11px] text-gray-400">
              {{ formatCostMultiplier(planCostEstimate(row).priceQuotaMultiplier) }} × {{ formatCostMultiplier(planCostEstimate(row).rateMultiplier) }}
            </div>
          </div>
        </template>
        <template #cell-validity_days="{ value, row }">
          <span class="text-sm">{{ value }} {{ t('payment.admin.' + (row.validity_unit || 'days')) }}</span>
        </template>
        <template #cell-for_sale="{ value, row }">
          <button
            type="button"
            :class="[
              'inline-flex min-w-[82px] items-center justify-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-semibold shadow-sm transition-all focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
              value
                ? 'border-emerald-200 bg-emerald-50 text-emerald-700 shadow-emerald-100 hover:bg-emerald-100 dark:border-emerald-900/50 dark:bg-emerald-950/25 dark:text-emerald-300 dark:shadow-none'
                : 'border-slate-200 bg-slate-50 text-slate-500 hover:bg-slate-100 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-300'
            ]"
            @click="toggleForSale(row)"
          >
            <Icon :name="value ? 'checkCircle' : 'ban'" size="xs" :stroke-width="2" />
            {{ value ? t('payment.admin.onSale') : t('payment.admin.offSale') }}
          </button>
        </template>
        <!-- 排序：value 为 0 时（后端 omitempty 导致 undefined）默认显示 0，避免空白看上去像数据缺失 -->
        <template #cell-sort_order="{ value }">
          <span class="text-sm tabular-nums">{{ value ?? 0 }}</span>
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center justify-center gap-2">
            <button @click="openPlanEdit(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400">
              <Icon name="edit" size="sm" />
              <span class="text-xs">{{ t('common.edit') }}</span>
            </button>
            <button @click="confirmDeletePlan(row)" class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400">
              <Icon name="trash" size="sm" />
              <span class="text-xs">{{ t('common.delete') }}</span>
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Plan Edit Dialog -->
    <PlanEditDialog
      :show="showPlanDialog"
      :plan="editingPlan"
      :groups="groups"
      :prefill="planPrefill"
      @close="closePlanDialog"
      @saved="onPlanSaved"
    />

    <ConfirmDialog :show="showDeletePlanDialog" :title="t('payment.admin.deletePlan')" :message="t('payment.admin.deletePlanConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDeletePlan" @cancel="showDeletePlanDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import adminAPI from '@/api/admin'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlanEditDialog from './PlanEditDialog.vue'
import { platformTextClass } from '@/utils/platformColors'
import { derivePlanCardType, cardTypeBadgeClass } from '@/utils/planCardType'
import { calculatePlanCostEstimate, formatCostMultiplier } from '@/utils/planCost'
import { badgeToneClass } from '@/utils/badgeTone'

const { t } = useI18n()
const appStore = useAppStore()
const route = useRoute()
const router = useRouter()

/**
 * 定价助手"用此建议创建套餐"链路：route.query 带 prefill=1 + name/price/limit_usd 时，
 * 自动打开创建对话框并预填这些值。打开后立即清掉 query，避免刷新页面重复触发。
 */
const planPrefill = ref<{
  name?: string
  price?: number
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
  rate_multiplier?: number
} | null>(null)

function tryConsumePrefillFromQuery() {
  if (route.query.prefill !== '1') return
  const num = (v: unknown): number | undefined => {
    if (v === null || v === undefined) return undefined
    const n = Number(v)
    return Number.isFinite(n) && n >= 0 ? n : undefined
  }
  planPrefill.value = {
    name: typeof route.query.name === 'string' ? route.query.name : undefined,
    price: num(route.query.price),
    daily_limit_usd: num(route.query.daily_limit_usd),
    weekly_limit_usd: num(route.query.weekly_limit_usd),
    monthly_limit_usd: num(route.query.monthly_limit_usd),
    rate_multiplier: num(route.query.rate_multiplier),
  }
  editingPlan.value = null
  showPlanDialog.value = true
  // 清掉 query，刷新或返回时不会再次触发
  void router.replace({ path: route.path, query: {} })
}

function closePlanDialog() {
  showPlanDialog.value = false
  // 关闭后清掉 prefill，下次手动点"创建"时是空白表单
  planPrefill.value = null
}

function onPlanSaved() {
  planPrefill.value = null
  loadPlans()
}

// ==================== Groups ====================

const groups = ref<AdminGroup[]>([])

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch { /* ignore */ }
}

function getGroup(id: number): AdminGroup | undefined {
  return groups.value.find(g => g.id === id)
}

function isGroupMissing(id: number): boolean {
  return id > 0 && !groups.value.find(g => g.id === id)
}

function getPlanNameClass(groupId: number): string {
  const group = getGroup(groupId)
  return group ? platformTextClass(group.platform) : 'text-gray-900 dark:text-white'
}

// 判断 plan 是否有自定义限额覆盖（用于"独立档位"徽章）
// 注：管理员列表 API 返回的 plan 直接来自 ent，所以这 4 个字段是 plan 自身值（区别于用户端 checkoutPlan 的合并值）
function hasPlanLimitOverride(plan: SubscriptionPlan): boolean {
  // checkoutPlan 已经标好了 has_plan_limit_override；管理员 API 没有该字段时回退到原始字段判断
  if (typeof plan.has_plan_limit_override === 'boolean') return plan.has_plan_limit_override
  return Boolean(plan.daily_limit_usd || plan.weekly_limit_usd || plan.monthly_limit_usd || plan.rate_multiplier)
}

function planBadgeText(plan: SubscriptionPlan): string {
  return (plan.badge_text || '').trim() || (plan.is_popular ? t('payment.admin.popularBadgeShort') : '')
}

function planCostEstimate(plan: SubscriptionPlan) {
  return calculatePlanCostEstimate(plan, getGroup(plan.group_id))
}

function costMultiplierTitle(plan: SubscriptionPlan): string {
  const estimate = planCostEstimate(plan)
  if (estimate.effectiveCostMultiplier === null) return t('payment.admin.costMultiplierUnavailable')
  const quota = estimate.periodLimitUSD === null ? '-' : Number(estimate.periodLimitUSD.toFixed(4))
  return t('payment.admin.costMultiplierTitle', {
    quota,
    quotaRate: formatCostMultiplier(estimate.priceQuotaMultiplier),
    billingRate: formatCostMultiplier(estimate.rateMultiplier),
    effectiveRate: formatCostMultiplier(estimate.effectiveCostMultiplier),
  })
}


// ==================== Plans ====================

const plansLoading = ref(false)
const plans = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const showDeletePlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const deletingPlanId = ref<number | null>(null)

// 批量选择：跨筛选保留（用 Set 而非数组）。开启/关闭多档套餐时显著加速。
const selectedIds = ref<Set<number>>(new Set())
function toggleSelect(id: number) {
  const s = new Set(selectedIds.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  selectedIds.value = s
}
function clearSelection() {
  selectedIds.value = new Set()
}
function toggleSelectAllVisible() {
  // 一致性：当前所有可见行都已选 → 全部取消；否则把可见行加入选择
  const allSelected = filteredPlans.value.length > 0 && filteredPlans.value.every((p) => selectedIds.value.has(p.id))
  if (allSelected) {
    const s = new Set(selectedIds.value)
    filteredPlans.value.forEach((p) => s.delete(p.id))
    selectedIds.value = s
  } else {
    const s = new Set(selectedIds.value)
    filteredPlans.value.forEach((p) => s.add(p.id))
    selectedIds.value = s
  }
}
const allVisibleSelected = computed(
  () => filteredPlans.value.length > 0 && filteredPlans.value.every((p) => selectedIds.value.has(p.id))
)
const someVisibleSelected = computed(
  () => !allVisibleSelected.value && filteredPlans.value.some((p) => selectedIds.value.has(p.id))
)

const bulkActionLoading = ref(false)
async function bulkSetForSale(forSale: boolean) {
  if (selectedIds.value.size === 0 || bulkActionLoading.value) return
  bulkActionLoading.value = true
  // 并行触发但限流——避免一次 100 个 PUT 把后端打爆
  const ids = Array.from(selectedIds.value)
  let succeeded = 0
  try {
    for (let i = 0; i < ids.length; i += 5) {
      const chunk = ids.slice(i, i + 5)
      await Promise.all(
        chunk.map((id) =>
          adminPaymentAPI
            .updatePlan(id, { for_sale: forSale })
            .then(() => {
              const target = plans.value.find((p) => p.id === id)
              if (target) target.for_sale = forSale
              succeeded++
            })
            .catch(() => {
              /* 单条失败不阻断；最后给汇总错误 */
            })
        )
      )
    }
    if (succeeded > 0) {
      appStore.showSuccess(t('payment.admin.bulkActionDone', { n: succeeded }))
    }
    if (succeeded < ids.length) {
      appStore.showError(t('payment.admin.bulkActionPartialFail', { n: ids.length - succeeded }))
    }
    clearSelection()
  } finally {
    bulkActionLoading.value = false
  }
}

// 搜索：模糊匹配 name / description / features / 关联分组名
const searchQuery = ref('')
const filteredPlans = computed<SubscriptionPlan[]>(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return plans.value
  return plans.value.filter((p) => {
    if (p.name.toLowerCase().includes(q)) return true
    if (planBadgeText(p).toLowerCase().includes(q)) return true
    if (p.description && p.description.toLowerCase().includes(q)) return true
    if (p.features && p.features.some((f) => f.toLowerCase().includes(q))) return true
    const g = getGroup(p.group_id)
    if (g && g.name.toLowerCase().includes(q)) return true
    return false
  })
})

const planColumns = computed((): Column[] => [
  // __select 列：批量选择用的复选框，不可排序、不参与排序键持久化
  { key: '__select', label: '', align: 'center' },
  { key: 'id', label: 'ID', numeric: true, sortable: true },
  { key: 'name', label: t('payment.admin.planName'), sortable: true },
  { key: 'group_id', label: t('payment.admin.group') },
  { key: 'price', label: t('payment.admin.price'), numeric: true, sortable: true },
  { key: 'cost_multiplier', label: t('payment.admin.costMultiplier'), numeric: true },
  { key: 'validity_days', label: t('payment.admin.validityDays'), numeric: true, sortable: true },
  { key: 'for_sale', label: t('payment.admin.forSale'), align: 'center', sortable: true },
  { key: 'sort_order', label: t('payment.admin.sortOrder'), numeric: true, sortable: true },
  { key: 'actions', label: t('common.actions'), align: 'center' },
])

async function loadPlans() {
  plansLoading.value = true
  try {
    const res = await adminPaymentAPI.getPlans()
    // Backend returns features as newline-separated string; parse to array
    plans.value = (res.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
      ...p,
      features: typeof p.features === 'string'
        ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
        : (p.features || []),
    }))
  }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { plansLoading.value = false }
}

function openPlanEdit(plan: SubscriptionPlan | null) {
  editingPlan.value = plan
  showPlanDialog.value = true
}


/** Quick toggle for_sale from the list */
async function toggleForSale(plan: SubscriptionPlan) {
  try {
    await adminPaymentAPI.updatePlan(plan.id, { for_sale: !plan.for_sale })
    plan.for_sale = !plan.for_sale
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function confirmDeletePlan(plan: SubscriptionPlan) { deletingPlanId.value = plan.id; showDeletePlanDialog.value = true }
async function handleDeletePlan() {
  if (!deletingPlanId.value) return
  try { await adminPaymentAPI.deletePlan(deletingPlanId.value); appStore.showSuccess(t('common.deleted')); showDeletePlanDialog.value = false; loadPlans() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

// ==================== Lifecycle ====================

onMounted(async () => {
  // 先加载 groups + plans，再消费 prefill；这样对话框打开时 groups 已就位（select 能展示）
  await Promise.all([loadGroups(), loadPlans()])
  tryConsumePrefillFromQuery()
})
</script>

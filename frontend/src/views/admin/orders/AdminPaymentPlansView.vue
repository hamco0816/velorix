<template>
  <AppLayout>
    <div class="space-y-5">
      <!-- Hero：teal 渐变标题区，标识订阅套餐业务色（与订阅管理同色系） -->
      <header class="page-hero page-hero-teal">
        <div class="relative z-10 flex flex-wrap items-end justify-between gap-4">
          <div class="max-w-3xl">
            <span class="page-hero-tag page-hero-tag-teal">
              <Icon name="badge" size="sm" />
              {{ t('payment.admin.plansPageTitle') }}
            </span>
            <h1 class="mt-3 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white md:text-[28px]">
              {{ t('payment.admin.plansPageTitle') }}
            </h1>
            <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-600 dark:text-dark-200">
              {{ t('payment.admin.plansPageDesc') }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button @click="loadPlans" :disabled="plansLoading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="plansLoading ? 'animate-spin' : ''" />
            </button>
            <button @click="openPlanEdit(null)" class="btn btn-primary">{{ t('payment.admin.createPlan') }}</button>
          </div>
        </div>
      </header>

      <!-- Plans Table -->
      <DataTable :columns="planColumns" :data="plans" :loading="plansLoading">
        <template #cell-name="{ value, row }">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-sm font-medium" :class="getPlanNameClass(row.group_id)">{{ value }}</span>
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
        </template>
        <template #cell-group_id="{ value }">
          <span v-if="isGroupMissing(value)" class="text-sm">
            <span class="text-gray-400">#{{ value }}</span>
            <span class="ml-1 badge badge-danger">{{ t('payment.admin.groupMissing') }}</span>
          </span>
          <GroupBadge
            v-else-if="getGroup(value)"
            :name="getGroup(value)!.name"
            :platform="getGroup(value)!.platform"
            :rate-multiplier="getGroup(value)!.rate_multiplier"
          />
          <span v-else class="text-sm text-gray-400">-</span>
        </template>
        <template #cell-price="{ value, row }">
          <div class="text-sm">
            <span class="font-medium text-gray-900 dark:text-white">${{ (value ?? 0).toFixed(2) }}</span>
            <span v-if="row.original_price" class="ml-1 text-xs text-gray-400 line-through">${{ row.original_price.toFixed(2) }}</span>
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
    <PlanEditDialog :show="showPlanDialog" :plan="editingPlan" :groups="groups" @close="showPlanDialog = false" @saved="loadPlans" />

    <ConfirmDialog :show="showDeletePlanDialog" :title="t('payment.admin.deletePlan')" :message="t('payment.admin.deletePlanConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDeletePlan" @cancel="showDeletePlanDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()
const appStore = useAppStore()

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


// ==================== Plans ====================

const plansLoading = ref(false)
const plans = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const showDeletePlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const deletingPlanId = ref<number | null>(null)

const planColumns = computed((): Column[] => [
  { key: 'id', label: 'ID', numeric: true },
  { key: 'name', label: t('payment.admin.planName') },
  { key: 'group_id', label: t('payment.admin.group') },
  { key: 'price', label: t('payment.admin.price'), numeric: true },
  { key: 'validity_days', label: t('payment.admin.validityDays'), numeric: true },
  { key: 'for_sale', label: t('payment.admin.forSale'), align: 'center' },
  { key: 'sort_order', label: t('payment.admin.sortOrder'), numeric: true },
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

onMounted(() => {
  loadGroups()
  loadPlans()
})
</script>

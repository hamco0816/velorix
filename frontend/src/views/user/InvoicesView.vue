<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 可开票额度概览：页面加载即展示真实可开金额与明细，强化透明（标题由全局 AppHeader 提供） -->
      <div class="surface-card flex flex-wrap items-end justify-between gap-4 p-5">
        <div class="min-w-0">
          <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('invoice.summary.title') }}</p>
          <p class="mt-1 text-[1.75rem] font-bold leading-tight tracking-tight tabular-nums text-gray-900 dark:text-white">
            ¥{{ invoiceableSummary.available_amount.toFixed(2) }}
          </p>
          <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500 dark:text-dark-400">
            <span>{{ t('invoice.summary.balance') }}
              <span class="tabular-nums font-medium text-gray-700 dark:text-gray-300">¥{{ invoiceableSummary.balance_amount.toFixed(2) }}</span>
            </span>
            <span class="text-gray-300 dark:text-dark-600">·</span>
            <span>{{ t('invoice.summary.plan') }}
              <span class="tabular-nums font-medium text-gray-700 dark:text-gray-300">¥{{ invoiceableSummary.plan_amount.toFixed(2) }}</span>
            </span>
            <span class="text-gray-300 dark:text-dark-600">·</span>
            <span>{{ t('invoice.summary.invoiced') }}
              <span class="tabular-nums font-medium text-gray-700 dark:text-gray-300">¥{{ invoiceableSummary.invoiced_amount.toFixed(2) }}</span>
            </span>
          </div>
          <p class="mt-2 max-w-2xl text-xs text-gray-400 dark:text-dark-500">{{ t('invoice.apply.availableHint') }}</p>
        </div>
        <div class="flex shrink-0 items-center gap-2">
          <button @click="refreshAll" :disabled="loading || summaryLoading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
            <Icon name="refresh" size="sm" :class="(loading || summaryLoading) ? 'animate-spin' : ''" />
          </button>
          <button class="btn btn-primary btn-sm shrink-0 whitespace-nowrap" @click="openApplyDialog">
            <Icon name="plus" size="sm" class="mr-1.5" />
            <span>{{ t('invoice.apply.button') }}</span>
          </button>
        </div>
      </div>

      <!-- 列表：tbody 内按 加载骨架 > 加载失败 > 空 > 数据 四态展示 -->
      <div class="surface-card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50/60 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:bg-dark-800/60 dark:text-gray-400">
              <tr>
                <th class="px-4 py-3">#</th>
                <th class="px-4 py-3">{{ t('invoice.fields.titleName') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.amount') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.status') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.recipientEmail') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.createdAt') }}</th>
                <th class="px-4 py-3 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <!-- 加载中：骨架行占位 -->
              <tr v-if="loading" v-for="i in 5" :key="`skeleton-${i}`">
                <td v-for="col in 7" :key="col" class="px-4 py-3">
                  <div class="h-4 w-3/4 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
                </td>
              </tr>
              <!-- 加载失败：展示错误提示并支持重试 -->
              <tr v-else-if="loadFailed">
                <td colspan="7" class="py-6">
                  <ErrorState @retry="fetchInvoices" />
                </td>
              </tr>
              <!-- 暂无开票申请 -->
              <tr v-else-if="invoices.length === 0">
                <td colspan="7" class="py-6">
                  <EmptyState :title="t('invoice.empty')" />
                </td>
              </tr>
              <tr v-else v-for="row in invoices" :key="row.id" class="hover:bg-gray-50/60 dark:hover:bg-dark-800/40">
                <td class="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{{ row.id }}</td>
                <td class="px-4 py-3">
                  <div class="font-medium text-gray-900 dark:text-white">{{ row.title_name }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t(`invoice.titleType.${row.title_type}`) }}</div>
                </td>
                <td class="px-4 py-3 text-gray-900 dark:text-white">¥{{ row.amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <StatusTag :status="row.status" :label="t(`invoice.status.${row.status}`)" />
                </td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ row.recipient_email }}</td>
                <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ formatDate(row.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-2">
                    <button @click="openDetail(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-sky-600 hover:bg-sky-50 dark:text-sky-400 dark:hover:bg-sky-900/20">
                      <Icon name="eye" size="sm" />
                      <span>{{ t('common.view') }}</span>
                    </button>
                    <button v-if="row.status === 'pending'" @click="cancelTarget = row" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20">
                      <Icon name="x" size="sm" />
                      <span>{{ t('invoice.actions.cancel') }}</span>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- 申请开票弹窗 -->
    <BaseDialog :show="applyVisible" :title="t('invoice.apply.title')" width="wide" @close="applyVisible = false">
      <div class="space-y-5">
        <!-- 可开票额度 -->
        <div class="rounded-xl border border-teal-100 bg-teal-50 px-4 py-3 dark:border-teal-500/20 dark:bg-teal-900/20">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p class="text-sm font-medium text-teal-800 dark:text-teal-200">{{ t('invoice.apply.availableTitle') }}</p>
              <p class="mt-1 text-xs text-teal-700 dark:text-teal-300">{{ t('invoice.apply.availableHint') }}</p>
            </div>
            <div class="text-right">
              <div v-if="summaryLoading" class="text-sm text-teal-700 dark:text-teal-300">{{ t('common.loading') }}</div>
              <div v-else class="text-2xl font-semibold tabular-nums text-teal-800 dark:text-teal-100">¥{{ availableAmount.toFixed(2) }}</div>
            </div>
          </div>
        </div>

        <!-- 开票金额 -->
        <div>
          <label class="input-label">{{ t('invoice.apply.amountLabel') }}</label>
          <input
            v-model.number="form.amount"
            type="number"
            min="0"
            step="0.01"
            :max="availableAmount"
            class="input mt-1 w-full"
            :placeholder="t('invoice.apply.amountPlaceholder')"
          />
          <p class="input-hint">{{ t('invoice.apply.amountHint') }}</p>
        </div>

        <!-- 抬头类型 -->
        <div>
          <label class="input-label">{{ t('invoice.fields.titleType') }}</label>
          <Select v-model="form.title_type" :options="titleTypeOptions" class="mt-1 w-full" />
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('invoice.fields.titleName') }}</label>
            <input v-model.trim="form.title_name" type="text" class="input mt-1 w-full" :placeholder="form.title_type === 'company' ? t('invoice.apply.companyNamePlaceholder') : t('invoice.apply.personalNamePlaceholder')" />
          </div>
          <div v-if="form.title_type === 'company'">
            <label class="input-label">{{ t('invoice.fields.taxId') }}</label>
            <input v-model.trim="form.tax_id" type="text" class="input mt-1 w-full" :placeholder="t('invoice.apply.taxIdPlaceholder')" />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('invoice.fields.recipientEmail') }}</label>
          <input v-model.trim="form.recipient_email" type="email" class="input mt-1 w-full" :placeholder="t('invoice.apply.emailPlaceholder')" />
        </div>

        <div>
          <label class="input-label">{{ t('invoice.fields.userRemark') }}</label>
          <textarea v-model.trim="form.user_remark" rows="2" class="input mt-1 w-full" :placeholder="t('invoice.apply.remarkPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="applyVisible = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="submitLoading || !canSubmit" @click="submitApply">
            {{ submitLoading ? t('common.processing') : t('invoice.apply.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- 详情弹窗 -->
    <BaseDialog :show="!!detailTarget" :title="t('invoice.detail.title')" width="wide" @close="closeDetail">
      <div v-if="detailTarget" class="space-y-4">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <DetailRow :label="t('invoice.fields.status')">
            <StatusTag :status="detailTarget.status" :label="t(`invoice.status.${detailTarget.status}`)" />
          </DetailRow>
          <DetailRow :label="t('invoice.fields.amount')">¥{{ detailTarget.amount.toFixed(2) }}</DetailRow>
          <DetailRow :label="t('invoice.fields.titleType')">{{ t(`invoice.titleType.${detailTarget.title_type}`) }}</DetailRow>
          <DetailRow :label="t('invoice.fields.titleName')">{{ detailTarget.title_name }}</DetailRow>
          <DetailRow v-if="detailTarget.tax_id" :label="t('invoice.fields.taxId')">{{ detailTarget.tax_id }}</DetailRow>
          <DetailRow :label="t('invoice.fields.recipientEmail')">{{ detailTarget.recipient_email }}</DetailRow>
          <DetailRow v-if="detailTarget.invoice_number" :label="t('invoice.fields.invoiceNumber')">{{ detailTarget.invoice_number }}</DetailRow>
          <DetailRow v-if="detailTarget.issued_at" :label="t('invoice.fields.issuedAt')">{{ formatDate(detailTarget.issued_at) }}</DetailRow>
        </div>

        <div v-if="detailTarget.status === 'rejected' && detailTarget.reject_reason" class="rounded-xl bg-red-50 px-4 py-3 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-300">
          <span class="font-medium">{{ t('invoice.fields.rejectReason') }}：</span>{{ detailTarget.reject_reason }}
        </div>
        <div v-if="detailTarget.status === 'issued'" class="rounded-xl bg-emerald-50 px-4 py-3 text-sm text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
          {{ detailTarget.email_sent ? t('invoice.detail.emailSent') : t('invoice.detail.emailPending') }}
        </div>

        <div v-if="detailOrders.length > 0">
          <h4 class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('invoice.detail.relatedOrders') }}</h4>
          <div class="overflow-hidden rounded-xl border border-gray-100 dark:border-dark-700">
            <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="order in detailOrders" :key="order.id">
                  <td class="px-3 py-2 font-mono text-xs text-gray-500 dark:text-gray-400">#{{ order.id }}</td>
                  <td class="px-3 py-2 text-gray-700 dark:text-gray-300">{{ t(`invoice.orderType.${order.order_type}`, order.order_type) }}</td>
                  <td class="px-3 py-2 text-right text-gray-900 dark:text-white">¥{{ order.pay_amount.toFixed(2) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <button class="btn btn-secondary" @click="closeDetail">{{ t('common.close') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- 取消确认 -->
    <BaseDialog :show="!!cancelTarget" :title="t('invoice.actions.cancel')" width="narrow" @close="cancelTarget = null">
      <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('invoice.actions.confirmCancel') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="cancelTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">{{ actionLoading ? t('common.processing') : t('invoice.actions.cancel') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch, h, defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { invoiceAPI } from '@/api/invoice'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { InvoiceItem, InvoiceableOrder, InvoiceableSummary, InvoiceTitleType } from '@/types/invoice'
import AppLayout from '@/components/layout/AppLayout.vue'
import StatusTag from '@/components/common/StatusTag.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

// 详情行：label + 内容的轻量展示组件，避免重复模板
const DetailRow = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots }) {
    return () =>
      h('div', { class: 'flex flex-col gap-0.5' }, [
        h('span', { class: 'text-xs text-gray-500 dark:text-gray-400' }, props.label),
        h('span', { class: 'text-sm text-gray-900 dark:text-white' }, slots.default?.()),
      ])
  },
})

const loading = ref(false)
// 发票列表是否加载失败，用于展示可重试的错误态
const loadFailed = ref(false)
const invoices = ref<InvoiceItem[]>([])
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const applyVisible = ref(false)
const summaryLoading = ref(false)
const emptyInvoiceableSummary = (): InvoiceableSummary => ({
  available_amount: 0,
  balance_amount: 0,
  plan_amount: 0,
  invoiced_amount: 0,
})
const invoiceableSummary = ref<InvoiceableSummary>(emptyInvoiceableSummary())
const submitLoading = ref(false)

const detailTarget = ref<InvoiceItem | null>(null)
const detailOrders = ref<InvoiceableOrder[]>([])

const cancelTarget = ref<InvoiceItem | null>(null)
const actionLoading = ref(false)

const form = reactive<{
  recipient_email: string
  title_type: InvoiceTitleType
  title_name: string
  tax_id: string
  user_remark: string
  amount: number | null
}>({
  recipient_email: '',
  title_type: 'personal',
  title_name: '',
  tax_id: '',
  user_remark: '',
  amount: null,
})

const titleTypeOptions = computed(() => [
  { value: 'personal', label: t('invoice.titleType.personal') },
  { value: 'company', label: t('invoice.titleType.company') },
])

const availableAmount = computed(() => invoiceableSummary.value.available_amount || 0)

const canSubmit = computed(() => {
  if (availableAmount.value <= 0) return false
  if (!form.recipient_email || !form.title_name) return false
  if (form.title_type === 'company' && !form.tax_id) return false
  // amount 为空表示全额开票；填了则必须 >0 且不超过可开额度
  if (form.amount != null && form.amount !== 0) {
    if (form.amount < 0 || form.amount > availableAmount.value + 0.001) return false
  }
  return true
})

// 个人抬头时清空税号，避免误带
watch(
  () => form.title_type,
  (type) => {
    if (type === 'personal') form.tax_id = ''
  },
)

function formatDate(value?: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

async function fetchInvoices() {
  loading.value = true
  loadFailed.value = false
  try {
    const res = await invoiceAPI.getMyInvoices({ page: pagination.page, page_size: pagination.page_size })
    invoices.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    loadFailed.value = true
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchInvoices() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchInvoices() }

async function openApplyDialog() {
  applyVisible.value = true
  invoiceableSummary.value = emptyInvoiceableSummary()
  form.recipient_email = authStore.user?.email || ''
  form.title_type = 'personal'
  form.title_name = ''
  form.tax_id = ''
  form.user_remark = ''
  form.amount = null
  await fetchInvoiceableSummary()
}

async function fetchInvoiceableSummary() {
  summaryLoading.value = true
  try {
    const res = await invoiceAPI.getInvoiceableSummary()
    invoiceableSummary.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    summaryLoading.value = false
  }
}

async function submitApply() {
  if (!canSubmit.value) return
  submitLoading.value = true
  try {
    await invoiceAPI.apply({
      recipient_email: form.recipient_email,
      title_type: form.title_type,
      title_name: form.title_name,
      tax_id: form.title_type === 'company' ? form.tax_id : undefined,
      user_remark: form.user_remark || undefined,
      // 省略或 0 = 全额开票
      amount: form.amount && form.amount > 0 ? form.amount : undefined,
    })
    appStore.showSuccess(t('invoice.apply.success'))
    applyVisible.value = false
    pagination.page = 1
    await fetchInvoices()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    submitLoading.value = false
  }
}

async function openDetail(row: InvoiceItem) {
  detailTarget.value = row
  detailOrders.value = []
  try {
    const res = await invoiceAPI.getInvoice(row.id)
    detailTarget.value = res.data.invoice
    detailOrders.value = res.data.orders || []
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  }
}

function closeDetail() {
  detailTarget.value = null
  detailOrders.value = []
}

async function confirmCancel() {
  if (!cancelTarget.value) return
  actionLoading.value = true
  try {
    await invoiceAPI.cancel(cancelTarget.value.id)
    appStore.showSuccess(t('common.success'))
    cancelTarget.value = null
    await fetchInvoices()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

// 同时刷新发票列表与可开票额度概览
function refreshAll() {
  fetchInvoices()
  fetchInvoiceableSummary()
}

onMounted(refreshAll)
</script>

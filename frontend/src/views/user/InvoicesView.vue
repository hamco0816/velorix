<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 顶部操作栏 -->
      <div class="surface-card card-teal p-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('invoice.title') }}</h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.subtitle') }}</p>
          </div>
          <div class="flex items-center gap-2">
            <button @click="fetchInvoices" :disabled="loading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary btn-sm shrink-0 whitespace-nowrap" @click="openApplyDialog">
              <Icon name="plus" size="sm" class="mr-1.5" />
              <span>{{ t('invoice.apply.button') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 列表 -->
      <div class="surface-card overflow-hidden">
        <div v-if="loading && invoices.length === 0" class="p-10 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="invoices.length === 0" class="p-10 text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('invoice.empty') }}</p>
        </div>
        <div v-else class="overflow-x-auto">
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
              <tr v-for="row in invoices" :key="row.id" class="hover:bg-gray-50/60 dark:hover:bg-dark-800/40">
                <td class="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{{ row.id }}</td>
                <td class="px-4 py-3">
                  <div class="font-medium text-gray-900 dark:text-white">{{ row.title_name }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t(`invoice.titleType.${row.title_type}`) }}</div>
                </td>
                <td class="px-4 py-3 text-gray-900 dark:text-white">¥{{ row.amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="statusClass(row.status)" class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium">
                    {{ t(`invoice.status.${row.status}`) }}
                  </span>
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
        <!-- 可开票订单选择 -->
        <div>
          <div class="mb-2 flex items-center justify-between">
            <label class="input-label">{{ t('invoice.apply.selectOrders') }}</label>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.apply.amountHint') }}</span>
          </div>
          <div v-if="ordersLoading" class="rounded-xl border border-gray-100 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
            {{ t('common.loading') }}
          </div>
          <div v-else-if="invoiceableOrders.length === 0" class="rounded-xl border border-gray-100 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
            {{ t('invoice.apply.noOrders') }}
          </div>
          <div v-else class="max-h-60 space-y-2 overflow-y-auto rounded-xl border border-gray-100 p-2 dark:border-dark-700">
            <label
              v-for="order in invoiceableOrders"
              :key="order.id"
              class="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 hover:bg-gray-50 dark:hover:bg-dark-800/60"
            >
              <input type="checkbox" :value="order.id" v-model="selectedOrderIds" class="h-4 w-4 rounded border-gray-300 text-teal-600 focus:ring-teal-500" />
              <div class="flex flex-1 items-center justify-between">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t(`invoice.orderType.${order.order_type}`, order.order_type) }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">#{{ order.id }} · {{ formatDate(order.paid_at || order.created_at) }}</div>
                </div>
                <div class="text-sm font-medium text-gray-900 dark:text-white">¥{{ order.pay_amount.toFixed(2) }}</div>
              </div>
            </label>
          </div>
        </div>

        <!-- 合计 -->
        <div class="flex items-center justify-between rounded-xl bg-teal-50 px-4 py-3 dark:bg-teal-900/20">
          <span class="text-sm text-teal-700 dark:text-teal-300">{{ t('invoice.apply.total') }}</span>
          <span class="text-lg font-semibold text-teal-700 dark:text-teal-300">¥{{ selectedTotal.toFixed(2) }}</span>
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
            <span :class="statusClass(detailTarget.status)" class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium">
              {{ t(`invoice.status.${detailTarget.status}`) }}
            </span>
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
import type { InvoiceItem, InvoiceableOrder, InvoiceTitleType } from '@/types/invoice'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
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
const invoices = ref<InvoiceItem[]>([])
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const applyVisible = ref(false)
const ordersLoading = ref(false)
const invoiceableOrders = ref<InvoiceableOrder[]>([])
const selectedOrderIds = ref<number[]>([])
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
}>({
  recipient_email: '',
  title_type: 'personal',
  title_name: '',
  tax_id: '',
  user_remark: '',
})

const titleTypeOptions = computed(() => [
  { value: 'personal', label: t('invoice.titleType.personal') },
  { value: 'company', label: t('invoice.titleType.company') },
])

const selectedTotal = computed(() =>
  invoiceableOrders.value
    .filter((o) => selectedOrderIds.value.includes(o.id))
    .reduce((sum, o) => sum + o.pay_amount, 0),
)

const canSubmit = computed(() => {
  if (selectedOrderIds.value.length === 0) return false
  if (!form.recipient_email || !form.title_name) return false
  if (form.title_type === 'company' && !form.tax_id) return false
  return true
})

// 个人抬头时清空税号，避免误带
watch(
  () => form.title_type,
  (type) => {
    if (type === 'personal') form.tax_id = ''
  },
)

function statusClass(status: string): string {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300'
    case 'issued':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
    case 'rejected':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
    case 'cancelled':
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
  }
}

function formatDate(value?: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

async function fetchInvoices() {
  loading.value = true
  try {
    const res = await invoiceAPI.getMyInvoices({ page: pagination.page, page_size: pagination.page_size })
    invoices.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchInvoices() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchInvoices() }

async function openApplyDialog() {
  applyVisible.value = true
  selectedOrderIds.value = []
  form.recipient_email = authStore.user?.email || ''
  form.title_type = 'personal'
  form.title_name = ''
  form.tax_id = ''
  form.user_remark = ''
  await fetchInvoiceableOrders()
}

async function fetchInvoiceableOrders() {
  ordersLoading.value = true
  try {
    // 一次性拉取较多可开票订单供勾选；正常用户量级不会过大
    const res = await invoiceAPI.getInvoiceableOrders({ page: 1, page_size: 100 })
    invoiceableOrders.value = res.data.items || []
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    ordersLoading.value = false
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
      order_ids: selectedOrderIds.value,
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

onMounted(fetchInvoices)
</script>

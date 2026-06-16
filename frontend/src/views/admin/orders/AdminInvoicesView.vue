<template>
  <AppLayout wide>
    <div class="space-y-5">
      <!-- 筛选 -->
      <div class="surface-card p-5">
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex-1 sm:max-w-64">
            <input v-model="keyword" type="text" :placeholder="t('invoice.admin.searchPlaceholder')" class="input" @input="debounceLoad" />
          </div>
          <div class="w-[10rem]">
            <Select v-model="statusFilter" :options="statusFilterOptions" @change="reload" />
          </div>
          <div class="ml-auto flex items-center gap-2">
            <button @click="reload" :disabled="loading" class="btn btn-secondary btn-sm" :title="t('common.refresh')">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </div>

      <!-- 列表 -->
      <div class="surface-card overflow-hidden">
        <div v-if="loading && invoices.length === 0" class="p-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
        <div v-else-if="invoices.length === 0" class="p-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('invoice.empty') }}</div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
            <thead class="bg-gray-50/60 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:bg-dark-800/60 dark:text-gray-400">
              <tr>
                <th class="px-4 py-3">#</th>
                <th class="px-4 py-3">{{ t('invoice.admin.applicant') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.titleName') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.amount') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.status') }}</th>
                <th class="px-4 py-3">{{ t('invoice.fields.createdAt') }}</th>
                <th class="px-4 py-3 text-right">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in invoices" :key="row.id" class="hover:bg-gray-50/60 dark:hover:bg-dark-800/40">
                <td class="px-4 py-3 font-mono text-gray-500 dark:text-gray-400">{{ row.id }}</td>
                <td class="px-4 py-3">
                  <div class="font-medium text-gray-900 dark:text-white">{{ row.user_name || row.user_email }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ row.user_email }}</div>
                </td>
                <td class="px-4 py-3">
                  <div class="text-gray-900 dark:text-white">{{ row.title_name }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t(`invoice.titleType.${row.title_type}`) }}<template v-if="row.tax_id"> · {{ row.tax_id }}</template></div>
                </td>
                <td class="px-4 py-3 text-gray-900 dark:text-white">¥{{ row.amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="statusClass(row.status)" class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium">{{ t(`invoice.status.${row.status}`) }}</span>
                </td>
                <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ formatDate(row.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openDetail(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-600">
                      <Icon name="eye" size="sm" />{{ t('common.view') }}
                    </button>
                    <button v-if="row.status === 'pending'" @click="openIssueDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-success hover:bg-success-soft dark:text-tea-300 dark:hover:bg-success/15">
                      <Icon name="check" size="sm" />{{ t('invoice.admin.issue') }}
                    </button>
                    <button v-if="row.status === 'pending'" @click="openRejectDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-danger hover:bg-danger-soft dark:text-danger dark:hover:bg-danger/15">
                      <Icon name="x" size="sm" />{{ t('invoice.admin.reject') }}
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

    <!-- 详情弹窗 -->
    <BaseDialog :show="!!detailTarget" :title="t('invoice.detail.title')" width="wide" @close="detailTarget = null">
      <div v-if="detailTarget" class="space-y-4">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <DetailRow :label="t('invoice.admin.applicant')">{{ detailTarget.user_name || detailTarget.user_email }}</DetailRow>
          <DetailRow :label="t('invoice.fields.recipientEmail')">{{ detailTarget.recipient_email }}</DetailRow>
          <DetailRow :label="t('invoice.fields.titleType')">{{ t(`invoice.titleType.${detailTarget.title_type}`) }}</DetailRow>
          <DetailRow :label="t('invoice.fields.titleName')">{{ detailTarget.title_name }}</DetailRow>
          <DetailRow v-if="detailTarget.tax_id" :label="t('invoice.fields.taxId')">{{ detailTarget.tax_id }}</DetailRow>
          <DetailRow :label="t('invoice.fields.amount')">¥{{ detailTarget.amount.toFixed(2) }}</DetailRow>
          <DetailRow v-if="detailTarget.invoice_number" :label="t('invoice.fields.invoiceNumber')">{{ detailTarget.invoice_number }}</DetailRow>
          <DetailRow v-if="detailTarget.issued_at" :label="t('invoice.fields.issuedAt')">{{ formatDate(detailTarget.issued_at) }}</DetailRow>
        </div>
        <div v-if="detailTarget.user_remark" class="rounded-xl bg-gray-50 px-4 py-3 text-sm text-gray-700 dark:bg-dark-800 dark:text-gray-300">
          <span class="font-medium">{{ t('invoice.fields.userRemark') }}：</span>{{ detailTarget.user_remark }}
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
        <div class="flex justify-end"><button class="btn btn-secondary" @click="detailTarget = null">{{ t('common.close') }}</button></div>
      </template>
    </BaseDialog>

    <!-- 开票弹窗 -->
    <BaseDialog :show="!!issueTarget" :title="t('invoice.admin.issueTitle')" width="wide" @close="closeIssueDialog">
      <div v-if="issueTarget" class="space-y-4">
        <div class="rounded-xl bg-gray-50 px-4 py-3 text-sm dark:bg-dark-800">
          <div class="flex justify-between"><span class="text-gray-500 dark:text-gray-400">{{ t('invoice.fields.titleName') }}</span><span class="text-gray-900 dark:text-white">{{ issueTarget.title_name }}</span></div>
          <div class="mt-1 flex justify-between"><span class="text-gray-500 dark:text-gray-400">{{ t('invoice.fields.amount') }}</span><span class="text-gray-900 dark:text-white">¥{{ issueTarget.amount.toFixed(2) }}</span></div>
          <div class="mt-1 flex justify-between"><span class="text-gray-500 dark:text-gray-400">{{ t('invoice.fields.recipientEmail') }}</span><span class="text-gray-900 dark:text-white">{{ issueTarget.recipient_email }}</span></div>
        </div>

        <!-- PDF 上传 + 识别预填 -->
        <div>
          <label class="input-label">{{ t('invoice.admin.uploadPdf') }}</label>
          <p class="mb-2 text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.admin.uploadHint') }}</p>
          <input ref="fileInput" type="file" accept="application/pdf,.pdf" class="hidden" @change="onFileChange" />
          <div class="flex items-center gap-2">
            <button class="btn btn-secondary btn-sm" @click="fileInput?.click()">
              <Icon name="upload" size="sm" class="mr-1.5" />{{ issueFile ? t('invoice.admin.changeFile') : t('invoice.admin.chooseFile') }}
            </button>
            <span v-if="issueFile" class="truncate text-sm text-gray-600 dark:text-gray-300">{{ issueFile.name }}</span>
            <span v-if="parsing" class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.admin.parsing') }}</span>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <label class="input-label">{{ t('invoice.fields.invoiceNumber') }} <span class="text-danger">*</span></label>
            <input v-model.trim="issueForm.invoice_number" type="text" class="input mt-1 w-full" :placeholder="t('invoice.admin.invoiceNumberPlaceholder')" />
          </div>
          <div>
            <label class="input-label">{{ t('invoice.fields.invoiceDate') }}</label>
            <input v-model="issueForm.invoice_date" type="date" class="input mt-1 w-full" />
          </div>
          <div>
            <label class="input-label">{{ t('invoice.fields.invoiceAmount') }}</label>
            <input v-model="issueForm.invoice_amount" type="number" step="0.01" class="input mt-1 w-full" :placeholder="issueTarget.amount.toFixed(2)" />
          </div>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.admin.issueNote') }}</p>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="closeIssueDialog">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="issueLoading || !issueFile || !issueForm.invoice_number" @click="confirmIssue">
            {{ issueLoading ? t('invoice.admin.issuing') : t('invoice.admin.confirmIssue') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- 驳回弹窗 -->
    <BaseDialog :show="!!rejectTarget" :title="t('invoice.admin.rejectTitle')" width="narrow" @close="rejectTarget = null">
      <div v-if="rejectTarget" class="space-y-3">
        <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('invoice.admin.rejectNote') }}</p>
        <div>
          <label class="input-label">{{ t('invoice.fields.rejectReason') }}</label>
          <textarea v-model.trim="rejectReason" rows="3" class="input mt-1 w-full" :placeholder="t('invoice.admin.rejectReasonPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="rejectTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="rejectLoading || !rejectReason.trim()" @click="confirmReject">{{ rejectLoading ? t('common.processing') : t('invoice.admin.reject') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, h, defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { adminAPI } from '@/api/admin'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { InvoiceItem, InvoiceableOrder } from '@/types/invoice'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

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
const keyword = ref('')
const statusFilter = ref('')

const detailTarget = ref<InvoiceItem | null>(null)
const detailOrders = ref<InvoiceableOrder[]>([])

const issueTarget = ref<InvoiceItem | null>(null)
const issueFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const parsing = ref(false)
const issueLoading = ref(false)
const issueForm = reactive<{ invoice_number: string; invoice_date: string; invoice_amount: string }>({
  invoice_number: '',
  invoice_date: '',
  invoice_amount: '',
})

const rejectTarget = ref<InvoiceItem | null>(null)
const rejectReason = ref('')
const rejectLoading = ref(false)

const statusFilterOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'pending', label: t('invoice.status.pending') },
  { value: 'issued', label: t('invoice.status.issued') },
  { value: 'rejected', label: t('invoice.status.rejected') },
  { value: 'cancelled', label: t('invoice.status.cancelled') },
])

function statusClass(status: string): string {
  switch (status) {
    case 'pending':
      return 'bg-warning-soft text-warning dark:bg-warning/15 dark:text-brand-300'
    case 'issued':
      return 'bg-success-soft text-success dark:bg-success/15 dark:text-tea-300'
    case 'rejected':
      return 'bg-danger-soft text-danger dark:bg-danger/15 dark:text-danger'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
  }
}

function formatDate(value?: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function debounceLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { pagination.page = 1; load() }, 400)
}

function reload() { pagination.page = 1; load() }

async function load() {
  loading.value = true
  try {
    const res = await adminAPI.invoices.list({
      page: pagination.page,
      page_size: pagination.page_size,
      status: statusFilter.value || undefined,
      keyword: keyword.value || undefined,
    })
    invoices.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; load() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; load() }

async function openDetail(row: InvoiceItem) {
  detailTarget.value = row
  detailOrders.value = []
  try {
    const res = await adminAPI.invoices.getDetail(row.id)
    detailTarget.value = res.data.invoice
    detailOrders.value = res.data.orders || []
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  }
}

function openIssueDialog(row: InvoiceItem) {
  issueTarget.value = row
  issueFile.value = null
  issueForm.invoice_number = ''
  issueForm.invoice_date = ''
  issueForm.invoice_amount = ''
}

function closeIssueDialog() {
  issueTarget.value = null
  issueFile.value = null
}

// 选择 PDF 后立即尝试识别，预填发票号码/日期/金额（识别失败不阻断手动填写）
async function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] || null
  issueFile.value = file
  input.value = ''
  if (!file || !issueTarget.value) return
  parsing.value = true
  try {
    const res = await adminAPI.invoices.parsePdf(issueTarget.value.id, file)
    const parsed = res.data
    if (parsed.invoice_number) issueForm.invoice_number = parsed.invoice_number
    if (parsed.invoice_date) issueForm.invoice_date = parsed.invoice_date
    if (parsed.invoice_amount != null) issueForm.invoice_amount = String(parsed.invoice_amount)
  } catch {
    // 识别失败保持静默，由管理员手动填写
  } finally {
    parsing.value = false
  }
}

async function confirmIssue() {
  if (!issueTarget.value || !issueFile.value || !issueForm.invoice_number) return
  issueLoading.value = true
  try {
    await adminAPI.invoices.issue(issueTarget.value.id, {
      file: issueFile.value,
      invoice_number: issueForm.invoice_number,
      invoice_date: issueForm.invoice_date || undefined,
      invoice_amount: issueForm.invoice_amount ? Number(issueForm.invoice_amount) : undefined,
    })
    appStore.showSuccess(t('invoice.admin.issueSuccess'))
    closeIssueDialog()
    await load()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    issueLoading.value = false
  }
}

function openRejectDialog(row: InvoiceItem) {
  rejectTarget.value = row
  rejectReason.value = ''
}

async function confirmReject() {
  if (!rejectTarget.value || !rejectReason.value.trim()) return
  rejectLoading.value = true
  try {
    await adminAPI.invoices.reject(rejectTarget.value.id, rejectReason.value.trim())
    appStore.showSuccess(t('common.success'))
    rejectTarget.value = null
    await load()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    rejectLoading.value = false
  }
}

onMounted(load)
</script>

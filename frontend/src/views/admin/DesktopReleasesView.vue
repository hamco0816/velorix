<template>
  <AppLayout wide>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('admin.desktopReleases.searchPlaceholder')"
            class="input w-full sm:w-64"
            @input="handleSearch"
          />
          <div class="flex items-center gap-2 sm:flex-1">
            <Select v-model="filters.status" :options="statusFilterOptions" class="w-32 sm:w-40" @change="handleFilterChange" />
            <div class="flex flex-1 items-center justify-end gap-2">
              <button @click="loadReleases" :disabled="loading" class="btn btn-secondary shrink-0" :title="t('common.refresh')">
                <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              </button>
              <button @click="openUploadDialog" class="btn btn-primary shrink-0 whitespace-nowrap">
                <Icon name="plus" size="md" class="mr-1" />
                <span class="hidden sm:inline">{{ t('admin.desktopReleases.uploadPackage') }}</span>
                <span class="sm:hidden">{{ t('common.create') }}</span>
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="releases"
          :loading="loading"
          :error="loadError"
          @retry="loadReleases"
        >
          <template #cell-version="{ row }">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-medium text-gray-900 dark:text-white">v{{ row.version }}</span>
                <span class="rounded bg-gray-100 px-1.5 py-0.5 text-2xs text-gray-500 dark:bg-dark-700 dark:text-gray-400">{{ row.channel }}</span>
              </div>
              <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ formatFileSize(row.file_size) }}</div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span
              class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
              :class="value === 'active'
                ? 'bg-success-soft text-success dark:bg-success/15 dark:text-tea-300'
                : value === 'rolledback'
                  ? 'bg-warning-soft text-warning dark:bg-warning/15 dark:text-warning'
                  : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'"
            >
              <span class="relative flex h-1.5 w-1.5">
                <span v-if="value === 'active'" class="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-70"></span>
                <span class="relative inline-flex h-1.5 w-1.5 rounded-full" :class="value === 'active' ? 'bg-success' : value === 'rolledback' ? 'bg-warning' : 'bg-gray-400'"></span>
              </span>
              {{ statusLabel(value) }}
            </span>
          </template>

          <template #cell-mandatory="{ value }">
            <span
              v-if="value"
              class="inline-flex items-center gap-1 rounded-md bg-danger-soft px-2 py-0.5 text-xs font-medium text-danger dark:bg-danger/15 dark:text-danger"
            >
              <Icon name="exclamationTriangle" size="xs" />
              {{ t('admin.desktopReleases.mandatoryYes') }}
            </span>
            <span v-else class="text-xs text-gray-400 dark:text-dark-500">{{ t('admin.desktopReleases.mandatoryNo') }}</span>
          </template>

          <template #cell-notes="{ value }">
            <span class="block max-w-[280px] truncate text-sm text-gray-600 dark:text-gray-300" :title="value">{{ value || '—' }}</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-center space-x-1">
              <button
                v-if="row.status !== 'active'"
                @click="handleRollback(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-info-soft hover:text-info dark:hover:bg-info/15 dark:hover:text-info"
                :title="t('admin.desktopReleases.rollback')"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <button
                v-if="row.status !== 'active'"
                @click="handleDelete(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-danger-soft hover:text-danger dark:hover:bg-danger/15 dark:hover:text-danger"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
              <span v-if="row.status === 'active'" class="text-2xs text-success dark:text-tea-300">{{ t('admin.desktopReleases.currentTag') }}</span>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.desktopReleases.emptyTitle')"
              :description="t('admin.desktopReleases.emptyDesc')"
              :action-text="t('admin.desktopReleases.uploadPackage')"
              @action="openUploadDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- 上传新版本 -->
    <BaseDialog :show="showUploadDialog" :title="t('admin.desktopReleases.uploadPackage')" width="wide" @close="closeUpload">
      <form id="release-form" @submit.prevent="handleUpload" class="space-y-4">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.desktopReleases.form.version') }}</label>
            <input v-model="form.version" type="text" class="input" placeholder="0.2.0" required />
          </div>
          <div>
            <label class="input-label">{{ t('admin.desktopReleases.form.channel') }}</label>
            <Select v-model="form.channel" :options="channelOptions" />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.desktopReleases.form.mandatory') }}</label>
          <Select v-model="form.mandatory" :options="mandatoryOptions" />
          <p class="input-hint">{{ t('admin.desktopReleases.form.mandatoryHint') }}</p>
        </div>

        <div>
          <label class="input-label">{{ t('admin.desktopReleases.form.notes') }}</label>
          <textarea v-model="form.notes" rows="4" class="input" :placeholder="t('admin.desktopReleases.form.notesHint')"></textarea>
        </div>

        <div class="space-y-3 rounded-lg border border-gray-200 bg-gray-50/60 p-3 dark:border-dark-700 dark:bg-dark-800/40">
          <div>
            <label class="input-label">{{ t('admin.desktopReleases.form.setupFile') }} <span class="text-danger">*</span></label>
            <input type="file" accept=".exe" class="input" @change="onFileChange($event, 'setup')" />
            <p class="input-hint">{{ t('admin.desktopReleases.form.setupHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.desktopReleases.form.latestYml') }} <span class="text-danger">*</span></label>
            <input type="file" accept=".yml,.yaml" class="input" @change="onFileChange($event, 'latestYml')" />
            <p class="input-hint">{{ t('admin.desktopReleases.form.latestHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.desktopReleases.form.blockmap') }}</label>
            <input type="file" accept=".blockmap" class="input" @change="onFileChange($event, 'blockmap')" />
            <p class="input-hint">{{ t('admin.desktopReleases.form.blockmapHint') }}</p>
          </div>
        </div>
        <p class="input-hint">{{ t('admin.desktopReleases.form.sourceHint') }}</p>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="closeUpload" class="btn btn-secondary">{{ t('common.cancel') }}</button>
          <button type="submit" form="release-form" :disabled="uploading" class="btn btn-primary">
            {{ uploading ? t('admin.desktopReleases.uploading') : t('admin.desktopReleases.publish') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="showRollbackDialog"
      :title="t('admin.desktopReleases.rollbackTitle')"
      :message="t('admin.desktopReleases.rollbackConfirm', { version: targetRelease?.version || '' })"
      :confirm-text="t('admin.desktopReleases.rollback')"
      :cancel-text="t('common.cancel')"
      @confirm="confirmRollback"
      @cancel="showRollbackDialog = false"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.desktopReleases.deleteTitle')"
      :message="t('admin.desktopReleases.deleteConfirm', { version: targetRelease?.version || '' })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { DesktopRelease } from '@/api/admin/desktopReleases'
import type { Column } from '@/components/common/types'

import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const releases = ref<DesktopRelease[]>([])
const loading = ref(false)
const loadError = ref(false)

const filters = reactive({ status: '' })
const searchQuery = ref('')
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0, pages: 0 })

const statusFilterOptions = computed(() => [
  { value: '', label: t('admin.desktopReleases.allStatus') },
  { value: 'active', label: t('admin.desktopReleases.statusLabels.active') },
  { value: 'archived', label: t('admin.desktopReleases.statusLabels.archived') },
  { value: 'rolledback', label: t('admin.desktopReleases.statusLabels.rolledback') },
])
const channelOptions = computed(() => [
  { value: 'stable', label: t('admin.desktopReleases.channelStable') },
  { value: 'beta', label: t('admin.desktopReleases.channelBeta') },
])
const mandatoryOptions = computed(() => [
  { value: 'false', label: t('admin.desktopReleases.mandatoryNo') },
  { value: 'true', label: t('admin.desktopReleases.mandatoryYes') },
])

const columns = computed<Column[]>(() => [
  { key: 'version', label: t('admin.desktopReleases.columns.version') },
  { key: 'status', label: t('admin.desktopReleases.columns.status'), align: 'center' },
  { key: 'mandatory', label: t('admin.desktopReleases.columns.mandatory'), align: 'center' },
  { key: 'notes', label: t('admin.desktopReleases.columns.notes') },
  { key: 'created_at', label: t('admin.desktopReleases.columns.createdAt') },
  { key: 'actions', label: t('admin.desktopReleases.columns.actions'), align: 'center' },
])

function statusLabel(status: string): string {
  if (status === 'active') return t('admin.desktopReleases.statusLabels.active')
  if (status === 'archived') return t('admin.desktopReleases.statusLabels.archived')
  if (status === 'rolledback') return t('admin.desktopReleases.statusLabels.rolledback')
  return status
}

function formatFileSize(bytes: number): string {
  if (!bytes) return '—'
  const mb = bytes / (1024 * 1024)
  return mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
}

// ===== 列表 =====
let currentController: AbortController | null = null

async function loadReleases() {
  currentController?.abort()
  const requestController = new AbortController()
  currentController = requestController
  const { signal } = requestController
  try {
    loading.value = true
    loadError.value = false
    const res = await adminAPI.desktopReleases.list(
      pagination.page,
      pagination.page_size,
      { status: filters.status || undefined, search: searchQuery.value || undefined },
      { signal },
    )
    if (signal.aborted || currentController !== requestController) return
    releases.value = res.items
    pagination.total = res.total
    pagination.pages = res.pages
    pagination.page = res.page
    pagination.page_size = res.page_size
  } catch (error: any) {
    if (signal.aborted || currentController !== requestController || error?.name === 'AbortError' || error?.code === 'ERR_CANCELED') return
    loadError.value = true
    appStore.showError(error.response?.data?.detail || t('admin.desktopReleases.failedToLoad'))
  } finally {
    if (currentController === requestController) {
      loading.value = false
      currentController = null
    }
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadReleases()
}
function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadReleases()
}
function handleFilterChange() {
  pagination.page = 1
  loadReleases()
}
let searchDebounceTimer: number | null = null
function handleSearch() {
  if (searchDebounceTimer) window.clearTimeout(searchDebounceTimer)
  searchDebounceTimer = window.setTimeout(() => {
    pagination.page = 1
    loadReleases()
  }, 300)
}

// ===== 上传 =====
const showUploadDialog = ref(false)
const uploading = ref(false)
const form = reactive({ version: '', channel: 'stable', mandatory: 'false', notes: '' })
const files = reactive<{ setup: File | null; latestYml: File | null; blockmap: File | null }>({
  setup: null,
  latestYml: null,
  blockmap: null,
})

function onFileChange(e: Event, field: 'setup' | 'latestYml' | 'blockmap') {
  files[field] = (e.target as HTMLInputElement).files?.[0] ?? null
}

function openUploadDialog() {
  form.version = ''
  form.channel = 'stable'
  form.mandatory = 'false'
  form.notes = ''
  files.setup = null
  files.latestYml = null
  files.blockmap = null
  showUploadDialog.value = true
}
function closeUpload() {
  showUploadDialog.value = false
}

async function handleUpload() {
  if (!form.version.trim()) {
    appStore.showError(t('admin.desktopReleases.versionRequired'))
    return
  }
  if (!files.setup || !files.latestYml) {
    appStore.showError(t('admin.desktopReleases.filesRequired'))
    return
  }
  uploading.value = true
  try {
    await adminAPI.desktopReleases.create({
      version: form.version.trim(),
      channel: form.channel,
      mandatory: form.mandatory === 'true',
      notes: form.notes,
      setup: files.setup,
      latestYml: files.latestYml,
      blockmap: files.blockmap,
    })
    appStore.showSuccess(t('common.success'))
    showUploadDialog.value = false
    await loadReleases()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.desktopReleases.failedToUpload'))
  } finally {
    uploading.value = false
  }
}

// ===== 回滚 / 删除 =====
const showRollbackDialog = ref(false)
const showDeleteDialog = ref(false)
const targetRelease = ref<DesktopRelease | null>(null)

function handleRollback(row: DesktopRelease) {
  targetRelease.value = row
  showRollbackDialog.value = true
}
async function confirmRollback() {
  if (!targetRelease.value) return
  try {
    await adminAPI.desktopReleases.rollback(targetRelease.value.id)
    appStore.showSuccess(t('common.success'))
    showRollbackDialog.value = false
    targetRelease.value = null
    await loadReleases()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.desktopReleases.failedToRollback'))
  }
}

function handleDelete(row: DesktopRelease) {
  targetRelease.value = row
  showDeleteDialog.value = true
}
async function confirmDelete() {
  if (!targetRelease.value) return
  try {
    await adminAPI.desktopReleases.delete(targetRelease.value.id)
    appStore.showSuccess(t('common.success'))
    showDeleteDialog.value = false
    targetRelease.value = null
    await loadReleases()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.desktopReleases.failedToDelete'))
  }
}

onMounted(loadReleases)
onUnmounted(() => {
  if (searchDebounceTimer) window.clearTimeout(searchDebounceTimer)
  currentController?.abort()
})
</script>

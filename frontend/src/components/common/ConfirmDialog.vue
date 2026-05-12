<template>
  <BaseDialog :show="show" :title="title" width="narrow" @close="loading ? null : handleCancel()">
    <div class="flex items-start gap-4">
      <!-- danger 模式：红色三角警示图标圈；普通：sky 提示图标圈 — 让确认对话框有清晰视觉锚点 -->
      <div
        :class="[
          'flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ring-1 ring-inset',
          danger
            ? 'bg-rose-50 text-rose-600 ring-rose-200/70 dark:bg-rose-500/15 dark:text-rose-300 dark:ring-rose-500/30'
            : 'bg-sky-50 text-sky-600 ring-sky-200/70 dark:bg-sky-500/15 dark:text-sky-300 dark:ring-sky-500/30',
        ]"
      >
        <Icon :name="danger ? 'exclamationTriangle' : 'infoCircle'" size="md" />
      </div>
      <div class="min-w-0 flex-1 space-y-3 pt-0.5">
        <p class="text-sm leading-relaxed text-gray-600 dark:text-gray-300">{{ message }}</p>
        <slot></slot>
      </div>
    </div>

    <template #footer>
      <button
        @click="handleCancel"
        type="button"
        :disabled="loading"
        class="inline-flex items-center rounded-lg border border-gray-200 bg-white px-3.5 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:bg-dark-700 dark:focus:ring-offset-dark-800"
      >
        {{ cancelText }}
      </button>
      <button
        @click="handleConfirm"
        type="button"
        :disabled="loading"
        :class="[
          'inline-flex items-center rounded-lg px-3.5 py-1.5 text-sm font-semibold text-white shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-70 dark:focus:ring-offset-dark-800',
          danger
            ? 'bg-rose-600 hover:bg-rose-700 focus:ring-rose-500'
            : 'bg-primary-600 hover:bg-primary-700 focus:ring-primary-500',
        ]"
      >
        <!-- loading=true 时显示转圈 icon + 切换文案为 common.processing；防双击 + 用户感知"在转" -->
        <svg
          v-if="loading"
          class="-ml-0.5 mr-1.5 h-4 w-4 animate-spin text-white"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
        </svg>
        {{ loading ? processingText : confirmText }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
  // 异步确认动作进行中时，disable 两个按钮 + 在确认按钮上显示转圈 icon + 切换文案
  loading?: boolean
  // 可选覆盖 loading 文案，默认走 common.processing
  processingText?: string
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false,
  loading: false,
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))
const processingText = computed(() => props.processingText || t('common.processing'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  if (props.loading) return
  emit('confirm')
}

const handleCancel = () => {
  if (props.loading) return
  emit('cancel')
}
</script>

<!--
  ErrorState 错误态原子：与 EmptyState 同款排版，红色图标 + 标题 + 描述 + 重试按钮。
  补齐"加载 / 空 / 错误"三态中的错误态，替换各页自写的报错块。

  用法：
    <ErrorState v-if="error" :description="errMsg" @retry="reload" />
-->
<template>
  <div class="empty-state">
    <div
      class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-red-50 text-red-500 ring-1 ring-inset ring-red-200/70 dark:bg-red-500/15 dark:text-red-300 dark:ring-red-500/30"
    >
      <slot name="icon">
        <Icon name="xCircle" class="h-7 w-7" :stroke-width="1.75" />
      </slot>
    </div>

    <h3 class="empty-state-title">{{ title || t('common.loadFailed') }}</h3>
    <p v-if="description" class="empty-state-description">{{ description }}</p>

    <div v-if="showRetry || $slots.action" class="mt-5">
      <slot name="action">
        <button v-if="showRetry" class="btn btn-secondary btn-sm" @click="$emit('retry')">
          <Icon name="refresh" size="sm" class="mr-1.5" />
          {{ retryText || t('common.retry') }}
        </button>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

withDefaults(
  defineProps<{
    title?: string
    description?: string
    retryText?: string
    showRetry?: boolean
  }>(),
  { showRetry: true },
)

defineEmits<{ (e: 'retry'): void }>()

const { t } = useI18n()
</script>

<!--
  PageHeader 统一页面头部原子：标题 + 副标题 + 可选返回箭头 + 右侧操作区。
  全站每个页面顶部都用它，保证头部观感一致。

  用法：
    <PageHeader :title="t('invoice.title')" :subtitle="t('invoice.subtitle')">
      <template #meta><StatusPill .../></template>     右侧紧贴标题的小标记（状态等）
      <template #actions><button class="btn btn-primary btn-sm">操作</button></template>
    </PageHeader>
-->
<template>
  <div
    :class="[
      'flex flex-wrap items-center justify-between gap-3',
      card ? 'surface-card px-5 py-4' : '',
    ]"
  >
    <div class="flex min-w-0 items-center gap-3">
      <button
        v-if="showBack"
        type="button"
        class="btn btn-ghost btn-icon shrink-0"
        :aria-label="t('common.back')"
        @click="$emit('back')"
      >
        <Icon name="chevronLeft" size="sm" :stroke-width="2" />
      </button>
      <div class="min-w-0">
        <div class="flex items-center gap-2.5">
          <h1 class="truncate text-lg font-semibold tracking-tight text-gray-900 dark:text-white">
            {{ title }}
          </h1>
          <slot name="meta" />
        </div>
        <p v-if="subtitle" class="mt-0.5 truncate text-sm text-gray-500 dark:text-dark-400">
          {{ subtitle }}
        </p>
      </div>
    </div>
    <div v-if="$slots.actions" class="flex shrink-0 flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    // 是否显示返回箭头（详情页常用）
    showBack?: boolean
    // 是否包成 surface-card 卡片头（默认是）；false 时渲染为无背景的纯头部
    card?: boolean
  }>(),
  {
    showBack: false,
    card: true,
  },
)

defineEmits<{ (e: 'back'): void }>()

const { t } = useI18n()
</script>

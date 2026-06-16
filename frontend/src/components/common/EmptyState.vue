<template>
  <div class="empty-state">
    <!-- Icon 框：圆角方块 + 主题色软底 + ring inset，比之前的 gray 实底更精致 -->
    <div :class="['empty-state-icon-frame', iconFrameClass]">
      <slot name="icon">
        <component
          v-if="icon"
          :is="icon"
          class="empty-state-icon"
          aria-hidden="true"
        />
        <svg
          v-else
          class="empty-state-icon"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
          />
        </svg>
      </slot>
    </div>

    <h3 class="empty-state-title">
      {{ displayTitle }}
    </h3>

    <p v-if="description || message" class="empty-state-description">
      {{ description || message }}
    </p>

    <!-- Action 按钮 -->
    <div v-if="actionText || $slots.action" class="mt-5">
      <slot name="action">
        <component
          :is="actionTo ? 'RouterLink' : 'button'"
          v-if="actionText"
          :to="actionTo"
          @click="!actionTo && $emit('action')"
          class="btn btn-primary btn-sm"
        >
          <Icon v-if="actionIcon" name="plus" size="sm" class="mr-1.5" />
          {{ actionText }}
        </component>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Component } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

type Variant = 'gray' | 'brand' | 'emerald' | 'sky' | 'violet' | 'rose' | 'amber' | 'indigo'

interface Props {
  icon?: Component | string
  title?: string
  description?: string
  actionText?: string
  actionTo?: string | object
  actionIcon?: boolean
  message?: string
  /** Icon 框的主题色，让空状态跟页面业务色对齐（默认 gray） */
  variant?: Variant
}

const props = withDefaults(defineProps<Props>(), {
  description: '',
  actionIcon: true,
  variant: 'gray',
})

const displayTitle = computed(() => props.title || t('common.noData'))

// 不同业务用不同主题色 — 让空状态不再千篇一律灰色
const iconFrameClass = computed(() => `empty-state-icon-frame--${props.variant}`)

defineEmits(['action'])
</script>

<style scoped>
.empty-state {
  @apply flex flex-col items-center justify-center px-4 py-10 text-center;
}

/* Icon 框：圆角方块 + ring inset，比 gray 实色更精致；不同 variant 配不同色 */
.empty-state-icon-frame {
  @apply mb-4 flex h-14 w-14 items-center justify-center rounded-2xl ring-1 ring-inset;
}
.empty-state-icon-frame--gray {
  @apply bg-gray-50 text-gray-400 ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-400 dark:ring-dark-700/60;
}
.empty-state-icon-frame--brand {
  @apply bg-brand-50 text-brand-500 ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30;
}
/* emerald → success（竹青·正向） */
.empty-state-icon-frame--emerald {
  @apply bg-success-soft text-success ring-success/20 dark:bg-success/15 dark:text-tea-300 dark:ring-success/30;
}
/* sky → info（黛蓝·信息/同步） */
.empty-state-icon-frame--sky {
  @apply bg-info-soft text-info ring-info/20 dark:bg-info/15 dark:text-info dark:ring-info/30;
}
/* violet → info（装饰无明确语义，归最接近的信息态） */
.empty-state-icon-frame--violet {
  @apply bg-info-soft text-info ring-info/20 dark:bg-info/15 dark:text-info dark:ring-info/30;
}
/* rose → danger（印章红·错误/危险） */
.empty-state-icon-frame--rose {
  @apply bg-danger-soft text-danger ring-danger/20 dark:bg-danger/15 dark:text-danger dark:ring-danger/30;
}
/* amber → warning（茶橘·警告/告警） */
.empty-state-icon-frame--amber {
  @apply bg-warning-soft text-warning ring-warning/20 dark:bg-warning/15 dark:text-warning dark:ring-warning/30;
}
/* indigo → info（信息装饰，归最接近的信息态） */
.empty-state-icon-frame--indigo {
  @apply bg-info-soft text-info ring-info/20 dark:bg-info/15 dark:text-info dark:ring-info/30;
}

.empty-state-icon,
:slotted(.empty-state-icon) {
  @apply h-7 w-7;
}

.empty-state-title {
  @apply mb-1 text-base font-semibold tracking-tight text-gray-900 dark:text-white;
}

.empty-state-description {
  @apply max-w-sm text-sm leading-relaxed text-gray-500 dark:text-dark-400;
}
</style>

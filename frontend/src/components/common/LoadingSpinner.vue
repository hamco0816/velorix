<template>
  <div
    :class="['spinner', sizeClasses, colorClass]"
    role="status"
    :aria-label="t('common.loading')"
  >
    <span class="sr-only">{{ t('common.loading') }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

type SpinnerSize = 'sm' | 'base' | 'md' | 'lg' | 'xl'
type SpinnerColor = 'primary' | 'secondary' | 'white' | 'gray' | 'current'

interface Props {
  size?: SpinnerSize
  color?: SpinnerColor
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  color: 'primary'
})

const sizeClasses = computed(() => {
  const sizes: Record<SpinnerSize, string> = {
    sm: 'w-4 h-4 border-2',
    // base：按钮内联场景专用的 20px 档
    base: 'w-5 h-5 border-2',
    md: 'w-8 h-8 border-2',
    lg: 'w-12 h-12 border-[3px]',
    xl: 'w-16 h-16 border-4'
  }
  return sizes[props.size]
})

const colorClass = computed(() => {
  const colors: Record<SpinnerColor, string> = {
    primary: 'text-primary-500',
    secondary: 'text-gray-500 dark:text-dark-400',
    white: 'text-white',
    gray: 'text-gray-400 dark:text-dark-500',
    // current：继承父元素文字颜色，替代按钮内 stroke=currentColor 的内联 spinner SVG
    current: 'text-current'
  }
  return colors[props.color]
})
</script>

<style scoped>
.spinner {
  @apply inline-block rounded-full border-solid border-current border-r-transparent;
  animation: spin 0.75s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

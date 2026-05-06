<template>
  <BrandIcon v-if="brand" :brand="brand" :size="pixelSize" />
  <!-- Antigravity logo (cloud) -->
  <svg v-else-if="platform === 'antigravity'" :class="sizeClass" viewBox="0 0 24 24" fill="currentColor">
    <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96z" />
  </svg>
  <!-- Fallback: generic platform icon -->
  <svg v-else :class="sizeClass" fill="currentColor" viewBox="0 0 24 24">
    <path
      d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BrandIcon from '@/components/common/BrandIcon.vue'
import type { GroupPlatform } from '@/types'

interface Props {
  platform?: GroupPlatform
  size?: 'xs' | 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'sm'
})

type BrandName = 'claude' | 'openai' | 'gemini'

const brand = computed<BrandName | null>(() => {
  if (props.platform === 'anthropic') return 'claude'
  if (props.platform === 'openai') return 'openai'
  if (props.platform === 'gemini') return 'gemini'
  return null
})

const pixelSize = computed(() => {
  const sizes = {
    xs: '12px',
    sm: '14px',
    md: '16px',
    lg: '20px'
  }
  return sizes[props.size]
})

const sizeClass = computed(() => {
  const sizes = {
    xs: 'w-3 h-3',
    sm: 'w-3.5 h-3.5',
    md: 'w-4 h-4',
    lg: 'w-5 h-5'
  }
  return sizes[props.size] + ' flex-shrink-0'
})
</script>

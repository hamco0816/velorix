<template>
  <BrandIcon v-if="brand" :brand="brand" :size="`${size}px`" />
  <span
    v-else
    class="inline-flex items-center justify-center font-bold text-gray-500"
    :style="{ width: `${size}px`, height: `${size}px`, fontSize: `${Math.round(size * 0.5)}px` }"
  >
    {{ fallbackText }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BrandIcon from '@/components/common/BrandIcon.vue'
import type { Provider } from '@/api/admin/channelMonitor'

const props = withDefaults(defineProps<{
  provider: Provider | string
  size?: number
}>(), {
  size: 20,
})

type BrandName = 'claude' | 'openai' | 'gemini'

const providerKey = computed(() => String(props.provider || '').toLowerCase())

const brand = computed<BrandName | null>(() => {
  if (providerKey.value === 'anthropic') return 'claude'
  if (providerKey.value === 'openai') return 'openai'
  if (providerKey.value === 'gemini') return 'gemini'
  return null
})

const fallbackText = computed(() =>
  (props.provider || '?').charAt(0).toUpperCase()
)
</script>

<template>
  <span
    v-if="iconSrc"
    class="inline-flex shrink-0 items-center justify-center"
    :style="{ width: resolvedSize, height: resolvedSize }"
  >
    <img
      :src="iconSrc"
      :alt="altText"
      class="h-full w-full object-contain"
      draggable="false"
    />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { normalizePaymentBrand, paymentBrandIcon } from './paymentBrand'

const props = withDefaults(defineProps<{
  type: string
  alt?: string
  size?: string | number
}>(), {
  size: '18px',
})

const iconSrc = computed(() => paymentBrandIcon(props.type))

const resolvedSize = computed(() =>
  typeof props.size === 'number' ? `${props.size}px` : props.size,
)

const altText = computed(() => {
  if (props.alt) return props.alt
  const brand = normalizePaymentBrand(props.type)
  if (brand === 'alipay') return 'Alipay'
  if (brand === 'wxpay') return 'WeChat Pay'
  if (brand === 'stripe') return 'Stripe'
  if (brand === 'easypay') return 'EasyPay'
  return ''
})
</script>

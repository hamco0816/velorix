<template>
  <span
    v-if="value !== null && Number.isFinite(value)"
    :class="['trend-chip', value >= 0 ? 'trend-up' : 'trend-down']"
  >
    <Icon :name="value >= 0 ? 'arrowUp' : 'arrowDown'" size="xs" :stroke-width="2.5" />
    {{ Math.abs(value).toFixed(1) }}%
  </span>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  /** 变化百分比；null/NaN 时不渲染 chip */
  value: number | null
}>()
</script>

<style scoped>
/* 升降趋势小标签：emerald=上升、rose=下降，配箭头 + 百分比，Stripe / Linear 同款"指标增长指示器" */
.trend-chip {
  @apply inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-[11px] font-semibold tabular-nums;
}
.trend-up {
  @apply bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300;
}
.trend-down {
  @apply bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300;
}
</style>

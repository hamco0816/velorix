<!-- 额度/限额进度条：按用量比例阈值变色（正常绿 / 接近上限黄 / 超限红），密钥列表与编辑弹窗共用 -->
<template>
  <div>
    <!-- 金额行：左侧窗口标签，右侧 已用/限额，可核对数字一律 tabular-nums -->
    <div
      v-if="label || showText"
      class="flex items-baseline justify-between gap-2 tabular-nums"
      :class="dense ? 'text-2xs' : 'text-xs'"
    >
      <span class="text-gray-400 dark:text-dark-500">{{ label }}</span>
      <span class="font-medium" :class="amountClass">
        ${{ used.toFixed(decimals) }} / ${{ limit.toFixed(2) }}
      </span>
    </div>
    <div
      class="mt-1 w-full overflow-hidden rounded-full bg-gray-200/70 dark:bg-dark-700"
      :class="dense ? 'h-1' : 'h-1.5'"
    >
      <div
        class="h-full rounded-full transition-all duration-200 ease-out motion-reduce:transition-none"
        :class="barClass"
        :style="{ width: percent + '%' }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// 用量达到限额 80% 时进入警示区间（黄色）
const WARN_RATIO = 0.8

const props = withDefaults(
  defineProps<{
    /** 已用金额（USD） */
    used: number
    /** 限额金额（USD） */
    limit: number
    /** 左侧标签（如 5h / 1d / 额度） */
    label?: string
    /** 是否显示金额文字行 */
    showText?: boolean
    /** 紧凑模式：表格单元格内使用更小字号与更细的进度条 */
    dense?: boolean
    /** 已用金额的小数位数（限额固定保留 2 位） */
    decimals?: number
  }>(),
  { label: '', showText: true, dense: false, decimals: 2 }
)

const ratio = computed(() => (props.limit > 0 ? props.used / props.limit : 0))
const percent = computed(() => Math.min(ratio.value * 100, 100))

// 金额文字颜色：超限红 / 接近上限黄 / 正常灰
const amountClass = computed(() => {
  if (ratio.value >= 1) return 'text-red-600 dark:text-red-400'
  if (ratio.value >= WARN_RATIO) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-700 dark:text-gray-300'
})

// 进度条填充色：与金额文字同一套阈值
const barClass = computed(() => {
  if (ratio.value >= 1) return 'bg-red-500'
  if (ratio.value >= WARN_RATIO) return 'bg-amber-500'
  return 'bg-emerald-500'
})
</script>

<template>
  <svg
    :viewBox="`0 0 ${width} ${height}`"
    :width="width"
    :height="height"
    preserveAspectRatio="none"
    class="block"
  >
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.25" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>
    <!-- 填充区域：让 sparkline 视觉更柔和有质感 -->
    <path v-if="areaPath" :d="areaPath" :fill="`url(#${gradId})`" />
    <!-- 折线 -->
    <path
      v-if="linePath"
      :d="linePath"
      fill="none"
      :stroke="color"
      stroke-width="1.5"
      stroke-linecap="round"
      stroke-linejoin="round"
      vector-effect="non-scaling-stroke"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 数据点序列（长度 < 2 时不渲染） */
    data: number[]
    /** 折线颜色（hex 或 rgb） */
    color?: string
    /** 渲染宽度（viewBox 宽度，配合 preserveAspectRatio=none 拉伸到容器宽） */
    width?: number
    /** 渲染高度 */
    height?: number
  }>(),
  {
    color: '#f97316',
    width: 100,
    height: 36
  }
)

// 每个实例独立的渐变 id，避免多 sparkline 同页时渐变定义冲突
const gradId = `spark-${Math.random().toString(36).slice(2, 9)}`

const points = computed(() => {
  if (!props.data?.length || props.data.length < 2) return []
  const max = Math.max(...props.data)
  const min = Math.min(...props.data)
  const range = max - min || 1
  const stepX = props.width / (props.data.length - 1)
  // 上下各留 2px 让 line 不顶到边界
  const padY = 2
  return props.data.map((v, i) => ({
    x: i * stepX,
    y: props.height - padY - ((v - min) / range) * (props.height - padY * 2)
  }))
})

const linePath = computed(() => {
  if (!points.value.length) return ''
  return points.value.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x.toFixed(2)} ${p.y.toFixed(2)}`).join(' ')
})

const areaPath = computed(() => {
  if (!points.value.length) return ''
  return `${linePath.value} L ${props.width} ${props.height} L 0 ${props.height} Z`
})
</script>

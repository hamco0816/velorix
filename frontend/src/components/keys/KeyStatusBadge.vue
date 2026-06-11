<!-- 密钥状态徽章：活跃 / 已停用 / 额度耗尽 / 已过期，统一为「圆点 + 浅底胶囊」样式 -->
<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium"
    :class="cfg.badge"
  >
    <span class="relative flex h-1.5 w-1.5">
      <!-- 活跃态呼吸点：reduced-motion 下隐藏动画层 -->
      <span
        v-if="cfg.live"
        class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-70 motion-reduce:hidden"
      ></span>
      <span class="relative inline-flex h-1.5 w-1.5 rounded-full" :class="cfg.dot"></span>
    </span>
    {{ t('keys.status.' + status) }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ status: string }>()

const { t } = useI18n()

// 状态 → 配色映射：active=emerald（带呼吸点）/ quota_exhausted=amber / expired=red / 其他=灰
const STATUS_CONFIG: Record<string, { badge: string; dot: string; live?: boolean }> = {
  active: {
    badge: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300',
    dot: 'bg-emerald-500',
    live: true
  },
  quota_exhausted: {
    badge: 'bg-amber-50 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300',
    dot: 'bg-amber-500'
  },
  expired: {
    badge: 'bg-red-50 text-red-700 dark:bg-red-500/15 dark:text-red-300',
    dot: 'bg-red-500'
  },
  inactive: {
    badge: 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300',
    dot: 'bg-gray-400'
  }
}

const cfg = computed(() => STATUS_CONFIG[props.status] ?? STATUS_CONFIG.inactive)
</script>

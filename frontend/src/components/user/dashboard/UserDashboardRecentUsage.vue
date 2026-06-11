<!--
  仪表盘最近使用列表：展示最近 5 条调用记录（模型 mono 字体、金额等宽对齐）。
  loading 用行形骨架占位，empty 用 EmptyState 教学引导。
-->
<template>
  <div class="surface-card overflow-hidden">
    <div class="flex items-center justify-between border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
      <div class="flex items-center gap-2.5">
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('dashboard.recentUsage') }}</h2>
        <span class="rounded-full bg-gray-100 px-2 py-0.5 text-2xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-300">{{ t('dashboard.last7Days') }}</span>
      </div>
      <router-link to="/usage" class="flex items-center gap-1 text-sm font-medium text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">
        {{ t('dashboard.viewAllUsage') }}
        <Icon name="chevronRight" size="xs" />
      </router-link>
    </div>
    <div class="p-2">
      <!-- loading：行形骨架，与真实行节奏一致 -->
      <div v-if="loading" class="space-y-1 px-2 py-1.5">
        <div v-for="i in 5" :key="`skel-${i}`" class="flex items-center gap-3 px-2 py-3">
          <div class="h-9 w-9 animate-pulse rounded-xl bg-gray-100 dark:bg-dark-700/70"></div>
          <div class="flex-1 space-y-1.5">
            <div class="h-3 w-40 animate-pulse rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-2.5 w-24 animate-pulse rounded bg-gray-100 dark:bg-dark-700/70"></div>
          </div>
          <div class="h-3 w-20 animate-pulse rounded bg-gray-100 dark:bg-dark-700/70"></div>
        </div>
      </div>
      <div v-else-if="data.length === 0" class="py-6">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <ul v-else class="divide-y divide-gray-100 dark:divide-dark-700/50">
        <li v-for="log in data" :key="log.id" class="flex items-center justify-between gap-3 px-4 py-3.5 transition-colors hover:bg-gray-50 dark:hover:bg-dark-800/60">
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-gray-100 text-gray-600 dark:bg-dark-700/70 dark:text-dark-300">
              <Icon name="cube" size="sm" :stroke-width="1.75" />
            </div>
            <div class="min-w-0">
              <p class="truncate font-mono text-sm font-medium text-gray-900 dark:text-white" :title="log.model">{{ log.model }}</p>
              <p class="mt-0.5 text-xs tabular-nums text-gray-500 dark:text-dark-400">{{ formatDateTime(log.created_at) }}</p>
            </div>
          </div>
          <div class="flex-shrink-0 text-right">
            <p class="text-sm font-semibold tabular-nums">
              <span class="text-emerald-600 dark:text-emerald-400" :title="t('dashboard.actual')">${{ formatCost(log.actual_cost) }}</span>
              <span class="ml-1 text-xs font-normal text-gray-400 dark:text-gray-500" :title="t('dashboard.standard')">/ ${{ formatCost(log.total_cost) }}</span>
            </p>
            <p class="mt-0.5 text-xs tabular-nums text-gray-500 dark:text-dark-400">{{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens</p>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()
const { t } = useI18n()
const formatCost = (c: number) => c.toFixed(4)
</script>

<style scoped>
.surface-card {
  @apply rounded-2xl border border-gray-200/70 bg-white shadow-card;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}
</style>

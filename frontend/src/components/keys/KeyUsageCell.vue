<!-- 密钥列表「用量」单元格：今日 / 近 30 天消费金额 + 可选额度进度条 -->
<template>
  <div class="min-w-[180px] text-sm">
    <div class="flex items-baseline gap-1.5 tabular-nums">
      <span class="font-semibold text-gray-900 dark:text-white">${{ todayCost.toFixed(4) }}</span>
      <span class="text-2xs text-gray-400 dark:text-dark-500">{{ t('keys.today') }}</span>
    </div>
    <div class="mt-0.5 flex items-baseline gap-1.5 text-xs tabular-nums">
      <span class="font-medium text-gray-500 dark:text-dark-400">${{ totalCost.toFixed(4) }}</span>
      <span class="text-2xs text-gray-400 dark:text-dark-500">{{ t('keys.total') }}</span>
    </div>
    <!-- 设置了额度上限时显示额度进度 -->
    <QuotaProgressBar
      v-if="apiKey.quota > 0"
      class="mt-2"
      :used="apiKey.quota_used || 0"
      :limit="apiKey.quota"
      :label="t('keys.quota')"
      dense
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import QuotaProgressBar from './QuotaProgressBar.vue'
import type { ApiKey } from '@/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'

const props = defineProps<{
  apiKey: ApiKey
  /** 批量用量统计，加载完成前可能为空 */
  stats?: BatchApiKeyUsageStats
}>()

const { t } = useI18n()

const todayCost = computed(() => props.stats?.today_actual_cost ?? 0)
const totalCost = computed(() => props.stats?.total_actual_cost ?? 0)
</script>

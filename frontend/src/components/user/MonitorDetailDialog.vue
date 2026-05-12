<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="py-8 text-center text-sm text-gray-500">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="!detail" class="py-8 text-center text-sm text-gray-500">
      {{ t('channelStatus.detailLoadError') }}
    </div>
    <div v-else class="overflow-x-auto">
      <table class="w-full text-left text-sm">
        <thead>
          <tr class="border-b border-gray-200/70 text-[12px] font-medium text-gray-500 dark:border-dark-700/60 dark:text-dark-400">
            <th class="py-2.5 pr-3 font-medium">{{ t('channelStatus.detailColumns.model') }}</th>
            <th class="py-2.5 pr-3 text-center font-medium">{{ t('channelStatus.detailColumns.latestStatus') }}</th>
            <th class="py-2.5 pr-3 text-right font-medium">{{ t('channelStatus.detailColumns.latestLatency') }}</th>
            <th class="py-2.5 pr-3 text-right font-medium">{{ t('channelStatus.detailColumns.availability7d') }}</th>
            <th class="py-2.5 pr-3 text-right font-medium">{{ t('channelStatus.detailColumns.availability15d') }}</th>
            <th class="py-2.5 pr-3 text-right font-medium">{{ t('channelStatus.detailColumns.availability30d') }}</th>
            <th class="py-2.5 pr-3 text-right font-medium">{{ t('channelStatus.detailColumns.avgLatency7d') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="m in detail.models"
            :key="m.model"
            class="border-b border-gray-100 transition-colors hover:bg-gray-50/60 dark:border-dark-700/40 dark:hover:bg-dark-800/40"
          >
            <td class="py-2.5 pr-3 font-medium text-gray-900 dark:text-gray-100">{{ m.model }}</td>
            <td class="py-2.5 pr-3 text-center">
              <span
                class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-medium"
                :class="statusBadgeClass(m.latest_status)"
              >
                <span class="h-1.5 w-1.5 rounded-full" :class="statusDotClass(m.latest_status)"></span>
                {{ statusLabel(m.latest_status) }}
              </span>
            </td>
            <td class="py-2.5 pr-3 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatLatency(m.latest_latency_ms) }}</td>
            <td class="py-2.5 pr-3 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_7d) }}</td>
            <td class="py-2.5 pr-3 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_15d) }}</td>
            <td class="py-2.5 pr-3 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatPercent(m.availability_30d) }}</td>
            <td class="py-2.5 pr-3 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatLatency(m.avg_latency_7d_ms) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <template #footer>
      <button @click="$emit('close')" class="btn btn-secondary">
        {{ t('channelStatus.closeDetail') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  status as fetchChannelMonitorDetail,
  type UserMonitorDetail,
} from '@/api/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = defineProps<{
  show: boolean
  monitorId: number | null
  title: string
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { statusLabel, statusBadgeClass, statusDotClass, formatLatency, formatPercent } = useChannelMonitorFormat()

const detail = ref<UserMonitorDetail | null>(null)
const loading = ref(false)

async function load(id: number) {
  detail.value = null
  loading.value = true
  try {
    detail.value = await fetchChannelMonitorDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.monitorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (id != null) void load(id)
  },
  { immediate: true },
)
</script>

<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-2 p-4">
      <button @click="router.push('/keys')" class="group flex w-full items-center gap-4 rounded-md border border-gray-200 bg-white p-4 text-left transition-colors hover:border-primary-400 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-500 dark:hover:bg-dark-700/50">
        <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded bg-primary-50 text-primary-600 dark:bg-primary-500/10 dark:text-primary-400">
          <Icon name="key" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-primary-500 dark:text-dark-500"
        />
      </button>

      <button @click="router.push('/usage')" class="group flex w-full items-center gap-4 rounded-md border border-gray-200 bg-white p-4 text-left transition-colors hover:border-primary-400 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-500 dark:hover:bg-dark-700/50">
        <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded bg-emerald-50 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-400">
          <Icon name="chart" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-emerald-500 dark:text-dark-500"
        />
      </button>

      <button @click="router.push('/redeem')" class="group flex w-full items-center gap-4 rounded-md border border-gray-200 bg-white p-4 text-left transition-colors hover:border-primary-400 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-500 dark:hover:bg-dark-700/50">
        <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-400">
          <Icon name="gift" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="text-gray-400 transition-colors group-hover:text-amber-500 dark:text-dark-500"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
const router = useRouter()
const { t } = useI18n()
</script>

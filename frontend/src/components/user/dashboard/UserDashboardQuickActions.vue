<template>
  <div class="surface-card overflow-hidden">
    <div class="border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
      <h2 class="text-[15px] font-semibold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-1.5 p-2">
      <button @click="router.push('/keys')" class="group flex w-full items-center gap-3 rounded-xl px-4 py-3 text-left transition-colors hover:bg-gray-50 dark:hover:bg-dark-800/60">
        <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-sky-50 text-sky-600 dark:bg-sky-500/15 dark:text-sky-300">
          <Icon name="key" size="sm" :stroke-width="1.75" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="text-gray-300 transition-all group-hover:translate-x-0.5 group-hover:text-gray-500 dark:text-dark-500 dark:group-hover:text-dark-300" />
      </button>

      <button @click="router.push('/usage')" class="group flex w-full items-center gap-3 rounded-xl px-4 py-3 text-left transition-colors hover:bg-gray-50 dark:hover:bg-dark-800/60">
        <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300">
          <Icon name="chart" size="sm" :stroke-width="1.75" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="text-gray-300 transition-all group-hover:translate-x-0.5 group-hover:text-gray-500 dark:text-dark-500 dark:group-hover:text-dark-300" />
      </button>

      <button @click="router.push('/redeem')" class="group flex w-full items-center gap-3 rounded-xl px-4 py-3 text-left transition-colors hover:bg-gray-50 dark:hover:bg-dark-800/60">
        <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-amber-50 text-amber-600 dark:bg-amber-500/15 dark:text-amber-300">
          <Icon name="gift" size="sm" :stroke-width="1.75" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="text-gray-300 transition-all group-hover:translate-x-0.5 group-hover:text-gray-500 dark:text-dark-500 dark:group-hover:text-dark-300" />
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

<style scoped>
/* 快速操作：amber 主题（操作 / 充值氛围，跟 PaymentView 的钱包色一致） */
.surface-card {
  @apply rounded-2xl;
  border: 1px solid rgb(254 215 170 / 0.7);
  background:
    radial-gradient(circle at 100% 0%, rgb(245 158 11 / 0.08), transparent 32%),
    radial-gradient(circle at 0% 100%, rgb(249 115 22 / 0.04), transparent 38%),
    linear-gradient(135deg, rgb(255 251 235), rgb(255 255 255) 55%, rgb(254 243 199));
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}
:root.dark .surface-card {
  border-color: rgb(120 53 15 / 0.4);
  background:
    radial-gradient(circle at 100% 0%, rgb(245 158 11 / 0.08), transparent 32%),
    radial-gradient(circle at 0% 100%, rgb(249 115 22 / 0.04), transparent 38%),
    linear-gradient(135deg, rgb(30 41 59 / 0.6), rgb(17 24 39 / 0.5));
}
</style>

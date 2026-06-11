<!--
  仪表盘快捷操作：高频入口（建 Key / 查用量 / 兑换码）。
  图标 chip 在 hover 时反转为黑底白字（与全站黑 CTA 同语言），箭头滑入并点亮品牌橙。
-->
<template>
  <div class="surface-card overflow-hidden">
    <div class="border-b border-gray-200/60 px-6 py-4 dark:border-dark-700/60">
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('dashboard.quickActions') }}</h2>
    </div>
    <div class="space-y-2 p-3">
      <button
        v-for="action in actions"
        :key="action.to"
        type="button"
        class="action-item group"
        @click="router.push(action.to)"
      >
        <span class="action-icon">
          <Icon :name="action.icon" size="sm" :stroke-width="1.75" />
        </span>
        <span class="min-w-0 flex-1">
          <span class="block truncate text-sm font-medium text-gray-900 dark:text-white">{{ action.title }}</span>
          <span class="mt-0.5 block truncate text-xs text-gray-500 dark:text-dark-400">{{ action.description }}</span>
        </span>
        <Icon name="chevronRight" size="sm" class="action-arrow" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const router = useRouter()
const { t } = useI18n()

// 快捷入口定义：图标 + 标题 + 描述 + 目标路由
const actions = computed(() => [
  {
    icon: 'key' as const,
    title: t('dashboard.createApiKey'),
    description: t('dashboard.generateNewKey'),
    to: '/keys'
  },
  {
    icon: 'chart' as const,
    title: t('dashboard.viewUsage'),
    description: t('dashboard.checkDetailedLogs'),
    to: '/usage'
  },
  {
    icon: 'gift' as const,
    title: t('dashboard.redeemCode'),
    description: t('dashboard.addBalanceWithCode'),
    to: '/redeem'
  }
])
</script>

<style scoped>
.surface-card {
  @apply rounded-2xl border border-gray-200/70 bg-white shadow-card;
  @apply dark:border-dark-700/60 dark:bg-dark-800/40;
}

/* 单条入口：透明描边占位，hover 时浮现边框 + 浅底 + 轻投影 */
.action-item {
  @apply flex w-full items-center gap-3 rounded-xl border border-transparent px-3.5 py-3 text-left;
  @apply transition-all duration-150 ease-out;
  @apply hover:border-gray-200 hover:bg-gray-50 hover:shadow-card;
  @apply dark:hover:border-dark-700 dark:hover:bg-dark-800/60;
  @apply focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600/40 dark:focus-visible:ring-brand-400/40;
}

/* 图标 chip：常态中性灰，hover 反转为黑底白字（暗色模式反相白底黑字） */
.action-icon {
  @apply flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl;
  @apply bg-gray-100 text-gray-600 transition-colors duration-150 ease-out;
  @apply group-hover:bg-primary-950 group-hover:text-white;
  @apply dark:bg-dark-700/70 dark:text-dark-300;
  @apply dark:group-hover:bg-white dark:group-hover:text-primary-950;
}

/* 箭头：hover 右滑 + 点亮品牌橙 */
.action-arrow {
  @apply flex-shrink-0 text-gray-300 transition-all duration-150 ease-out;
  @apply group-hover:translate-x-0.5 group-hover:text-brand-600;
  @apply dark:text-dark-500 dark:group-hover:text-brand-400;
}

/* 降级：减少动态偏好时取消位移反馈 */
@media (prefers-reduced-motion: reduce) {
  .action-arrow {
    @apply transition-none group-hover:translate-x-0;
  }
}
</style>

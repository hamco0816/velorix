<template>
  <!-- 营销首页页脚：品牌信息 + 产品/支持链接 -->
  <footer class="border-t border-gray-200 bg-white dark:border-dark-800 dark:bg-dark-950">
    <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div class="grid gap-10 sm:grid-cols-2 lg:grid-cols-3">
        <!-- 品牌列 -->
        <div class="lg:col-span-1">
          <div class="flex items-center gap-2">
            <div class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-md ring-1 ring-gray-200 dark:ring-dark-700">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <span class="font-display text-sm font-semibold italic tracking-tight text-gray-900 dark:text-white">{{ siteName }}</span>
          </div>
          <p class="mt-3 text-xs leading-relaxed text-gray-500 dark:text-dark-400">
            {{ t('home.heroLead') }}
          </p>
          <p class="mt-4 text-2xs text-gray-400 dark:text-dark-500">
            &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
          </p>
        </div>

        <!-- 产品链接 -->
        <div>
          <h4 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.nav.product') }}</h4>
          <ul class="mt-3 space-y-2.5 text-sm">
            <li><a href="#features" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.steps.title') }}</a></li>
            <li v-if="hasPricing"><a href="#pricing" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.pricing') }}</a></li>
            <li><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.docs') }}</a></li>
          </ul>
        </div>

        <!-- 支持链接 -->
        <div>
          <h4 class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('home.nav.support') }}</h4>
          <ul class="mt-3 space-y-2.5 text-sm">
            <li><router-link to="/login" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.login') }}</router-link></li>
            <li><router-link to="/login" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.register') }}</router-link></li>
            <li><a href="#cta" class="text-gray-500 transition-colors hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">{{ t('home.nav.support') }}</a></li>
          </ul>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

defineProps<{
  siteName: string
  siteLogo: string
  docUrl: string
  /** 后台是否配置了套餐（决定是否渲染价格锚点） */
  hasPricing: boolean
}>()

const { t } = useI18n()

const currentYear = computed(() => new Date().getFullYear())
</script>

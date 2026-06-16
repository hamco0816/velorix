<template>
  <!-- 营销首页吸顶导航：品牌 + 锚点菜单 + 主题/语言切换 + 登录注册入口 -->
  <header
    class="sticky top-0 z-30 border-b border-gray-200/60 bg-white/85 backdrop-blur-md transition-shadow dark:border-dark-800/60 dark:bg-dark-950/85"
    :class="{ 'shadow-card': scrolled }"
  >
    <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
      <!-- 品牌 -->
      <div class="flex items-center gap-2.5">
        <div class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-lg ring-1 ring-gray-200 dark:ring-dark-700">
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <span class="font-display text-base font-semibold italic tracking-tight text-gray-900 dark:text-white">
          {{ siteName }}
        </span>
      </div>

      <!-- 中部锚点菜单（桌面端） -->
      <div class="hidden items-center gap-7 md:flex">
        <a href="#features" class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
          {{ t('home.nav.product') }}
        </a>
        <a href="#pricing" class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
          {{ t('home.nav.pricing') }}
        </a>
        <router-link to="/download" class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
          {{ t('home.nav.download') }}
        </router-link>
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
        >
          {{ t('home.nav.docs') }}
        </a>
        <a href="#providers" class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
          {{ t('home.nav.status') }}
        </a>
        <a href="#cta" class="text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white">
          {{ t('home.nav.support') }}
        </a>
      </div>

      <!-- 右侧操作区 -->
      <div class="flex items-center gap-1 sm:gap-2">
        <LocaleSwitcher />
        <button
          @click="toggleTheme"
          class="rounded-md p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
        >
          <Icon v-if="isDark" name="sun" size="sm" />
          <Icon v-else name="moon" size="sm" />
        </button>
        <router-link
          v-if="isAuthenticated"
          :to="dashboardPath"
          class="btn btn-primary btn-sm ml-1"
        >
          {{ t('home.dashboard') }}
          <Icon name="arrowRight" size="xs" :stroke-width="2" />
        </router-link>
        <template v-else>
          <router-link to="/login" class="btn btn-ghost btn-sm hidden sm:inline-flex">
            {{ t('home.login') }}
          </router-link>
          <router-link to="/login" class="btn btn-primary btn-sm ml-1">
            {{ t('home.nav.register') }}
          </router-link>
        </template>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { useTheme } from '@/composables/useTheme'

defineProps<{
  siteName: string
  siteLogo: string
  docUrl: string
  isAuthenticated: boolean
  dashboardPath: string
}>()

const { t } = useI18n()
const { isDark, toggleTheme } = useTheme()

// 滚动后给吸顶导航加投影，与页面内容形成层级
const scrolled = ref(false)

function onScroll(): void {
  scrolled.value = window.scrollY > 4
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

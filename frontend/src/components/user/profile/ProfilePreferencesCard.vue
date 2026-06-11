<!-- 个人资料页「偏好设置」卡片：切换界面语言与明暗主题，设置只保存在当前浏览器 -->
<template>
  <section
    data-testid="profile-preferences-card"
    class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
  >
    <div class="flex items-center gap-3 border-b border-gray-100 bg-gray-50/60 px-5 py-4 dark:border-dark-700/60 dark:bg-dark-800/30">
      <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-gray-50 text-gray-600 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-200 dark:ring-dark-700/60">
        <Icon name="cog" size="sm" />
      </div>
      <div class="min-w-0">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
          {{ t('profile.preferences.title') }}
        </h3>
        <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
          {{ t('profile.preferences.description') }}
        </p>
      </div>
    </div>

    <div class="divide-y divide-gray-100 dark:divide-dark-700/60">
      <!-- 界面语言：中 / 英分段切换 -->
      <div class="flex flex-wrap items-center justify-between gap-3 px-5 py-4">
        <div class="flex min-w-0 items-center gap-2.5">
          <Icon name="globe" size="sm" class="shrink-0 text-gray-400 dark:text-dark-400" />
          <span class="text-sm text-gray-700 dark:text-gray-300">
            {{ t('profile.preferences.language') }}
          </span>
        </div>
        <div
          class="inline-flex shrink-0 rounded-lg bg-gray-100 p-0.5 dark:bg-dark-800"
          role="group"
          :aria-label="t('profile.preferences.language')"
        >
          <button
            v-for="option in availableLocales"
            :key="option.code"
            type="button"
            class="rounded-md px-3 py-1.5 text-xs font-medium transition-colors duration-150 motion-reduce:transition-none"
            :class="currentLocale === option.code ? segmentActiveClass : segmentInactiveClass"
            :aria-pressed="currentLocale === option.code"
            :disabled="localeSwitching"
            @click="handleLocaleChange(option.code)"
          >
            {{ option.name }}
          </button>
        </div>
      </div>

      <!-- 外观主题：浅色 / 深色分段切换 -->
      <div class="flex flex-wrap items-center justify-between gap-3 px-5 py-4">
        <div class="flex min-w-0 items-center gap-2.5">
          <Icon :name="isDark ? 'moon' : 'sun'" size="sm" class="shrink-0 text-gray-400 dark:text-dark-400" />
          <span class="text-sm text-gray-700 dark:text-gray-300">
            {{ t('profile.preferences.theme') }}
          </span>
        </div>
        <div
          class="inline-flex shrink-0 rounded-lg bg-gray-100 p-0.5 dark:bg-dark-800"
          role="group"
          :aria-label="t('profile.preferences.theme')"
        >
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors duration-150 motion-reduce:transition-none"
            :class="!isDark ? segmentActiveClass : segmentInactiveClass"
            :aria-pressed="!isDark"
            @click="applyTheme(false)"
          >
            <Icon name="sun" size="xs" />
            {{ t('nav.lightMode') }}
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors duration-150 motion-reduce:transition-none"
            :class="isDark ? segmentActiveClass : segmentInactiveClass"
            :aria-pressed="isDark"
            @click="applyTheme(true)"
          >
            <Icon name="moon" size="xs" />
            {{ t('nav.darkMode') }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { availableLocales, getLocale, setLocale } from '@/i18n'
import { useTheme } from '@/composables/useTheme'

const { t } = useI18n()
const { isDark, applyTheme } = useTheme()

// 分段切换器激活态 = 中性 zinc 黑底反白（与全站分段控件语言一致）
const segmentActiveClass = 'bg-gray-900 text-white shadow-sm dark:bg-white dark:text-gray-900'
const segmentInactiveClass = 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'

const currentLocale = ref(getLocale())
const localeSwitching = ref(false)

// 切换界面语言并持久化到 localStorage（setLocale 内部完成懒加载语言包与标题刷新）
async function handleLocaleChange(code: string) {
  if (localeSwitching.value || code === currentLocale.value) {
    return
  }
  localeSwitching.value = true
  try {
    await setLocale(code)
    currentLocale.value = getLocale()
  } finally {
    localeSwitching.value = false
  }
}
</script>

<template>
  <!-- 登录/注册分屏外壳：左品牌叙事面板（lg+）+ 右表单列；窄屏退化为单列 -->
  <div class="flex min-h-screen bg-white dark:bg-dark-950">
    <AuthBrandPanel />

    <!-- 右侧：表单列 + 底栏 -->
    <div class="flex min-w-0 flex-1 flex-col">
      <main class="flex flex-1 flex-col items-center justify-center px-5 py-10 sm:py-14">
        <div class="w-full max-w-[400px]">
          <!-- 移动端品牌 logo（lg+ 由左侧品牌面板承载） -->
          <router-link to="/" class="mb-7 inline-block lg:hidden">
            <div class="h-12 w-12 overflow-hidden rounded-xl ring-1 ring-gray-200 dark:ring-dark-700">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
          </router-link>

          <!-- 页面动作标题 + 副标题 -->
          <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
            {{ title }}
          </h1>
          <p v-if="subtitle" class="mt-1.5 text-sm text-gray-500 dark:text-dark-400">
            {{ subtitle }}
          </p>

          <div class="mt-8">
            <slot />
          </div>

          <!-- footer 插槽（注册/已注册之类的尾部信息） -->
          <div v-if="$slots.footer" class="mt-7 text-center text-sm">
            <slot name="footer" />
          </div>
        </div>
      </main>

      <!-- 底栏：API 文档 + 版权 + 语言切换器（贴边、低调） -->
      <footer class="flex flex-col items-center gap-2 px-4 pb-5 text-xs text-gray-400 dark:text-dark-500 sm:flex-row sm:justify-between sm:px-6 sm:pb-5">
        <div class="order-2 flex items-center gap-3 sm:order-1">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="transition-colors hover:text-gray-700 dark:hover:text-dark-300"
          >
            {{ t('auth.apiDocs') }}
          </a>
          <span>&copy; {{ currentYear }} {{ siteName }}</span>
        </div>
        <div class="order-1 sm:order-2">
          <LocaleSwitcher />
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import AuthBrandPanel from '@/components/auth/AuthBrandPanel.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'

defineProps<{
  /** 表单列动作标题（如"登录"/"创建账户"） */
  title: string
  /** 标题下的一句话副标题（可选） */
  subtitle?: string
}>()

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => appStore.docUrl || '')
const currentYear = computed(() => new Date().getFullYear())

// 进入页面即拉取站点公开设置（站名/logo/文档地址等品牌信息）
onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

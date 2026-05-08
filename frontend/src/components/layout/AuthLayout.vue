<template>
  <!-- 极简单列居中布局：logo + 表单 + 注册提示 整体作为单一视觉中心 -->
  <div class="flex min-h-screen flex-col bg-white dark:bg-dark-950">
    <!-- 中部：logo + 表单插槽 整体居中（最大宽 380，移动端自适应） -->
    <main class="flex flex-1 flex-col items-center justify-center px-5 py-10 sm:py-14">
      <div class="w-full max-w-[380px]">
        <!-- 站点品牌（紧贴表单上方，视觉与表单连成一体） -->
        <div class="mb-12 flex justify-center">
          <div class="flex items-center gap-2.5">
            <div class="h-8 w-8 overflow-hidden rounded-md ring-1 ring-gray-200 dark:ring-dark-700">
              <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
            </div>
            <span class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">{{ siteName }}</span>
          </div>
        </div>

        <slot />

        <!-- footer 插槽（注册/已注册之类的尾部信息） -->
        <div v-if="$slots.footer" class="mt-7 text-center text-sm">
          <slot name="footer" />
        </div>
      </div>
    </main>

    <!-- 底栏：API 文档 + 版权 + 语言切换器（贴边、低调） -->
    <footer class="flex flex-col items-center gap-2 px-4 pb-5 text-xs text-gray-400 dark:text-dark-500 sm:flex-row sm:justify-between sm:px-6 sm:pb-5">
      <div class="flex items-center gap-3 order-2 sm:order-1">
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
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'

const appStore = useAppStore()
const { t } = useI18n()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => appStore.docUrl || '')
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

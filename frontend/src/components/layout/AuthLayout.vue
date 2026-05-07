<template>
  <div class="auth-shell relative flex min-h-screen items-center justify-center overflow-hidden p-4">
    <!-- Background -->
    <div
      class="absolute inset-0 bg-slate-50 dark:bg-dark-950"
    ></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Grid Pattern -->
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.045)_1px,transparent_1px)] bg-[size:64px_64px] dark:bg-[linear-gradient(rgba(148,163,184,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.05)_1px,transparent_1px)]"
      ></div>
      <!-- 顶部暖色微光：用 brand 系，给登录页一点品牌温度，避免冷灰沉闷 -->
      <div class="absolute inset-x-0 top-0 h-40 bg-gradient-to-b from-brand-100/40 to-transparent dark:from-brand-500/8"></div>
      <div class="absolute inset-x-0 bottom-0 h-48 bg-gradient-to-t from-white to-transparent dark:from-dark-950"></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-[440px]">
      <!-- Logo/Brand -->
      <div class="mb-7 flex justify-center">
        <div
          class="inline-flex h-20 w-20 items-center justify-center overflow-hidden rounded-lg bg-white p-1 shadow-card ring-1 ring-slate-200 dark:bg-dark-900 dark:ring-dark-700"
        >
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </div>
      </div>

      <!-- Card Container：扁平卡片节奏，与主应用 .card 一致 -->
      <div class="rounded-lg border border-gray-200 bg-white p-7 shadow-card-hover dark:border-dark-700 dark:bg-dark-900">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 text-center text-xs text-gray-400 dark:text-dark-500">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped></style>

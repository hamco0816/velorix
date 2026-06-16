<template>
  <div class="relative min-h-screen overflow-hidden bg-white text-gray-900 dark:bg-dark-950 dark:text-gray-100">
    <!-- 顶栏：品牌 + 返回首页 + 登录 -->
    <header class="sticky top-0 z-30 border-b border-gray-200/60 bg-white/85 backdrop-blur-md dark:border-dark-800/60 dark:bg-dark-950/85">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <router-link to="/" class="flex items-center gap-2.5">
          <div class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-lg ring-1 ring-gray-200 dark:ring-dark-700">
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="font-display text-base font-semibold italic tracking-tight text-gray-900 dark:text-white">{{ siteName }}</span>
        </router-link>
        <div class="flex items-center gap-2">
          <router-link to="/" class="btn btn-ghost btn-sm hidden sm:inline-flex">{{ t('download.backHome') }}</router-link>
          <router-link to="/login" class="btn btn-primary btn-sm">{{ t('download.login') }}</router-link>
        </div>
      </nav>
    </header>

    <!-- ============== Hero ============== -->
    <section class="relative overflow-hidden">
      <div class="pointer-events-none absolute inset-0">
        <div class="absolute inset-x-0 top-0 h-[560px] bg-gradient-to-b from-brand-50/70 via-brand-100/30 to-transparent dark:from-brand-500/[0.04] dark:via-brand-500/[0.02]"></div>
        <div class="absolute inset-0 bg-[linear-gradient(theme(colors.primary.500/0.05)_1px,transparent_1px),linear-gradient(90deg,theme(colors.primary.500/0.05)_1px,transparent_1px)] bg-[size:64px_64px] [mask-image:radial-gradient(ellipse_at_top,black_5%,transparent_70%)]"></div>
        <div class="absolute -right-32 top-24 h-[440px] w-[440px] rounded-full bg-brand-300/25 blur-[100px] dark:bg-brand-500/10"></div>
      </div>

      <div class="relative mx-auto max-w-7xl px-4 pb-16 pt-12 sm:px-6 sm:pt-16 lg:px-8 lg:pt-20">
        <div class="grid items-center gap-10 lg:grid-cols-12 lg:gap-12">
          <!-- 左：文案 + 下载 -->
          <div class="lg:col-span-7">
            <span class="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-3 py-1 text-xs font-medium text-brand-700 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30">
              <span class="h-1.5 w-1.5 rounded-full bg-brand-500"></span>
              {{ t('download.badge') }}
            </span>
            <h1 class="mt-5 text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl dark:text-white">
              {{ t('download.heroTitle', { name: siteName }) }}
            </h1>
            <p class="mt-5 max-w-xl text-base leading-relaxed text-gray-600 sm:text-lg dark:text-dark-300">
              {{ t('download.heroSubtitle') }}
            </p>

            <div class="mt-8 flex flex-wrap items-center gap-4">
              <a
                v-if="latest?.available && downloadHref"
                :href="downloadHref"
                class="btn btn-primary inline-flex h-12 items-center gap-2 px-6 text-base"
                download
              >
                <Icon name="download" size="md" :stroke-width="2" />
                {{ t('download.downloadWin') }}
              </a>
              <button v-else class="btn btn-primary inline-flex h-12 items-center gap-2 px-6 text-base opacity-60" disabled>
                <Icon name="download" size="md" :stroke-width="2" />
                {{ loading ? t('download.loading') : t('download.comingSoon') }}
              </button>

              <div v-if="latest?.available" class="text-sm text-gray-500 dark:text-dark-400">
                <span class="font-medium text-gray-700 dark:text-dark-200">v{{ latest.version }}</span>
                <span class="mx-1.5 text-gray-300 dark:text-dark-700">·</span>
                <span>{{ formatFileSize(latest.file_size) }}</span>
                <span class="mx-1.5 text-gray-300 dark:text-dark-700">·</span>
                <span>Windows 10/11 (x64)</span>
              </div>
            </div>

            <p class="mt-4 text-xs leading-relaxed text-gray-400 dark:text-dark-500">
              {{ t('download.smartScreenHint') }}
            </p>
          </div>

          <!-- 右：品牌图 + 当前版本卡 -->
          <div class="lg:col-span-5">
            <div class="rounded-2xl border border-gray-200/70 bg-white p-6 shadow-card dark:border-dark-700/60 dark:bg-dark-800/40">
              <div class="flex items-center gap-4">
                <img :src="siteLogo" alt="Logo" class="h-14 w-14 rounded-xl object-contain ring-1 ring-gray-200 dark:ring-dark-700" />
                <div>
                  <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ siteName }}</p>
                  <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('download.cardSubtitle') }}</p>
                </div>
              </div>

              <div v-if="latest?.available" class="mt-5 space-y-3 border-t border-gray-100 pt-5 dark:border-dark-700/60">
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-500 dark:text-dark-400">{{ t('download.versionLabel') }}</span>
                  <span class="font-medium tabular-nums text-gray-900 dark:text-white">v{{ latest.version }}</span>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-500 dark:text-dark-400">{{ t('download.sizeLabel') }}</span>
                  <span class="font-medium tabular-nums text-gray-900 dark:text-white">{{ formatFileSize(latest.file_size) }}</span>
                </div>
                <div v-if="latest.released_at" class="flex items-center justify-between text-sm">
                  <span class="text-gray-500 dark:text-dark-400">{{ t('download.dateLabel') }}</span>
                  <span class="font-medium tabular-nums text-gray-900 dark:text-white">{{ formatDate(latest.released_at) }}</span>
                </div>
                <!-- 更新说明 -->
                <div v-if="latest.notes" class="rounded-lg bg-gray-50/70 p-3 dark:bg-dark-800/60">
                  <p class="mb-1 text-xs font-semibold text-gray-500 dark:text-dark-400">{{ t('download.notesLabel') }}</p>
                  <p class="whitespace-pre-wrap text-sm leading-relaxed text-gray-700 dark:text-dark-200">{{ latest.notes }}</p>
                </div>
              </div>
              <p v-else class="mt-5 border-t border-gray-100 pt-5 text-sm text-gray-500 dark:border-dark-700/60 dark:text-dark-400">
                {{ t('download.comingSoonDesc') }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ============== 功能亮点 ============== -->
    <section class="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
      <div class="mx-auto max-w-2xl text-center">
        <h2 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl dark:text-white">{{ t('download.featuresTitle') }}</h2>
        <p class="mt-3 text-base text-gray-600 dark:text-dark-300">{{ t('download.featuresSubtitle') }}</p>
      </div>

      <div class="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="f in features"
          :key="f.key"
          class="rounded-xl border border-gray-200/70 bg-white p-5 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-card dark:border-dark-700/60 dark:bg-dark-800/40"
        >
          <div :class="['flex h-11 w-11 items-center justify-center rounded-xl', f.iconBg]">
            <Icon :name="f.icon" size="md" :stroke-width="2" :class="f.iconColor" />
          </div>
          <h3 class="mt-4 text-base font-semibold text-gray-900 dark:text-white">{{ t(`download.features.${f.key}.title`) }}</h3>
          <p class="mt-1.5 text-sm leading-relaxed text-gray-600 dark:text-dark-300">{{ t(`download.features.${f.key}.desc`) }}</p>
        </div>
      </div>
    </section>

    <!-- ============== 系统要求 ============== -->
    <section class="border-t border-gray-100 bg-gray-50/50 dark:border-dark-800/60 dark:bg-dark-900/30">
      <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div class="grid gap-6 sm:grid-cols-3">
          <div v-for="r in requirements" :key="r.key" class="flex items-start gap-3">
            <Icon :name="r.icon" size="md" class="mt-0.5 shrink-0 text-brand-600 dark:text-brand-400" :stroke-width="2" />
            <div>
              <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t(`download.requirements.${r.key}.title`) }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-dark-300">{{ t(`download.requirements.${r.key}.desc`) }}</p>
            </div>
          </div>
        </div>

        <!-- 底部再来一次下载 CTA -->
        <div class="mt-10 flex flex-col items-center justify-between gap-4 rounded-2xl border border-gray-200/70 bg-white px-6 py-6 sm:flex-row dark:border-dark-700/60 dark:bg-dark-800/40">
          <div>
            <p class="text-base font-semibold text-gray-900 dark:text-white">{{ t('download.ctaTitle') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-dark-300">{{ t('download.ctaSubtitle') }}</p>
          </div>
          <a
            v-if="latest?.available && downloadHref"
            :href="downloadHref"
            class="btn btn-primary inline-flex h-11 shrink-0 items-center gap-2 px-6"
            download
          >
            <Icon name="download" size="sm" :stroke-width="2" />
            {{ t('download.downloadWin') }}
          </a>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import desktopAPI, { type DesktopLatest } from '@/api/desktop'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const latest = ref<DesktopLatest | null>(null)
const loading = ref(true)

const siteName = computed(() => (appStore.cachedPublicSettings as any)?.site_name || 'Velorix')
const siteLogo = computed(() => (appStore.cachedPublicSettings as any)?.site_logo || '/logo.png')

// setup_url 是相对路径（/desktop/updates/xxx.exe），同源直接用
const downloadHref = computed(() => latest.value?.setup_url || '')

const features = [
  { key: 'cli', icon: 'terminal', iconBg: 'bg-brand-50 dark:bg-brand-500/15', iconColor: 'text-brand-600 dark:text-brand-400' },
  { key: 'workspace', icon: 'grid', iconBg: 'bg-info-soft dark:bg-info/15', iconColor: 'text-info dark:text-info' },
  { key: 'chat', icon: 'chat', iconBg: 'bg-success-soft dark:bg-success/15', iconColor: 'text-success dark:text-success' },
  { key: 'image', icon: 'sparkles', iconBg: 'bg-tea-50 dark:bg-tea-500/15', iconColor: 'text-tea-600 dark:text-tea-400' },
  { key: 'billing', icon: 'creditCard', iconBg: 'bg-warning-soft dark:bg-warning/15', iconColor: 'text-warning dark:text-warning' },
  { key: 'update', icon: 'refresh', iconBg: 'bg-info-soft dark:bg-info/15', iconColor: 'text-info dark:text-info' },
] as const

const requirements = [
  { key: 'os', icon: 'cube' },
  { key: 'network', icon: 'globe' },
  { key: 'account', icon: 'key' },
] as const

function formatFileSize(bytes?: number): string {
  if (!bytes) return '—'
  const mb = bytes / (1024 * 1024)
  return mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
}

function formatDate(value?: string): string {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

onMounted(async () => {
  try {
    latest.value = await desktopAPI.getLatest('stable')
  } catch {
    latest.value = { available: false }
  } finally {
    loading.value = false
  }
})
</script>

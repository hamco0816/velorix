<template>
  <!-- 登录/注册左侧品牌叙事面板：深色底 + 品牌微光，承载价值主张与信任点（仅 lg+ 显示） -->
  <aside class="relative hidden overflow-hidden bg-gray-950 lg:flex lg:w-[440px] lg:flex-shrink-0 lg:flex-col xl:w-[500px]">
    <!-- 装饰：细网格 + 品牌暖光（与首页 CTA 区同语言） -->
    <div class="pointer-events-none absolute inset-0">
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] bg-[size:44px_44px] [mask-image:radial-gradient(ellipse_at_top_left,black_10%,transparent_70%)]"
      ></div>
      <div class="absolute -left-28 -top-28 h-[380px] w-[380px] rounded-full bg-brand-500/20 blur-[100px]"></div>
      <div class="absolute -bottom-32 -right-10 h-[320px] w-[320px] rounded-full bg-brand-400/10 blur-[90px]"></div>
    </div>

    <div class="relative flex flex-1 flex-col p-10 xl:p-12">
      <!-- 顶部：站点品牌（点击回首页） -->
      <router-link to="/" class="inline-flex items-center gap-2.5 self-start">
        <div class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-lg bg-white/10 ring-1 ring-white/15">
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <span class="font-display text-base font-semibold italic tracking-tight text-white">{{ siteName }}</span>
      </router-link>

      <!-- 中部：价值主张 + 信任点 -->
      <div class="flex flex-1 flex-col justify-center py-12">
        <h2 class="text-[2rem] font-semibold leading-tight tracking-tight text-white [text-wrap:balance] xl:text-4xl">
          {{ t('auth.brand.headlineMain') }}<br />
          <span class="text-brand-400">{{ t('auth.brand.headlineSub') }}</span>
        </h2>
        <p class="mt-4 max-w-sm text-sm leading-relaxed text-gray-400">
          {{ t('auth.brand.description') }}
        </p>

        <ul class="mt-10 space-y-5">
          <li v-for="point in trustPoints" :key="point.key" class="flex items-start gap-3.5">
            <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-white/[0.06] ring-1 ring-inset ring-white/10">
              <Icon :name="point.icon" size="sm" class="text-brand-400" :stroke-width="2" />
            </div>
            <div class="min-w-0">
              <div class="text-sm font-semibold tracking-tight text-white">{{ point.title }}</div>
              <div class="mt-0.5 text-xs text-gray-400">{{ point.desc }}</div>
            </div>
          </li>
        </ul>
      </div>

      <!-- 底部：服务质量数据 -->
      <div class="flex items-center gap-8 border-t border-white/10 pt-6">
        <div v-for="stat in brandStats" :key="stat.key">
          <div class="text-lg font-semibold tabular-nums tracking-tight text-white">{{ stat.value }}</div>
          <div class="mt-0.5 text-2xs text-gray-400">{{ stat.label }}</div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))

type TrustIcon = 'cube' | 'shield' | 'dollar'

// 三个信任点：多模型聚合 / 稳定可靠 / 按量计费
const trustPoints = computed<{ key: string; icon: TrustIcon; title: string; desc: string }[]>(() => [
  { key: 'unified', icon: 'cube', title: t('auth.brand.features.unifiedTitle'), desc: t('auth.brand.features.unifiedDesc') },
  { key: 'stable', icon: 'shield', title: t('auth.brand.features.stableTitle'), desc: t('auth.brand.features.stableDesc') },
  { key: 'pay', icon: 'dollar', title: t('auth.brand.features.payTitle'), desc: t('auth.brand.features.payDesc') }
])

// 底部服务质量数据（静态展示）
const brandStats = computed(() => [
  { key: 'uptime', value: t('auth.brand.stats.uptime'), label: t('auth.brand.stats.uptimeLabel') },
  { key: 'requests', value: t('auth.brand.stats.requests'), label: t('auth.brand.stats.requestsLabel') },
  { key: 'models', value: t('auth.brand.stats.models'), label: t('auth.brand.stats.modelsLabel') }
])
</script>

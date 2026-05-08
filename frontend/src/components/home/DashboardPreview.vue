<template>
  <!-- Dashboard Preview Card：Hero 右侧的"产品截图"，静态展示，模拟真实控制台 -->
  <div class="dash-preview relative">
    <!-- 暖色光晕装饰（卡片下方 + 右侧） -->
    <div class="pointer-events-none absolute -inset-6 -z-10 rounded-3xl bg-gradient-to-br from-brand-500/15 via-amber-200/15 to-transparent blur-2xl dark:from-brand-500/15"></div>

    <!-- 主卡片 -->
    <div class="relative overflow-hidden rounded-xl border border-gray-200 bg-white shadow-card-hover dark:border-dark-700 dark:bg-dark-900">
      <!-- 顶部窗口栏 -->
      <div class="flex items-center gap-2 border-b border-gray-100 bg-gray-50/60 px-4 py-2.5 dark:border-dark-700 dark:bg-dark-900/40">
        <div class="flex items-center gap-1.5">
          <span class="h-2.5 w-2.5 rounded-full bg-red-400"></span>
          <span class="h-2.5 w-2.5 rounded-full bg-yellow-400"></span>
          <span class="h-2.5 w-2.5 rounded-full bg-emerald-400"></span>
        </div>
        <div class="ml-2 text-[11px] font-medium text-gray-500 dark:text-dark-400">
          {{ t('home.dashPreview.title') }}
        </div>
      </div>

      <!-- 主体：左 sidebar + 右 main -->
      <div class="flex">
        <!-- Sidebar -->
        <aside class="w-32 flex-shrink-0 border-r border-gray-100 bg-gray-50/40 px-2 py-3 dark:border-dark-700 dark:bg-dark-900/30 sm:w-36">
          <nav class="space-y-0.5">
            <button
              v-for="(item, i) in navItems"
              :key="item.key"
              class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-left text-[11px] font-medium transition-colors"
              :class="i === 0
                ? 'bg-brand-50 text-brand-700 dark:bg-brand-500/10 dark:text-brand-300'
                : 'text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800'"
            >
              <component :is="item.iconSvg" class="h-3.5 w-3.5 flex-shrink-0" />
              <span class="truncate">{{ item.label }}</span>
            </button>
          </nav>
        </aside>

        <!-- Main content -->
        <div class="flex-1 p-4 sm:p-5">
          <!-- Top row: Model Status + 3 metric cards -->
          <div class="grid gap-3 sm:grid-cols-5 sm:gap-3">
            <!-- Model status list (3/5) -->
            <div class="rounded-lg border border-gray-100 bg-gray-50/50 p-3 dark:border-dark-700 dark:bg-dark-800/40 sm:col-span-3">
              <div class="text-[11px] font-semibold text-gray-700 dark:text-dark-200">
                {{ t('home.dashPreview.modelStatus') }}
              </div>
              <ul class="mt-2.5 space-y-2">
                <li
                  v-for="m in models"
                  :key="m.name"
                  class="flex items-center gap-2 rounded-md bg-white px-2.5 py-1.5 dark:bg-dark-900/60"
                >
                  <div :class="['flex h-5 w-5 flex-shrink-0 items-center justify-center rounded', m.iconBg]">
                    <BrandIcon v-if="m.brand" :brand="m.brand" size="12px" />
                    <span v-else :class="['text-[9px] font-bold', m.textClass]">{{ m.letter }}</span>
                  </div>
                  <span class="flex-1 truncate text-[11px] font-medium text-gray-700 dark:text-dark-200">{{ m.name }}</span>
                  <span class="inline-flex items-center gap-1 text-[10px] font-medium text-emerald-600 dark:text-emerald-400">
                    <span class="h-1 w-1 rounded-full bg-emerald-500"></span>
                    {{ t('home.dashPreview.stateAvailable') }}
                  </span>
                  <span class="font-mono text-[10px] tabular-nums text-gray-500 dark:text-dark-400">{{ m.value }}</span>
                </li>
              </ul>
            </div>

            <!-- Right metrics column -->
            <div class="space-y-2 sm:col-span-2">
              <!-- Today usage -->
              <div class="rounded-lg border border-gray-100 bg-white p-3 dark:border-dark-700 dark:bg-dark-800/40">
                <div class="text-[10px] font-medium text-gray-500 dark:text-dark-400">
                  {{ t('home.dashPreview.todayUsage') }}
                </div>
                <div class="mt-1 text-lg font-bold tabular-nums tracking-tight text-gray-900 dark:text-white">
                  ¥ 12.86
                </div>
                <div class="mt-0.5 inline-flex items-center gap-0.5 text-[10px] font-medium text-emerald-600 dark:text-emerald-400">
                  <svg class="h-2.5 w-2.5" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M5 15l7-7 7 7" stroke="currentColor" stroke-width="3" fill="none" stroke-linecap="round" />
                  </svg>
                  {{ t('home.dashPreview.change') }} +8.5%
                </div>
              </div>
              <!-- Two small stats -->
              <div class="grid grid-cols-2 gap-2">
                <div class="rounded-lg border border-gray-100 bg-white p-2.5 dark:border-dark-700 dark:bg-dark-800/40">
                  <div class="text-[10px] font-medium text-gray-500 dark:text-dark-400">
                    {{ t('home.dashPreview.totalRequests') }}
                  </div>
                  <div class="mt-0.5 text-sm font-bold tabular-nums tracking-tight text-gray-900 dark:text-white">
                    2,128
                  </div>
                </div>
                <div class="rounded-lg border border-gray-100 bg-white p-2.5 dark:border-dark-700 dark:bg-dark-800/40">
                  <div class="text-[10px] font-medium text-gray-500 dark:text-dark-400">
                    {{ t('home.dashPreview.successRate') }}
                  </div>
                  <div class="mt-0.5 text-sm font-bold tabular-nums tracking-tight text-emerald-600 dark:text-emerald-400">
                    99.8%
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- API Key + Trend Chart -->
          <div class="mt-3 grid gap-3 sm:grid-cols-5">
            <!-- API Key input -->
            <div class="rounded-lg border border-gray-100 bg-gray-50/40 p-3 dark:border-dark-700 dark:bg-dark-800/40 sm:col-span-2">
              <div class="text-[10px] font-medium text-gray-500 dark:text-dark-400">
                {{ t('home.dashPreview.apiKey') }}
              </div>
              <div class="mt-1.5 flex items-center gap-1.5 rounded-md bg-white px-2 py-1.5 ring-1 ring-gray-200 dark:bg-dark-900 dark:ring-dark-700">
                <span class="font-mono text-[10px] tabular-nums text-gray-700 dark:text-dark-200">sk-***********</span>
                <span class="ml-auto text-gray-400 dark:text-dark-500">
                  <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178zM15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </span>
              </div>
            </div>

            <!-- Trend chart -->
            <div class="rounded-lg border border-gray-100 bg-white p-3 dark:border-dark-700 dark:bg-dark-800/40 sm:col-span-3">
              <div class="text-[10px] font-medium text-gray-500 dark:text-dark-400">
                {{ t('home.dashPreview.recent7d') }}
              </div>
              <!-- 简单 SVG 趋势线 -->
              <svg class="mt-1.5 h-12 w-full" viewBox="0 0 200 40" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="trendFill" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0%" stop-color="rgb(234,88,12)" stop-opacity="0.18" />
                    <stop offset="100%" stop-color="rgb(234,88,12)" stop-opacity="0" />
                  </linearGradient>
                </defs>
                <!-- Fill -->
                <path
                  d="M0,30 L30,26 L60,28 L90,18 L120,22 L150,12 L180,15 L200,8 L200,40 L0,40 Z"
                  fill="url(#trendFill)"
                />
                <!-- Stroke -->
                <path
                  d="M0,30 L30,26 L60,28 L90,18 L120,22 L150,12 L180,15 L200,8"
                  fill="none"
                  stroke="rgb(234,88,12)"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <!-- Last point marker -->
                <circle cx="200" cy="8" r="2.5" fill="rgb(234,88,12)" />
                <circle cx="200" cy="8" r="5" fill="rgb(234,88,12)" fill-opacity="0.2" />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import BrandIcon from '@/components/common/BrandIcon.vue'

const { t } = useI18n()

// Sidebar 图标用内联 SVG（保持小巧无依赖）
const makeIcon = (d: string) =>
  h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '2' }, [
    h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d })
  ])

const navItems = computed(() => [
  { key: 'overview', label: t('home.dashPreview.overview'), iconSvg: makeIcon('M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z') },
  { key: 'tokens', label: t('home.dashPreview.tokenManage'), iconSvg: makeIcon('M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z') },
  { key: 'usage', label: t('home.dashPreview.usageStats'), iconSvg: makeIcon('M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75z') },
  { key: 'billing', label: t('home.dashPreview.billing'), iconSvg: makeIcon('M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z') },
  { key: 'team', label: t('home.dashPreview.team'), iconSvg: makeIcon('M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0z') },
  { key: 'settings', label: t('home.dashPreview.settings'), iconSvg: makeIcon('M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992.022.379.022.762 0 1.141-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94H10.7a1.125 1.125 0 01-1.11-.94l-.213-1.281a1.5 1.5 0 00-.644-.87 6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49L3.557 14.7a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992A6.93 6.93 0 015.25 12c0-.193.005-.385.015-.575.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43L4.871 5.93a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281zM15 12a3 3 0 11-6 0 3 3 0 016 0z') }
])

interface ModelRow {
  name: string
  value: string
  brand?: 'claude' | 'openai' | 'gemini'
  letter?: string
  iconBg: string
  textClass?: string
}

const models = computed<ModelRow[]>(() => [
  { name: 'Claude Opus 4.7', value: '99.9%', brand: 'claude', iconBg: 'bg-orange-50 dark:bg-orange-500/15' },
  { name: 'ChatGPT-5.5', value: '99.0%', brand: 'openai', iconBg: 'bg-gray-100 dark:bg-white/15' },
  { name: 'Gemini 3.1 Pro', value: '99.6%', brand: 'gemini', iconBg: 'bg-blue-50 dark:bg-blue-500/15' },
  { name: 'DeepSeek V3.2', value: '99.7%', letter: 'D', iconBg: 'bg-violet-100 dark:bg-violet-500/20', textClass: 'text-violet-700 dark:text-violet-300' }
])
</script>

<style scoped>
.dash-preview {
  /* 给整个卡片轻微透视感（参考图风格更稳重，所以不大幅倾斜） */
  transition: transform 0.4s ease;
}
.dash-preview:hover {
  transform: translateY(-2px);
}
</style>

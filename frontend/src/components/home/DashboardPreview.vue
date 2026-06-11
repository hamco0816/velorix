<template>
  <!-- Dashboard 预览卡：Hero 右侧的"产品截图"，静态模拟真实控制台 -->
  <div class="dash-preview relative">
    <!-- 暖色光晕装饰（卡片下方 + 右侧） -->
    <div class="pointer-events-none absolute -inset-6 -z-10 rounded-3xl bg-gradient-to-br from-brand-500/15 via-amber-200/15 to-transparent blur-2xl dark:from-brand-500/15"></div>

    <!-- 主卡片 -->
    <div class="relative overflow-hidden rounded-xl border border-gray-200 bg-white shadow-card-hover ring-1 ring-gray-950/5 dark:border-dark-700 dark:bg-dark-900 dark:ring-white/5">
      <!-- 顶部窗口栏：红绿灯 + 居中地址栏 -->
      <div class="relative flex items-center border-b border-gray-100 bg-gray-50/60 px-4 py-2.5 dark:border-dark-700 dark:bg-dark-900/40">
        <div class="flex items-center gap-1.5">
          <span class="h-2.5 w-2.5 rounded-full bg-red-400"></span>
          <span class="h-2.5 w-2.5 rounded-full bg-yellow-400"></span>
          <span class="h-2.5 w-2.5 rounded-full bg-emerald-400"></span>
        </div>
        <div class="absolute left-1/2 -translate-x-1/2">
          <div class="flex items-center gap-1.5 rounded-md bg-white px-3 py-1 text-2xs font-medium text-gray-500 ring-1 ring-gray-200 dark:bg-dark-800 dark:text-dark-300 dark:ring-dark-600">
            <Icon name="lock" size="xs" class="h-2.5 w-2.5 text-emerald-500" :stroke-width="2" />
            {{ t('home.dashPreview.title') }}
          </div>
        </div>
      </div>

      <!-- 主体：左 sidebar + 右 main -->
      <div class="flex">
        <!-- Sidebar：激活项黑底反白，与真实控制台同语言 -->
        <aside class="flex w-32 flex-shrink-0 flex-col border-r border-gray-100 bg-gray-50/40 px-2 py-3 dark:border-dark-700 dark:bg-dark-900/30 sm:w-36">
          <nav class="space-y-0.5">
            <button
              v-for="(item, i) in navItems"
              :key="item.key"
              class="flex w-full items-center gap-2 rounded-md px-2.5 py-1.5 text-left text-2xs font-medium transition-colors"
              :class="i === 0
                ? 'bg-gray-900 text-white shadow-sm dark:bg-white dark:text-gray-900'
                : 'text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800'"
            >
              <Icon :name="item.icon" size="xs" class="h-3.5 w-3.5 flex-shrink-0" :stroke-width="2" />
              <span class="truncate">{{ item.label }}</span>
            </button>
          </nav>
          <!-- 底部模拟账号行 -->
          <div class="mt-auto flex items-center gap-2 rounded-md px-2 pt-3">
            <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-brand-100 text-2xs font-semibold text-brand-700 dark:bg-brand-500/20 dark:text-brand-300">D</span>
            <span class="truncate font-mono text-2xs text-gray-400 dark:text-dark-500">dev@***</span>
          </div>
        </aside>

        <!-- 主内容区 -->
        <div class="flex-1 p-4 sm:p-5">
          <!-- 第一行：模型状态 + 指标卡 -->
          <div class="grid gap-3 sm:grid-cols-5 sm:gap-3">
            <!-- 模型状态列表 (3/5) -->
            <div class="rounded-lg border border-gray-100 bg-gray-50/50 p-3 dark:border-dark-700 dark:bg-dark-800/40 sm:col-span-3">
              <div class="text-2xs font-semibold text-gray-700 dark:text-dark-200">
                {{ t('home.dashPreview.modelStatus') }}
              </div>
              <ul class="mt-2.5 space-y-2">
                <li
                  v-for="m in models"
                  :key="m.name"
                  class="flex items-center gap-2 rounded-md bg-white px-2.5 py-1.5 ring-1 ring-gray-100 dark:bg-dark-900/60 dark:ring-dark-700/60"
                >
                  <div :class="['flex h-5 w-5 flex-shrink-0 items-center justify-center rounded', m.iconBg]">
                    <BrandIcon :brand="m.brand" size="12px" />
                  </div>
                  <span class="flex-1 truncate text-2xs font-medium text-gray-700 dark:text-dark-200">{{ m.name }}</span>
                  <span class="inline-flex items-center gap-1.5 text-2xs font-medium text-emerald-600 dark:text-emerald-400">
                    <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                    {{ t('home.dashPreview.stateAvailable') }}
                  </span>
                  <span class="font-mono text-2xs tabular-nums text-gray-500 dark:text-dark-400">{{ m.value }}</span>
                </li>
              </ul>
            </div>

            <!-- 右侧指标列 -->
            <div class="space-y-2 sm:col-span-2">
              <!-- 今日使用 -->
              <div class="rounded-lg border border-gray-100 bg-white p-3 dark:border-dark-700 dark:bg-dark-800/40">
                <div class="text-2xs font-medium text-gray-500 dark:text-dark-400">
                  {{ t('home.dashPreview.todayUsage') }}
                </div>
                <div class="mt-1 text-lg font-semibold tabular-nums tracking-tight text-emerald-600 dark:text-emerald-400">
                  ¥ 12.86
                </div>
                <div class="mt-0.5 inline-flex items-center gap-1 text-2xs font-medium text-emerald-600 dark:text-emerald-400">
                  <Icon name="trendingUp" size="xs" class="h-2.5 w-2.5" :stroke-width="2.5" />
                  {{ t('home.dashPreview.change') }} +8.5%
                </div>
              </div>
              <!-- 两个小统计 -->
              <div class="grid grid-cols-2 gap-2">
                <div class="rounded-lg border border-gray-100 bg-white p-2.5 dark:border-dark-700 dark:bg-dark-800/40">
                  <div class="text-2xs font-medium text-gray-500 dark:text-dark-400">
                    {{ t('home.dashPreview.totalRequests') }}
                  </div>
                  <div class="mt-0.5 text-sm font-semibold tabular-nums tracking-tight text-gray-900 dark:text-white">
                    2,128
                  </div>
                </div>
                <div class="rounded-lg border border-gray-100 bg-white p-2.5 dark:border-dark-700 dark:bg-dark-800/40">
                  <div class="text-2xs font-medium text-gray-500 dark:text-dark-400">
                    {{ t('home.dashPreview.successRate') }}
                  </div>
                  <div class="mt-0.5 text-sm font-semibold tabular-nums tracking-tight text-emerald-600 dark:text-emerald-400">
                    99.8%
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 第二行：API Key + 趋势图 -->
          <div class="mt-3 grid gap-3 sm:grid-cols-5">
            <!-- API Key -->
            <div class="rounded-lg border border-gray-100 bg-gray-50/40 p-3 dark:border-dark-700 dark:bg-dark-800/40 sm:col-span-2">
              <div class="text-2xs font-medium text-gray-500 dark:text-dark-400">
                {{ t('home.dashPreview.apiKey') }}
              </div>
              <div class="mt-1.5 flex items-center gap-1.5 rounded-md bg-white px-2 py-1.5 ring-1 ring-gray-200 dark:bg-dark-900 dark:ring-dark-700">
                <span class="font-mono text-2xs tabular-nums text-gray-700 dark:text-dark-200">sk-***********</span>
                <Icon name="eye" size="xs" class="ml-auto h-3 w-3 text-gray-400 dark:text-dark-500" :stroke-width="2" />
              </div>
            </div>

            <!-- 近 7 日趋势：平滑曲线 + 基线网格 -->
            <div class="rounded-lg border border-gray-100 bg-white p-3 dark:border-dark-700 dark:bg-dark-800/40 sm:col-span-3">
              <div class="text-2xs font-medium text-gray-500 dark:text-dark-400">
                {{ t('home.dashPreview.recent7d') }}
              </div>
              <svg class="mt-1.5 h-12 w-full text-gray-200 dark:text-dark-700" viewBox="0 0 200 40" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="trendFill" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0%" stop-color="rgb(234,88,12)" stop-opacity="0.18" />
                    <stop offset="100%" stop-color="rgb(234,88,12)" stop-opacity="0" />
                  </linearGradient>
                </defs>
                <!-- 基线网格 -->
                <line x1="0" y1="14" x2="200" y2="14" stroke="currentColor" stroke-width="0.5" stroke-dasharray="3 4" />
                <line x1="0" y1="27" x2="200" y2="27" stroke="currentColor" stroke-width="0.5" stroke-dasharray="3 4" />
                <!-- 面积填充 -->
                <path
                  d="M0,31 C14,29.5 22,27 33,27.5 C46,28 54,22 66,19.5 C76,17.5 86,21 100,21.5 C114,22 124,14 138,12.5 C150,11.3 160,15.5 172,14 C184,12.7 192,9.5 200,8 L200,40 L0,40 Z"
                  fill="url(#trendFill)"
                />
                <!-- 曲线 -->
                <path
                  d="M0,31 C14,29.5 22,27 33,27.5 C46,28 54,22 66,19.5 C76,17.5 86,21 100,21.5 C114,22 124,14 138,12.5 C150,11.3 160,15.5 172,14 C184,12.7 192,9.5 200,8"
                  fill="none"
                  stroke="rgb(234,88,12)"
                  stroke-width="1.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <!-- 末端数据点 -->
                <circle cx="200" cy="8" r="2.5" fill="rgb(234,88,12)" />
                <circle cx="200" cy="8" r="5" fill="rgb(234,88,12)" fill-opacity="0.2" />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 悬浮充值回执卡：增加截图的"正在被使用"的真实感 -->
    <div class="absolute -bottom-5 left-4 flex items-center gap-2.5 rounded-xl border border-gray-200/80 bg-white px-3.5 py-2.5 shadow-elevated dark:border-dark-700 dark:bg-dark-900 sm:-left-5">
      <span class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-emerald-50 dark:bg-emerald-500/15">
        <Icon name="checkCircle" size="sm" class="text-emerald-500" :stroke-width="2" />
      </span>
      <div class="leading-tight">
        <div class="text-2xs font-semibold text-gray-900 dark:text-white">{{ t('home.dashPreview.topupToast') }}</div>
        <div class="font-mono text-2xs font-medium tabular-nums text-emerald-600 dark:text-emerald-400">+ ¥100.00</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import BrandIcon from '@/components/common/BrandIcon.vue'

const { t } = useI18n()

type NavIcon = 'grid' | 'key' | 'chart' | 'creditCard' | 'users' | 'cog'

const navItems = computed<{ key: string; label: string; icon: NavIcon }[]>(() => [
  { key: 'overview', label: t('home.dashPreview.overview'), icon: 'grid' },
  { key: 'tokens', label: t('home.dashPreview.tokenManage'), icon: 'key' },
  { key: 'usage', label: t('home.dashPreview.usageStats'), icon: 'chart' },
  { key: 'billing', label: t('home.dashPreview.billing'), icon: 'creditCard' },
  { key: 'team', label: t('home.dashPreview.team'), icon: 'users' },
  { key: 'settings', label: t('home.dashPreview.settings'), icon: 'cog' }
])

interface ModelRow {
  name: string
  value: string
  brand: 'claude' | 'openai' | 'gemini' | 'deepseek'
  iconBg: string
}

const models = computed<ModelRow[]>(() => [
  { name: 'Claude Opus 4.8', value: '99.9%', brand: 'claude', iconBg: 'bg-orange-50 dark:bg-orange-500/15' },
  { name: 'ChatGPT-5.5', value: '99.0%', brand: 'openai', iconBg: 'bg-gray-100 dark:bg-white/15' },
  { name: 'Gemini 3.1 Pro', value: '99.6%', brand: 'gemini', iconBg: 'bg-blue-50 dark:bg-blue-500/15' },
  { name: 'DeepSeek V3.2', value: '99.7%', brand: 'deepseek', iconBg: 'bg-blue-50 dark:bg-blue-500/15' }
])
</script>

<style scoped>
/* 静态展示卡，不做 hover 位移以保持克制 */
</style>

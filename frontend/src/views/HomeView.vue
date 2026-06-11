<template>
  <!-- 自定义首页内容：站长可在后台配置整页 HTML 或外链 -->
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- 默认营销首页 -->
  <div
    v-else
    class="relative min-h-screen overflow-hidden bg-white text-gray-900 dark:bg-dark-950 dark:text-gray-100"
  >
    <HomeNavBar
      :site-name="siteName"
      :site-logo="siteLogo"
      :doc-url="docUrl"
      :is-authenticated="isAuthenticated"
      :dashboard-path="dashboardPath"
    />

    <!-- ============== Hero ============== -->
    <section class="relative overflow-hidden">
      <!-- 装饰：暖色 wash + 网格 + 右上 brand glow -->
      <div class="pointer-events-none absolute inset-0">
        <div class="absolute inset-x-0 top-0 h-[640px] bg-gradient-to-b from-amber-50/70 via-orange-50/30 to-transparent dark:from-brand-500/[0.04] dark:via-amber-500/[0.02]"></div>
        <div
          class="absolute inset-0 bg-[linear-gradient(rgba(120,113,108,0.05)_1px,transparent_1px),linear-gradient(90deg,rgba(120,113,108,0.05)_1px,transparent_1px)] bg-[size:64px_64px] [mask-image:radial-gradient(ellipse_at_top,black_5%,transparent_70%)]"
        ></div>
        <div class="absolute -right-32 top-32 h-[480px] w-[480px] rounded-full bg-brand-300/25 blur-[100px] dark:bg-brand-500/10"></div>
      </div>

      <div class="relative mx-auto max-w-7xl px-4 pb-20 pt-12 sm:px-6 sm:pb-24 sm:pt-16 lg:px-8 lg:pt-20">
        <div class="grid items-center gap-10 lg:grid-cols-12 lg:gap-12">
          <!-- 左：主文案 -->
          <div class="lg:col-span-6">
            <div class="inline-flex items-center gap-2 rounded-full border border-brand-200 bg-brand-50/80 px-3 py-1 text-xs font-medium text-brand-800 dark:border-brand-500/30 dark:bg-brand-500/10 dark:text-brand-300">
              <span class="h-1.5 w-1.5 rounded-full bg-brand-500 dark:bg-brand-400"></span>
              {{ t('home.heroEyebrowFull') }}
            </div>

            <!-- Display 标题：大字号 + 橙色关键词高亮 -->
            <h1 class="mt-6 text-[2.625rem] font-semibold leading-[1.08] tracking-[-0.03em] text-gray-900 [text-wrap:balance] dark:text-white sm:text-6xl sm:leading-[1.05] lg:text-[4.25rem]">
              {{ t('home.heroDisplayPrimary') }}<br />
              {{ t('home.heroDisplayLine2') }}
              <!-- 高亮词整体不拆行，避免"模/型"这类词中断行的孤字 -->
              <span class="whitespace-nowrap text-brand-600 dark:text-brand-400">{{ t('home.heroDisplayHighlight') }}</span>
            </h1>

            <p class="mt-6 max-w-xl text-base leading-relaxed text-gray-600 dark:text-dark-300 sm:text-lg">
              {{ t('home.heroLead') }}
            </p>

            <!-- 主行动区：注册 CTA + 价格入口 + 免费额度提示 -->
            <div class="mt-8 flex flex-wrap items-center gap-3">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary btn-lg group"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon
                  name="arrowRight"
                  size="sm"
                  class="transition-transform duration-200 group-hover:translate-x-0.5"
                  :stroke-width="2"
                />
              </router-link>
              <a href="#pricing" class="btn btn-secondary btn-lg">
                {{ t('home.heroSecondaryCta') }}
              </a>
            </div>
            <p class="mt-3.5 inline-flex items-center gap-1.5 text-xs font-medium text-gray-500 dark:text-dark-400">
              <Icon name="checkCircle" size="xs" class="text-emerald-500" :stroke-width="2" />
              {{ t('home.pricing.plans.free.priceNote') }}
            </p>

            <!-- 三个特性 chip -->
            <div class="mt-10 grid grid-cols-1 gap-5 sm:grid-cols-3">
              <div v-for="chip in heroChips" :key="chip.key" class="flex items-center gap-3">
                <div :class="['flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl ring-1 ring-inset', chip.boxClass]">
                  <Icon :name="chip.icon" size="sm" :class="chip.iconClass" :stroke-width="chip.strokeWidth" />
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ chip.title }}</div>
                  <div class="text-xs text-gray-500 dark:text-dark-400">{{ chip.sub }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 右：控制台预览卡 -->
          <div class="lg:col-span-6">
            <DashboardPreview />
          </div>
        </div>
      </div>
    </section>

    <PainPointsSection />
    <StepsSection :site-name="siteName" />
    <ComparisonSection :site-name="siteName" />
    <PricingSection v-if="plans.length > 0" :plans="plans" :is-authenticated="isAuthenticated" />
    <FaqSection />
    <CtaSection :is-authenticated="isAuthenticated" :dashboard-path="dashboardPath" :doc-url="docUrl" />
    <HomeFooter :site-name="siteName" :site-logo="siteLogo" :doc-url="docUrl" :has-pricing="plans.length > 0" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { useTheme } from '@/composables/useTheme'
import Icon from '@/components/icons/Icon.vue'
import HomeNavBar from '@/components/home/HomeNavBar.vue'
import DashboardPreview from '@/components/home/DashboardPreview.vue'
import PainPointsSection from '@/components/home/PainPointsSection.vue'
import StepsSection from '@/components/home/StepsSection.vue'
import ComparisonSection from '@/components/home/ComparisonSection.vue'
import PricingSection, { type DisplayPlan } from '@/components/home/PricingSection.vue'
import FaqSection from '@/components/home/FaqSection.vue'
import CtaSection from '@/components/home/CtaSection.vue'
import HomeFooter from '@/components/home/HomeFooter.vue'
import { paymentAPI, type PublicPlan } from '@/api/payment'

const { t } = useI18n()
const { initTheme } = useTheme()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

// Hero 下方三个特性 chip（兼容 / 计费 / 稳定）
type HeroChipIcon = 'check' | 'dollar' | 'shield'
interface HeroChip {
  key: string
  icon: HeroChipIcon
  boxClass: string
  iconClass: string
  strokeWidth: number
  title: string
  sub: string
}
const heroChips = computed<HeroChip[]>(() => [
  {
    key: 'compatible', icon: 'check', strokeWidth: 2.5,
    boxClass: 'bg-emerald-50 ring-emerald-200/70 dark:bg-emerald-500/15 dark:ring-emerald-500/30',
    iconClass: 'text-emerald-600 dark:text-emerald-300',
    title: t('home.stats.compatible'), sub: t('home.stats.compatibleSub')
  },
  {
    key: 'payAsYouGo', icon: 'dollar', strokeWidth: 2,
    boxClass: 'bg-brand-50 ring-brand-200/70 dark:bg-brand-500/15 dark:ring-brand-500/30',
    iconClass: 'text-brand-600 dark:text-brand-300',
    title: t('home.stats.payAsYouGo'), sub: t('home.stats.payAsYouGoSub')
  },
  {
    key: 'stable', icon: 'shield', strokeWidth: 2,
    boxClass: 'bg-sky-50 ring-sky-200/70 dark:bg-sky-500/15 dark:ring-sky-500/30',
    iconClass: 'text-sky-600 dark:text-sky-300',
    title: t('home.stats.stable'), sub: t('home.stats.stableSub')
  }
])

// ========= 后端驱动的订阅套餐 =========

const plans = ref<DisplayPlan[]>([])

function parsePlanFeatures(raw: string): string[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed.filter(Boolean).map(String)
  } catch {
    // 后台可能存的是逗号/换行分隔字符串
  }
  return raw.split(/[\n,，]/).map(s => s.trim()).filter(Boolean)
}

function formatValidity(days: number, unit: string): string {
  if (!days || days <= 0) return ''
  const u = (unit || 'day').toLowerCase()
  if (u.startsWith('day')) return `${days} 天`
  if (u.startsWith('month')) return `${days} 个月`
  if (u.startsWith('year')) return `${days} 年`
  return `${days} ${unit}`
}

// 拉取后台公开套餐并映射为展示数据；加载完成后补扫一次 reveal（套餐区是异步渲染的）
async function loadPlans(): Promise<void> {
  try {
    const list = await paymentAPI.getPlansPublic()
    if (Array.isArray(list) && list.length > 0) {
      const sorted = [...list].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
      plans.value = sorted.map((p: PublicPlan): DisplayPlan => {
        const badgeText = (p.badge_text || '').trim() || (p.is_popular ? t('home.pricing.recommended') : '')
        return {
          id: p.id,
          name: p.name,
          description: p.description || '',
          price: p.price,
          originalPrice: p.original_price,
          validityLabel: formatValidity(p.validity_days, p.validity_unit),
          features: parsePlanFeatures(p.features),
          productName: p.product_name,
          recommended: badgeText !== '',
          badgeText,
        }
      })
    } else {
      plans.value = []
    }
  } catch {
    plans.value = []
  }
  await nextTick()
  scanReveal()
}

// ========= 滚动进入动效（纯增强） =========
// 内容任何时候都默认可见（不做预隐藏）：元素首次进入视口时才加 .reveal-in
// 播放一次性 CSS animation（动画自带从透明升起的初始帧）。这样无 JS、
// 打印、后台标签页、截图等 IntersectionObserver 不触发的场景内容照常可见。

let revealObserver: IntersectionObserver | null = null

function scanReveal(): void {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
  if (!('IntersectionObserver' in window)) return

  if (!revealObserver) {
    revealObserver = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue
          entry.target.classList.add('reveal-in')
          revealObserver?.unobserve(entry.target)
        }
      },
      { rootMargin: '0px 0px -8% 0px', threshold: 0.1 }
    )
  }

  const nodes = document.querySelectorAll<HTMLElement>('[data-reveal]:not([data-reveal-bound])')
  nodes.forEach((el) => {
    el.dataset.revealBound = '1'
    // 已处于视口内的内容不参与动效，避免首屏闪烁
    if (el.getBoundingClientRect().top < window.innerHeight * 0.92) return
    const delay = el.dataset.reveal
    if (delay) el.style.setProperty('--reveal-delay', delay)
    revealObserver?.observe(el)
  })
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  loadPlans()
  scanReveal()
})

onBeforeUnmount(() => {
  revealObserver?.disconnect()
  revealObserver = null
})
</script>

<style>
/* 滚动 reveal 动效（全局：作用于各 section 子组件内打了 data-reveal 的元素）。
 * 元素默认始终可见；.reveal-in 只在进入视口时由 JS 添加，播放一次
 * 从透明升起的 animation（初始帧由 keyframes 提供，不做常驻预隐藏）。 */
.reveal-in {
  animation: home-reveal-up 0.55s cubic-bezier(0.16, 1, 0.3, 1) both;
  animation-delay: var(--reveal-delay, 0ms);
}
@keyframes home-reveal-up {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}
@media (prefers-reduced-motion: reduce) {
  .reveal-in {
    animation: none;
  }
}
</style>

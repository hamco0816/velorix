<template>
  <div class="app-bg min-h-screen">
    <!-- 多色 mesh gradient：4 个角各一个 5-8% 透明度的色彩 blob，给"空白处"加温度和层次 -->
    <div class="bg-mesh pointer-events-none fixed inset-0"></div>
    <!-- 极淡 dot pattern：24px 网格让纯白区域也有"纸质质感"，类似 Notion / Linear 的细节 -->
    <div class="bg-dots pointer-events-none fixed inset-0"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content：默认 1760px 版心；wide=true 时上限 2200px（超 2K/4K 屏不让内容铺散，保持可读）-->
      <main class="px-4 py-5 sm:px-6 sm:py-6 lg:px-8">
        <div :class="['mx-auto w-full', wide ? 'max-w-[2200px]' : 'max-w-[1760px]']">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

withDefaults(defineProps<{
  /** 占满剩余宽度（默认 false，保留 1760px 版心避免超宽屏内容铺散）。仪表盘/分析类页面需要时显式传 true */
  wide?: boolean
}>(), { wide: false })

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>

<style scoped>
/* ============ 全局背景：暖底 + 多色 mesh + 极淡 dot 纹理 ============ */
/* 底色：stone-50 暖白（#fafaf9），比 gray-50 更温和，避免"消毒水冷感" */
.app-bg {
  background-color: #fafaf9;
}

:root.dark .app-bg {
  background-color: rgb(9, 9, 11); /* zinc-950 */
}

/* mesh：4 个角 radial blob 各取一个色相，透明度 5-8%。
   构成"温暖+冷静"色彩平衡，让大面积留白也有温度，但不抢卡片视线 */
.bg-mesh {
  background-image:
    radial-gradient(circle at 12% 18%, rgba(249, 115, 22, 0.08) 0px, transparent 45%),
    radial-gradient(circle at 88% 12%, rgba(59, 130, 246, 0.05) 0px, transparent 45%),
    radial-gradient(circle at 80% 88%, rgba(139, 92, 246, 0.05) 0px, transparent 50%),
    radial-gradient(circle at 18% 82%, rgba(16, 185, 129, 0.05) 0px, transparent 45%);
}

:root.dark .bg-mesh {
  background-image:
    radial-gradient(circle at 12% 18%, rgba(249, 115, 22, 0.10) 0px, transparent 45%),
    radial-gradient(circle at 88% 12%, rgba(59, 130, 246, 0.06) 0px, transparent 45%),
    radial-gradient(circle at 80% 88%, rgba(139, 92, 246, 0.06) 0px, transparent 50%),
    radial-gradient(circle at 18% 82%, rgba(16, 185, 129, 0.06) 0px, transparent 45%);
}

/* dot pattern：24px 间距、1px 圆点、3% 黑透明度，肉眼几乎不可见但让"空白有纸感"
   Linear/Vercel 类深 SaaS 都有这层细节 */
.bg-dots {
  background-image: radial-gradient(circle, rgba(15, 23, 42, 0.045) 1px, transparent 1px);
  background-size: 24px 24px;
}

:root.dark .bg-dots {
  background-image: radial-gradient(circle, rgba(255, 255, 255, 0.04) 1px, transparent 1px);
  background-size: 24px 24px;
}
</style>

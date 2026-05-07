<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <!-- Background Decoration: 极淡中性 radial 替代彩色 mesh，避免 SaaS 模板感 -->
    <div class="pointer-events-none fixed inset-0 bg-[radial-gradient(at_40%_20%,rgba(113,113,122,0.04)_0px,transparent_50%),radial-gradient(at_80%_0%,rgba(234,88,12,0.025)_0px,transparent_50%)]"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content：超宽屏内容居中收口（1536px），避免铺满产生空旷感 -->
      <main class="px-4 py-5 sm:px-6 sm:py-6 lg:px-8">
        <div class="mx-auto w-full max-w-screen-2xl">
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

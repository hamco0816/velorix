<template>
  <!-- 收尾 CTA 区块：深色面板 + 网格纹理 + 品牌暖光，引导注册/进入控制台 -->
  <section id="cta" class="px-4 pb-20 sm:px-6 lg:px-8">
    <div data-reveal class="relative mx-auto max-w-7xl overflow-hidden rounded-2xl bg-gray-950">
      <div class="pointer-events-none absolute inset-0">
        <div
          class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.025)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.025)_1px,transparent_1px)] bg-[size:48px_48px] [mask-image:radial-gradient(ellipse_at_top_right,black_5%,transparent_75%)]"
        ></div>
        <div class="absolute -right-32 -top-32 h-[420px] w-[420px] rounded-full bg-brand-500/25 blur-[100px]"></div>
        <div class="absolute -bottom-40 -left-20 h-[420px] w-[420px] rounded-full bg-amber-400/12 blur-[100px]"></div>
      </div>

      <div class="relative grid items-center gap-8 p-10 sm:grid-cols-12 sm:p-14">
        <div class="sm:col-span-7">
          <h2 class="text-[1.75rem] font-semibold leading-tight tracking-tight text-white [text-wrap:balance] sm:text-3xl lg:text-4xl">
            {{ t('home.cta.title') }}
          </h2>
          <p class="mt-3 text-sm leading-relaxed text-gray-300/85 sm:text-base">
            {{ t('home.cta.description') }}
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-3 sm:col-span-5 sm:justify-end">
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="group inline-flex items-center justify-center gap-2 rounded-md bg-white px-5 py-2.5 text-sm font-semibold text-gray-900 transition-colors hover:bg-gray-100"
          >
            {{ isAuthenticated ? t('home.goToDashboard') : t('home.cta.button') }}
            <Icon name="arrowRight" size="xs" class="transition-transform duration-200 group-hover:translate-x-0.5" :stroke-width="2" />
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center justify-center gap-2 rounded-md bg-white/[0.08] px-5 py-2.5 text-sm font-semibold text-white ring-1 ring-inset ring-white/15 transition-colors hover:bg-white/[0.12]"
          >
            <Icon name="book" size="xs" :stroke-width="2" />
            {{ t('home.viewDocs') }}
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  isAuthenticated: boolean
  dashboardPath: string
  docUrl: string
}>()

const { t } = useI18n()
</script>

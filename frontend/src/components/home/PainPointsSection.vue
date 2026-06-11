<template>
  <!-- 用户痛点区块：四张卡片陈列接入 AI 时的常见负担 -->
  <section id="pain-points" class="border-y border-gray-200 bg-stone-50/70 py-20 dark:border-dark-800 dark:bg-dark-900/30 sm:py-24">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <SectionHeading
        data-reveal
        :title="t('home.painPoints.title')"
        :subtitle="t('home.painPoints.subtitle')"
      />

      <div class="mt-10 grid gap-3 sm:grid-cols-2 lg:grid-cols-4 lg:gap-4">
        <div
          v-for="(item, index) in painPointItems"
          :key="item.key"
          :data-reveal="`${index * 70}ms`"
          class="group rounded-2xl border border-gray-200/70 bg-white p-6 transition-all duration-200 hover:-translate-y-0.5 hover:border-brand-200/80 hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-brand-500/30"
        >
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50 text-brand-600 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30">
            <Icon :name="item.icon" size="sm" :stroke-width="2" />
          </div>
          <h3 class="mt-4 text-base font-semibold tracking-tight text-gray-900 dark:text-white">
            {{ item.title }}
          </h3>
          <p class="mt-1.5 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
            {{ item.desc }}
          </p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import SectionHeading from '@/components/home/SectionHeading.vue'

const { t } = useI18n()

type PainPointIcon = 'dollar' | 'users' | 'chart' | 'shield'

interface PainPoint {
  key: string
  icon: PainPointIcon
  title: string
  desc: string
}

const painPointItems = computed<PainPoint[]>(() => [
  { key: 'expensive', icon: 'dollar', title: t('home.painPoints.items.expensive.title'), desc: t('home.painPoints.items.expensive.desc') },
  { key: 'complex', icon: 'users', title: t('home.painPoints.items.complex.title'), desc: t('home.painPoints.items.complex.desc') },
  { key: 'unstable', icon: 'chart', title: t('home.painPoints.items.unstable.title'), desc: t('home.painPoints.items.unstable.desc') },
  { key: 'noControl', icon: 'shield', title: t('home.painPoints.items.noControl.title'), desc: t('home.painPoints.items.noControl.desc') }
])
</script>

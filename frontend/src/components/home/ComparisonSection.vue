<template>
  <!-- 优势对比区块：左侧对照表 + 右侧已支持模型网格 -->
  <section class="border-y border-gray-200 bg-stone-50/70 py-20 dark:border-dark-800 dark:bg-dark-900/30 sm:py-24">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <SectionHeading
        data-reveal
        :title="t('home.comparison.title')"
        :subtitle="t('home.comparison.subtitle')"
      />

      <div class="mt-10 grid gap-6 lg:grid-cols-2 lg:gap-8">
        <!-- 对照表 -->
        <div
          data-reveal
          class="overflow-hidden rounded-2xl border border-gray-200/70 bg-white dark:border-dark-700/60 dark:bg-dark-800/40"
        >
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-50/60 dark:bg-dark-800/40">
                <th class="px-4 py-3.5 text-left text-xs font-medium text-gray-500 dark:text-dark-400 sm:px-5">
                  {{ t('home.comparison.headers.feature') }}
                </th>
                <th class="px-4 py-3.5 text-left text-xs font-medium text-gray-500 dark:text-dark-400 sm:px-5">
                  {{ t('home.comparison.headers.official') }}
                </th>
                <th class="bg-brand-50/60 px-4 py-3.5 text-left text-xs font-medium text-brand-700 dark:bg-brand-500/10 dark:text-brand-300 sm:px-5">
                  {{ siteName }} {{ t('home.comparison.headers.us') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700/60">
              <tr v-for="row in comparisonItems" :key="row.key">
                <td class="px-4 py-3.5 font-medium text-gray-900 dark:text-white sm:px-5">{{ row.feature }}</td>
                <td class="px-4 py-3.5 text-sm text-gray-500 dark:text-dark-400 sm:px-5">
                  <span class="inline-flex items-start gap-1.5">
                    <Icon name="x" size="xs" class="mt-0.5 flex-shrink-0 text-gray-400" :stroke-width="2.5" />
                    <span>{{ row.official }}</span>
                  </span>
                </td>
                <td class="bg-brand-50/30 px-4 py-3.5 text-sm text-gray-900 dark:bg-brand-500/[0.06] dark:text-gray-100 sm:px-5">
                  <span class="inline-flex items-start gap-1.5">
                    <Icon name="check" size="xs" class="mt-0.5 flex-shrink-0 text-brand-600 dark:text-brand-400" :stroke-width="2.5" />
                    <span class="font-medium">{{ row.us }}</span>
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 已支持模型网格 -->
        <div id="providers" data-reveal="120ms">
          <h3 class="text-base font-semibold tracking-tight text-gray-900 dark:text-white">
            {{ t('home.providers.title') }}
          </h3>
          <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
            {{ t('home.providers.description') }}
          </p>

          <div class="mt-5 grid grid-cols-2 gap-2.5">
            <div
              v-for="provider in providerCards"
              :key="provider.key"
              class="group flex items-center gap-3 rounded-xl border border-gray-200/70 bg-white p-3.5 transition-all duration-200 hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/40"
            >
              <div :class="['flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl', provider.bg]">
                <BrandIcon :brand="provider.brand" size="22px" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ provider.name }}</p>
                <p class="mt-0.5 inline-flex items-center gap-1.5 text-2xs font-medium text-emerald-600 dark:text-emerald-400">
                  <span class="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                  {{ t('home.providers.supported') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import BrandIcon from '@/components/common/BrandIcon.vue'
import SectionHeading from '@/components/home/SectionHeading.vue'

defineProps<{
  siteName: string
}>()

const { t } = useI18n()

interface ComparisonRow {
  key: string
  feature: string
  official: string
  us: string
}

const comparisonItems = computed<ComparisonRow[]>(() => [
  { key: 'pricing', feature: t('home.comparison.items.pricing.feature'), official: t('home.comparison.items.pricing.official'), us: t('home.comparison.items.pricing.us') },
  { key: 'ban', feature: t('home.comparison.items.ban.feature'), official: t('home.comparison.items.ban.official'), us: t('home.comparison.items.ban.us') },
  { key: 'models', feature: t('home.comparison.items.models.feature'), official: t('home.comparison.items.models.official'), us: t('home.comparison.items.models.us') },
  { key: 'management', feature: t('home.comparison.items.management.feature'), official: t('home.comparison.items.management.official'), us: t('home.comparison.items.management.us') },
  { key: 'stability', feature: t('home.comparison.items.stability.feature'), official: t('home.comparison.items.stability.official'), us: t('home.comparison.items.stability.us') },
  { key: 'control', feature: t('home.comparison.items.control.feature'), official: t('home.comparison.items.control.official'), us: t('home.comparison.items.control.us') }
])

interface ProviderCard {
  key: string
  name: string
  brand: 'claude' | 'openai' | 'gemini' | 'deepseek'
  bg: string
}

const providerCards = computed<ProviderCard[]>(() => [
  { key: 'claude', name: t('home.providers.claude'), brand: 'claude', bg: 'bg-orange-50 ring-1 ring-inset ring-orange-100 dark:bg-orange-500/10 dark:ring-orange-500/20' },
  { key: 'gpt', name: 'ChatGPT', brand: 'openai', bg: 'bg-gray-100 ring-1 ring-inset ring-gray-200 dark:bg-white/10 dark:ring-white/15' },
  { key: 'gemini', name: t('home.providers.gemini'), brand: 'gemini', bg: 'bg-blue-50 ring-1 ring-inset ring-blue-100 dark:bg-blue-500/10 dark:ring-blue-500/20' },
  { key: 'deepseek', name: t('home.providersExt.deepseek'), brand: 'deepseek', bg: 'bg-blue-50 ring-1 ring-inset ring-blue-100 dark:bg-blue-500/10 dark:ring-blue-500/20' }
])
</script>

<template>
  <!-- 信任 FAQ 区块：直面"会不会跑路"等顾虑 -->
  <section id="faq" class="border-y border-gray-200 bg-stone-50/70 py-20 dark:border-dark-800 dark:bg-dark-900/30 sm:py-24">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <SectionHeading
        data-reveal
        :title="t('home.faq.title')"
        :subtitle="t('home.faq.subtitle')"
      />

      <div class="mt-10 grid gap-3 sm:grid-cols-2 lg:gap-4">
        <div
          v-for="(item, index) in faqItems"
          :key="item.key"
          :data-reveal="`${(index % 2) * 70}ms`"
          class="rounded-2xl border border-gray-200/70 bg-white p-6 transition-all duration-200 hover:-translate-y-0.5 hover:border-brand-200/80 hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/40 dark:hover:border-brand-500/30"
        >
          <h3 class="flex items-start gap-2 text-base font-semibold tracking-tight text-gray-900 dark:text-white">
            <Icon name="checkCircle" size="sm" class="mt-0.5 shrink-0 text-brand-500 dark:text-brand-400" :stroke-width="2" />
            {{ item.q }}
          </h3>
          <p class="mt-2 pl-7 text-sm leading-relaxed text-gray-600 dark:text-dark-300">
            {{ item.a }}
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

interface FaqItem {
  key: string
  q: string
  a: string
}

const faqItems = computed<FaqItem[]>(() => [
  { key: 'runaway', q: t('home.faq.items.runaway.q'), a: t('home.faq.items.runaway.a') },
  { key: 'invoice', q: t('home.faq.items.invoice.q'), a: t('home.faq.items.invoice.a') },
  { key: 'refund', q: t('home.faq.items.refund.q'), a: t('home.faq.items.refund.a') },
  { key: 'billing', q: t('home.faq.items.billing.q'), a: t('home.faq.items.billing.a') },
  { key: 'security', q: t('home.faq.items.security.q'), a: t('home.faq.items.security.a') },
  { key: 'support', q: t('home.faq.items.support.q'), a: t('home.faq.items.support.a') }
])
</script>

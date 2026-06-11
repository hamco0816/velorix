<template>
  <!-- 定价区块：后台配置驱动的套餐卡片 + 右侧购买保障说明 -->
  <section id="pricing" class="py-20 sm:py-24">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <SectionHeading
        data-reveal
        :title="t('home.pricing.title')"
        :subtitle="t('home.pricing.subtitle')"
      />

      <div class="mt-10 grid gap-6 lg:grid-cols-2 xl:grid-cols-[2fr_1fr]">
        <!-- 套餐卡片网格 -->
        <div class="grid gap-4" :class="planGridColsClass">
          <div
            v-for="(plan, index) in plans"
            :key="plan.id"
            :data-reveal="`${index * 80}ms`"
            class="relative flex flex-col rounded-2xl p-6 transition-all duration-200"
            :class="plan.recommended
              ? 'border border-gray-900 bg-gray-950 text-white dark:border-white'
              : 'border border-gray-200/70 bg-white hover:-translate-y-0.5 hover:border-gray-300 hover:shadow-card-hover dark:border-dark-700/60 dark:bg-dark-800/40'"
          >
            <span
              v-if="plan.recommended"
              class="absolute -top-2.5 right-6 inline-flex items-center rounded-full bg-brand-500 px-2.5 py-0.5 text-2xs font-semibold text-white"
            >
              {{ plan.badgeText }}
            </span>

            <h3 class="text-base font-semibold tracking-tight" :class="plan.recommended ? 'text-white' : 'text-gray-900 dark:text-white'">
              {{ plan.name }}
            </h3>
            <p v-if="plan.description" class="mt-1 text-xs" :class="plan.recommended ? 'text-gray-400' : 'text-gray-500 dark:text-dark-400'">
              {{ plan.description }}
            </p>

            <div class="mt-5 flex items-baseline gap-2">
              <span class="text-3xl font-semibold tabular-nums tracking-tight" :class="plan.recommended ? 'text-white' : 'text-gray-900 dark:text-white'">
                ¥ {{ plan.price }}
              </span>
              <span v-if="plan.originalPrice && plan.originalPrice > plan.price" class="text-sm text-gray-400 line-through tabular-nums">
                ¥ {{ plan.originalPrice }}
              </span>
            </div>
            <p v-if="plan.validityLabel" class="mt-1 text-xs" :class="plan.recommended ? 'text-gray-400' : 'text-gray-500 dark:text-dark-400'">
              {{ plan.validityLabel }}
            </p>

            <ul v-if="plan.features.length > 0" class="mt-6 space-y-2.5 text-sm" :class="plan.recommended ? 'text-gray-200' : 'text-gray-700 dark:text-dark-200'">
              <li v-for="feat in plan.features" :key="feat" class="flex items-start gap-2">
                <Icon name="check" size="xs" class="mt-0.5 flex-shrink-0" :class="plan.recommended ? 'text-brand-400' : 'text-emerald-500'" :stroke-width="2.5" />
                <span>{{ feat }}</span>
              </li>
            </ul>

            <router-link
              :to="isAuthenticated ? '/purchase' : '/login?redirect=/purchase'"
              class="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-md px-3.5 py-2 text-sm font-semibold transition-colors"
              :class="plan.recommended
                ? 'bg-white text-gray-900 hover:bg-gray-100'
                : 'bg-primary-950 text-white hover:bg-primary-800 dark:bg-white dark:text-primary-950 dark:hover:bg-primary-100'"
            >
              {{ isAuthenticated ? t('home.pricing.plans.developer.cta') : t('home.cta.button') }}
              <Icon name="arrowRight" size="xs" :stroke-width="2" />
            </router-link>
          </div>
        </div>

        <!-- 右侧购买保障说明 -->
        <div data-reveal="160ms" class="grid gap-x-5 gap-y-5 sm:grid-cols-2 xl:grid-cols-1 xl:gap-y-6 xl:self-start">
          <div v-for="benefit in benefitItems" :key="benefit.key" class="flex items-start gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-brand-50 text-brand-600 ring-1 ring-inset ring-brand-200/70 dark:bg-brand-500/15 dark:text-brand-300 dark:ring-brand-500/30">
              <Icon :name="benefit.icon" size="sm" :stroke-width="2" />
            </div>
            <div class="min-w-0">
              <div class="text-sm font-semibold tracking-tight text-gray-900 dark:text-white">{{ benefit.title }}</div>
              <div class="mt-0.5 text-xs leading-relaxed text-gray-500 dark:text-dark-400">{{ benefit.desc }}</div>
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
import SectionHeading from '@/components/home/SectionHeading.vue'

/** 营销页展示用的套餐数据（由 HomeView 拉取后端公开套餐映射而来） */
export interface DisplayPlan {
  id: number
  name: string
  description: string
  price: number
  originalPrice?: number
  validityLabel: string
  features: string[]
  productName: string
  recommended: boolean
  badgeText: string
}

const props = defineProps<{
  plans: DisplayPlan[]
  isAuthenticated: boolean
}>()

const { t } = useI18n()

// 套餐卡片网格列数：根据后台配置的套餐数自适应（1/2/3 列）
const planGridColsClass = computed(() => {
  const count = props.plans.length
  if (count <= 1) return 'grid-cols-1'
  if (count === 2) return 'grid-cols-1 sm:grid-cols-2'
  return 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3'
})

type BenefitIcon = 'dollar' | 'chart' | 'shield' | 'clock'

interface BenefitItem {
  key: string
  icon: BenefitIcon
  title: string
  desc: string
}

const benefitItems = computed<BenefitItem[]>(() => [
  { key: 'payAsYouGo', icon: 'dollar', title: t('home.pricing.benefits.payAsYouGo.title'), desc: t('home.pricing.benefits.payAsYouGo.desc') },
  { key: 'tieredDiscount', icon: 'chart', title: t('home.pricing.benefits.tieredDiscount.title'), desc: t('home.pricing.benefits.tieredDiscount.desc') },
  { key: 'secure', icon: 'shield', title: t('home.pricing.benefits.secure.title'), desc: t('home.pricing.benefits.secure.desc') },
  { key: 'support247', icon: 'clock', title: t('home.pricing.benefits.support247.title'), desc: t('home.pricing.benefits.support247.desc') }
])
</script>

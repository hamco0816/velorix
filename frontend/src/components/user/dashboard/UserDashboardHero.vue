<!--
  账户英雄卡：仪表盘首屏视觉焦点。
  左侧展示账户余额（超大等宽数字）+ 今日消费/请求摘要 + 充值入口；
  右侧在用户存在有效订阅时展示用量最高的订阅配额进度（品牌橙进度条），拿不到数据则整块隐藏。
  浅色模式为黑底白字，暗色模式反相为白底黑字（与 .btn-primary 同一语言）。
-->
<template>
  <section class="hero-card">
    <div class="flex flex-col gap-6 p-6 sm:p-7 lg:flex-row lg:items-center lg:gap-10">
      <!-- 左侧：余额 + 今日摘要 + 充值 CTA -->
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-3">
          <span class="hero-icon">
            <Icon name="dollar" size="sm" :stroke-width="1.75" />
          </span>
          <p class="hero-label">{{ t('dashboard.balance') }} · {{ t('common.available') }}</p>
        </div>
        <p class="mt-4 flex items-baseline tabular-nums tracking-tight">
          <span class="hero-value text-[26px] leading-none">$</span>
          <span class="hero-value ml-0.5 text-[44px] leading-none">{{ balanceParts.int }}</span>
          <span class="hero-value-dim text-[24px] leading-none">.{{ balanceParts.dec }}</span>
        </p>
        <p class="hero-meta mt-3">
          <span>{{ t('dashboard.todayCost') }} <strong>${{ formatCost(todayCost) }}</strong></span>
          <span class="hero-divider" aria-hidden="true"></span>
          <span>{{ t('dashboard.todayRequests') }} <strong>{{ formatNumber(todayRequests) }}</strong></span>
        </p>
        <button type="button" class="hero-cta mt-5" @click="goRecharge">
          <Icon name="creditCard" size="sm" :stroke-width="1.75" />
          {{ t('dashboard.recharge') }}
        </button>
      </div>

      <!-- 右侧：活跃订阅配额进度（无订阅时不渲染，不造假数据） -->
      <div v-if="topSubscription" class="hero-quota">
        <div class="flex items-center justify-between gap-2">
          <p class="quota-title">{{ t('dashboard.subscriptionQuota') }}</p>
          <span v-if="expiryLabel" class="quota-expiry">{{ expiryLabel }}</span>
        </div>
        <p class="quota-group">{{ topSubscriptionName }}</p>

        <!-- 无限额订阅：展示无限制标识；有限额：逐条画进度条 -->
        <div v-if="quotaBars.length === 0" class="quota-unlimited">
          <span class="text-base leading-none">∞</span>
          {{ t('subscriptionProgress.unlimited') }}
        </div>
        <div v-else class="mt-3 space-y-2.5">
          <div v-for="bar in quotaBars" :key="bar.key" class="flex items-center gap-2.5">
            <span class="quota-bar-label">{{ bar.label }}</span>
            <div class="quota-track" role="progressbar" :aria-valuenow="Math.round(bar.percent)" aria-valuemin="0" aria-valuemax="100">
              <div class="quota-fill" :class="bar.fillClass" :style="{ width: bar.width }"></div>
            </div>
            <span class="quota-bar-nums">${{ bar.used }}/${{ bar.limit }}</span>
          </div>
        </div>

        <router-link to="/subscriptions" class="quota-link">
          {{ t('subscriptionProgress.viewAll') }}
          <span v-if="activeSubscriptions.length > 1" class="tabular-nums">({{ activeSubscriptions.length }})</span>
          <Icon name="chevronRight" size="xs" />
        </router-link>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useSubscriptionStore } from '@/stores'
import type { UserSubscription } from '@/types'

const props = defineProps<{
  balance: number
  todayCost: number
  todayRequests: number
}>()

const { t } = useI18n()
const router = useRouter()
const subscriptionStore = useSubscriptionStore()

// 跳转到充值页
const goRecharge = () => router.push({ path: '/purchase', query: { tab: 'recharge' } })

// 余额拆为整数/小数两段：整数放大、小数缩小，大数字更易读
const balanceParts = computed(() => {
  const formatted = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(props.balance || 0)
  const [int, dec] = formatted.split('.')
  return { int, dec }
})

const formatCost = (c: number) => (c || 0).toFixed(4)
const formatNumber = (n: number) => (n || 0).toLocaleString()

const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)

// 限额有效性：>0 才算有限额；订阅自身限额优先，缺省回落到分组限额
const validLimit = (v?: number | null) => (typeof v === 'number' && v > 0 ? v : null)

// 取一个订阅的三档（日/周/月）配额定义，只保留配置了限额的档位
const quotaDefsOf = (sub: UserSubscription) => [
  { key: 'daily', label: t('subscriptionProgress.daily'), used: sub.daily_usage_usd || 0, limit: validLimit(sub.daily_limit_usd) ?? validLimit(sub.group?.daily_limit_usd) },
  { key: 'weekly', label: t('subscriptionProgress.weekly'), used: sub.weekly_usage_usd || 0, limit: validLimit(sub.weekly_limit_usd) ?? validLimit(sub.group?.weekly_limit_usd) },
  { key: 'monthly', label: t('subscriptionProgress.monthly'), used: sub.monthly_usage_usd || 0, limit: validLimit(sub.monthly_limit_usd) ?? validLimit(sub.group?.monthly_limit_usd) }
].filter((d) => d.limit !== null)

// 订阅最高用量百分比，用于挑出"最紧张"的订阅放在英雄卡展示
const maxUsagePercent = (sub: UserSubscription) => {
  const percents = quotaDefsOf(sub).map((d) => ((d.used || 0) / (d.limit as number)) * 100)
  return percents.length ? Math.max(...percents) : 0
}

const topSubscription = computed<UserSubscription | null>(() => {
  if (!activeSubscriptions.value.length) return null
  return [...activeSubscriptions.value].sort((a, b) => maxUsagePercent(b) - maxUsagePercent(a))[0]
})

const topSubscriptionName = computed(() => {
  const sub = topSubscription.value
  return sub ? sub.group?.name || `Group #${sub.group_id}` : ''
})

// 进度条数据：宽度封顶 100%，>=90% 红色告警、>=70% 琥珀提醒、其余品牌橙
const quotaBars = computed(() => {
  if (!topSubscription.value) return []
  return quotaDefsOf(topSubscription.value).map((d) => {
    const percent = ((d.used || 0) / (d.limit as number)) * 100
    let fillClass = 'bg-brand-500'
    if (percent >= 90) fillClass = 'bg-red-500'
    else if (percent >= 70) fillClass = 'bg-amber-500'
    return {
      key: d.key,
      label: d.label,
      percent,
      width: `${Math.min(percent, 100)}%`,
      used: (d.used || 0).toFixed(2),
      limit: (d.limit as number).toFixed(2),
      fillClass
    }
  })
})

// 到期提示：过期/今天/明天/剩余 N 天
const expiryLabel = computed(() => {
  const expiresAt = topSubscription.value?.expires_at
  if (!expiresAt) return ''
  const diff = new Date(expiresAt).getTime() - Date.now()
  if (diff < 0) return t('subscriptionProgress.expired')
  const days = Math.ceil(diff / 86400000)
  if (days === 0) return t('subscriptionProgress.expiresToday')
  if (days === 1) return t('subscriptionProgress.expiresTomorrow')
  return t('subscriptionProgress.daysRemaining', { days })
})

onMounted(() => {
  // 订阅数据带 60s 缓存与去重，App 全局已有加载，这里兜底触发一次
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load subscriptions for dashboard hero:', error)
  })
})
</script>

<style scoped>
/* 英雄卡主体：浅色模式黑底（与 .btn-primary 同语言），暗色模式反相为白底 */
.hero-card {
  @apply relative overflow-hidden rounded-2xl border border-primary-950 bg-primary-950 shadow-inner-top;
  @apply dark:border-primary-200 dark:bg-white dark:shadow-card;
}

.hero-icon {
  @apply flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-white/10 text-white;
  @apply dark:bg-primary-950/[0.06] dark:text-primary-950;
}
.hero-label { @apply text-sm font-medium text-white/60 dark:text-primary-950/60; }
.hero-value { @apply font-semibold text-white dark:text-primary-950; }
.hero-value-dim { @apply font-semibold text-white/60 dark:text-primary-950/50; }

.hero-meta { @apply flex flex-wrap items-center gap-x-2.5 gap-y-1 text-xs text-white/55 dark:text-primary-950/55; }
.hero-meta strong { @apply font-semibold tabular-nums text-white/90 dark:text-primary-950/90; }
.hero-divider { @apply h-3 w-px bg-white/20 dark:bg-primary-950/15; }

/* 充值 CTA：在黑底上反白（暗色模式反黑），是英雄卡内唯一的强按钮 */
.hero-cta {
  @apply inline-flex items-center justify-center gap-2 rounded-lg bg-white px-3.5 py-2 text-sm font-medium text-primary-950;
  @apply transition duration-150 ease-out hover:bg-primary-100 active:scale-[0.98];
  @apply focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-400/60;
  @apply dark:bg-primary-950 dark:text-white dark:shadow-inner-top dark:hover:bg-primary-800;
  min-height: 2.25rem;
}

/* 订阅配额面板：英雄卡内的半透明子面板 */
.hero-quota {
  @apply w-full rounded-xl border border-white/10 bg-white/[0.06] p-4 lg:w-[340px] lg:flex-shrink-0;
  @apply dark:border-primary-950/10 dark:bg-primary-950/[0.03];
}
.quota-title { @apply text-xs font-semibold uppercase tracking-wide text-white/60 dark:text-primary-950/60; }
.quota-expiry { @apply text-2xs tabular-nums text-white/50 dark:text-primary-950/50; }
.quota-group { @apply mt-1 truncate text-sm font-medium text-white dark:text-primary-950; }

.quota-unlimited {
  @apply mt-3 inline-flex items-center gap-1.5 rounded-lg bg-emerald-500/15 px-2.5 py-1.5 text-xs font-medium text-emerald-300;
  @apply dark:bg-emerald-50 dark:text-emerald-700;
}

.quota-bar-label { @apply w-8 flex-shrink-0 text-2xs text-white/55 dark:text-primary-950/55; }
.quota-track { @apply h-1.5 min-w-0 flex-1 overflow-hidden rounded-full bg-white/15 dark:bg-primary-950/10; }
.quota-fill { @apply h-full rounded-full transition-all duration-200 ease-out; }
.quota-bar-nums { @apply w-24 flex-shrink-0 text-right text-2xs tabular-nums text-white/70 dark:text-primary-950/60; }

.quota-link {
  @apply mt-3 inline-flex items-center gap-1 text-xs font-medium text-white/60 transition-colors hover:text-white;
  @apply dark:text-primary-950/60 dark:hover:text-primary-950;
}
</style>

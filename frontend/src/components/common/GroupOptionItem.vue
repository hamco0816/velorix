<template>
  <div class="flex min-w-0 flex-1 items-start justify-between gap-3">
    <!-- Left: name + description -->
    <div
      class="flex min-w-0 flex-1 flex-col items-start"
      :title="description || undefined"
    >
      <!-- Row 1: platform badge (name bold) -->
      <GroupBadge
        :name="name"
        :platform="platform"
        :subscription-type="subscriptionType"
        :show-rate="false"
        class="groupOptionItemBadge"
      />
      <!-- Row 2: description with top spacing -->
      <span
        v-if="description"
        class="mt-1.5 w-full text-left text-xs leading-relaxed text-gray-500 dark:text-gray-400 line-clamp-2"
      >
        {{ description }}
      </span>
    </div>

    <!-- Right: rate pill + checkmark (vertically centered to first row) -->
    <div class="flex shrink-0 items-center gap-2 pt-0.5">
      <!-- Rate pill (platform color) -->
      <span v-if="rateMultiplier !== undefined" :class="['inline-flex items-center whitespace-nowrap rounded-full px-3 py-1 text-xs font-semibold', ratePillClass]">
        <template v-if="hasCustomRate">
          <span class="mr-1 line-through opacity-50">{{ rateMultiplier }}x</span>
          <span class="font-bold">{{ userRateMultiplier }}x</span>
        </template>
        <template v-else-if="showPromo">
          <!-- 限时倍率：原倍率删除线 + 折后倍率 + 倒计时 -->
          <span class="mr-1 line-through opacity-50">{{ rateMultiplier }}x</span>
          <span class="font-bold">{{ promoRateMultiplier }}x</span>
          <span v-if="promoCountdown" class="ml-1.5 tabular-nums opacity-90">{{ promoCountdown }}</span>
        </template>
        <template v-else>
          {{ rateMultiplier }}x 倍率
        </template>
      </span>
      <!-- Checkmark -->
      <Icon
        v-if="showCheckmark && selected"
        name="check"
        size="sm"
        class="shrink-0 text-primary-600 dark:text-primary-400"
        :stroke-width="2"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GroupBadge from './GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import { usePromoRate } from '@/composables/usePromoRate'
import type { SubscriptionType, GroupPlatform } from '@/types'

interface Props {
  name: string
  platform: GroupPlatform
  subscriptionType?: SubscriptionType
  rateMultiplier?: number
  userRateMultiplier?: number | null
  // 限时倍率（promo rate）：窗口内显示折后价 + 倒计时，不传则行为不变
  promoRateMultiplier?: number | null
  promoStartsAt?: string | null
  promoEndsAt?: string | null
  description?: string | null
  selected?: boolean
  showCheckmark?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: 'standard',
  selected: false,
  showCheckmark: true,
  userRateMultiplier: null,
  promoRateMultiplier: null,
  promoStartsAt: null,
  promoEndsAt: null
})

const { promoActive, promoCountdown } = usePromoRate(() => ({
  promoRateMultiplier: props.promoRateMultiplier,
  promoStartsAt: props.promoStartsAt,
  promoEndsAt: props.promoEndsAt
}))

// Whether user has a custom rate different from default
const hasCustomRate = computed(() => {
  return (
    props.userRateMultiplier !== null &&
    props.userRateMultiplier !== undefined &&
    props.rateMultiplier !== undefined &&
    props.userRateMultiplier !== props.rateMultiplier
  )
})

// 是否展示限时倍率：专属倍率优先级更高
const showPromo = computed(() => {
  if (!promoActive.value) return false
  if (hasCustomRate.value) return false
  return props.rateMultiplier !== undefined
})

// Rate pill color matches platform badge color
const ratePillClass = computed(() => {
  // 限时活动：用 rose 强调色突出折扣（与价格页一致）
  if (showPromo.value) {
    return 'bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
  }
  switch (props.platform) {
    case 'anthropic':
      return 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-400'
    case 'openai':
      return 'bg-green-50 text-green-700 dark:bg-green-900/20 dark:text-green-400'
    case 'gemini':
      return 'bg-sky-50 text-sky-700 dark:bg-sky-900/20 dark:text-sky-400'
    default: // antigravity and others
      return 'bg-violet-50 text-violet-700 dark:bg-violet-900/20 dark:text-violet-400'
  }
})
</script>

<style scoped>
/* Bold the group name inside GroupBadge when used in dropdown option */
.groupOptionItemBadge :deep(span.truncate) {
  font-weight: 600;
}
</style>

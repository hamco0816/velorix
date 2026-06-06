<!--
  TrustRibbon 信任条原子：一排"正规/可查/可开票/客服"等信任信号，用分隔点串起。
  内容完全由调用方通过 items 传入（避免把数据源写死进组件），按需展示。

  用法：
    <TrustRibbon :items="[
      { label: '用量实时可查', icon: 'checkCircle' },
      { label: '支持企业开票', icon: 'receipt' },
    ]" />
-->
<template>
  <div
    v-if="items.length"
    class="surface-card flex flex-wrap items-center gap-x-5 gap-y-2 px-5 py-3"
  >
    <template v-for="(item, index) in items" :key="item.label">
      <span
        v-if="index > 0"
        class="h-3 w-px bg-gray-200 dark:bg-dark-700"
        aria-hidden="true"
      ></span>
      <span
        :class="[
          'inline-flex items-center gap-1.5 whitespace-nowrap text-xs',
          item.strong
            ? 'font-semibold text-gray-900 dark:text-white'
            : 'text-gray-600 dark:text-dark-300',
        ]"
      >
        <Icon
          v-if="item.icon"
          :name="(item.icon as IconName)"
          size="xs"
          class="text-gray-400 dark:text-dark-500"
          :stroke-width="2"
        />
        {{ item.label }}
      </span>
    </template>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

// Icon 组件接受的名称联合类型（用于把外部传入的 string 收窄到合法图标名）
type IconName = InstanceType<typeof Icon>['$props']['name']

interface TrustItem {
  label: string
  // Icon 组件名（可选）
  icon?: string
  // 是否加粗强调（通常用于第一项"正规运营"这类总括）
  strong?: boolean
}

defineProps<{ items: TrustItem[] }>()
</script>

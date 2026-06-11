<template>
  <!-- 营销页 section 标题：大号标题 + 一句副标题的左对齐组合，全站营销 section 统一使用 -->
  <div class="max-w-3xl">
    <h2
      class="text-[1.75rem] font-semibold leading-tight tracking-tight text-gray-900 [text-wrap:balance] dark:text-white sm:text-3xl lg:text-4xl"
    >
      {{ title }}
    </h2>
    <p
      v-if="subtitle"
      class="mt-3 text-base leading-relaxed text-gray-600 dark:text-dark-300"
    >
      {{ subtitle }}
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  /** 标题文案 */
  title: string
  /** 一句话副标题（可选） */
  subtitle?: string
}>()
</script>

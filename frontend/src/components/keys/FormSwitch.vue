<!-- 表单开关：密钥创建/编辑弹窗内统一使用，选中态用品牌橙（全站「选中=橙」语言） -->
<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :class="[
      'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent',
      'transition-colors duration-200 ease-out',
      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600/40 dark:focus-visible:ring-brand-400/40',
      modelValue ? 'bg-brand-600 dark:bg-brand-500' : 'bg-gray-300 dark:bg-dark-600'
    ]"
    @click="$emit('update:modelValue', !modelValue)"
  >
    <span
      :class="[
        'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0',
        'transition-transform duration-200 ease-out motion-reduce:transition-none',
        modelValue ? 'translate-x-4' : 'translate-x-0'
      ]"
    />
  </button>
</template>

<script setup lang="ts">
defineProps<{ modelValue: boolean }>()

defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()
</script>

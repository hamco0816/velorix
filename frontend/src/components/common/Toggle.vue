<template>
  <button
    type="button"
    @click="toggle"
    class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500 dark:focus:ring-offset-dark-800"
    :class="[modelValue ? 'bg-emerald-500 hover:bg-emerald-600 dark:bg-emerald-600' : 'bg-gray-300 hover:bg-gray-400 dark:bg-dark-600 dark:hover:bg-dark-500']"
    role="switch"
    :aria-checked="modelValue"
  >
    <!-- ON 用 emerald 而非 primary：跟 iOS/GitHub 风格一致，肉眼能 100% 区分开启/关闭；
         原来用 primary（紫色）在深色背景或弱光下与灰色对比度太低，运营反映"开关都是灰的看不出" -->
    <span
      class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
      :class="[modelValue ? 'translate-x-5' : 'translate-x-0']"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

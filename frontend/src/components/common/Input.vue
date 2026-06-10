<template>
  <div class="w-full">
    <label v-if="label" :for="inputId" class="input-label mb-1.5 block">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <div class="relative">
      <!-- Prefix Icon Slot -->
      <div
        v-if="$slots.prefix"
        class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-gray-400 dark:text-dark-400"
      >
        <slot name="prefix"></slot>
      </div>

      <input
        :id="inputId"
        ref="inputRef"
        :type="type"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        :placeholder="placeholderText"
        :autocomplete="autocomplete"
        :readonly="readonly"
        :aria-invalid="error ? 'true' : undefined"
        :aria-describedby="describedBy"
        :class="[
          'input w-full transition-all duration-200',
          $slots.prefix ? 'pl-11' : '',
          $slots.suffix ? 'pr-11' : '',
          error ? 'input-error ring-2 ring-red-500/20' : '',
          disabled ? 'cursor-not-allowed bg-gray-100 opacity-60 dark:bg-dark-900' : ''
        ]"
        @input="onInput"
        @change="$emit('change', ($event.target as HTMLInputElement).value)"
        @blur="$emit('blur', $event)"
        @focus="$emit('focus', $event)"
        @keyup.enter="$emit('enter', $event)"
      />

      <!-- Suffix Slot (e.g. Password Toggle or Clear Button) -->
      <div
        v-if="$slots.suffix"
        class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 dark:text-dark-400"
      >
        <slot name="suffix"></slot>
      </div>
    </div>
    <!-- Hint / Error Text -->
    <p v-if="error" :id="`${inputId}-error`" class="input-error-text mt-1.5">
      {{ error }}
    </p>
    <p v-else-if="hint" :id="`${inputId}-hint`" class="input-hint mt-1.5">
      {{ hint }}
    </p>
  </div>
</template>

<script lang="ts">
// 模块级计数器：调用方不传 id 时自动生成唯一 id，保证 label 与 input 的关联不失效
let inputIdCounter = 0
</script>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  modelValue: string | number | null | undefined
  type?: string
  label?: string
  placeholder?: string
  disabled?: boolean
  required?: boolean
  readonly?: boolean
  error?: string
  hint?: string
  id?: string
  autocomplete?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false,
  required: false,
  readonly: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
  (e: 'blur', event: FocusEvent): void
  (e: 'focus', event: FocusEvent): void
  (e: 'enter', event: KeyboardEvent): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const placeholderText = computed(() => props.placeholder || '')

// 实际使用的 id：优先用调用方传入的，否则用自动生成的
const autoId = `app-input-${++inputIdCounter}`
const inputId = computed(() => props.id || autoId)

// 错误/提示文案与输入框的读屏关联：有错误时指向错误文案，否则指向提示文案
const describedBy = computed(() => {
  if (props.error) return `${inputId.value}-error`
  if (props.hint) return `${inputId.value}-hint`
  return undefined
})

const onInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  emit('update:modelValue', value)
}

// Expose focus method
defineExpose({
  focus: () => inputRef.value?.focus(),
  select: () => inputRef.value?.select()
})
</script>

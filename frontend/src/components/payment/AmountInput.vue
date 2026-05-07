<template>
  <div class="space-y-4">
    <!-- Quick Amount Buttons -->
    <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
      <div class="mb-3 flex items-center justify-between gap-3">
        <label class="block text-sm font-semibold text-gray-800 dark:text-gray-100">
          {{ t('payment.quickAmounts') }}
        </label>
        <span v-if="limitText" class="rounded-full bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-500 ring-1 ring-gray-100 dark:bg-dark-800 dark:text-dark-300 dark:ring-dark-700">
          {{ limitText }}
        </span>
      </div>
      <div class="grid grid-cols-3 gap-2 sm:grid-cols-5">
        <button
          v-for="amt in filteredAmounts"
          :key="amt"
          type="button"
          :class="[
            'min-h-[50px] rounded-lg border px-3 py-2 text-center font-semibold transition-colors',
            modelValue === amt
              ? 'border-gray-900 bg-gray-900 text-white shadow-sm dark:border-white dark:bg-white dark:text-gray-950'
              : 'border-gray-200 bg-white text-gray-700 hover:border-primary-300 hover:text-primary-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-primary-500/70 dark:hover:text-primary-300',
          ]"
          @click="selectAmount(amt)"
        >
          <span :class="['block text-xs', modelValue === amt ? 'text-white/65 dark:text-gray-600' : 'text-gray-400 dark:text-dark-400']">CNY</span>
          <span class="block text-base">{{ amt }}</span>
        </button>
      </div>
    </div>

    <!-- Custom Amount Input -->
    <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
      <label class="mb-2 block text-sm font-semibold text-gray-800 dark:text-gray-100">
        {{ t('payment.customAmount') }}
      </label>
      <div class="relative rounded-lg border border-gray-200 bg-white transition-colors focus-within:border-primary-400 focus-within:ring-2 focus-within:ring-primary-500/10 dark:border-dark-600 dark:bg-dark-900">
        <span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm font-semibold text-gray-400 dark:text-dark-400">
          CNY
        </span>
        <input
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="w-full rounded-lg border-0 bg-transparent py-3 pl-14 pr-4 text-lg font-semibold text-gray-900 outline-none placeholder:text-gray-300 focus:ring-0 dark:text-white dark:placeholder:text-dark-500"
          @input="handleInput"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(defineProps<{
  amounts?: number[]
  modelValue: number | null
  min?: number
  max?: number
}>(), {
  amounts: () => [10, 20, 50, 100, 200, 500, 1000, 2000, 5000],
  min: 0,
  max: 0,
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')

// 0 = no limit
const filteredAmounts = computed(() =>
  props.amounts.filter((a) => (props.min <= 0 || a >= props.min) && (props.max <= 0 || a <= props.max))
)

const formatLimit = (operator: '>=' | '<=', value: number) => `${operator} ${value}`

const placeholderText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return formatLimit('>=', props.min)
  if (props.max > 0) return formatLimit('<=', props.max)
  return t('payment.enterAmount')
})

const limitText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return formatLimit('>=', props.min)
  if (props.max > 0) return formatLimit('<=', props.max)
  return ''
})

const AMOUNT_PATTERN = /^\d*(\.\d{0,2})?$/

function selectAmount(amt: number) {
  customText.value = String(amt)
  emit('update:modelValue', amt)
}

function handleInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  if (!AMOUNT_PATTERN.test(val)) return
  customText.value = val
  if (val === '') {
    emit('update:modelValue', null)
    return
  }
  const num = parseFloat(val)
  if (!isNaN(num) && num > 0) {
    emit('update:modelValue', num)
  } else {
    emit('update:modelValue', null)
  }
}

watch(() => props.modelValue, (v) => {
  if (v !== null && String(v) !== customText.value) {
    customText.value = String(v)
  }
}, { immediate: true })
</script>

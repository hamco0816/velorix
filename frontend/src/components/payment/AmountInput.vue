<template>
  <div class="space-y-4">
    <!-- 快捷金额：不再嵌"卡中卡"，让父容器决定背景。3 列 / 4 列响应式，9 个金额完美对齐 -->
    <div>
      <div class="mb-2.5 flex items-center justify-between gap-3">
        <label class="block text-sm font-semibold text-gray-800 dark:text-gray-100">
          {{ t('payment.quickAmounts') }}
        </label>
        <span v-if="limitText" class="rounded-full bg-gray-50 px-2.5 py-0.5 text-xs font-medium tabular-nums text-gray-500 ring-1 ring-inset ring-gray-200/70 dark:bg-dark-800/60 dark:text-dark-300 dark:ring-dark-700/60">
          {{ limitText }}
        </span>
      </div>
      <!-- 3 列布局（9 个数据正好 3×3 对齐）；选中态用品牌橙描边，与支付方式选中态同语言 -->
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="amt in filteredAmounts"
          :key="amt"
          type="button"
          :class="[
            'relative flex flex-col items-center justify-center rounded-xl border px-3 py-2.5 text-center font-semibold transition-colors duration-150 ease-out',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600/40 dark:focus-visible:ring-brand-400/40',
            modelValue === amt
              ? 'border-brand-500 bg-brand-50 text-gray-900 ring-1 ring-inset ring-brand-500/40 dark:border-brand-400 dark:bg-brand-500/10 dark:text-white dark:ring-brand-400/30'
              : 'border-gray-200 bg-white text-gray-700 hover:border-gray-400 hover:text-gray-900 dark:border-dark-700/60 dark:bg-dark-800/60 dark:text-gray-200 dark:hover:border-dark-500 dark:hover:text-white',
          ]"
          @click="selectAmount(amt)"
        >
          <span :class="['text-2xs font-medium tracking-wide', modelValue === amt ? 'text-brand-600 dark:text-brand-400' : 'text-gray-400 dark:text-dark-400']">¥</span>
          <span class="text-base tabular-nums leading-tight">{{ amt }}</span>
        </button>
      </div>
    </div>

    <!-- 自定义金额：大号等宽数字输入，focus 用品牌橙描边 -->
    <div>
      <label class="mb-2 block text-sm font-semibold text-gray-800 dark:text-gray-100">
        {{ t('payment.customAmount') }}
      </label>
      <div class="relative rounded-xl border border-gray-200 bg-white transition-colors duration-150 ease-out focus-within:border-brand-500 focus-within:ring-2 focus-within:ring-brand-500/20 dark:border-dark-700/60 dark:bg-dark-800/60 dark:focus-within:border-brand-400 dark:focus-within:ring-brand-400/20">
        <span class="absolute left-4 top-1/2 -translate-y-1/2 text-base font-semibold text-gray-400 dark:text-dark-400">
          ¥
        </span>
        <input
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="w-full rounded-xl border-0 bg-transparent py-3 pl-10 pr-4 text-2xl font-semibold tabular-nums tracking-tight text-gray-900 outline-none placeholder:text-base placeholder:font-normal placeholder:text-gray-300 focus:ring-0 dark:text-white dark:placeholder:text-dark-500"
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

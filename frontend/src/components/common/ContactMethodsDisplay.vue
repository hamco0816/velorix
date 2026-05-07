<template>
  <div
    v-if="resolvedMethods.length"
    :class="[
      compact ? 'flex flex-wrap items-center gap-1.5' : 'space-y-2',
    ]"
  >
    <component
      :is="method.url ? 'a' : 'div'"
      v-for="(method, index) in resolvedMethods"
      :key="`${method.type}-${method.value}-${index}`"
      :href="method.url || undefined"
      :target="method.url ? '_blank' : undefined"
      :rel="method.url ? 'noopener noreferrer' : undefined"
      :class="[
        'inline-flex min-w-0 items-center gap-2 rounded-lg border border-gray-200 bg-white text-gray-800 shadow-sm dark:border-dark-700 dark:bg-dark-800 dark:text-dark-100',
        compact ? 'max-w-full px-2 py-1 text-xs' : 'w-full px-3 py-2 text-sm',
        method.url && 'transition-colors hover:border-primary-300 hover:text-primary-700 dark:hover:border-primary-700 dark:hover:text-primary-300',
      ]"
    >
      <ContactMethodIcon :type="method.type" :size="compact ? '18px' : '24px'" />
      <span v-if="showLabel" class="shrink-0 font-semibold">
        {{ method.label || defaultContactMethodLabel(method.type, locale) }}
      </span>
      <span class="min-w-0 truncate font-medium">
        {{ method.value || method.url }}
      </span>
    </component>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ContactMethod } from '@/types'
import ContactMethodIcon from '@/components/common/ContactMethodIcon.vue'
import { defaultContactMethodLabel, normalizeContactMethods } from '@/utils/contactMethods'

const props = withDefaults(defineProps<{
  methods?: ContactMethod[]
  legacyInfo?: string
  compact?: boolean
  showLabel?: boolean
}>(), {
  methods: () => [],
  legacyInfo: '',
  compact: false,
  showLabel: true,
})

const { locale } = useI18n()

const resolvedMethods = computed(() =>
  normalizeContactMethods(props.methods, props.legacyInfo, locale.value),
)
</script>

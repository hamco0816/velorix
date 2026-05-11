<template>
  <div
    v-if="resolvedMethods.length"
    :class="[
      compact ? 'flex flex-wrap items-center gap-2' : 'space-y-2',
    ]"
  >
    <!-- 三种点击行为：
         1) 带 image_data（如 QQ群二维码）：点击弹出二维码大图 modal
         2) 带 url：跳转外部链接
         3) 纯文本：不可点击，靠 title 提示
         qrTarget !== null 时启用 teleport modal 展示当前条目的图 -->
    <component
      :is="getTag(method)"
      v-for="(method, index) in resolvedMethods"
      :key="`${method.type}-${method.value || 'img'}-${index}`"
      :href="!method.image_data && method.url ? method.url : undefined"
      :target="!method.image_data && method.url ? '_blank' : undefined"
      :rel="!method.image_data && method.url ? 'noopener noreferrer' : undefined"
      :type="method.image_data ? 'button' : undefined"
      :title="contactTitle(method)"
      :aria-label="contactTitle(method)"
      :class="[
        'inline-flex min-w-0 items-center gap-2 rounded-lg border border-gray-200 bg-white text-gray-800 shadow-sm dark:border-dark-700 dark:bg-dark-800 dark:text-dark-100',
        iconOnly
          ? 'h-10 w-10 justify-center p-0'
          : compact
            ? 'max-w-full px-2.5 py-1.5 text-xs'
            : 'w-full px-3 py-2 text-sm',
        (method.url || method.image_data) && 'cursor-pointer transition-colors hover:border-primary-300 hover:text-primary-700 dark:hover:border-primary-700 dark:hover:text-primary-300',
      ]"
      @click="method.image_data ? (qrTarget = method) : undefined"
    >
      <ContactMethodIcon :type="method.type" :size="iconOnly ? '24px' : compact ? '18px' : '24px'" />
      <span v-if="!iconOnly && showLabel" class="shrink-0 font-semibold">
        {{ method.label || defaultContactMethodLabel(method.type, locale) }}
      </span>
      <span v-if="!iconOnly" class="min-w-0 truncate font-medium">
        {{ method.value || method.url || (method.image_data ? t('contactMethods.viewQrCode') : '') }}
      </span>
      <!-- 二维码角标：明示"这条点击会弹二维码"，避免用户以为是普通链接 -->
      <span
        v-if="method.image_data && !iconOnly"
        class="ml-auto inline-flex shrink-0 items-center rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300"
      >
        QR
      </span>
    </component>

    <!-- 二维码大图 modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="qrTarget"
          class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4"
          @click.self="qrTarget = null"
        >
          <div class="w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
            <div class="mb-4 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <ContactMethodIcon :type="qrTarget.type" size="22px" />
                <span class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ qrTarget.label || defaultContactMethodLabel(qrTarget.type, locale) }}
                </span>
              </div>
              <button
                class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                :aria-label="t('common.close')"
                @click="qrTarget = null"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <img
              :src="qrTarget.image_data"
              :alt="qrTarget.label || 'QR'"
              class="mx-auto block max-h-80 w-full rounded-lg bg-white object-contain"
            />
            <p v-if="qrTarget.value" class="mt-3 break-all text-center text-sm text-gray-600 dark:text-gray-300">
              {{ qrTarget.value }}
            </p>
            <p class="mt-2 text-center text-xs text-gray-400">
              {{ t('contactMethods.scanQrHint') }}
            </p>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ContactMethod } from '@/types'
import ContactMethodIcon from '@/components/common/ContactMethodIcon.vue'
import { defaultContactMethodLabel, normalizeContactMethods } from '@/utils/contactMethods'

const props = withDefaults(defineProps<{
  methods?: ContactMethod[]
  legacyInfo?: string
  compact?: boolean
  showLabel?: boolean
  iconOnly?: boolean
}>(), {
  methods: () => [],
  legacyInfo: '',
  compact: false,
  showLabel: true,
  iconOnly: false,
})

const { t, locale } = useI18n()

const resolvedMethods = computed(() =>
  normalizeContactMethods(props.methods, props.legacyInfo, locale.value),
)

const qrTarget = ref<ContactMethod | null>(null)

function getTag(method: ContactMethod): string {
  if (method.image_data) return 'button'
  if (method.url) return 'a'
  return 'div'
}

function contactTitle(method: ContactMethod): string {
  const label = method.label || defaultContactMethodLabel(method.type, locale.value)
  const value = method.value || method.url || ''
  if (method.image_data) {
    return value ? `${label}: ${value} (${t('contactMethods.viewQrCode')})` : `${label}: ${t('contactMethods.viewQrCode')}`
  }
  return value ? `${label}: ${value}` : label
}
</script>

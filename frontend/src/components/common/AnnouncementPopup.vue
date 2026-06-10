<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="announcementStore.currentPopup"
        class="fixed inset-0 z-[120] flex items-start justify-center overflow-y-auto bg-black/60 p-4 pt-[8vh]"
      >
        <div
          class="w-full max-w-[680px] overflow-hidden rounded-md border border-gray-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-800"
          @click.stop
        >
          <!-- Header（保持公告语义的暖色但去除装饰渐变球） -->
          <div class="border-b border-gray-200 bg-white px-8 py-5 dark:border-dark-700 dark:bg-dark-800">
            <div class="mb-3 flex items-center gap-2">
              <div class="flex h-9 w-9 items-center justify-center rounded bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-400">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
              </div>
              <span class="inline-flex items-center gap-1.5 rounded bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
                <span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>
                {{ t('announcements.unread') }}
              </span>
            </div>

            <!-- Title -->
            <h2 class="mb-2 text-xl font-semibold leading-tight text-gray-900 dark:text-white">
              {{ announcementStore.currentPopup.title }}
            </h2>

            <!-- Time -->
            <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
              <Icon name="clock" size="sm" :stroke-width="1.8" />
              <time>{{ formatRelativeWithDateTime(announcementStore.currentPopup.created_at) }}</time>
            </div>
          </div>

          <!-- Body -->
          <div class="max-h-[50vh] overflow-y-auto bg-white px-8 py-8 dark:bg-dark-800">
            <div class="border-l-2 border-amber-500 pl-6">
              <div
                class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                v-html="renderedContent"
              ></div>
            </div>
          </div>

          <!-- Footer -->
          <div class="border-t border-gray-200 bg-gray-50 px-8 py-4 dark:border-dark-700 dark:bg-dark-900/40">
            <div class="flex items-center justify-end">
              <button
                @click="handleDismiss"
                class="inline-flex items-center gap-2 rounded bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
              >
                <Icon name="check" size="sm" :stroke-width="2" />
                {{ t('announcements.markRead') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = announcementStore.currentPopup?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  announcementStore.dismissPopup()
}

// Manage body overflow — only set, never unset (bell component handles restore)
watch(
  () => announcementStore.currentPopup,
  (popup) => {
    if (popup) {
      document.body.style.overflow = 'hidden'
    }
  }
)
</script>

<style scoped>
.popup-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.popup-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
}

.popup-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.popup-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #374151);
}
</style>

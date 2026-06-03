<template>
  <div
    :class="standalone
      ? 'flex h-screen min-h-0 flex-col overflow-hidden bg-white text-gray-900 dark:bg-dark-950 dark:text-white'
      : 'relative'"
  >
    <button
      v-if="!standalone"
      type="button"
      :class="triggerClass"
      :aria-label="t('support.openChat')"
      @click="openChatWindow"
    >
      <Icon name="chatBubble" size="sm" />
      <span :class="embedded ? '' : 'hidden sm:inline'">{{ t('support.onlineSupport') }}</span>
      <span v-if="embedded" class="ml-auto text-xs text-gray-400 dark:text-dark-500">{{ t('support.openChat') }}</span>
      <span
        v-if="unreadCount > 0"
        class="absolute -right-1 -top-1 min-w-[18px] rounded-full bg-rose-500 px-1 text-center text-[10px] font-semibold leading-[18px] text-white"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
    </button>

    <section
      v-else
      class="flex h-screen min-h-0 flex-col overflow-hidden bg-white dark:bg-dark-900"
    >
      <header class="flex items-center justify-between border-b border-gray-100 px-4 py-3 dark:border-dark-800">
        <div>
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('support.onlineSupport') }}</span>
            <span
              class="h-2 w-2 rounded-full"
              :class="wsStatus === 'connected' ? 'bg-emerald-500' : 'bg-amber-500'"
            ></span>
          </div>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
            {{ wsStatusLabel }}
          </p>
        </div>
        <button
          type="button"
          class="btn-ghost btn-icon"
          :aria-label="t('common.close')"
          @click="closeChatWindow"
        >
          <Icon name="x" size="sm" />
        </button>
      </header>

      <div ref="messageListRef" class="min-h-0 flex-1 space-y-3 overflow-y-auto bg-gray-50 px-4 py-4 dark:bg-dark-950">
        <div v-if="loading" class="py-8 text-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('common.loading') }}
        </div>

        <div v-else-if="messages.length === 0" class="rounded-lg border border-gray-200 bg-white p-4 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300">
          <div class="mb-2 flex items-center gap-2 font-semibold text-gray-900 dark:text-white">
            <Icon name="chat" size="sm" />
            {{ t('support.welcomeTitle') }}
          </div>
          <p class="leading-6">{{ t('support.welcomeText') }}</p>
        </div>

        <div
          v-for="message in messages"
          :key="message.id"
          class="flex"
          :class="message.sender_type === 'user' ? 'justify-end' : 'justify-start'"
        >
          <div
            class="max-w-[82%] rounded-lg px-3 py-2 text-sm shadow-sm"
            :class="message.sender_type === 'user'
              ? 'bg-primary-600 text-white'
              : 'border border-gray-200 bg-white text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-white'"
          >
            <div class="whitespace-pre-wrap break-words leading-6">{{ message.content }}</div>
            <div
              class="mt-1 text-[11px]"
              :class="message.sender_type === 'user' ? 'text-primary-100' : 'text-gray-400 dark:text-dark-400'"
            >
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>
        </div>
      </div>

      <footer class="border-t border-gray-100 bg-white p-3 dark:border-dark-800 dark:bg-dark-900">
        <form class="flex items-end gap-2" @submit.prevent="handleSend">
          <textarea
            v-model="draft"
            rows="1"
            class="input max-h-28 min-h-[40px] resize-none"
            :placeholder="t('support.inputPlaceholder')"
            :disabled="sending"
            @keydown.enter.exact.prevent="handleSend"
          ></textarea>
          <button
            type="submit"
            class="btn btn-primary h-10 shrink-0 px-3"
            :disabled="sending || !draft.trim()"
            :title="t('support.send')"
          >
            <Icon name="arrowRight" size="sm" />
          </button>
        </form>
        <p v-if="errorMessage" class="mt-2 text-xs text-rose-600 dark:text-rose-400">
          {{ errorMessage }}
        </p>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import supportAPI from '@/api/support'
import type { SupportConversation, SupportMessage, SupportRealtimeEvent, SupportWSStatus } from '@/api/support'
import { formatDateTime, formatRelativeTime } from '@/utils/format'

const props = withDefaults(defineProps<{
  embedded?: boolean
  standalone?: boolean
}>(), {
  embedded: false,
  standalone: false
})
const emit = defineEmits<{
  (e: 'opened'): void
}>()

const { t } = useI18n()
const router = useRouter()

const messageListRef = ref<HTMLElement | null>(null)
const loading = ref(false)
const sending = ref(false)
const draft = ref('')
const messages = ref<SupportMessage[]>([])
const conversation = ref<SupportConversation | null>(null)
const unreadCount = ref(0)
const errorMessage = ref('')
const wsStatus = ref<SupportWSStatus>('closed')
let closeWS: (() => void) | null = null
let readTimer: ReturnType<typeof setTimeout> | null = null

const embedded = computed(() => props.embedded)
const standalone = computed(() => props.standalone)
const triggerClass = computed(() => {
  if (embedded.value) {
    return 'relative flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 hover:text-gray-900 dark:text-dark-200 dark:hover:bg-dark-800 dark:hover:text-white'
  }
  return 'relative flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white'
})

const wsStatusLabel = computed(() => {
  switch (wsStatus.value) {
    case 'connected':
      return t('support.status.connected')
    case 'reconnecting':
      return t('support.status.reconnecting')
    case 'offline':
      return t('support.status.offline')
    case 'connecting':
      return t('support.status.connecting')
    default:
      return t('support.status.closed')
  }
})

function openChatWindow() {
  const width = 430
  const height = 650
  const left = Math.max(0, Math.round(window.screenX + window.outerWidth - width - 32))
  const top = Math.max(0, Math.round(window.screenY + 72))
  const features = [
    'popup=yes',
    'resizable=yes',
    'scrollbars=no',
    `width=${width}`,
    `height=${height}`,
    `left=${left}`,
    `top=${top}`,
  ].join(',')
  const href = router.resolve({ name: 'SupportChatWindow' }).href
  const popup = window.open(href, 'velorix_support_chat', features)
  popup?.focus()
  emit('opened')
}

function closeChatWindow() {
  window.close()
}

async function loadConversation() {
  loading.value = true
  errorMessage.value = ''
  try {
    const detail = await supportAPI.getConversation()
    conversation.value = detail.conversation
    messages.value = detail.messages || []
    unreadCount.value = detail.conversation?.user_unread_count || 0
    await nextTick()
    scrollToBottom()
    await markReadQuiet()
  } catch (error: any) {
    errorMessage.value = error?.message || t('support.loadFailed')
  } finally {
    loading.value = false
  }
}

async function handleSend() {
  const content = draft.value.trim()
  if (!content || sending.value) return

  sending.value = true
  errorMessage.value = ''
  try {
    draft.value = ''
    const result = await supportAPI.sendMessage(content)
    conversation.value = result.conversation
    mergeMessage(result.message)
    await nextTick()
    scrollToBottom()
  } catch (error: any) {
    draft.value = content
    errorMessage.value = error?.message || t('support.sendFailed')
  } finally {
    sending.value = false
  }
}

function mergeMessage(message?: SupportMessage) {
  if (!message) return
  if (messages.value.some((item) => item.id === message.id)) return
  messages.value = [...messages.value, message].sort((a, b) => a.id - b.id)
}

function handleRealtime(event: SupportRealtimeEvent) {
  if (event.conversation) {
    conversation.value = event.conversation
    unreadCount.value = event.conversation.user_unread_count || 0
  }
  if (event.type === 'support.ready' && typeof event.unread_count === 'number') {
    unreadCount.value = event.unread_count
  }
  if (event.type === 'support.message' && event.message) {
    mergeMessage(event.message)
    if (event.message.sender_type === 'admin' && !standalone.value) {
      unreadCount.value = event.conversation?.user_unread_count ?? unreadCount.value + 1
    }
    if (standalone.value) {
      scheduleMarkRead()
      nextTick(scrollToBottom)
    }
  }
  if (event.type === 'support.unread' && typeof event.unread_count === 'number') {
    unreadCount.value = event.unread_count
  }
}

function scheduleMarkRead() {
  if (readTimer) clearTimeout(readTimer)
  readTimer = setTimeout(() => {
    markReadQuiet()
  }, 300)
}

async function markReadQuiet() {
  if (!conversation.value) return
  try {
    const updated = await supportAPI.markRead()
    conversation.value = updated
    unreadCount.value = updated.user_unread_count || 0
  } catch {
    // Ignore read marker failures; messages remain visible and durable.
  }
}

function scrollToBottom() {
  const el = messageListRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

function formatMessageTime(value: string): string {
  return formatRelativeTime(value) || formatDateTime(value)
}

onMounted(() => {
  closeWS = supportAPI.openWebSocket({
    onMessage: handleRealtime,
    onStatus: (status) => {
      wsStatus.value = status
    }
  })
  if (standalone.value) {
    loadConversation()
  }
})

onBeforeUnmount(() => {
  if (readTimer) clearTimeout(readTimer)
  closeWS?.()
})
</script>

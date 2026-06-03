<template>
  <AppLayout wide>
    <div class="grid h-[calc(100vh-8rem)] min-h-[620px] grid-cols-1 gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
      <aside class="flex min-h-0 flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <div class="border-b border-gray-100 p-4 dark:border-dark-800">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('support.admin.conversations') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ wsStatusLabel }}</p>
            </div>
            <button class="btn btn-secondary btn-sm" :disabled="loading" @click="loadConversations">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>

          <div class="mt-3 grid grid-cols-[1fr_112px] gap-2">
            <input
              v-model="search"
              type="text"
              class="input"
              :placeholder="t('support.admin.searchPlaceholder')"
              @keyup.enter="loadConversations"
            />
            <select v-model="statusFilter" class="input" @change="loadConversations">
              <option value="open">{{ t('support.statuses.open') }}</option>
              <option value="closed">{{ t('support.statuses.closed') }}</option>
              <option value="all">{{ t('common.all') }}</option>
            </select>
          </div>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto p-2">
          <div v-if="loading && conversations.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('common.loading') }}
          </div>
          <div v-else-if="conversations.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('support.admin.noConversations') }}
          </div>
          <button
            v-for="item in conversations"
            :key="item.id"
            type="button"
            class="mb-1 w-full rounded-lg px-3 py-3 text-left transition-colors"
            :class="selectedConversation?.id === item.id
              ? 'bg-primary-50 text-primary-900 ring-1 ring-inset ring-primary-200 dark:bg-primary-500/15 dark:text-primary-100 dark:ring-primary-500/30'
              : 'hover:bg-gray-50 dark:hover:bg-dark-800'"
            @click="selectConversation(item)"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                  {{ displayUser(item) }}
                </div>
                <div class="mt-1 truncate text-xs text-gray-500 dark:text-dark-400">
                  #{{ item.id }} - {{ item.last_message_at ? formatRelativeTime(item.last_message_at) : formatRelativeTime(item.created_at) }}
                </div>
              </div>
              <span
                v-if="item.admin_unread_count > 0"
                class="min-w-[20px] rounded-full bg-rose-500 px-1.5 text-center text-[11px] font-semibold leading-5 text-white"
              >
                {{ item.admin_unread_count > 99 ? '99+' : item.admin_unread_count }}
              </span>
            </div>
            <p class="mt-2 line-clamp-2 text-xs leading-5 text-gray-600 dark:text-dark-300">
              {{ item.last_message || t('support.admin.emptyPreview') }}
            </p>
          </button>
        </div>
      </aside>

      <section class="flex min-h-0 flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <template v-if="selectedConversation">
          <header class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 px-5 py-4 dark:border-dark-800">
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ displayUser(selectedConversation) }}</h2>
                <span
                  class="rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="selectedConversation.status === 'open'
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
                    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
                >
                  {{ selectedConversation.status === 'open' ? t('support.statuses.open') : t('support.statuses.closed') }}
                </span>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                {{ selectedConversation.user_email || `User #${selectedConversation.user_id}` }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <button class="btn btn-secondary btn-sm" @click="markSelectedRead">
                <Icon name="check" size="sm" />
                {{ t('support.admin.markRead') }}
              </button>
              <button
                v-if="selectedConversation.status === 'open'"
                class="btn btn-secondary btn-sm"
                @click="closeSelected"
              >
                {{ t('support.admin.closeConversation') }}
              </button>
              <button v-else class="btn btn-primary btn-sm" @click="reopenSelected">
                {{ t('support.admin.reopenConversation') }}
              </button>
            </div>
          </header>

          <div ref="messageListRef" class="min-h-0 flex-1 space-y-3 overflow-y-auto bg-gray-50 px-5 py-5 dark:bg-dark-950">
            <div v-if="messageLoading" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
              {{ t('common.loading') }}
            </div>
            <div
              v-for="message in messages"
              :key="message.id"
              class="flex"
              :class="message.sender_type === 'admin' ? 'justify-end' : 'justify-start'"
            >
              <div
                class="max-w-[72%] rounded-lg px-3 py-2 text-sm shadow-sm"
                :class="message.sender_type === 'admin'
                  ? 'bg-primary-600 text-white'
                  : 'border border-gray-200 bg-white text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-white'"
              >
                <div class="whitespace-pre-wrap break-words leading-6">{{ message.content }}</div>
                <div
                  class="mt-1 text-[11px]"
                  :class="message.sender_type === 'admin' ? 'text-primary-100' : 'text-gray-400 dark:text-dark-400'"
                >
                  {{ formatDateTime(message.created_at) }}
                </div>
              </div>
            </div>
          </div>

          <footer class="border-t border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
            <form class="flex items-end gap-2" @submit.prevent="sendReply">
              <textarea
                v-model="reply"
                rows="2"
                class="input max-h-32 resize-none"
                :placeholder="t('support.admin.replyPlaceholder')"
                :disabled="sending || selectedConversation.status !== 'open'"
                @keydown.enter.exact.prevent="sendReply"
              ></textarea>
              <button
                type="submit"
                class="btn btn-primary h-10 shrink-0"
                :disabled="sending || !reply.trim() || selectedConversation.status !== 'open'"
              >
                <Icon name="arrowRight" size="sm" />
                {{ t('support.send') }}
              </button>
            </form>
            <p v-if="errorMessage" class="mt-2 text-xs text-rose-600 dark:text-rose-400">{{ errorMessage }}</p>
          </footer>
        </template>

        <div v-else class="flex flex-1 items-center justify-center p-8 text-center text-sm text-gray-500 dark:text-dark-400">
          {{ t('support.admin.selectConversation') }}
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { SupportConversation, SupportMessage, SupportRealtimeEvent, SupportWSStatus } from '@/api/admin/support'
import { formatDateTime, formatRelativeTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const conversations = ref<SupportConversation[]>([])
const selectedConversation = ref<SupportConversation | null>(null)
const messages = ref<SupportMessage[]>([])
const loading = ref(false)
const messageLoading = ref(false)
const sending = ref(false)
const search = ref('')
const statusFilter = ref('open')
const reply = ref('')
const errorMessage = ref('')
const wsStatus = ref<SupportWSStatus>('closed')
const messageListRef = ref<HTMLElement | null>(null)
let closeWS: (() => void) | null = null

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

async function loadConversations() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = await adminAPI.support.listConversations(1, 50, {
      status: statusFilter.value,
      search: search.value.trim() || undefined
    })
    conversations.value = result.items || []
    if (selectedConversation.value) {
      selectedConversation.value = conversations.value.find((item) => item.id === selectedConversation.value?.id) || selectedConversation.value
    }
  } catch (error: any) {
    errorMessage.value = error?.message || t('support.admin.loadFailed')
  } finally {
    loading.value = false
  }
}

async function selectConversation(item: SupportConversation) {
  selectedConversation.value = item
  messageLoading.value = true
  errorMessage.value = ''
  try {
    messages.value = await adminAPI.support.listMessages(item.id)
    await nextTick()
    scrollToBottom()
    await markSelectedRead()
  } catch (error: any) {
    errorMessage.value = error?.message || t('support.admin.loadMessagesFailed')
  } finally {
    messageLoading.value = false
  }
}

async function sendReply() {
  if (!selectedConversation.value || sending.value) return
  const content = reply.value.trim()
  if (!content) return

  sending.value = true
  errorMessage.value = ''
  try {
    reply.value = ''
    const result = await adminAPI.support.sendMessage(selectedConversation.value.id, content)
    mergeConversation(result.conversation)
    mergeMessage(result.message)
    await nextTick()
    scrollToBottom()
  } catch (error: any) {
    reply.value = content
    errorMessage.value = error?.message || t('support.sendFailed')
  } finally {
    sending.value = false
  }
}

async function markSelectedRead() {
  if (!selectedConversation.value) return
  try {
    const updated = await adminAPI.support.markRead(selectedConversation.value.id)
    mergeConversation(updated)
  } catch {
    // Keep the selected conversation usable even if read state update fails.
  }
}

async function closeSelected() {
  if (!selectedConversation.value) return
  errorMessage.value = ''
  try {
    const updated = await adminAPI.support.closeConversation(selectedConversation.value.id)
    mergeConversation(updated)
  } catch (error: any) {
    errorMessage.value = error?.message || t('support.admin.closeFailed')
  }
}

async function reopenSelected() {
  if (!selectedConversation.value) return
  errorMessage.value = ''
  try {
    const updated = await adminAPI.support.reopenConversation(selectedConversation.value.id)
    mergeConversation(updated)
  } catch (error: any) {
    errorMessage.value = error?.message || t('support.admin.reopenFailed')
  }
}

function handleRealtime(event: SupportRealtimeEvent) {
  if (event.conversation) {
    mergeConversation(event.conversation)
  }
  if (event.type === 'support.message' && event.message) {
      if (selectedConversation.value?.id === event.message.conversation_id) {
      mergeMessage(event.message)
      nextTick(scrollToBottom)
      if (event.message.sender_type === 'user' && isWorkbenchActive()) {
        markSelectedRead()
      }
    }
  }
}

function mergeConversation(conversation?: SupportConversation) {
  if (!conversation) return
  if (selectedConversation.value?.id === conversation.id) {
    selectedConversation.value = conversation
  }
  const next = conversations.value.filter((item) => item.id !== conversation.id)
  if (statusFilter.value === 'all' || conversation.status === statusFilter.value) {
    next.unshift(conversation)
  }
  conversations.value = next.sort((a, b) => {
    if (a.admin_unread_count !== b.admin_unread_count) return b.admin_unread_count - a.admin_unread_count
    const at = new Date(a.last_message_at || a.updated_at).getTime()
    const bt = new Date(b.last_message_at || b.updated_at).getTime()
    return bt - at
  })
}

function mergeMessage(message?: SupportMessage) {
  if (!message) return
  if (messages.value.some((item) => item.id === message.id)) return
  messages.value = [...messages.value, message].sort((a, b) => a.id - b.id)
}

function scrollToBottom() {
  const el = messageListRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

function displayUser(item: SupportConversation): string {
  return item.username || item.user_email || `User #${item.user_id}`
}

function isWorkbenchActive(): boolean {
  return document.visibilityState === 'visible' && document.hasFocus()
}

onMounted(() => {
  loadConversations()
  closeWS = adminAPI.support.openWebSocket({
    onMessage: handleRealtime,
    onStatus: (status) => {
      wsStatus.value = status
    }
  })
})

onBeforeUnmount(() => {
  closeWS?.()
})
</script>

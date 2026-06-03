<template></template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { adminAPI } from '@/api/admin'
import type { SupportConversation, SupportMessage, SupportRealtimeEvent } from '@/api/admin/support'

const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()

let closeWS: (() => void) | null = null
let permissionListenerInstalled = false
const notifiedMessageIds = new Set<number>()

function supportsBrowserNotification(): boolean {
  return typeof window !== 'undefined' && 'Notification' in window
}

function requestNotificationPermission() {
  if (!supportsBrowserNotification() || Notification.permission !== 'default') return
  Notification.requestPermission().catch(() => {
    // Browsers may reject permission prompts outside supported contexts.
  })
}

function installPermissionGestureListener() {
  if (!supportsBrowserNotification() || Notification.permission !== 'default' || permissionListenerInstalled) return
  permissionListenerInstalled = true
  document.addEventListener('click', handlePermissionGesture, { once: true })
  document.addEventListener('keydown', handlePermissionGesture, { once: true })
}

function removePermissionGestureListener() {
  if (!permissionListenerInstalled) return
  permissionListenerInstalled = false
  document.removeEventListener('click', handlePermissionGesture)
  document.removeEventListener('keydown', handlePermissionGesture)
}

function handlePermissionGesture() {
  removePermissionGestureListener()
  requestNotificationPermission()
}

function handleRealtime(event: SupportRealtimeEvent) {
  if (event.type !== 'support.message' || !event.message || event.message.sender_type !== 'user') {
    return
  }
  notifySupportMessage(event.conversation, event.message)
}

function notifySupportMessage(conversation: SupportConversation | undefined, message: SupportMessage) {
  if (notifiedMessageIds.has(message.id)) return
  notifiedMessageIds.add(message.id)
  if (notifiedMessageIds.size > 200) {
    const first = notifiedMessageIds.values().next().value
    if (typeof first === 'number') notifiedMessageIds.delete(first)
  }

  const sender = displayUser(conversation, message)
  const body = truncateMessage(message.content)
  const title = `${t('support.admin.title')} - ${sender}`

  if (supportsBrowserNotification() && Notification.permission === 'granted') {
    const notification = new Notification(title, {
      body,
      tag: `support-conversation-${message.conversation_id}`,
      icon: appStore.siteLogo || undefined
    })
    notification.onclick = () => {
      window.focus()
      notification.close()
      router.push('/admin/support')
    }
    return
  }

  installPermissionGestureListener()
  appStore.showInfo(`${sender}: ${body}`, 6000)
}

function displayUser(conversation: SupportConversation | undefined, message: SupportMessage): string {
  return conversation?.username || conversation?.user_email || `User #${conversation?.user_id || message.sender_id || message.conversation_id}`
}

function truncateMessage(value: string): string {
  const normalized = value.replace(/\s+/g, ' ').trim()
  if (normalized.length <= 120) return normalized
  return `${normalized.slice(0, 120)}...`
}

onMounted(() => {
  installPermissionGestureListener()
  closeWS = adminAPI.support.openWebSocket({
    onMessage: handleRealtime
  })
})

onBeforeUnmount(() => {
  removePermissionGestureListener()
  closeWS?.()
})
</script>

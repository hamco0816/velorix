import { apiClient } from './client'

export type SupportConversationStatus = 'open' | 'closed'
export type SupportSenderType = 'user' | 'admin' | 'system'
export type SupportWSStatus = 'connecting' | 'connected' | 'reconnecting' | 'offline' | 'closed'

export interface SupportConversation {
  id: number
  user_id: number
  user_email?: string
  username?: string
  status: SupportConversationStatus
  subject: string
  last_message: string
  last_message_at?: string | null
  user_unread_count: number
  admin_unread_count: number
  user_last_read_at?: string | null
  admin_last_read_at?: string | null
  closed_at?: string | null
  created_at: string
  updated_at: string
}

export interface SupportMessage {
  id: number
  conversation_id: number
  sender_type: SupportSenderType
  sender_id?: number | null
  content: string
  created_at: string
}

export interface SupportConversationDetail {
  conversation: SupportConversation | null
  messages: SupportMessage[]
}

export interface SupportMessageResult {
  conversation: SupportConversation
  message: SupportMessage
}

export interface SupportRealtimeEvent {
  type: 'support.ready' | 'support.message' | 'support.conversation_updated' | 'support.unread' | string
  conversation_id?: number
  user_id?: number
  message?: SupportMessage
  conversation?: SupportConversation
  unread_count?: number
  timestamp?: string
}

interface SupportWSOptions {
  token?: string
  wsBaseUrl?: string
  admin?: boolean
  onMessage: (event: SupportRealtimeEvent) => void
  onStatus?: (status: SupportWSStatus) => void
  onError?: (error: Event) => void
}

const SUPPORT_WS_PROTOCOL = 'sub2api-support'
const SUPPORT_ADMIN_WS_PROTOCOL = 'sub2api-support-admin'

export async function getConversation(limit = 50): Promise<SupportConversationDetail> {
  const { data } = await apiClient.get<SupportConversationDetail>('/support/conversation', {
    params: { limit }
  })
  return data
}

export async function listMessages(beforeId?: number, limit = 50): Promise<SupportConversationDetail> {
  const { data } = await apiClient.get<SupportConversationDetail>('/support/messages', {
    params: {
      before_id: beforeId || undefined,
      limit
    }
  })
  return data
}

export async function sendMessage(content: string): Promise<SupportMessageResult> {
  const { data } = await apiClient.post<SupportMessageResult>('/support/messages', { content })
  return data
}

export async function markRead(): Promise<SupportConversation> {
  const { data } = await apiClient.post<SupportConversation>('/support/read')
  return data
}

export function openSupportWebSocket(options: SupportWSOptions): () => void {
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  let shouldReconnect = true
  let connectedOnce = false

  const setStatus = (status: SupportWSStatus) => options.onStatus?.(status)
  const clearReconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
  }

  const scheduleReconnect = () => {
    if (!shouldReconnect) return
    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 15000) + Math.floor(Math.random() * 300)
    reconnectAttempts += 1
    setStatus(connectedOnce ? 'reconnecting' : 'connecting')
    clearReconnect()
    reconnectTimer = setTimeout(connect, delay)
  }

  const connect = () => {
    if (!shouldReconnect) return
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsBaseUrl = options.wsBaseUrl || import.meta.env.VITE_WS_BASE_URL || window.location.host
    const path = options.admin ? '/api/v1/admin/support/ws' : '/api/v1/support/ws'
    const wsURL = new URL(`${protocol}//${wsBaseUrl}${path}`)
    const rawToken = String(options.token ?? localStorage.getItem('auth_token') ?? '').trim()
    const protocols = [options.admin ? SUPPORT_ADMIN_WS_PROTOCOL : SUPPORT_WS_PROTOCOL]
    if (rawToken) protocols.push(`jwt.${rawToken}`)

    setStatus(connectedOnce ? 'reconnecting' : 'connecting')
    ws = new WebSocket(wsURL.toString(), protocols)

    ws.onopen = () => {
      connectedOnce = true
      reconnectAttempts = 0
      clearReconnect()
      setStatus('connected')
    }

    ws.onmessage = (event) => {
      try {
        options.onMessage(JSON.parse(event.data) as SupportRealtimeEvent)
      } catch (error) {
        console.warn('[SupportWS] Failed to parse message:', error)
      }
    }

    ws.onerror = (error) => {
      options.onError?.(error)
    }

    ws.onclose = () => {
      ws = null
      scheduleReconnect()
    }
  }

  const handleOnline = () => connect()
  const handleOffline = () => setStatus('offline')

  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
  connect()

  return () => {
    shouldReconnect = false
    window.removeEventListener('online', handleOnline)
    window.removeEventListener('offline', handleOffline)
    clearReconnect()
    if (ws) ws.close()
    ws = null
    setStatus('closed')
  }
}

export const supportAPI = {
  getConversation,
  listMessages,
  sendMessage,
  markRead,
  openWebSocket: openSupportWebSocket
}

export default supportAPI

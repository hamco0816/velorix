import { apiClient } from '../client'
import { openSupportWebSocket } from '../support'
import type {
  SupportConversation,
  SupportMessage,
  SupportMessageResult,
  SupportRealtimeEvent,
  SupportWSStatus
} from '../support'
import type { BasePaginationResponse } from '@/types'

export type {
  SupportConversation,
  SupportMessage,
  SupportMessageResult,
  SupportRealtimeEvent,
  SupportWSStatus
}

export async function listConversations(
  page = 1,
  pageSize = 20,
  filters?: {
    status?: string
    search?: string
  }
): Promise<BasePaginationResponse<SupportConversation>> {
  const { data } = await apiClient.get<BasePaginationResponse<SupportConversation>>('/admin/support/conversations', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    }
  })
  return data
}

export async function listMessages(conversationId: number, beforeId?: number, limit = 50): Promise<SupportMessage[]> {
  const { data } = await apiClient.get<SupportMessage[]>(`/admin/support/conversations/${conversationId}/messages`, {
    params: {
      before_id: beforeId || undefined,
      limit
    }
  })
  return data
}

export async function sendMessage(conversationId: number, content: string): Promise<SupportMessageResult> {
  const { data } = await apiClient.post<SupportMessageResult>(`/admin/support/conversations/${conversationId}/messages`, {
    content
  })
  return data
}

export async function markRead(conversationId: number): Promise<SupportConversation> {
  const { data } = await apiClient.post<SupportConversation>(`/admin/support/conversations/${conversationId}/read`)
  return data
}

export async function closeConversation(conversationId: number): Promise<SupportConversation> {
  const { data } = await apiClient.post<SupportConversation>(`/admin/support/conversations/${conversationId}/close`)
  return data
}

export async function reopenConversation(conversationId: number): Promise<SupportConversation> {
  const { data } = await apiClient.post<SupportConversation>(`/admin/support/conversations/${conversationId}/reopen`)
  return data
}

export async function unread(): Promise<{ unread_count: number }> {
  const { data } = await apiClient.get<{ unread_count: number }>('/admin/support/unread')
  return data
}

export function openAdminSupportWebSocket(options: {
  token?: string
  onMessage: (event: SupportRealtimeEvent) => void
  onStatus?: (status: SupportWSStatus) => void
  onError?: (error: Event) => void
}): () => void {
  return openSupportWebSocket({
    ...options,
    admin: true
  })
}

const supportAdminAPI = {
  listConversations,
  listMessages,
  sendMessage,
  markRead,
  closeConversation,
  reopenConversation,
  unread,
  openWebSocket: openAdminSupportWebSocket
}

export default supportAdminAPI

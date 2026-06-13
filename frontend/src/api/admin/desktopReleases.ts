/**
 * Admin Desktop Releases API endpoints
 * 桌面客户端版本发布接口
 */

import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'

export interface DesktopRelease {
  id: number
  version: string
  channel: string
  mandatory: boolean
  notes: string
  setup_file: string
  blockmap_file: string
  file_size: number
  status: string
  created_at: string
  updated_at: string
}

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    channel?: string
    status?: string
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: { signal?: AbortSignal },
): Promise<BasePaginationResponse<DesktopRelease>> {
  const { data } = await apiClient.get<BasePaginationResponse<DesktopRelease>>('/admin/desktop-releases', {
    params: { page, page_size: pageSize, ...filters },
    signal: options?.signal,
  })
  return data
}

export async function create(payload: {
  version: string
  channel?: string
  mandatory: boolean
  notes?: string
  setup: File
  latestYml: File
  blockmap?: File | null
}): Promise<DesktopRelease> {
  const formData = new FormData()
  formData.append('version', payload.version)
  if (payload.channel) formData.append('channel', payload.channel)
  formData.append('mandatory', payload.mandatory ? 'true' : 'false')
  if (payload.notes) formData.append('notes', payload.notes)
  formData.append('setup', payload.setup)
  formData.append('latest_yml', payload.latestYml)
  if (payload.blockmap) formData.append('blockmap', payload.blockmap)
  const { data } = await apiClient.post<DesktopRelease>('/admin/desktop-releases', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export async function rollback(id: number): Promise<DesktopRelease> {
  const { data } = await apiClient.post<DesktopRelease>(`/admin/desktop-releases/${id}/rollback`)
  return data
}

export async function deleteRelease(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/desktop-releases/${id}`)
  return data
}

const desktopReleasesAPI = {
  list,
  create,
  rollback,
  delete: deleteRelease,
}

export default desktopReleasesAPI

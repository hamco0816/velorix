/**
 * 桌面客户端公开接口（下载页用）
 */
import { apiClient } from './client'

export interface DesktopLatest {
  available: boolean
  version?: string
  channel?: string
  notes?: string
  mandatory?: boolean
  file_size?: number
  setup_url?: string
  released_at?: string
}

export async function getLatest(channel = 'stable'): Promise<DesktopLatest> {
  const { data } = await apiClient.get<DesktopLatest>('/desktop/latest', { params: { channel } })
  return data
}

export default { getLatest }

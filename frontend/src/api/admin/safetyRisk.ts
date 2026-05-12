import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type SafetyRiskStatus = 'pending' | 'reviewed' | 'cleared' | string

export interface SafetyRiskEvent {
  id: number
  created_at: string
  user_id?: number | null
  user_email: string
  api_key_id?: number | null
  api_key_name: string
  group_id?: number | null
  group_name: string
  request_id: string
  client_request_id: string
  method: string
  path: string
  client_ip: string
  user_agent: string
  rule_source: string
  rule_word: string
  rule_path: string
  category: string
  severity: string
  action: string
  ai_reviewed: boolean
  ai_review_provider: string
  ai_review_result: string
  status: SafetyRiskStatus
  prompt_preview: string
  reviewed_by_user_id?: number | null
  reviewed_at?: string | null
  review_note: string
  cleared_at?: string | null
}

export interface SafetyRiskQueryParams {
  page?: number
  page_size?: number
  time_range?: '5m' | '30m' | '1h' | '6h' | '24h' | '7d' | '30d'
  start_time?: string
  end_time?: string
  status?: string
  action?: string
  severity?: string
  source?: string
  user_id?: number | null
  api_key_id?: number | null
  group_id?: number | null
  q?: string
}

export type SafetyRiskListResponse = PaginatedResponse<SafetyRiskEvent>

export async function listSafetyRiskEvents(params: SafetyRiskQueryParams): Promise<SafetyRiskListResponse> {
  const { data } = await apiClient.get<SafetyRiskListResponse>('/admin/ops/safety-risk-events', { params })
  return data
}

export async function reviewSafetyRiskEvent(id: number, payload: { status: SafetyRiskStatus; review_note?: string }): Promise<void> {
  await apiClient.put(`/admin/ops/safety-risk-events/${id}/review`, payload)
}

export async function clearSafetyRiskEventsForUser(userId: number, reviewNote?: string): Promise<{ cleared: number }> {
  const { data } = await apiClient.post<{ cleared: number }>('/admin/ops/safety-risk-events/clear-user', {
    user_id: userId,
    review_note: reviewNote,
  })
  return data
}

export async function listSafetyAllowlist(): Promise<number[]> {
  const { data } = await apiClient.get<{ user_ids: number[] }>('/admin/ops/safety-risk/allowlist')
  return data?.user_ids || []
}

export async function addSafetyAllowlist(userId: number): Promise<void> {
  await apiClient.post('/admin/ops/safety-risk/allowlist', { user_id: userId })
}

export async function removeSafetyAllowlist(userId: number): Promise<void> {
  await apiClient.delete(`/admin/ops/safety-risk/allowlist/${userId}`)
}

export interface AIReviewTestResult {
  verdict: string
  category: string
  reason: string
  raw?: string
}

export async function testSafetyAIReview(prompt: string): Promise<AIReviewTestResult> {
  const { data } = await apiClient.post<AIReviewTestResult>('/admin/ops/safety-risk/ai-review/test', { prompt })
  return data
}

export interface SafetyRiskRuleStat {
  rule_word: string
  rule_source: string
  total_hits: number
  blocked_count: number
  warned_count: number
  cleared_count: number
  ai_pass_count: number
  ai_reject_count: number
  ai_flag_count: number
  last_hit_at?: string
}

export interface SafetyRiskRuleStatsResponse {
  stats: SafetyRiskRuleStat[]
  hours: number
}

export async function listSafetyRiskRuleStats(hours = 168, limit = 50): Promise<SafetyRiskRuleStatsResponse> {
  const { data } = await apiClient.get<SafetyRiskRuleStatsResponse>('/admin/ops/safety-risk/rule-stats', {
    params: { hours, limit },
  })
  return data
}

export const safetyRiskAPI = {
  listSafetyRiskEvents,
  reviewSafetyRiskEvent,
  clearSafetyRiskEventsForUser,
  listSafetyAllowlist,
  addSafetyAllowlist,
  removeSafetyAllowlist,
  testSafetyAIReview,
  listSafetyRiskRuleStats,
}

export default safetyRiskAPI

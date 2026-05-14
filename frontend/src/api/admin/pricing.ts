/**
 * Admin Model Pricing API endpoints
 *
 * 提供 admin "模型定价总览"页面的数据来源：拉取 PricingService 已加载的全量模型 + 元信息。
 * 前端按需在客户端叠加 group/account 倍率进行换算，避免后端做组合数据爆炸。
 */

import { apiClient } from '../client'

export interface PricingModelEntry {
  model: string
  provider: string
  mode: string
  input_cost_per_token: number
  output_cost_per_token: number
  cache_read_input_token_cost: number
  cache_creation_input_token_cost: number
  output_cost_per_image: number       // 每张图片单价（部分模型有，如 gemini-2.5-flash-image）
  output_cost_per_image_token: number // 每个图片 token 单价（GPT 系列图片模型主要用这个，如 gpt-image-1.5）
  supports_prompt_caching: boolean
  supports_service_tier: boolean
}

export interface PricingMetadata {
  remote_url: string
  hash_url: string
  last_updated: string
  local_hash: string
  model_count: number
}

export interface PricingListResponse {
  models: PricingModelEntry[]
  metadata: PricingMetadata
}

export async function listAllModelPricing(): Promise<PricingListResponse> {
  const { data } = await apiClient.get<PricingListResponse>('/admin/pricing/models')
  return data
}

export const pricingAPI = {
  listAllModelPricing,
}

export default pricingAPI

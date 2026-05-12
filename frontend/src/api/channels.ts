/**
 * User Channels API endpoints (non-admin)
 * 用户侧「可用渠道」聚合查询：渠道 + 用户可访问的分组 + 支持模型（含定价）。
 */

import { apiClient } from './client'
import type { BillingMode } from '@/constants/channel'

export interface UserAvailableGroup {
  id: number
  name: string
  platform: string
  /** 'standard' | 'subscription' — 订阅分组视觉加深，和 API 密钥页保持一致。 */
  subscription_type: string
  /** 分组默认倍率。用户专属倍率（若有）通过 /groups/rates 获取后在前端 join。 */
  rate_multiplier: number
  /** true = 专属分组（小范围授权）；false = 公开分组。 */
  is_exclusive: boolean
}

export interface UserPricingInterval {
  min_tokens: number
  max_tokens: number | null
  tier_label?: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  per_request_price: number | null
}

export interface UserSupportedModelPricing {
  billing_mode: BillingMode
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  image_output_price: number | null
  per_request_price: number | null
  intervals: UserPricingInterval[]
}

export interface UserSupportedModel {
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
}

/**
 * 渠道下单个平台的子视图：用户可访问的分组 + 该平台支持的模型。
 * 后端把一个渠道按平台聚合成 sections，前端可以把渠道名作为 row-group
 * 一次渲染，后面按 sections 顺序用 rowspan 铺开。
 */
export interface UserChannelPlatformSection {
  platform: string
  groups: UserAvailableGroup[]
  supported_models: UserSupportedModel[]
}

export interface UserAvailableChannel {
  name: string
  description: string
  platforms: UserChannelPlatformSection[]
}

/** 列出当前用户可见的「可用渠道」（与 /groups/available 保持一致，返回平数组）。 */
export async function getAvailable(options?: { signal?: AbortSignal }): Promise<UserAvailableChannel[]> {
  const { data } = await apiClient.get<UserAvailableChannel[]>('/channels/available', {
    signal: options?.signal
  })
  return data
}

/**
 * 「计费标准」总览：返回当前 PricingService 已加载的全量模型定价（从 LiteLLM 同步）。
 * 用户端可直接使用此接口列出所有平台的所有可定价模型，不依赖 admin 手动维护渠道的 supported_models。
 *
 * 价格单位：USD per token（与 admin /admin/pricing/models 一致），前端按需换算为 $/MTok。
 */
export interface PricingListEntry {
  model: string
  provider: string
  mode: string
  input_cost_per_token: number
  output_cost_per_token: number
  cache_read_input_token_cost: number
  cache_creation_input_token_cost: number
  output_cost_per_image: number
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

export interface ListPricingResponse {
  models: PricingListEntry[]
  metadata: PricingMetadata
}

export async function listAllPricing(options?: { signal?: AbortSignal }): Promise<ListPricingResponse> {
  const { data } = await apiClient.get<ListPricingResponse>('/pricing/models', {
    signal: options?.signal,
  })
  return data
}

export const userChannelsAPI = { getAvailable, listAllPricing }

export default userChannelsAPI

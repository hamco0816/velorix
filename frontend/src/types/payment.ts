/**
 * Payment System Type Definitions
 */

// ==================== Enums / Union Types ====================

export type OrderStatus =
  | 'PENDING'
  | 'PAID'
  | 'RECHARGING'
  | 'COMPLETED'
  | 'EXPIRED'
  | 'CANCELLED'
  | 'FAILED'
  | 'REFUND_REQUESTED'
  | 'REFUNDING'
  | 'PARTIALLY_REFUNDED'
  | 'REFUNDED'
  | 'REFUND_FAILED'

export type PaymentType = 'alipay' | 'wxpay' | 'alipay_direct' | 'wxpay_direct' | 'stripe' | 'easypay' | 'xunhupay'

export type OrderType = 'balance' | 'subscription'

// ==================== Configuration ====================

export interface PaymentConfig {
  payment_enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  max_pending_orders: number
  order_timeout_minutes: number
  balance_disabled: boolean
  balance_recharge_multiplier: number
  enabled_payment_types: PaymentType[]
  quick_amounts: number[]
  help_image_url: string
  help_text: string
  stripe_publishable_key: string
}

export interface MethodLimit {
  daily_limit: number
  daily_used: number
  daily_remaining: number
  single_min: number
  single_max: number
  fee_rate: number
  available: boolean
}

/** Response from /payment/limits API */
export interface MethodLimitsResponse {
  methods: Record<string, MethodLimit>
  global_min: number  // widest min across all methods; 0 = no minimum
  global_max: number  // widest max across all methods; 0 = no maximum
}

/** Response from /payment/checkout-info API — single call for the payment page */
export interface CheckoutInfoResponse {
  methods: Record<string, MethodLimit>
  global_min: number
  global_max: number
  plans: SubscriptionPlan[]
  balance_disabled: boolean
  balance_recharge_multiplier: number
  recharge_fee_rate: number
  quick_amounts: number[]
  help_text: string
  help_image_url: string
  stripe_publishable_key: string
}

// ==================== Orders ====================

export interface PaymentOrder {
  id: number
  user_id: number
  amount: number
  pay_amount: number
  fee_rate: number
  payment_type: string
  out_trade_no: string
  status: OrderStatus
  order_type: OrderType
  created_at: string
  expires_at: string
  paid_at?: string
  completed_at?: string
  refund_amount: number
  refund_reason?: string
  refund_at?: string
  refund_requested_at?: string
  refund_requested_by?: number
  refund_request_reason?: string
  plan_id?: number
  provider_instance_id?: string
  /** 待支付订单可恢复扫码所需的字段 */
  pay_url?: string
  qr_code?: string
  qr_code_img?: string
}

// ==================== Plans & Channels ====================

export interface SubscriptionPlan {
  id: number
  group_id: number
  group_platform?: string
  group_name?: string
  rate_multiplier?: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  /** 该套餐自带限额/倍率覆盖（migration 138）；true = 独立档位，限额跟同 group 其他 plan 不同 */
  has_plan_limit_override?: boolean
  supported_model_scopes?: string[]
  name: string
  description: string
  price: number
  original_price?: number
  validity_days: number
  validity_unit: string
  /** Stored as JSON string in backend; API layer should parse before use */
  features: string[]
  for_sale: boolean
  sort_order: number
  /** Short custom plan badge shown on the user plan card. Empty means no badge. */
  badge_text?: string
  /** 角标预设色板 key（gold/obsidian/purple/emerald/sapphire/rose），空 = gold */
  badge_color?: string
  /** 档位名（对比表列头用，如 Lite/Pro），空 = 前端从套餐名自动推导 */
  plan_label?: string
  /** 档位样式预设 key（basic/standard/advanced/flagship/luxury/supreme），空 = basic */
  tier_style?: string
  /** Legacy popular marker retained for API compatibility. New UI uses badge_text. */
  is_popular?: boolean
  /** 'shared' = 共享池套餐；'exclusive' = 独享池套餐（购买后独占一个账号） */
  kind?: 'shared' | 'exclusive'
  /** 独享池套餐的剩余库存数；shared 套餐为 undefined */
  stock_available?: number
  /** 该套餐所属分组消费是否可开票（购买页提示用户） */
  invoice_eligible?: boolean
}

export type PlanKind = 'shared' | 'exclusive'

// 独享池 seat 状态
export type ExclusiveSeatStatus = 'active' | 'expired' | 'refunded' | 'cancelled'

// 用户视角的「我的独享号」一行
// 后端用户接口（SeatHandler）只下发脱敏 account_label，不暴露内部 account_id；
// 管理员路径有独立的 AdminSeatView 类型。
export interface ExclusiveSeat {
  id: number
  user_id: number
  group_id: number
  group_name?: string
  group_platform?: string
  plan_id: number
  plan_name?: string
  account_label: string
  status: ExclusiveSeatStatus
  starts_at: string
  expires_at: string
  assigned_at: string
  last_renewal_at?: string
  usage_usd: number
  notes?: string
  // 日/周/月窗口用量与上限（来自 SeatView）。limit 为空 = 该窗口不限额，
  // 前端按"有 limit 才画进度条 / 无 limit 只显示累计值"决定 UI。
  daily_usage_usd?: number
  weekly_usage_usd?: number
  monthly_usage_usd?: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
}

// 后台视角：管理员接口（AdminSeatView）下发完整 account_id 用于运维操作（释放/换号等）。
export interface AdminExclusiveSeat extends ExclusiveSeat {
  account_id: number
  assigned_by?: number
}

// 独享池库存
export interface ExclusivePoolInventory {
  group_id: number
  total: number
  free: number
  used: number
  /** 当下立即可分配的账号数（剔除限流/过载/临时不可用 + 已占用），是真实"可售卖"指标 */
  schedulable?: number
  expiring_in_7: number
}

export interface PaymentChannel {
  id: number
  group_id?: number
  name: string
  platform: string
  rate_multiplier: number
  description: string
  models: string[]
  features: string[]
  enabled: boolean
}

// ==================== Providers ====================

export interface ProviderInstance {
  id: number
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types: string[]
  enabled: boolean
  payment_mode: string
  refund_enabled: boolean
  allow_user_refund: boolean
  limits: string
  sort_order: number
}

// ==================== Request / Response ====================

export interface RenewalSeatRequest {
  /** 续费目标 seat ID（>0 时本订单走 RenewSeat 路径，不消耗库存） */
  renewal_seat_id?: number
}

export interface CreateOrderRequest {
  amount: number
  payment_type: string
  order_type: string
  plan_id?: number
  return_url?: string
  payment_source?: string
  openid?: string
  wechat_resume_token?: string
  is_mobile?: boolean
  /** 续费目标 seat ID（>0 时本订单走 RenewSeat 路径） */
  renewal_seat_id?: number
}

export type CreateOrderResultType = 'order_created' | 'oauth_required' | 'jsapi_ready'

export interface WechatOAuthInfo {
  authorize_url?: string
  appid?: string
  openid?: string
  scope?: string
  state?: string
  redirect_url?: string
}

export interface WechatJSAPIPayload {
  appId?: string
  timeStamp?: string
  nonceStr?: string
  package?: string
  signType?: string
  paySign?: string
}

export interface CreateOrderResult {
  order_id: number
  amount: number
  pay_url?: string
  qr_code?: string
  qr_code_image?: string
  client_secret?: string
  pay_amount: number
  fee_rate: number
  expires_at: string
  result_type?: CreateOrderResultType
  payment_type?: string
  out_trade_no?: string
  payment_mode?: string
  resume_token?: string
  oauth?: WechatOAuthInfo
  jsapi?: WechatJSAPIPayload
  jsapi_payload?: WechatJSAPIPayload
}

export interface DashboardStats {
  today_amount: number
  total_amount: number
  today_count: number
  total_count: number
  avg_amount: number
  daily_series: { date: string; amount: number; count: number }[]
  payment_methods: { type: string; amount: number; count: number }[]
  top_users: { user_id: number; email: string; amount: number }[]
}

export interface FinanceRevenueBucket {
  date: string
  amount: number
  gross_amount: number
  refund_amount: number
  count: number
}

export interface FinanceRevenueStats {
  period: 'day' | 'month' | string
  start_date: string
  end_date: string
  total_amount: number
  gross_amount: number
  refund_amount: number
  total_count: number
  avg_amount: number
  series: FinanceRevenueBucket[]
}

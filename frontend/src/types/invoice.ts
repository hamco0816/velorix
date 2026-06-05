/**
 * 发票（开票申请）相关类型定义。
 */

export type InvoiceTitleType = 'personal' | 'company'
export type InvoiceStatus = 'pending' | 'issued' | 'rejected' | 'cancelled'

/** 发票申请单 */
export interface InvoiceItem {
  id: number
  recipient_email: string
  title_type: InvoiceTitleType
  title_name: string
  tax_id?: string
  user_remark?: string
  amount: number
  status: InvoiceStatus
  invoice_number?: string
  invoice_date?: string
  invoice_amount?: number
  issued_at?: string
  reject_reason?: string
  email_sent: boolean
  created_at: string
  updated_at: string
  // 管理端附带的申请人信息
  user_id?: number
  user_email?: string
  user_name?: string
}

/** 可开票 / 申请单关联的订单（精简展示） */
export interface InvoiceableOrder {
  id: number
  order_type: string
  amount: number
  pay_amount: number
  payment_type: string
  status: string
  paid_at?: string
  created_at: string
}

/** 可开票额度明细（按消费来源核定，人民币） */
export interface InvoiceableSummary {
  available_amount: number // 可开票总额
  balance_amount: number // 其中：余额按量消费可开部分
  plan_amount: number // 其中：套餐购买可开部分
  invoiced_amount: number // 已被待开/已开申请占用（已扣除）
}

/** 提交开票申请的入参 */
export interface ApplyInvoicePayload {
  recipient_email: string
  title_type: InvoiceTitleType
  title_name: string
  tax_id?: string
  user_remark?: string
  // 申请开票金额（人民币）；省略或 <=0 表示按可开票额度全额开票
  amount?: number
}

/** 申请单详情（含关联订单） */
export interface InvoiceDetail {
  invoice: InvoiceItem
  orders: InvoiceableOrder[]
}

/** PDF 识别出的发票字段（用于开票弹窗预填） */
export interface ParsedInvoicePDF {
  invoice_number: string
  invoice_date: string | null
  invoice_amount: number | null
}

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

/** 提交开票申请的入参 */
export interface ApplyInvoicePayload {
  recipient_email: string
  title_type: InvoiceTitleType
  title_name: string
  tax_id?: string
  user_remark?: string
  order_ids: number[]
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

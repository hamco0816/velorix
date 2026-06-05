/**
 * User Invoice API endpoints
 * 用户端发票（开票申请）接口
 */

import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'
import type {
  InvoiceItem,
  InvoiceableSummary,
  ApplyInvoicePayload,
  InvoiceDetail,
} from '@/types/invoice'

export const invoiceAPI = {
  /** 当前用户的可开票额度明细（按支持开票分组的真实消费核定） */
  getInvoiceableSummary() {
    return apiClient.get<InvoiceableSummary>('/invoices/invoiceable-summary')
  },

  /** 提交开票申请 */
  apply(payload: ApplyInvoicePayload) {
    return apiClient.post<InvoiceItem>('/invoices', payload)
  },

  /** 我的发票申请单列表 */
  getMyInvoices(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<InvoiceItem>>('/invoices/my', {
      params,
    })
  },

  /** 申请单详情（含关联订单） */
  getInvoice(id: number) {
    return apiClient.get<InvoiceDetail>(`/invoices/${id}`)
  },

  /** 取消待开票申请 */
  cancel(id: number) {
    return apiClient.post(`/invoices/${id}/cancel`)
  },
}

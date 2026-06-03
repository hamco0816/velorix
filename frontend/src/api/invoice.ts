/**
 * User Invoice API endpoints
 * 用户端发票（开票申请）接口
 */

import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'
import type {
  InvoiceItem,
  InvoiceableOrder,
  ApplyInvoicePayload,
  InvoiceDetail,
} from '@/types/invoice'

export const invoiceAPI = {
  /** 可开票订单（已完成、实付>0、未被占用） */
  getInvoiceableOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<InvoiceableOrder>>(
      '/invoices/invoiceable-orders',
      { params },
    )
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

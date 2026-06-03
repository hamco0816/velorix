/**
 * Admin Invoice API endpoints
 * 管理端发票（开票申请）接口
 */

import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type {
  InvoiceItem,
  InvoiceDetail,
  ParsedInvoicePDF,
} from '@/types/invoice'

const adminInvoiceAPI = {
  /** 分页查询全部开票申请（可按状态 / 关键词过滤） */
  list(params?: { page?: number; page_size?: number; status?: string; keyword?: string }) {
    return apiClient.get<BasePaginationResponse<InvoiceItem>>('/admin/invoices', { params })
  },

  /** 申请单详情（含关联订单） */
  getDetail(id: number) {
    return apiClient.get<InvoiceDetail>(`/admin/invoices/${id}`)
  },

  /** 上传 PDF 仅做识别预填，不发送、不留存 */
  parsePdf(id: number, file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.post<ParsedInvoicePDF>(`/admin/invoices/${id}/parse-pdf`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  /** 开票：上传 PDF + 确认后的发票号码等，发邮件给客户并留存元数据 */
  issue(
    id: number,
    payload: { file: File; invoice_number: string; invoice_date?: string; invoice_amount?: number },
  ) {
    const formData = new FormData()
    formData.append('file', payload.file)
    formData.append('invoice_number', payload.invoice_number)
    if (payload.invoice_date) formData.append('invoice_date', payload.invoice_date)
    if (payload.invoice_amount != null) {
      formData.append('invoice_amount', String(payload.invoice_amount))
    }
    return apiClient.post<InvoiceItem>(`/admin/invoices/${id}/issue`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  /** 驳回待开票申请并释放订单 */
  reject(id: number, reason: string) {
    return apiClient.post<InvoiceItem>(`/admin/invoices/${id}/reject`, { reason })
  },
}

export default adminInvoiceAPI

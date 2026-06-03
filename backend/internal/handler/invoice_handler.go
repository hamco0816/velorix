package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// InvoiceHandler 用户端发票（开票申请）接口。
type InvoiceHandler struct {
	invoiceService *service.InvoiceService
	settingService *service.SettingService
}

// NewInvoiceHandler 创建用户端发票 handler。
func NewInvoiceHandler(invoiceService *service.InvoiceService, settingService *service.SettingService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService, settingService: settingService}
}

var errInvoiceDisabled = infraerrors.Forbidden("INVOICE_DISABLED", "invoice feature is disabled")

// applyInvoiceBody 提交开票申请的请求体。
type applyInvoiceBody struct {
	RecipientEmail string  `json:"recipient_email" binding:"required"`
	TitleType      string  `json:"title_type" binding:"required,oneof=personal company"`
	TitleName      string  `json:"title_name" binding:"required"`
	TaxID          string  `json:"tax_id"`
	UserRemark     string  `json:"user_remark"`
	OrderIDs       []int64 `json:"order_ids" binding:"required,min=1"`
}

// ensureEnabled 校验发票功能开关，未开启则写 403 并返回 false。
func (h *InvoiceHandler) ensureEnabled(c *gin.Context) bool {
	if !h.settingService.IsInvoiceEnabled(c.Request.Context()) {
		response.ErrorFrom(c, errInvoiceDisabled)
		return false
	}
	return true
}

// GetInvoiceableOrders 返回当前用户「可开票」的订单（已完成、实付>0、未被占用）。
// GET /api/v1/invoices/invoiceable-orders
func (h *InvoiceHandler) GetInvoiceableOrders(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok || !h.ensureEnabled(c) {
		return
	}
	page, pageSize := response.ParsePagination(c)
	orders, total, err := h.invoiceService.ListInvoiceableOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.InvoiceOrdersFromEnt(orders), int64(total), page, pageSize)
}

// ApplyInvoice 提交开票申请。
// POST /api/v1/invoices
func (h *InvoiceHandler) ApplyInvoice(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok || !h.ensureEnabled(c) {
		return
	}
	var body applyInvoiceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}
	ir, err := h.invoiceService.ApplyInvoice(c.Request.Context(), subject.UserID, service.ApplyInvoiceRequest{
		RecipientEmail: body.RecipientEmail,
		TitleType:      body.TitleType,
		TitleName:      body.TitleName,
		TaxID:          body.TaxID,
		UserRemark:     body.UserRemark,
		OrderIDs:       body.OrderIDs,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.InvoiceFromEnt(ir, false))
}

// GetMyInvoices 分页返回当前用户的发票申请单。
// GET /api/v1/invoices/my
func (h *InvoiceHandler) GetMyInvoices(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok || !h.ensureEnabled(c) {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.invoiceService.ListMyInvoices(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.InvoicesFromEnt(items, false), int64(total), page, pageSize)
}

// GetMyInvoice 返回某张申请单详情及其关联订单。
// GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetMyInvoice(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok || !h.ensureEnabled(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}
	ir, orders, err := h.invoiceService.GetMyInvoice(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"invoice": dto.InvoiceFromEnt(ir, false),
		"orders":  dto.InvoiceOrdersFromEnt(orders),
	})
}

// CancelInvoice 取消「待开票」的申请。
// POST /api/v1/invoices/:id/cancel
func (h *InvoiceHandler) CancelInvoice(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok || !h.ensureEnabled(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}
	if err := h.invoiceService.CancelInvoice(c.Request.Context(), subject.UserID, id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "cancelled"})
}

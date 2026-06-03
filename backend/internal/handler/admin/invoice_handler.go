package admin

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// maxInvoicePDFSize 上传发票 PDF 的大小上限（10 MB）。
const maxInvoicePDFSize = 10 << 20

// InvoiceHandler 管理端发票管理接口。
type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

// NewInvoiceHandler 创建管理端发票 handler。
func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// List 分页查询全部发票申请单（支持 status / keyword 过滤）。
// GET /api/v1/admin/invoices
func (h *InvoiceHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.invoiceService.ListInvoices(c.Request.Context(), service.InvoiceListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, dto.InvoicesFromEnt(items, true), int64(total), page, pageSize)
}

// GetDetail 返回某张申请单及其关联订单。
// GET /api/v1/admin/invoices/:id
func (h *InvoiceHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}
	ir, orders, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"invoice": dto.InvoiceFromEnt(ir, true),
		"orders":  dto.InvoiceOrdersFromEnt(orders),
	})
}

// ParsePDF 上传发票 PDF 并尽力识别发票号码/日期/金额，用于开票弹窗预填。不发送、不持久化。
// POST /api/v1/admin/invoices/:id/parse-pdf  (multipart, field: file)
func (h *InvoiceHandler) ParsePDF(c *gin.Context) {
	pdf, err := readUploadedPDF(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	parsed := h.invoiceService.ParseInvoicePDF(pdf)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"invoice_number": parsed.InvoiceNumber,
			"invoice_date":   parsed.InvoiceDate,
			"invoice_amount": parsed.InvoiceAmount,
		},
	})
}

// Issue 开票：上传 PDF 并提交确认后的发票号码等，发邮件给客户并留存元数据。
// POST /api/v1/admin/invoices/:id/issue  (multipart: file + invoice_number/invoice_date/invoice_amount)
func (h *InvoiceHandler) Issue(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}
	pdf, err := readUploadedPDF(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	input := service.IssueInvoiceInput{
		PDF:           pdf,
		InvoiceNumber: c.PostForm("invoice_number"),
		InvoiceDate:   parseInvoiceFormDate(c.PostForm("invoice_date")),
		InvoiceAmount: parseInvoiceFormAmount(c.PostForm("invoice_amount")),
	}
	ir, err := h.invoiceService.IssueInvoice(c.Request.Context(), subject.UserID, id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceFromEnt(ir, true))
}

// rejectInvoiceBody 驳回请求体。
type rejectInvoiceBody struct {
	Reason string `json:"reason"`
}

// Reject 驳回「待开票」申请并释放订单。
// POST /api/v1/admin/invoices/:id/reject
func (h *InvoiceHandler) Reject(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid invoice ID")
		return
	}
	var body rejectInvoiceBody
	_ = c.ShouldBindJSON(&body)
	ir, err := h.invoiceService.RejectInvoice(c.Request.Context(), subject.UserID, id, body.Reason)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.InvoiceFromEnt(ir, true))
}

// --- helpers ---

// readUploadedPDF 从 multipart 表单字段 file 读取 PDF 字节并校验大小/类型。
func readUploadedPDF(c *gin.Context) ([]byte, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, infraerrors.BadRequest("INVOICE_PDF_REQUIRED", "invoice pdf file is required")
	}
	if fileHeader.Size > maxInvoicePDFSize {
		return nil, infraerrors.BadRequest("INVOICE_PDF_TOO_LARGE", "invoice pdf exceeds 10MB limit")
	}
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".pdf") {
		return nil, infraerrors.BadRequest("INVOICE_PDF_INVALID", "only .pdf files are allowed")
	}
	f, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded pdf: %w", err)
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxInvoicePDFSize+1))
	if err != nil {
		return nil, fmt.Errorf("read uploaded pdf: %w", err)
	}
	if len(data) == 0 {
		return nil, infraerrors.BadRequest("INVOICE_PDF_EMPTY", "invoice pdf is empty")
	}
	return data, nil
}

// parseInvoiceFormDate 解析表单里的开票日期（YYYY-MM-DD），空或非法返回 nil。
func parseInvoiceFormDate(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil
	}
	return &t
}

// parseInvoiceFormAmount 解析表单里的开票金额，空或非法返回 nil。
func parseInvoiceFormAmount(raw string) *float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}

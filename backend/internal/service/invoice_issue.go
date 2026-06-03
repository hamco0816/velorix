package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invoicerequest"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const invoicePDFMIME = "application/pdf"

// ErrInvoiceNumberRequired 开票时未能识别也未手动填写发票号码。
var ErrInvoiceNumberRequired = infraerrors.BadRequest("INVOICE_NUMBER_REQUIRED", "invoice number is required")

// InvoiceListParams 管理端发票列表筛选参数。
type InvoiceListParams struct {
	Page     int
	PageSize int
	Status   string // 可选：按状态过滤
	Keyword  string // 可选：匹配接收邮箱/用户邮箱/抬头/用户名
}

// IssueInvoiceInput 管理员开票入参（发票号码等字段由管理员在前端确认/修正后提交）。
type IssueInvoiceInput struct {
	PDF           []byte     // 发票 PDF 字节，仅用于发送，不落盘
	InvoiceNumber string     // 发票号码（必填，可由 PDF 识别预填）
	InvoiceDate   *time.Time // 开票日期（可选）
	InvoiceAmount *float64   // 开票金额（可选，用于核验）
}

// ParseInvoicePDF 供管理端「上传后预填」使用：尽力识别发票号码/日期/金额，不发送、不持久化。
func (s *InvoiceService) ParseInvoicePDF(pdf []byte) ParsedInvoice {
	return parseInvoicePDF(pdf)
}

// ListInvoices 管理端分页查询全部发票申请单（支持状态/关键字过滤）。
func (s *InvoiceService) ListInvoices(ctx context.Context, params InvoiceListParams) ([]*dbent.InvoiceRequest, int, error) {
	pageSize, page := applyPagination(params.PageSize, params.Page)
	query := s.entClient.InvoiceRequest.Query()
	if status := strings.TrimSpace(params.Status); status != "" {
		query = query.Where(invoicerequest.StatusEQ(status))
	}
	if kw := strings.TrimSpace(params.Keyword); kw != "" {
		query = query.Where(invoicerequest.Or(
			invoicerequest.RecipientEmailContainsFold(kw),
			invoicerequest.UserEmailContainsFold(kw),
			invoicerequest.TitleNameContainsFold(kw),
			invoicerequest.UserNameContainsFold(kw),
		))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count invoices: %w", err)
	}
	items, err := query.
		Order(dbent.Desc(invoicerequest.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list invoices: %w", err)
	}
	return items, total, nil
}

// GetInvoice 管理端获取某张申请单及其关联订单。
func (s *InvoiceService) GetInvoice(ctx context.Context, requestID int64) (*dbent.InvoiceRequest, []*dbent.PaymentOrder, error) {
	ir, err := s.entClient.InvoiceRequest.Get(ctx, requestID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrInvoiceNotFound
		}
		return nil, nil, fmt.Errorf("get invoice: %w", err)
	}
	orders, err := s.entClient.PaymentOrder.Query().
		Where(paymentorder.InvoiceRequestIDEQ(requestID)).
		Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list invoice orders: %w", err)
	}
	return ir, orders, nil
}

// IssueInvoice 管理员开票：把 PDF 作为附件发给客户邮箱，发送成功后留存发票号码等元数据。
// PDF 文件本身不写盘、不入库；发送失败则保持待开票状态并返回错误。
func (s *InvoiceService) IssueInvoice(ctx context.Context, adminID, requestID int64, input IssueInvoiceInput) (*dbent.InvoiceRequest, error) {
	invoiceNumber := strings.TrimSpace(input.InvoiceNumber)
	if invoiceNumber == "" {
		return nil, ErrInvoiceNumberRequired
	}
	if len(input.PDF) == 0 {
		return nil, infraerrors.BadRequest("INVOICE_PDF_REQUIRED", "invoice pdf is required")
	}

	ir, err := s.entClient.InvoiceRequest.Get(ctx, requestID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	if ir.Status != InvoiceStatusPending {
		return nil, ErrInvoiceNotPending
	}

	// 先发邮件（带 PDF 附件）。发送成功才落库为已开票，避免「标记已开但客户没收到」。
	if err := s.sendInvoiceEmail(ctx, ir, invoiceNumber, input.PDF); err != nil {
		return nil, infraerrors.ServiceUnavailable("INVOICE_EMAIL_FAILED", fmt.Sprintf("send invoice email failed: %v", err))
	}

	updated, err := s.entClient.InvoiceRequest.UpdateOneID(requestID).
		SetStatus(InvoiceStatusIssued).
		SetInvoiceNumber(invoiceNumber).
		SetNillableInvoiceDate(input.InvoiceDate).
		SetNillableInvoiceAmount(input.InvoiceAmount).
		SetIssuedAt(time.Now()).
		SetIssuedBy(adminID).
		SetEmailSent(true).
		ClearEmailError().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark invoice issued: %w", err)
	}
	return updated, nil
}

// RejectInvoice 管理员驳回「待开票」申请，释放被占用的订单。
func (s *InvoiceService) RejectInvoice(ctx context.Context, adminID, requestID int64, reason string) (*dbent.InvoiceRequest, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ir, err := lockInvoiceRequest(ctx, tx, requestID)
	if err != nil {
		return nil, err
	}
	if ir.Status != InvoiceStatusPending {
		return nil, ErrInvoiceNotPending
	}
	updated, err := tx.InvoiceRequest.UpdateOneID(requestID).
		SetStatus(InvoiceStatusRejected).
		SetNillableRejectReason(psNilIfEmpty(strings.TrimSpace(reason))).
		SetIssuedBy(adminID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("reject invoice: %w", err)
	}
	if err := releaseInvoiceOrders(ctx, tx, requestID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return updated, nil
}

// sendInvoiceEmail 把发票 PDF 作为附件发送到客户登记的接收邮箱。
func (s *InvoiceService) sendInvoiceEmail(ctx context.Context, ir *dbent.InvoiceRequest, invoiceNumber string, pdf []byte) error {
	if s.mailer == nil {
		return fmt.Errorf("mailer not configured")
	}
	siteName := s.siteName(ctx)
	subject := fmt.Sprintf("[%s] 您申请的发票已开具", siteName)
	body := buildInvoiceIssuedEmailBody(siteName, ir, invoiceNumber)
	attachmentName := fmt.Sprintf("invoice-%s.pdf", invoiceNumber)
	return s.mailer.SendEmailWithAttachment(ctx, ir.RecipientEmail, subject, body, attachmentName, pdf, invoicePDFMIME)
}

func (s *InvoiceService) siteName(ctx context.Context) string {
	if s.settingService == nil {
		return "Sub2API"
	}
	return s.settingService.GetSiteName(ctx)
}

// buildInvoiceIssuedEmailBody 构建「发票已开具」通知邮件正文（发票作为附件随邮件发送）。
func buildInvoiceIssuedEmailBody(siteName string, ir *dbent.InvoiceRequest, invoiceNumber string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#f5f5f5;margin:0;padding:24px;color:#1f2937;">
  <div style="max-width:600px;margin:0 auto;background:#fff;border-radius:8px;border:1px solid #e5e7eb;overflow:hidden;">
    <div style="padding:20px 28px;border-bottom:1px solid #e5e7eb;font-size:18px;font-weight:600;">%s 发票通知</div>
    <div style="padding:28px;line-height:1.7;font-size:14px;">
      <p>您好，您申请的发票已开具完成，详情如下：</p>
      <table style="width:100%%;border-collapse:collapse;margin:16px 0;font-size:14px;">
        <tr><td style="padding:8px 0;color:#6b7280;width:96px;">发票抬头</td><td style="padding:8px 0;">%s</td></tr>
        <tr><td style="padding:8px 0;color:#6b7280;">发票号码</td><td style="padding:8px 0;">%s</td></tr>
        <tr><td style="padding:8px 0;color:#6b7280;">开票金额</td><td style="padding:8px 0;">%.2f 元</td></tr>
      </table>
      <p>发票文件已作为附件随本邮件发送，请注意查收。如有疑问请联系我们。</p>
    </div>
    <div style="padding:16px 28px;background:#f9fafb;border-top:1px solid #e5e7eb;color:#9ca3af;font-size:12px;">此邮件为系统自动发送，请勿直接回复。</div>
  </div>
</body>
</html>`, siteName, ir.TitleName, invoiceNumber, ir.Amount)
}

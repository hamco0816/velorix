package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invoicerequest"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"entgo.io/ent/dialect"
	"github.com/shopspring/decimal"
)

// --- 发票状态 / 抬头类型常量 ---

const (
	InvoiceStatusPending   = "pending"   // 待开票
	InvoiceStatusIssued    = "issued"    // 已开票
	InvoiceStatusRejected  = "rejected"  // 已驳回
	InvoiceStatusCancelled = "cancelled" // 用户取消

	InvoiceTitlePersonal = "personal" // 个人抬头
	InvoiceTitleCompany  = "company"  // 企业抬头
)

// --- 错误定义 ---

var (
	ErrInvoiceNotFound       = infraerrors.NotFound("INVOICE_NOT_FOUND", "invoice request not found")
	ErrInvoiceNotOwned       = infraerrors.Forbidden("INVOICE_NOT_OWNED", "invoice request does not belong to current user")
	ErrInvoiceNotPending     = infraerrors.Conflict("INVOICE_NOT_PENDING", "invoice request is not in pending status")
	ErrInvoiceNoOrders       = infraerrors.BadRequest("INVOICE_NO_ORDERS", "no orders selected for invoicing")
	ErrInvoiceOrderNotUsable = infraerrors.Conflict("INVOICE_ORDER_NOT_USABLE", "one or more orders are not eligible for invoicing")
	ErrInvoiceTaxIDRequired  = infraerrors.BadRequest("INVOICE_TAX_ID_REQUIRED", "tax id is required for company invoice")
	ErrInvoiceInvalidTitle   = infraerrors.BadRequest("INVOICE_INVALID_TITLE", "invalid invoice title")
	ErrInvoiceInvalidEmail   = infraerrors.BadRequest("INVOICE_INVALID_EMAIL", "invalid recipient email")
)

// InvoiceMailer 抽象发票邮件发送，便于单测替换。生产环境由 EmailService 实现。
type InvoiceMailer interface {
	SendEmailWithAttachment(ctx context.Context, to, subject, htmlBody, attachmentName string, attachment []byte, attachmentMIME string) error
}

// InvoiceService 发票申请/开票服务。
//
// 直接持有 ent client 做跨表事务（与支付域 PaymentService 一致）：
// 申请/取消/开票/驳回都要同时改 invoice_requests 和 payment_orders.invoice_request_id。
type InvoiceService struct {
	entClient      *dbent.Client
	mailer         InvoiceMailer
	settingService *SettingService // 可空（单测）：用于读取站点名、前端地址
}

// NewInvoiceService 创建发票服务实例。
func NewInvoiceService(entClient *dbent.Client, mailer InvoiceMailer, settingService *SettingService) *InvoiceService {
	return &InvoiceService{entClient: entClient, mailer: mailer, settingService: settingService}
}

// ApplyInvoiceRequest 用户提交开票申请的入参。
type ApplyInvoiceRequest struct {
	RecipientEmail string
	TitleType      string
	TitleName      string
	TaxID          string
	UserRemark     string
	OrderIDs       []int64
}

// ListInvoiceableOrders 返回用户「可开票」的订单：已完成、实付金额大于 0、且未被任何申请单占用。
// 余额充值与订阅/套餐购买都算（按 created_at 倒序分页）。
func (s *InvoiceService) ListInvoiceableOrders(ctx context.Context, userID int64, page, pageSize int) ([]*dbent.PaymentOrder, int, error) {
	pageSize, page = applyPagination(pageSize, page)
	query := s.entClient.PaymentOrder.Query().Where(
		paymentorder.UserIDEQ(userID),
		paymentorder.StatusEQ(OrderStatusCompleted),
		paymentorder.PayAmountGT(0),
		paymentorder.InvoiceRequestIDIsNil(),
	)
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count invoiceable orders: %w", err)
	}
	orders, err := query.
		Order(dbent.Desc(paymentorder.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list invoiceable orders: %w", err)
	}
	return orders, total, nil
}

// ApplyInvoice 提交开票申请：校验所选订单，合并金额，创建申请单并锁定订单。
func (s *InvoiceService) ApplyInvoice(ctx context.Context, userID int64, req ApplyInvoiceRequest) (*dbent.InvoiceRequest, error) {
	if err := validateInvoiceForm(req); err != nil {
		return nil, err
	}
	orderIDs := dedupeInt64(req.OrderIDs)
	if len(orderIDs) == 0 {
		return nil, ErrInvoiceNoOrders
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	user, err := tx.User.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 事务内锁定并校验所选订单（PG 用 FOR UPDATE，SQLite 测试环境降级为普通查询）。
	orderQuery := tx.PaymentOrder.Query().Where(
		paymentorder.UserIDEQ(userID),
		paymentorder.IDIn(orderIDs...),
	)
	if txIsPostgres(tx) {
		orderQuery = orderQuery.ForUpdate()
	}
	orders, err := orderQuery.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("lock orders: %w", err)
	}
	if len(orders) != len(orderIDs) {
		return nil, ErrInvoiceOrderNotUsable
	}

	total := decimal.Zero
	for _, order := range orders {
		if order.Status != OrderStatusCompleted || order.PayAmount <= 0 || order.InvoiceRequestID != nil {
			return nil, ErrInvoiceOrderNotUsable
		}
		total = total.Add(decimal.NewFromFloat(order.PayAmount))
	}
	amount := total.Round(2).InexactFloat64()

	taxID := normalizeTaxID(req.TitleType, req.TaxID)
	created, err := tx.InvoiceRequest.Create().
		SetUserID(userID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetRecipientEmail(strings.TrimSpace(req.RecipientEmail)).
		SetTitleType(req.TitleType).
		SetTitleName(strings.TrimSpace(req.TitleName)).
		SetNillableTaxID(psNilIfEmpty(taxID)).
		SetNillableUserRemark(psNilIfEmpty(strings.TrimSpace(req.UserRemark))).
		SetAmount(amount).
		SetStatus(InvoiceStatusPending).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create invoice request: %w", err)
	}

	// 关联订单：写入 invoice_request_id，占用这些订单。
	if _, err := tx.PaymentOrder.Update().
		Where(paymentorder.IDIn(orderIDs...)).
		SetInvoiceRequestID(created.ID).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("link orders: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return created, nil
}

// CancelInvoice 用户取消「待开票」的申请，释放被占用的订单。
func (s *InvoiceService) CancelInvoice(ctx context.Context, userID, requestID int64) error {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ir, err := lockInvoiceRequest(ctx, tx, requestID)
	if err != nil {
		return err
	}
	if ir.UserID != userID {
		return ErrInvoiceNotOwned
	}
	if ir.Status != InvoiceStatusPending {
		return ErrInvoiceNotPending
	}
	if err := tx.InvoiceRequest.UpdateOneID(requestID).SetStatus(InvoiceStatusCancelled).Exec(ctx); err != nil {
		return fmt.Errorf("cancel invoice request: %w", err)
	}
	if err := releaseInvoiceOrders(ctx, tx, requestID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ListMyInvoices 分页返回用户自己的发票申请单（按 created_at 倒序）。
func (s *InvoiceService) ListMyInvoices(ctx context.Context, userID int64, page, pageSize int) ([]*dbent.InvoiceRequest, int, error) {
	pageSize, page = applyPagination(pageSize, page)
	query := s.entClient.InvoiceRequest.Query().Where(invoicerequest.UserIDEQ(userID))
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

// GetMyInvoice 返回用户某张申请单及其关联订单（校验归属）。
func (s *InvoiceService) GetMyInvoice(ctx context.Context, userID, requestID int64) (*dbent.InvoiceRequest, []*dbent.PaymentOrder, error) {
	ir, err := s.entClient.InvoiceRequest.Get(ctx, requestID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, ErrInvoiceNotFound
		}
		return nil, nil, fmt.Errorf("get invoice: %w", err)
	}
	if ir.UserID != userID {
		return nil, nil, ErrInvoiceNotOwned
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

// --- 内部 helpers ---

// validateInvoiceForm 校验开票表单：邮箱格式、抬头类型、企业必填税号。
func validateInvoiceForm(req ApplyInvoiceRequest) error {
	if _, err := mail.ParseAddress(strings.TrimSpace(req.RecipientEmail)); err != nil {
		return ErrInvoiceInvalidEmail
	}
	if strings.TrimSpace(req.TitleName) == "" {
		return ErrInvoiceInvalidTitle
	}
	switch req.TitleType {
	case InvoiceTitlePersonal:
	case InvoiceTitleCompany:
		if strings.TrimSpace(req.TaxID) == "" {
			return ErrInvoiceTaxIDRequired
		}
	default:
		return ErrInvoiceInvalidTitle
	}
	return nil
}

// normalizeTaxID 个人抬头不存税号；企业抬头去掉首尾空白。
func normalizeTaxID(titleType, taxID string) string {
	if titleType == InvoiceTitleCompany {
		return strings.TrimSpace(taxID)
	}
	return ""
}

// lockInvoiceRequest 事务内加锁读取申请单（PG 行锁，SQLite 降级）。
func lockInvoiceRequest(ctx context.Context, tx *dbent.Tx, requestID int64) (*dbent.InvoiceRequest, error) {
	query := tx.InvoiceRequest.Query().Where(invoicerequest.IDEQ(requestID))
	if txIsPostgres(tx) {
		query = query.ForUpdate()
	}
	ir, err := query.Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("lock invoice request: %w", err)
	}
	return ir, nil
}

// releaseInvoiceOrders 把申请单关联的订单 invoice_request_id 置空，使其可重新申请。
func releaseInvoiceOrders(ctx context.Context, tx *dbent.Tx, requestID int64) error {
	if _, err := tx.PaymentOrder.Update().
		Where(paymentorder.InvoiceRequestIDEQ(requestID)).
		ClearInvoiceRequestID().
		Save(ctx); err != nil {
		return fmt.Errorf("release invoice orders: %w", err)
	}
	return nil
}

// txIsPostgres 判断当前事务底层是否 PostgreSQL（SQLite 测试环境不支持 FOR UPDATE）。
func txIsPostgres(tx *dbent.Tx) bool {
	d, ok := tx.Client().Driver().(interface{ Dialect() string })
	return ok && d.Dialect() == dialect.Postgres
}

// dedupeInt64 去重并丢弃非正数 ID。
func dedupeInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

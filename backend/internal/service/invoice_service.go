package service

import (
	"context"
	"fmt"
	"math"
	"net/mail"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/invoicerequest"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/payment"
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
	ErrInvoiceNotFound      = infraerrors.NotFound("INVOICE_NOT_FOUND", "invoice request not found")
	ErrInvoiceNotOwned      = infraerrors.Forbidden("INVOICE_NOT_OWNED", "invoice request does not belong to current user")
	ErrInvoiceNotPending    = infraerrors.Conflict("INVOICE_NOT_PENDING", "invoice request is not in pending status")
	ErrInvoiceTaxIDRequired = infraerrors.BadRequest("INVOICE_TAX_ID_REQUIRED", "tax id is required for company invoice")
	ErrInvoiceInvalidTitle  = infraerrors.BadRequest("INVOICE_INVALID_TITLE", "invalid invoice title")
	ErrInvoiceInvalidEmail  = infraerrors.BadRequest("INVOICE_INVALID_EMAIL", "invalid recipient email")
	// ErrInvoiceNothingToInvoice 当前没有可开票金额（无支持开票分组的真实消费）。
	ErrInvoiceNothingToInvoice = infraerrors.BadRequest("INVOICE_NOTHING_TO_INVOICE", "no invoiceable amount available")
	// ErrInvoiceAmountExceedsAvailable 申请金额超过当前可开票额度。
	ErrInvoiceAmountExceedsAvailable = infraerrors.Conflict("INVOICE_AMOUNT_EXCEEDS_AVAILABLE", "requested amount exceeds available invoiceable amount")
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
	// Amount 申请开票金额（人民币）；<= 0 表示按当前可开票额度全额开票。
	Amount float64
}

// InvoiceableSummary 当前用户的可开票额度明细，用于申请前展示。
// 口径（保守 + 非赠送）：只算「支持开票分组」的真实消费，且不超过真实付费金额。
type InvoiceableSummary struct {
	AvailableAmount float64 `json:"available_amount"` // 可开票总额（人民币）
	BalanceAmount   float64 `json:"balance_amount"`   // 其中：余额按量消费可开部分
	PlanAmount      float64 `json:"plan_amount"`      // 其中：套餐购买可开部分
	InvoicedAmount  float64 `json:"invoiced_amount"`  // 已被待开/已开申请占用（已从可开额度中扣除）
}

// invoiceOrderClients 聚合可开票额度计算所需的 ent 查询入口，
// 让事务（*dbent.Tx）与非事务（*dbent.Client）路径复用同一套计算逻辑。
type invoiceOrderClients struct {
	orders   *dbent.PaymentOrderClient
	invoices *dbent.InvoiceRequestClient
	groups   *dbent.GroupClient
}

func (s *InvoiceService) clientInvoiceOrderClients() invoiceOrderClients {
	return invoiceOrderClients{orders: s.entClient.PaymentOrder, invoices: s.entClient.InvoiceRequest, groups: s.entClient.Group}
}

func txInvoiceOrderClients(tx *dbent.Tx) invoiceOrderClients {
	return invoiceOrderClients{orders: tx.PaymentOrder, invoices: tx.InvoiceRequest, groups: tx.Group}
}

// rechargeMultiplier 读取当前充值倍率（额度 = 人民币 × 倍率），用于把消费额度折算回人民币。
func (s *InvoiceService) rechargeMultiplier(ctx context.Context) float64 {
	if s.settingService == nil {
		return defaultBalanceRechargeMultiplier
	}
	return s.settingService.GetBalanceRechargeMultiplier(ctx)
}

// computeInvoiceable 计算用户当前可开票额度（人民币）。
//
// 口径（保守 + 非赠送）：
//   - 余额消费部分 = min(支持开票分组的消费额度 / 充值倍率, 真实余额充值付费)。
//     左项保证「只开已消费」，右项保证「不超过真实付费」（赠送/兑换码/手动加额都不计入）。
//   - 套餐部分 = 绑定支持开票分组的套餐订单实付金额（净退款）。
//   - 再减去已被待开/已开申请占用的金额。
func (s *InvoiceService) computeInvoiceable(ctx context.Context, c invoiceOrderClients, userID int64, consumedQuota, multiplier float64) (*InvoiceableSummary, error) {
	if multiplier <= 0 {
		multiplier = defaultBalanceRechargeMultiplier
	}

	// 支持开票的分组 ID（套餐部分按此过滤）
	eligibleGroupIDs, err := c.groups.Query().Where(group.InvoiceEligible(true)).IDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list invoice-eligible groups: %w", err)
	}

	// 真实余额充值付费（人民币，净退款）
	balancePaid, err := sumOrderNetPay(ctx, c.orders,
		paymentorder.UserIDEQ(userID),
		paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
		paymentorder.StatusEQ(OrderStatusCompleted))
	if err != nil {
		return nil, err
	}

	// 套餐可开部分（人民币，净退款）：仅支持开票分组
	planAmount := 0.0
	if len(eligibleGroupIDs) > 0 {
		planAmount, err = sumOrderNetPay(ctx, c.orders,
			paymentorder.UserIDEQ(userID),
			paymentorder.OrderTypeEQ(payment.OrderTypeSubscription),
			paymentorder.StatusEQ(OrderStatusCompleted),
			paymentorder.SubscriptionGroupIDIn(eligibleGroupIDs...))
		if err != nil {
			return nil, err
		}
	}

	// 余额消费折算人民币，并用真实付费封顶（排除赠送）
	consumedCNY := decimal.NewFromFloat(consumedQuota).Div(decimal.NewFromFloat(multiplier)).InexactFloat64()
	balanceInvoiceable := math.Min(consumedCNY, balancePaid)
	if balanceInvoiceable < 0 {
		balanceInvoiceable = 0
	}

	// 已被待开/已开申请占用的金额
	invoiced, err := sumInvoiceAmount(ctx, c.invoices, userID)
	if err != nil {
		return nil, err
	}

	available := decimal.NewFromFloat(balanceInvoiceable).
		Add(decimal.NewFromFloat(planAmount)).
		Sub(decimal.NewFromFloat(invoiced)).
		Round(2).InexactFloat64()
	if available < 0 {
		available = 0
	}

	return &InvoiceableSummary{
		AvailableAmount: available,
		BalanceAmount:   decimal.NewFromFloat(balanceInvoiceable).Round(2).InexactFloat64(),
		PlanAmount:      decimal.NewFromFloat(planAmount).Round(2).InexactFloat64(),
		InvoicedAmount:  decimal.NewFromFloat(invoiced).Round(2).InexactFloat64(),
	}, nil
}

// sumOrderNetPay 汇总满足条件订单的「实付 - 退款」之和（人民币），结果不为负。
func sumOrderNetPay(ctx context.Context, orders *dbent.PaymentOrderClient, preds ...predicate.PaymentOrder) (float64, error) {
	var rows []struct {
		Pay    *float64 `json:"pay"`
		Refund *float64 `json:"refund"`
	}
	if err := orders.Query().Where(preds...).Aggregate(
		dbent.As(dbent.Sum(paymentorder.FieldPayAmount), "pay"),
		dbent.As(dbent.Sum(paymentorder.FieldRefundAmount), "refund"),
	).Scan(ctx, &rows); err != nil {
		return 0, fmt.Errorf("sum order net pay: %w", err)
	}
	net := 0.0
	if len(rows) > 0 {
		net = derefFloat(rows[0].Pay) - derefFloat(rows[0].Refund)
	}
	if net < 0 {
		net = 0
	}
	return net, nil
}

// sumInvoiceAmount 汇总用户「待开票 + 已开票」申请单金额（已占用的可开额度）。
func sumInvoiceAmount(ctx context.Context, invoices *dbent.InvoiceRequestClient, userID int64) (float64, error) {
	var rows []struct {
		Sum *float64 `json:"sum"`
	}
	if err := invoices.Query().
		Where(invoicerequest.UserIDEQ(userID), invoicerequest.StatusIn(InvoiceStatusPending, InvoiceStatusIssued)).
		Aggregate(dbent.As(dbent.Sum(invoicerequest.FieldAmount), "sum")).
		Scan(ctx, &rows); err != nil {
		return 0, fmt.Errorf("sum invoiced amount: %w", err)
	}
	if len(rows) > 0 {
		return derefFloat(rows[0].Sum), nil
	}
	return 0, nil
}

func derefFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

// GetInvoiceableSummary 返回当前用户的可开票额度明细。
func (s *InvoiceService) GetInvoiceableSummary(ctx context.Context, userID int64) (*InvoiceableSummary, error) {
	u, err := s.entClient.User.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return s.computeInvoiceable(ctx, s.clientInvoiceOrderClients(), userID, u.InvoiceableConsumed, s.rechargeMultiplier(ctx))
}

// ApplyInvoice 提交开票申请：按「支持开票分组的真实消费」核定可开票额度，
// 校验申请金额不超过额度后创建申请单（金额制，不再绑定具体订单）。
func (s *InvoiceService) ApplyInvoice(ctx context.Context, userID int64, req ApplyInvoiceRequest) (*dbent.InvoiceRequest, error) {
	if err := validateInvoiceForm(req); err != nil {
		return nil, err
	}
	multiplier := s.rechargeMultiplier(ctx)

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 锁定用户行，串行化同一用户的并发开票，避免可开票额度被重复占用。
	userQuery := tx.User.Query().Where(user.IDEQ(userID))
	if txIsPostgres(tx) {
		userQuery = userQuery.ForUpdate()
	}
	u, err := userQuery.Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("lock user: %w", err)
	}

	summary, err := s.computeInvoiceable(ctx, txInvoiceOrderClients(tx), userID, u.InvoiceableConsumed, multiplier)
	if err != nil {
		return nil, err
	}
	available := summary.AvailableAmount
	if available <= 0 {
		return nil, ErrInvoiceNothingToInvoice
	}

	// 申请金额：<=0 表示全额开票；否则不得超过可开额度（留 0.01 容差吸收浮点误差）。
	amount := decimal.NewFromFloat(req.Amount).Round(2).InexactFloat64()
	if amount <= 0 {
		amount = available
	}
	if amount > available+0.001 {
		return nil, ErrInvoiceAmountExceedsAvailable
	}
	if amount > available {
		amount = available
	}

	taxID := normalizeTaxID(req.TitleType, req.TaxID)
	created, err := tx.InvoiceRequest.Create().
		SetUserID(userID).
		SetUserEmail(u.Email).
		SetUserName(u.Username).
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

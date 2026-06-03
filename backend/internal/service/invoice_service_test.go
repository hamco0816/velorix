//go:build unit

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

// fakeInvoiceMailer 记录最近一次发送的发票邮件，可注入失败用于测试。
type fakeInvoiceMailer struct {
	calls     int
	lastTo    string
	lastName  string
	lastBytes []byte
	lastMIME  string
	failWith  error
}

func (m *fakeInvoiceMailer) SendEmailWithAttachment(_ context.Context, to, _ string, _ string, attachmentName string, attachment []byte, attachmentMIME string) error {
	m.calls++
	m.lastTo = to
	m.lastName = attachmentName
	m.lastBytes = attachment
	m.lastMIME = attachmentMIME
	return m.failWith
}

// newInvoiceTestUser 创建一个测试用户。
func newInvoiceTestUser(t *testing.T, client *dbent.Client, email string) *dbent.User {
	t.Helper()
	user, err := client.User.Create().
		SetEmail(email).
		SetPasswordHash("hash").
		SetUsername(email).
		Save(context.Background())
	require.NoError(t, err)
	return user
}

// newCompletedOrder 创建一个已完成的付费订单（可开票）。
func newCompletedOrder(t *testing.T, client *dbent.Client, user *dbent.User, payAmount float64, suffix string) *dbent.PaymentOrder {
	t.Helper()
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(payAmount).
		SetPayAmount(payAmount).
		SetFeeRate(0).
		SetRechargeCode("INV-" + suffix).
		SetOutTradeNo("sub2_inv_" + suffix).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-" + suffix).
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(context.Background())
	require.NoError(t, err)
	return order
}

func personalInvoiceRequest(orderIDs []int64) ApplyInvoiceRequest {
	return ApplyInvoiceRequest{
		RecipientEmail: "recipient@example.com",
		TitleType:      InvoiceTitlePersonal,
		TitleName:      "个人",
		OrderIDs:       orderIDs,
	}
}

// 申请开票：金额按实付金额合计，订单被占用，状态为待开票。
func TestApplyInvoice_SumsPayAmountAndLocksOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "apply@example.com")
	o1 := newCompletedOrder(t, client, user, 100.00, "a1")
	o2 := newCompletedOrder(t, client, user, 49.99, "a2")

	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{o1.ID, o2.ID}))
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusPending, ir.Status)
	require.InDelta(t, 149.99, ir.Amount, 0.001)

	// 两个订单都被占用
	linked, err := client.PaymentOrder.Query().
		Where(paymentorder.InvoiceRequestIDEQ(ir.ID)).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, linked)
}

// 不传 order_ids 时，自动申请当前用户全部可开票订单。
func TestApplyInvoice_AutoIncludesAllInvoiceableOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "auto@example.com")
	other := newInvoiceTestUser(t, client, "other@example.com")
	o1 := newCompletedOrder(t, client, user, 40.25, "auto1")
	o2 := newCompletedOrder(t, client, user, 59.75, "auto2")
	_ = newCompletedOrder(t, client, other, 999, "auto-other")

	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 2, summary.TotalCount)
	require.InDelta(t, 100.00, summary.TotalAmount, 0.001)

	req := personalInvoiceRequest(nil)
	ir, err := svc.ApplyInvoice(ctx, user.ID, req)
	require.NoError(t, err)
	require.InDelta(t, 100.00, ir.Amount, 0.001)

	linked, err := client.PaymentOrder.Query().
		Where(paymentorder.InvoiceRequestIDEQ(ir.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, linked, 2)
	require.ElementsMatch(t, []int64{o1.ID, o2.ID}, []int64{linked[0].ID, linked[1].ID})
}

// 企业抬头缺税号应被拒绝。
func TestApplyInvoice_CompanyRequiresTaxID(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "company@example.com")
	order := newCompletedOrder(t, client, user, 200, "c1")

	_, err := svc.ApplyInvoice(ctx, user.ID, ApplyInvoiceRequest{
		RecipientEmail: "recipient@example.com",
		TitleType:      InvoiceTitleCompany,
		TitleName:      "某某公司",
		OrderIDs:       []int64{order.ID},
	})
	require.ErrorIs(t, err, ErrInvoiceTaxIDRequired)
}

// 非法邮箱应被拒绝。
func TestApplyInvoice_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "bademail@example.com")
	order := newCompletedOrder(t, client, user, 30, "e1")

	req := personalInvoiceRequest([]int64{order.ID})
	req.RecipientEmail = "not-an-email"
	_, err := svc.ApplyInvoice(ctx, user.ID, req)
	require.ErrorIs(t, err, ErrInvoiceInvalidEmail)
}

// 已被占用的订单不能再次申请。
func TestApplyInvoice_RejectsAlreadyLinkedOrder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "linked@example.com")
	order := newCompletedOrder(t, client, user, 60, "l1")

	_, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.NoError(t, err)

	_, err = svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.ErrorIs(t, err, ErrInvoiceOrderNotUsable)
}

// 可开票订单列表排除已占用和零实付订单。
func TestListInvoiceableOrders_ExcludesLinkedAndZeroPay(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "list@example.com")
	usable := newCompletedOrder(t, client, user, 100, "u1")
	zeroPay := newCompletedOrder(t, client, user, 0, "u2")
	require.NotNil(t, zeroPay)

	// 占用一笔
	linked := newCompletedOrder(t, client, user, 100, "u3")
	_, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{linked.ID}))
	require.NoError(t, err)

	orders, total, err := svc.ListInvoiceableOrders(ctx, user.ID, 1, 20)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, orders, 1)
	require.Equal(t, usable.ID, orders[0].ID)
}

// 取消申请释放订单。
func TestCancelInvoice_ReleasesOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "cancel@example.com")
	order := newCompletedOrder(t, client, user, 80, "x1")
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.NoError(t, err)

	require.NoError(t, svc.CancelInvoice(ctx, user.ID, ir.ID))

	reloaded, err := client.InvoiceRequest.Get(ctx, ir.ID)
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusCancelled, reloaded.Status)

	freed, err := client.PaymentOrder.Query().
		Where(paymentorder.InvoiceRequestIDIsNil(), paymentorder.IDEQ(order.ID)).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, freed)
}

// 开票：先发邮件（带 PDF 附件），成功后落库为已开票并留存发票号码。
func TestIssueInvoice_SendsEmailThenPersists(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	mailer := &fakeInvoiceMailer{}
	svc := NewInvoiceService(client, mailer, nil)

	user := newInvoiceTestUser(t, client, "issue@example.com")
	order := newCompletedOrder(t, client, user, 300, "i1")
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.NoError(t, err)

	pdf := []byte("%PDF-1.4 fake")
	issued, err := svc.IssueInvoice(ctx, 999, ir.ID, IssueInvoiceInput{
		PDF:           pdf,
		InvoiceNumber: "  25117000000123456789  ",
	})
	require.NoError(t, err)

	require.Equal(t, 1, mailer.calls)
	require.Equal(t, "recipient@example.com", mailer.lastTo)
	require.Equal(t, "application/pdf", mailer.lastMIME)
	require.Equal(t, pdf, mailer.lastBytes)

	require.Equal(t, InvoiceStatusIssued, issued.Status)
	require.NotNil(t, issued.InvoiceNumber)
	require.Equal(t, "25117000000123456789", *issued.InvoiceNumber) // 去除首尾空白
	require.True(t, issued.EmailSent)
	require.NotNil(t, issued.IssuedAt)
}

// 开票必须有发票号码。
func TestIssueInvoice_RequiresInvoiceNumber(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "noinv@example.com")
	order := newCompletedOrder(t, client, user, 50, "n1")
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.NoError(t, err)

	_, err = svc.IssueInvoice(ctx, 1, ir.ID, IssueInvoiceInput{PDF: []byte("x"), InvoiceNumber: "   "})
	require.ErrorIs(t, err, ErrInvoiceNumberRequired)
}

// 邮件发送失败时保持待开票状态，不标记已开。
func TestIssueInvoice_EmailFailureKeepsPending(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	mailer := &fakeInvoiceMailer{failWith: fmt.Errorf("smtp down")}
	svc := NewInvoiceService(client, mailer, nil)

	user := newInvoiceTestUser(t, client, "fail@example.com")
	order := newCompletedOrder(t, client, user, 70, "f1")
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.NoError(t, err)

	_, err = svc.IssueInvoice(ctx, 1, ir.ID, IssueInvoiceInput{PDF: []byte("x"), InvoiceNumber: "123"})
	require.Error(t, err)

	reloaded, err := client.InvoiceRequest.Get(ctx, ir.ID)
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusPending, reloaded.Status)
	require.False(t, reloaded.EmailSent)
}

// 驳回申请释放订单并写入原因。
func TestRejectInvoice_ReleasesOrders(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "reject@example.com")
	order := newCompletedOrder(t, client, user, 90, "r1")
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest([]int64{order.ID}))
	require.NoError(t, err)

	rejected, err := svc.RejectInvoice(ctx, 1, ir.ID, "信息不完整")
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusRejected, rejected.Status)

	freed, err := client.PaymentOrder.Query().
		Where(paymentorder.InvoiceRequestIDIsNil(), paymentorder.IDEQ(order.ID)).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, freed)
}

// 管理端关键字过滤命中抬头。
func TestListInvoices_KeywordFilter(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "kw@example.com")
	order := newCompletedOrder(t, client, user, 10, "k1")
	req := personalInvoiceRequest([]int64{order.ID})
	req.TitleName = "蓝天科技有限公司"
	req.TitleType = InvoiceTitleCompany
	req.TaxID = "91110000XXXXXXXX1A"
	_, err := svc.ApplyInvoice(ctx, user.ID, req)
	require.NoError(t, err)

	items, total, err := svc.ListInvoices(ctx, InvoiceListParams{Page: 1, PageSize: 20, Keyword: "蓝天"})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, items, 1)
	require.Equal(t, InvoiceTitleCompany, items[0].TitleType)
}

//go:build unit

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
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

// newBalanceRecharge 创建一笔已完成的余额充值订单（真实付费，作为可开票上限）。
func newBalanceRecharge(t *testing.T, client *dbent.Client, user *dbent.User, payAmount float64, suffix string) *dbent.PaymentOrder {
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

// newInvoiceEligibleGroup 创建一个支持开票的分组。
func newInvoiceEligibleGroup(t *testing.T, client *dbent.Client, name string, eligible bool) *dbent.Group {
	t.Helper()
	g, err := client.Group.Create().
		SetName(name).
		SetInvoiceEligible(eligible).
		Save(context.Background())
	require.NoError(t, err)
	return g
}

// newPlanOrder 创建一笔已完成的套餐订单，绑定指定分组。
func newPlanOrder(t *testing.T, client *dbent.Client, user *dbent.User, groupID int64, payAmount float64, suffix string) *dbent.PaymentOrder {
	t.Helper()
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(payAmount).
		SetPayAmount(payAmount).
		SetFeeRate(0).
		SetRechargeCode("PLAN-" + suffix).
		SetOutTradeNo("sub2_plan_" + suffix).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("plantrade-" + suffix).
		SetOrderType(payment.OrderTypeSubscription).
		SetSubscriptionGroupID(groupID).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(context.Background())
	require.NoError(t, err)
	return order
}

// setInvoiceableConsumed 设置用户累计的可开票消费额度（额度，倍率默认 1 时等于人民币）。
func setInvoiceableConsumed(t *testing.T, client *dbent.Client, userID int64, amount float64) {
	t.Helper()
	require.NoError(t, client.User.UpdateOneID(userID).SetInvoiceableConsumed(amount).Exec(context.Background()))
}

func personalInvoiceRequest(amount float64) ApplyInvoiceRequest {
	return ApplyInvoiceRequest{
		RecipientEmail: "recipient@example.com",
		TitleType:      InvoiceTitlePersonal,
		TitleName:      "个人",
		Amount:         amount,
	}
}

// multiplierSettingRepo 让发票服务读到指定充值倍率，用于验证额度→人民币折算。
type multiplierSettingRepo struct{ mult string }

func (r *multiplierSettingRepo) Get(context.Context, string) (*Setting, error) { return nil, nil }
func (r *multiplierSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	if key == SettingBalanceRechargeMult {
		return r.mult, nil
	}
	return "", nil
}
func (r *multiplierSettingRepo) Set(context.Context, string, string) error { return nil }
func (r *multiplierSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *multiplierSettingRepo) SetMultiple(context.Context, map[string]string) error { return nil }
func (r *multiplierSettingRepo) GetAll(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *multiplierSettingRepo) Delete(context.Context, string) error { return nil }

// 充值倍率 1.2（充100得120额度，含20赠送）：消费 60 额度在自建分组，
// 折算应为 60/1.2 = 50 元——精确剔除了赠送部分。
func TestInvoiceable_MultiplierExcludesGiftProportionally(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	settingSvc := NewSettingService(&multiplierSettingRepo{mult: "1.2"}, nil)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, settingSvc)

	user := newInvoiceTestUser(t, client, "mult@example.com")
	newBalanceRecharge(t, client, user, 100, "m1") // 真实付费 100 元 → 120 额度
	setInvoiceableConsumed(t, client, user.ID, 60) // 自建分组消费 60 额度

	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 50.00, summary.AvailableAmount, 0.001)
	require.InDelta(t, 50.00, summary.BalanceAmount, 0.001)
}

// 余额消费可开票：金额 = min(已消费额度/倍率, 真实充值付费)；金额省略时全额开票。
func TestApplyInvoice_BalanceConsumptionFullAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "apply@example.com")
	newBalanceRecharge(t, client, user, 150.00, "a1") // 真实付费 150
	setInvoiceableConsumed(t, client, user.ID, 149.99)

	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 149.99, summary.AvailableAmount, 0.001)

	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0)) // 0 = 全额
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusPending, ir.Status)
	require.InDelta(t, 149.99, ir.Amount, 0.001)
}

// 赠送排除：消费额度远超真实付费时，可开票被真实付费封顶。
func TestApplyInvoice_CapByRealPayment(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "gift@example.com")
	newBalanceRecharge(t, client, user, 100.00, "g1")  // 真实付费仅 100
	setInvoiceableConsumed(t, client, user.ID, 300.00) // 含 200 赠送额度被消费

	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 100.00, summary.AvailableAmount, 0.001)

	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
	require.NoError(t, err)
	require.InDelta(t, 100.00, ir.Amount, 0.001)
}

// 套餐：仅支持开票分组的套餐订单计入可开票额度。
func TestInvoiceable_EligiblePlanOnly(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "plan@example.com")
	eligible := newInvoiceEligibleGroup(t, client, "自建池", true)
	relayed := newInvoiceEligibleGroup(t, client, "对接池", false)
	newPlanOrder(t, client, user, eligible.ID, 50.00, "pe1")
	newPlanOrder(t, client, user, relayed.ID, 80.00, "pr1") // 不可开票分组，应排除

	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 50.00, summary.AvailableAmount, 0.001)
	require.InDelta(t, 50.00, summary.PlanAmount, 0.001)
}

// 部分开票：申请金额小于可开额度时，剩余额度仍可继续开票。
func TestApplyInvoice_PartialAmountLeavesRemainder(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "partial@example.com")
	newBalanceRecharge(t, client, user, 100.00, "p1")
	setInvoiceableConsumed(t, client, user.ID, 100.00)

	_, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(30.00))
	require.NoError(t, err)

	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 70.00, summary.AvailableAmount, 0.001)
}

// 申请金额超过可开额度应被拒绝。
func TestApplyInvoice_AmountExceedsAvailable(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "exceed@example.com")
	newBalanceRecharge(t, client, user, 100.00, "x1")
	setInvoiceableConsumed(t, client, user.ID, 100.00)

	_, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(150.00))
	require.ErrorIs(t, err, ErrInvoiceAmountExceedsAvailable)
}

// 无可开票金额时申请被拒绝。
func TestApplyInvoice_NothingToInvoice(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "empty@example.com")
	_, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
	require.ErrorIs(t, err, ErrInvoiceNothingToInvoice)
}

// 企业抬头缺税号应被拒绝。
func TestApplyInvoice_CompanyRequiresTaxID(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "company@example.com")
	newBalanceRecharge(t, client, user, 200, "c1")
	setInvoiceableConsumed(t, client, user.ID, 200)

	_, err := svc.ApplyInvoice(ctx, user.ID, ApplyInvoiceRequest{
		RecipientEmail: "recipient@example.com",
		TitleType:      InvoiceTitleCompany,
		TitleName:      "某某公司",
	})
	require.ErrorIs(t, err, ErrInvoiceTaxIDRequired)
}

// 非法邮箱应被拒绝。
func TestApplyInvoice_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "bademail@example.com")
	newBalanceRecharge(t, client, user, 30, "e1")
	setInvoiceableConsumed(t, client, user.ID, 30)

	req := personalInvoiceRequest(0)
	req.RecipientEmail = "not-an-email"
	_, err := svc.ApplyInvoice(ctx, user.ID, req)
	require.ErrorIs(t, err, ErrInvoiceInvalidEmail)
}

// 取消申请后额度释放，可重新开票。
func TestCancelInvoice_ReleasesAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "cancel@example.com")
	newBalanceRecharge(t, client, user, 80, "x1")
	setInvoiceableConsumed(t, client, user.ID, 80)

	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
	require.NoError(t, err)

	// 申请后额度被占用
	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 0, summary.AvailableAmount, 0.001)

	require.NoError(t, svc.CancelInvoice(ctx, user.ID, ir.ID))

	reloaded, err := client.InvoiceRequest.Get(ctx, ir.ID)
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusCancelled, reloaded.Status)

	// 取消后额度恢复
	summary, err = svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 80, summary.AvailableAmount, 0.001)
}

// 开票：先发邮件（带 PDF 附件），成功后落库为已开票并留存发票号码。
func TestIssueInvoice_SendsEmailThenPersists(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	mailer := &fakeInvoiceMailer{}
	svc := NewInvoiceService(client, mailer, nil)

	user := newInvoiceTestUser(t, client, "issue@example.com")
	newBalanceRecharge(t, client, user, 300, "i1")
	setInvoiceableConsumed(t, client, user.ID, 300)
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
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
	newBalanceRecharge(t, client, user, 50, "n1")
	setInvoiceableConsumed(t, client, user.ID, 50)
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
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
	newBalanceRecharge(t, client, user, 70, "f1")
	setInvoiceableConsumed(t, client, user.ID, 70)
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
	require.NoError(t, err)

	_, err = svc.IssueInvoice(ctx, 1, ir.ID, IssueInvoiceInput{PDF: []byte("x"), InvoiceNumber: "123"})
	require.Error(t, err)

	reloaded, err := client.InvoiceRequest.Get(ctx, ir.ID)
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusPending, reloaded.Status)
	require.False(t, reloaded.EmailSent)
}

// 驳回申请写入原因并释放额度。
func TestRejectInvoice_ReleasesAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "reject@example.com")
	newBalanceRecharge(t, client, user, 90, "r1")
	setInvoiceableConsumed(t, client, user.ID, 90)
	ir, err := svc.ApplyInvoice(ctx, user.ID, personalInvoiceRequest(0))
	require.NoError(t, err)

	rejected, err := svc.RejectInvoice(ctx, 1, ir.ID, "信息不完整")
	require.NoError(t, err)
	require.Equal(t, InvoiceStatusRejected, rejected.Status)

	// 驳回后额度恢复
	summary, err := svc.GetInvoiceableSummary(ctx, user.ID)
	require.NoError(t, err)
	require.InDelta(t, 90, summary.AvailableAmount, 0.001)
}

// 管理端关键字过滤命中抬头。
func TestListInvoices_KeywordFilter(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	svc := NewInvoiceService(client, &fakeInvoiceMailer{}, nil)

	user := newInvoiceTestUser(t, client, "kw@example.com")
	newBalanceRecharge(t, client, user, 10, "k1")
	setInvoiceableConsumed(t, client, user.ID, 10)
	req := personalInvoiceRequest(0)
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

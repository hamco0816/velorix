//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func newFinancePaymentOrder(t *testing.T, client *dbent.Client, user *dbent.User, status string, amount, payAmount, refundAmount float64, paidAt *time.Time, suffix string) *dbent.PaymentOrder {
	t.Helper()
	create := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(amount).
		SetPayAmount(payAmount).
		SetFeeRate(0).
		SetRechargeCode("FIN-" + suffix).
		SetOutTradeNo("sub2_finance_" + suffix).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-finance-" + suffix).
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(status).
		SetRefundAmount(refundAmount).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com")
	if paidAt != nil {
		create.SetPaidAt(*paidAt)
	}
	order, err := create.Save(context.Background())
	require.NoError(t, err)
	return order
}

func TestGetFinanceRevenueStats_UsesSettledNetIncome(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user := newInvoiceTestUser(t, client, "finance@example.com")
	svc := &PaymentService{entClient: client}

	june1 := time.Date(2026, 6, 1, 9, 0, 0, 0, time.Local)
	june2 := time.Date(2026, 6, 2, 10, 0, 0, 0, time.Local)
	newFinancePaymentOrder(t, client, user, OrderStatusCompleted, 100, 100, 0, &june1, "completed")
	newFinancePaymentOrder(t, client, user, OrderStatusPartiallyRefunded, 100, 110, 50, &june2, "partial")
	newFinancePaymentOrder(t, client, user, OrderStatusRefunded, 80, 80, 80, &june2, "refunded")
	newFinancePaymentOrder(t, client, user, OrderStatusPending, 60, 60, 0, nil, "pending")

	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	stats, err := svc.GetFinanceRevenueStats(ctx, start, end, "month")
	require.NoError(t, err)

	require.Equal(t, "month", stats.Period)
	require.Equal(t, "2026-06-01", stats.StartDate)
	require.Equal(t, "2026-06-30", stats.EndDate)
	require.Len(t, stats.Series, 30)
	require.Equal(t, 2, stats.TotalCount)
	require.InDelta(t, 155, stats.TotalAmount, 0.001)
	require.InDelta(t, 210, stats.GrossAmount, 0.001)
	require.InDelta(t, 55, stats.RefundAmount, 0.001)
	require.InDelta(t, 77.5, stats.AvgAmount, 0.001)
	require.InDelta(t, 100, stats.Series[0].Amount, 0.001)
	require.InDelta(t, 55, stats.Series[1].Amount, 0.001)
	require.InDelta(t, 55, stats.Series[1].RefundAmount, 0.001)
}

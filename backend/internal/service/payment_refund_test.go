//go:build unit

package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateRefundRequestRejectsLegacyGuessedProviderInstance(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-legacy@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-legacy-user").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-instance").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetAllowUserRefund(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("REFUND-LEGACY-ORDER").
		SetOutTradeNo("sub2_refund_legacy_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-legacy-refund").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient: client,
	}

	_, err = svc.validateRefundRequest(ctx, order.ID, user.ID)
	require.Error(t, err)
	require.Equal(t, "USER_REFUND_DISABLED", infraerrors.Reason(err))
}

func TestPrepareRefundRejectsLegacyGuessedProviderInstance(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-legacy-admin@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-legacy-admin-user").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-admin-instance").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetAllowUserRefund(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(188).
		SetPayAmount(188).
		SetFeeRate(0).
		SetRechargeCode("REFUND-LEGACY-ADMIN-ORDER").
		SetOutTradeNo("sub2_refund_legacy_admin_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-legacy-admin-refund").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient: client,
	}

	plan, result, err := svc.PrepareRefund(ctx, order.ID, 0, "", false, false)
	require.Nil(t, plan)
	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "REFUND_DISABLED", infraerrors.Reason(err))
}

// TestPrepareRefund_RefundRequestedTreatsAmountAsPending 验证 REFUND_REQUESTED 状态下
// 用户预先写入 refund_amount 的"申请金额"在管理员审批阶段不应被当成"已退金额"。
// 全额申请退款后，PrepareRefund 必须按 maxRefundable=order.Amount（已退=0）计算。
func TestPrepareRefund_RefundRequestedTreatsAmountAsPending(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-requested-amt@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-requested-amt").
		Save(ctx)
	require.NoError(t, err)

	inst, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-requested-amt").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetAllowUserRefund(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	instID := strconv.FormatInt(inst.ID, 10)
	// 模拟 RequestRefund 已经把"申请金额=订单全额"写入 refund_amount
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("REFUND-REQUESTED-AMT").
		SetOutTradeNo("sub2_refund_requested_amt").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-refund-requested-amt").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusRefundRequested).
		SetRefundAmount(100). // 申请金额 = 100
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID(instID).
		SetProviderKey(payment.TypeAlipay).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}

	// 不传 amt → 默认应取 maxRefundable = Amount - alreadyRefunded(REFUND_REQUESTED 视为 0) = 100
	plan, result, err := svc.PrepareRefund(ctx, order.ID, 0, "", false, false)
	require.NoError(t, err)
	require.Nil(t, result)
	require.NotNil(t, plan)
	require.InDelta(t, 100.0, plan.RefundAmount, 0.0001,
		"REFUND_REQUESTED 下默认退款金额应为订单全额，而不是 0（之前 bug：把申请金额当已退）")
}

// TestPrepareRefund_PartiallyRefundedAcceptsRemainingAmount 验证已部分退款的订单
// 还能继续退款，且按"剩余可退金额"计算上限。
func TestPrepareRefund_PartiallyRefundedAcceptsRemainingAmount(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-partial-second@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-partial-second").
		Save(ctx)
	require.NoError(t, err)

	inst, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-partial-second").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetAllowUserRefund(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	instID := strconv.FormatInt(inst.ID, 10)
	// 已部分退款：金额 100，已退 30，剩余 70
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("REFUND-PARTIAL-SECOND").
		SetOutTradeNo("sub2_refund_partial_second").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-refund-partial-second").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPartiallyRefunded).
		SetRefundAmount(30).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID(instID).
		SetProviderKey(payment.TypeAlipay).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}

	// 退 50，应该被允许（remaining=70）
	plan, _, err := svc.PrepareRefund(ctx, order.ID, 50, "second refund", false, false)
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.InDelta(t, 50.0, plan.RefundAmount, 0.0001)

	// 退 80，超过 remaining 70 应该被拒绝
	_, _, err = svc.PrepareRefund(ctx, order.ID, 80, "too much", false, false)
	require.Error(t, err)
	require.Equal(t, "REFUND_AMOUNT_EXCEEDED", infraerrors.Reason(err))
}

// TestRetrySeatReleaseForOrder_RejectsPartiallyRefunded 验证后台 worker / admin 手动重试入口
// 不会对部分退款订单触发完整 seat 释放（部分退款保留独享号权益）。
func TestRetrySeatReleaseForOrder_RejectsPartiallyRefunded(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("retry-partial-refund@example.com").
		SetPasswordHash("hash").
		SetUsername("retry-partial-refund").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("RETRY-PARTIAL-REFUND").
		SetOutTradeNo("sub2_retry_partial_refund").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-retry-partial-refund").
		SetOrderType(payment.OrderTypeSubscription).
		SetStatus(OrderStatusPartiallyRefunded).
		SetRefundAmount(40).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}

	err = svc.RetrySeatReleaseForOrder(ctx, order.ID)
	require.Error(t, err)
	require.Equal(t, "ORDER_NOT_FULLY_REFUNDED", infraerrors.Reason(err),
		"部分退款订单不允许重试 seat 释放：保留独享号权益的语义必须在所有入口一致")
}

// TestRestoreStatus_PreservesPartiallyRefunded 验证二次退款失败回滚时，
// 订单状态保持 PARTIALLY_REFUNDED 而不是默认的 COMPLETED。
func TestRestoreStatus_PreservesPartiallyRefunded(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("restore-partial@example.com").
		SetPasswordHash("hash").
		SetUsername("restore-partial").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("RESTORE-PARTIAL").
		SetOutTradeNo("sub2_restore_partial").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-restore-partial").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusPartiallyRefunded).
		SetRefundAmount(30).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	plan := &RefundPlan{
		OrderID: order.ID,
		Order:   order, // 保留原始 PartiallyRefunded 状态
	}
	svc.restoreStatus(ctx, plan)

	updated, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPartiallyRefunded, updated.Status,
		"二次退款失败的订单回滚后必须保持 PARTIALLY_REFUNDED，不能掉回 COMPLETED")
}

func TestGwRefundRejectsAlipayMerchantIdentitySnapshotMismatch(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-snapshot-mismatch@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-snapshot-mismatch-user").
		Save(ctx)
	require.NoError(t, err)

	inst, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-mismatch-instance").
		SetConfig(encryptWebhookProviderConfig(t, map[string]string{
			"appId":      "runtime-alipay-app",
			"privateKey": "runtime-private-key",
		})).
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	instID := strconv.FormatInt(inst.ID, 10)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("REFUND-SNAPSHOT-MISMATCH-ORDER").
		SetOutTradeNo("sub2_refund_snapshot_mismatch_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-refund-snapshot-mismatch").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID(instID).
		SetProviderKey(payment.TypeAlipay).
		SetProviderSnapshot(map[string]any{
			"schema_version":       2,
			"provider_instance_id": instID,
			"provider_key":         payment.TypeAlipay,
			"merchant_app_id":      "expected-alipay-app",
		}).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient:    client,
		loadBalancer: newWebhookProviderTestLoadBalancer(client),
	}

	err = svc.gwRefund(ctx, &RefundPlan{
		OrderID:       order.ID,
		Order:         order,
		RefundAmount:  order.Amount,
		GatewayAmount: order.Amount,
		Reason:        "snapshot mismatch",
	})
	require.ErrorContains(t, err, "alipay app_id mismatch")
}

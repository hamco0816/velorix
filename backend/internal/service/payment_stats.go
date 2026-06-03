package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
)

// --- Dashboard & Analytics ---

func (s *PaymentService) GetDashboardStats(ctx context.Context, days int) (*DashboardStats, error) {
	if days <= 0 {
		days = 30
	}
	now := time.Now()
	since := now.AddDate(0, 0, -days)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	paidStatuses := []string{OrderStatusCompleted, OrderStatusPaid, OrderStatusRecharging}

	orders, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusIn(paidStatuses...),
			paymentorder.PaidAtGTE(since),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	st := &DashboardStats{}
	computeBasicStats(st, orders, todayStart)

	st.PendingOrders, err = s.entClient.PaymentOrder.Query().
		Where(paymentorder.StatusEQ(OrderStatusPending)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	st.DailySeries = buildDailySeries(orders, since, days)
	st.PaymentMethods = buildMethodDistribution(orders)
	st.TopUsers = buildTopUsers(orders)

	return st, nil
}

func (s *PaymentService) GetFinanceRevenueStats(ctx context.Context, start, end time.Time, period string) (*FinanceRevenueStats, error) {
	if !end.After(start) {
		end = start.AddDate(0, 0, 1)
	}
	orders, err := s.entClient.PaymentOrder.Query().
		Where(
			paymentorder.StatusIn(OrderStatusCompleted, OrderStatusPartiallyRefunded),
			paymentorder.PaidAtGTE(start),
			paymentorder.PaidAtLT(end),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	st := &FinanceRevenueStats{
		Period:    period,
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.AddDate(0, 0, -1).Format("2006-01-02"),
		Series:    buildFinanceRevenueSeries(orders, start, end),
	}
	for _, bucket := range st.Series {
		st.TotalAmount += bucket.Amount
		st.GrossAmount += bucket.GrossAmount
		st.RefundAmount += bucket.RefundAmount
		st.TotalCount += bucket.Count
	}
	st.TotalAmount = roundMoney(st.TotalAmount)
	st.GrossAmount = roundMoney(st.GrossAmount)
	st.RefundAmount = roundMoney(st.RefundAmount)
	if st.TotalCount > 0 {
		st.AvgAmount = roundMoney(st.TotalAmount / float64(st.TotalCount))
	}
	return st, nil
}

func computeBasicStats(st *DashboardStats, orders []*dbent.PaymentOrder, todayStart time.Time) {
	var totalAmount, todayAmount float64
	var todayCount int
	for _, o := range orders {
		totalAmount += o.PayAmount
		if o.PaidAt != nil && !o.PaidAt.Before(todayStart) {
			todayAmount += o.PayAmount
			todayCount++
		}
	}
	st.TotalAmount = math.Round(totalAmount*100) / 100
	st.TodayAmount = math.Round(todayAmount*100) / 100
	st.TotalCount = len(orders)
	st.TodayCount = todayCount
	if st.TotalCount > 0 {
		st.AvgAmount = math.Round(totalAmount/float64(st.TotalCount)*100) / 100
	}
}

func buildFinanceRevenueSeries(orders []*dbent.PaymentOrder, start, end time.Time) []FinanceRevenueBucket {
	bucketMap := make(map[string]*FinanceRevenueBucket)
	for _, o := range orders {
		if o.PaidAt == nil {
			continue
		}
		date := o.PaidAt.Format("2006-01-02")
		bucket, ok := bucketMap[date]
		if !ok {
			bucket = &FinanceRevenueBucket{Date: date}
			bucketMap[date] = bucket
		}
		gross, refund, net := financeOrderAmounts(o)
		if net <= 0 {
			continue
		}
		bucket.GrossAmount += gross
		bucket.RefundAmount += refund
		bucket.Amount += net
		bucket.Count++
	}

	days := int(end.Sub(start).Hours() / 24)
	if days < 1 {
		days = 1
	}
	series := make([]FinanceRevenueBucket, 0, days)
	for i := 0; i < days; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		if bucket, ok := bucketMap[date]; ok {
			bucket.GrossAmount = roundMoney(bucket.GrossAmount)
			bucket.RefundAmount = roundMoney(bucket.RefundAmount)
			bucket.Amount = roundMoney(bucket.Amount)
			series = append(series, *bucket)
		} else {
			series = append(series, FinanceRevenueBucket{Date: date})
		}
	}
	return series
}

func financeOrderAmounts(o *dbent.PaymentOrder) (gross, refund, net float64) {
	gross = o.PayAmount
	if o.Status == OrderStatusPartiallyRefunded && o.RefundAmount > 0 {
		refund = calculateGatewayRefundAmount(o.Amount, o.PayAmount, o.RefundAmount)
	}
	net = gross - refund
	if net < 0 {
		net = 0
	}
	return roundMoney(gross), roundMoney(refund), roundMoney(net)
}

func buildDailySeries(orders []*dbent.PaymentOrder, since time.Time, days int) []DailyStats {
	dailyMap := make(map[string]*DailyStats)
	for _, o := range orders {
		if o.PaidAt == nil {
			continue
		}
		date := o.PaidAt.Format("2006-01-02")
		ds, ok := dailyMap[date]
		if !ok {
			ds = &DailyStats{Date: date}
			dailyMap[date] = ds
		}
		ds.Amount += o.PayAmount
		ds.Count++
	}
	series := make([]DailyStats, 0, days)
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i+1).Format("2006-01-02")
		if ds, ok := dailyMap[date]; ok {
			ds.Amount = math.Round(ds.Amount*100) / 100
			series = append(series, *ds)
		} else {
			series = append(series, DailyStats{Date: date})
		}
	}
	return series
}

func buildMethodDistribution(orders []*dbent.PaymentOrder) []PaymentMethodStat {
	methodMap := make(map[string]*PaymentMethodStat)
	for _, o := range orders {
		ms, ok := methodMap[o.PaymentType]
		if !ok {
			ms = &PaymentMethodStat{Type: o.PaymentType}
			methodMap[o.PaymentType] = ms
		}
		ms.Amount += o.PayAmount
		ms.Count++
	}
	methods := make([]PaymentMethodStat, 0, len(methodMap))
	for _, ms := range methodMap {
		ms.Amount = math.Round(ms.Amount*100) / 100
		methods = append(methods, *ms)
	}
	return methods
}

func buildTopUsers(orders []*dbent.PaymentOrder) []TopUserStat {
	userMap := make(map[int64]*TopUserStat)
	for _, o := range orders {
		us, ok := userMap[o.UserID]
		if !ok {
			us = &TopUserStat{UserID: o.UserID, Email: o.UserEmail}
			userMap[o.UserID] = us
		}
		us.Amount += o.PayAmount
	}
	userList := make([]*TopUserStat, 0, len(userMap))
	for _, us := range userMap {
		us.Amount = math.Round(us.Amount*100) / 100
		userList = append(userList, us)
	}
	sort.Slice(userList, func(i, j int) bool {
		return userList[i].Amount > userList[j].Amount
	})
	limit := topUsersLimit
	if len(userList) < limit {
		limit = len(userList)
	}
	result := make([]TopUserStat, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, *userList[i])
	}
	return result
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

// --- Audit Logs ---

func (s *PaymentService) writeAuditLog(ctx context.Context, oid int64, action, op string, detail map[string]any) {
	dj, _ := json.Marshal(detail)
	_, err := s.entClient.PaymentAuditLog.Create().SetOrderID(strconv.FormatInt(oid, 10)).SetAction(action).SetDetail(string(dj)).SetOperator(op).Save(ctx)
	if err != nil {
		slog.Error("audit log failed", "orderID", oid, "action", action, "error", err)
	}
}

func (s *PaymentService) GetOrderAuditLogs(ctx context.Context, oid int64) ([]*dbent.PaymentAuditLog, error) {
	return s.entClient.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(oid, 10))).Order(paymentauditlog.ByCreatedAt()).All(ctx)
}

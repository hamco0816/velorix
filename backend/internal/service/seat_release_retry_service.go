package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// SeatReleaseRetryService 周期性扫描"退款成功但 seat 释放失败"的订单，自动重试释放。
//
// 与 ExclusiveSeatExpiryService 同款模式：定时 ticker + best-effort batch 处理。
//
// 闭环逻辑：
//   - PaymentService.recordSeatReleaseFailure 在主流程失败时写 EXCLUSIVE_SEAT_RELEASE_FAILED 审计
//   - 本 worker 周期扫该 action，对未补到 RELEASED_ON_REFUND/RENEWAL_REVOKED 的项调
//     PaymentService.RetrySeatReleaseForOrder 重试一次
//   - 重试成功会写新的 RELEASED_ON_REFUND/REVOKED 审计，下一轮扫描就跳过；持续失败会再写一条
//     RELEASE_FAILED，方便告警系统/管理后台识别"长期未闭环"的订单
//
// 默认参数：每 5 分钟跑一次，扫描最近 24 小时内的失败项，每轮最多处理 100 单。
type SeatReleaseRetryService struct {
	paymentSvc *PaymentService
	interval   time.Duration
	since      time.Duration
	batchLimit int
	stopCh     chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

func NewSeatReleaseRetryService(paymentSvc *PaymentService, interval time.Duration) *SeatReleaseRetryService {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return &SeatReleaseRetryService{
		paymentSvc: paymentSvc,
		interval:   interval,
		since:      24 * time.Hour,
		batchLimit: 100,
		stopCh:     make(chan struct{}),
	}
}

func (s *SeatReleaseRetryService) Start() {
	if s == nil || s.paymentSvc == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		// 启动后立即跑一次，避免崩溃重启后等一个周期才修
		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *SeatReleaseRetryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}

func (s *SeatReleaseRetryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pending, err := s.paymentSvc.ListPendingSeatReleaseFailures(ctx, s.since, s.batchLimit)
	if err != nil {
		slog.Error("[SeatReleaseRetry] list pending failed", "error", err)
		return
	}
	if len(pending) == 0 {
		return
	}

	retried, succeeded, failed := 0, 0, 0
	for _, item := range pending {
		if item.OrderID <= 0 {
			continue
		}
		retried++
		if err := s.paymentSvc.RetrySeatReleaseForOrder(ctx, item.OrderID); err != nil {
			slog.Error("[SeatReleaseRetry] retry failed",
				"order_id", item.OrderID, "seat_id", item.SeatID, "stage", item.Stage, "error", err)
			failed++
			continue
		}
		// 注意：RetrySeatReleaseForOrder 内部不抛错也可能没真正修复（比如查不到 seat）
		// 闭环判定靠下一轮扫描时 hasAuditLog(RELEASED_ON_REFUND) 是否被填上
		succeeded++
	}
	slog.Info("[SeatReleaseRetry] cycle done",
		"pending", len(pending), "retried", retried, "succeeded", succeeded, "failed", failed)
}

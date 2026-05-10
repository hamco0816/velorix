package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// ExclusiveSeatExpiryService 定时扫描并回收过期的独享名额。
//
// 与 AccountExpiryService（上游账号过期）独立，专注独享名额的生命周期回收：
//   - 扫描 status='active' 且 expires_at < now 的 seat
//   - 调用 ExclusiveSeatService.ReleaseSeat 把 status 改 expired，清 account.assigned_seat_id
type ExclusiveSeatExpiryService struct {
	seatSvc    *ExclusiveSeatService
	interval   time.Duration
	batchLimit int
	stopCh     chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

func NewExclusiveSeatExpiryService(seatSvc *ExclusiveSeatService, interval time.Duration) *ExclusiveSeatExpiryService {
	return &ExclusiveSeatExpiryService{
		seatSvc:    seatSvc,
		interval:   interval,
		batchLimit: 200,
		stopCh:     make(chan struct{}),
	}
}

func (s *ExclusiveSeatExpiryService) Start() {
	if s == nil || s.seatSvc == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
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

func (s *ExclusiveSeatExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stopCh) })
	s.wg.Wait()
}

func (s *ExclusiveSeatExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	released, err := s.seatSvc.RunExpiryOnce(ctx, s.batchLimit)
	if err != nil {
		slog.Error("[ExclusiveSeatExpiry] run failed", "released", released, "error", err)
		return
	}
	if released > 0 {
		slog.Info("[ExclusiveSeatExpiry] released", "count", released)
	}
}

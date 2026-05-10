package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

// stubSeatLister 仅供 helper 单元测试使用，绕过 ent 与数据库。
type stubSeatResolver struct {
	accounts []*Account
	err      error
}

func (s *stubSeatResolver) GetByIDs(_ context.Context, _ []int64) ([]*Account, error) {
	return s.accounts, s.err
}

func makeAccount(id int64, priority int, schedulable bool, status string) *Account {
	return &Account{
		ID:          id,
		Priority:    priority,
		Schedulable: schedulable,
		Status:      status,
	}
}

func TestPickByPriorityThenLastUsed(t *testing.T) {
	now := time.Now()
	older := now.Add(-1 * time.Hour)

	cases := []struct {
		name     string
		input    []*Account
		expectID int64
	}{
		{
			name: "lower priority wins",
			input: []*Account{
				{ID: 1, Priority: 50},
				{ID: 2, Priority: 10},
				{ID: 3, Priority: 100},
			},
			expectID: 2,
		},
		{
			name: "same priority: nil last_used_at preferred (least recently used)",
			input: []*Account{
				{ID: 1, Priority: 10, LastUsedAt: &now},
				{ID: 2, Priority: 10}, // nil last_used_at
			},
			expectID: 2,
		},
		{
			name: "same priority: earlier last_used_at preferred",
			input: []*Account{
				{ID: 1, Priority: 10, LastUsedAt: &now},
				{ID: 2, Priority: 10, LastUsedAt: &older},
			},
			expectID: 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := pickByPriorityThenLastUsed(c.input)
			if got == nil || got.ID != c.expectID {
				t.Fatalf("got %v, want %d", got, c.expectID)
			}
		})
	}
}

func TestFilterSchedulableActiveAccounts(t *testing.T) {
	accs := []*Account{
		makeAccount(1, 10, true, domain.StatusActive),
		makeAccount(2, 10, false, domain.StatusActive), // schedulable=false
		makeAccount(3, 10, true, "disabled"),           // status disabled
	}
	out := filterSchedulableActiveAccounts(accs)
	if len(out) != 1 || out[0].ID != 1 {
		t.Fatalf("want only id=1, got %v", out)
	}
}

func TestTrySelectExclusiveSeat_NoSeatService(t *testing.T) {
	// nil seatSvc → noop（让调度走原共享池路径）
	gid := int64(1)
	acc, hit, err := trySelectExclusiveSeatAccount(context.Background(), nil, &stubSeatResolver{}, &gid, nil)
	if err != nil || hit || acc != nil {
		t.Fatalf("expected noop, got err=%v hit=%v acc=%v", err, hit, acc)
	}
}

func TestTrySelectExclusiveSeat_NoUserID(t *testing.T) {
	// ctx 里没有 user_id → noop
	gid := int64(1)
	dummySvc := &ExclusiveSeatService{} // 不会被调到（user_id 缺失时提前返回）
	acc, hit, err := trySelectExclusiveSeatAccount(context.Background(), dummySvc, &stubSeatResolver{}, &gid, nil)
	if err != nil || hit || acc != nil {
		t.Fatalf("expected noop without user_id, got err=%v hit=%v", err, hit)
	}
}

func TestErrNoUsableExclusiveAccountIsExported(t *testing.T) {
	// 仅检验 sentinel error 被正确导出（防止重命名后破坏调用方判断）
	if !errors.Is(ErrNoUsableExclusiveAccount, ErrNoUsableExclusiveAccount) {
		t.Fatalf("ErrNoUsableExclusiveAccount must equal itself")
	}
}

func TestCtxKeySub2APIUserIDExists(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxkey.Sub2APIUserID, int64(42))
	if v, ok := ctx.Value(ctxkey.Sub2APIUserID).(int64); !ok || v != 42 {
		t.Fatalf("expected to read user_id=42 from ctx, got %v ok=%v", v, ok)
	}
}

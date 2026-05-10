package service

import (
	"testing"
	"time"
)

// 阈值解析：负数/非数字/越界回退到 0；正常值原样返回；> 100 截断到 100
func TestParseSystemOverloadThreshold(t *testing.T) {
	cases := []struct {
		raw  string
		want float64
	}{
		{"", 0},
		{" 80 ", 80},
		{"75.5", 75.5},
		{"-1", 0},
		{"abc", 0},
		{"120", 100},
		{"99.99", 99.99},
	}
	for _, c := range cases {
		got := parseSystemOverloadThreshold(c.raw)
		if got != c.want {
			t.Fatalf("parseSystemOverloadThreshold(%q) = %v, want %v", c.raw, got, c.want)
		}
	}
}

// CheckOverload 三个维度任一超阈值都应返回 overloaded；阈值=0 表示该维度不参与判定
func TestCheckOverload(t *testing.T) {
	m := &SystemOverloadMonitor{}
	m.latest.Store(&SystemOverloadStatus{
		CPUUsage:    85,
		MemoryUsage: 60,
		DiskUsage:   40,
		UpdatedAt:   time.Now(),
	})

	t.Run("disabled returns no overload", func(t *testing.T) {
		ov, _, _, _ := m.CheckOverload(&SystemOverloadSettings{Enabled: false, CPUThreshold: 50})
		if ov {
			t.Fatal("expected no overload when settings disabled")
		}
	})

	t.Run("cpu over threshold", func(t *testing.T) {
		ov, dim, cur, th := m.CheckOverload(&SystemOverloadSettings{Enabled: true, CPUThreshold: 80})
		if !ov || dim != "cpu" || cur != 85 || th != 80 {
			t.Fatalf("expected cpu overload, got ov=%v dim=%q cur=%v th=%v", ov, dim, cur, th)
		}
	})

	t.Run("zero threshold disables that dimension", func(t *testing.T) {
		// CPU=85 but threshold=0 应跳过 CPU 维度，转到 memory（60 < 80）→ 不超载
		ov, _, _, _ := m.CheckOverload(&SystemOverloadSettings{Enabled: true, CPUThreshold: 0, MemoryThreshold: 80})
		if ov {
			t.Fatal("expected no overload when all enabled thresholds satisfied")
		}
	})

	t.Run("first overloaded dimension wins", func(t *testing.T) {
		// 顺序是 cpu → memory → disk；同时超载时返回 cpu
		ov, dim, _, _ := m.CheckOverload(&SystemOverloadSettings{
			Enabled: true, CPUThreshold: 50, MemoryThreshold: 50, DiskThreshold: 30,
		})
		if !ov || dim != "cpu" {
			t.Fatalf("expected cpu to win, got ov=%v dim=%q", ov, dim)
		}
	})

	t.Run("disk only", func(t *testing.T) {
		ov, dim, _, _ := m.CheckOverload(&SystemOverloadSettings{Enabled: true, DiskThreshold: 30})
		if !ov || dim != "disk" {
			t.Fatalf("expected disk overload, got ov=%v dim=%q", ov, dim)
		}
	})
}

// nil monitor 必须 fail-open（不能 panic 也不能误报）
func TestSystemOverloadMonitorNilSafe(t *testing.T) {
	var m *SystemOverloadMonitor
	if v := m.GetStatus(); v == nil {
		t.Fatal("nil monitor GetStatus must not return nil")
	}
	if ov, _, _, _ := m.CheckOverload(&SystemOverloadSettings{Enabled: true, CPUThreshold: 1}); ov {
		t.Fatal("nil monitor must fail-open")
	}
}

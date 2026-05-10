package service

import (
	"context"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// 系统过载保护：进程内采样 CPU/内存/磁盘 → atomic 缓存 → 中间件 O(1) 读取做实时门控
//
// 与 OpsMetricsCollector 的区别：
//   - OpsMetricsCollector：分布式 leader-only，60s 一次，写 DB 用于监控告警面板
//   - SystemOverloadMonitor：每实例独立，5s 一次，仅内存缓存用于雪崩保护门控
const (
	systemOverloadSampleInterval     = 5 * time.Second
	systemOverloadDiskSampleInterval = 30 * time.Second
)

// SystemOverloadStatus 单次采样结果
type SystemOverloadStatus struct {
	CPUUsage    float64
	MemoryUsage float64
	DiskUsage   float64
	UpdatedAt   time.Time
}

// SystemOverloadSettings 中间件门控阈值
type SystemOverloadSettings struct {
	Enabled         bool
	CPUThreshold    float64 // 0-100，0 表示该项不参与判定
	MemoryThreshold float64
	DiskThreshold   float64
}

// SystemOverloadMonitor 进程内系统资源采样器
type SystemOverloadMonitor struct {
	settingService *SettingService

	latest   atomic.Value // *SystemOverloadStatus
	settings atomic.Value // *SystemOverloadSettings —— 中间件每次请求读这里，避免回 DB

	stopCh    chan struct{}
	startOnce uint32 // 用 atomic.CompareAndSwapUint32 替代 sync.Once 以便测试重新启动

	diskPath string
}

// NewSystemOverloadMonitor 创建系统过载监控器（settingService 用于读取门控阈值）。
func NewSystemOverloadMonitor(settingService *SettingService) *SystemOverloadMonitor {
	m := &SystemOverloadMonitor{
		settingService: settingService,
		stopCh:         make(chan struct{}),
		diskPath:       defaultDiskMonitorPath(),
	}
	m.latest.Store(&SystemOverloadStatus{UpdatedAt: time.Now()})
	// 默认禁用，避免 settings 还没第一次加载时中间件 nil panic
	m.settings.Store(&SystemOverloadSettings{})
	return m
}

// Settings 返回最新一次后台刷新到的过载阈值，O(1) atomic 读。
// 中间件每次请求都会调，必须避免 DB query —— 真正的刷新由 run 循环里 5s 一次完成。
func (m *SystemOverloadMonitor) Settings() *SystemOverloadSettings {
	if m == nil {
		return &SystemOverloadSettings{}
	}
	if v, ok := m.settings.Load().(*SystemOverloadSettings); ok && v != nil {
		return v
	}
	return &SystemOverloadSettings{}
}

// refreshSettings 后台调用：从 settingService 读 DB 一次，存入 atomic 缓存。
// 失败时保留旧值（fail-stale 比把开启的保护悄悄禁用更安全）：
// 走 GetSystemOverloadSettingsChecked 拿到 (settings, ok)，仅 ok=true 时才覆盖缓存。
func (m *SystemOverloadMonitor) refreshSettings(ctx context.Context) {
	if m == nil || m.settingService == nil {
		return
	}
	cur, ok := m.settingService.GetSystemOverloadSettingsChecked(ctx)
	if !ok || cur == nil {
		// DB 抖动：保留上次成功值，不动 atomic 缓存
		return
	}
	m.settings.Store(cur)
}

// Start 启动后台采样 goroutine。重复调用安全。
func (m *SystemOverloadMonitor) Start(ctx context.Context) {
	if m == nil {
		return
	}
	if !atomic.CompareAndSwapUint32(&m.startOnce, 0, 1) {
		return
	}
	go m.run(ctx)
}

// Stop 停止后台 goroutine。
func (m *SystemOverloadMonitor) Stop() {
	if m == nil {
		return
	}
	select {
	case <-m.stopCh:
		return
	default:
		close(m.stopCh)
	}
}

// GetStatus 读取最新采样状态（O(1)）。
func (m *SystemOverloadMonitor) GetStatus() *SystemOverloadStatus {
	if m == nil {
		return &SystemOverloadStatus{}
	}
	v, _ := m.latest.Load().(*SystemOverloadStatus)
	if v == nil {
		return &SystemOverloadStatus{}
	}
	return v
}

// CheckOverload 判定当前资源是否超过任一启用阈值。
// 返回 (overloaded, 维度名, 当前值, 阈值)；overloaded=false 时其它返回值为零值。
func (m *SystemOverloadMonitor) CheckOverload(settings *SystemOverloadSettings) (bool, string, float64, float64) {
	if m == nil || settings == nil || !settings.Enabled {
		return false, "", 0, 0
	}
	status := m.GetStatus()
	if settings.CPUThreshold > 0 && status.CPUUsage > settings.CPUThreshold {
		return true, "cpu", status.CPUUsage, settings.CPUThreshold
	}
	if settings.MemoryThreshold > 0 && status.MemoryUsage > settings.MemoryThreshold {
		return true, "memory", status.MemoryUsage, settings.MemoryThreshold
	}
	if settings.DiskThreshold > 0 && status.DiskUsage > settings.DiskThreshold {
		return true, "disk", status.DiskUsage, settings.DiskThreshold
	}
	return false, "", 0, 0
}

// GetSystemOverloadSettings 从配置中读取门控阈值。任何错误都回退到禁用状态（fail-open）。
//
// 注意：调用方如果是后台缓存刷新（如 SystemOverloadMonitor.refreshSettings），
// 应该用 GetSystemOverloadSettingsChecked 区分"成功读到禁用配置"与"读 DB 失败"，
// 避免 DB 抖动时把已开启的保护静默切到 disabled。
func (s *SettingService) GetSystemOverloadSettings(ctx context.Context) *SystemOverloadSettings {
	settings, _ := s.GetSystemOverloadSettingsChecked(ctx)
	return settings
}

// GetSystemOverloadSettingsChecked 同上，但额外返回 ok 表示本次是否真正从 DB 读到了值。
//   - ok=true：settings 是 DB 当前真实值（即便 Enabled=false 也是用户的有效配置）
//   - ok=false：DB 调用失败，settings 是 fail-open 占位（disabled），调用方按需决定保留旧值还是落到默认
func (s *SettingService) GetSystemOverloadSettingsChecked(ctx context.Context) (*SystemOverloadSettings, bool) {
	if s == nil || s.settingRepo == nil {
		return &SystemOverloadSettings{}, false
	}
	dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), gatewaySensitiveFilterDBTimeout)
	defer cancel()
	values, err := s.settingRepo.GetMultiple(dbCtx, []string{
		SettingKeySystemOverloadProtectionEnabled,
		SettingKeySystemOverloadCPUThreshold,
		SettingKeySystemOverloadMemoryThreshold,
		SettingKeySystemOverloadDiskThreshold,
	})
	if err != nil {
		slog.Warn("system_overload: failed to read settings, treating as disabled (caller may keep stale)", "error", err)
		return &SystemOverloadSettings{}, false
	}
	return &SystemOverloadSettings{
		Enabled:         strings.EqualFold(strings.TrimSpace(values[SettingKeySystemOverloadProtectionEnabled]), "true"),
		CPUThreshold:    parseSystemOverloadThreshold(values[SettingKeySystemOverloadCPUThreshold]),
		MemoryThreshold: parseSystemOverloadThreshold(values[SettingKeySystemOverloadMemoryThreshold]),
		DiskThreshold:   parseSystemOverloadThreshold(values[SettingKeySystemOverloadDiskThreshold]),
	}, true
}

func parseSystemOverloadThreshold(raw string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func (m *SystemOverloadMonitor) run(ctx context.Context) {
	cpuTicker := time.NewTicker(systemOverloadSampleInterval)
	defer cpuTicker.Stop()
	diskTicker := time.NewTicker(systemOverloadDiskSampleInterval)
	defer diskTicker.Stop()

	// 立即采样一次 + 加载 settings 缓存，避免中间件第一次读到全零或回 DB
	m.sampleAll(ctx)
	m.refreshSettings(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-cpuTicker.C:
			m.sampleCPUAndMem(ctx)
			// 与 CPU/内存采样同节奏刷新阈值缓存（5s）—— 系统过载时会有大量请求打到中间件，
			// 不能每次请求都回 settingRepo.GetMultiple 给 DB 反向加压
			m.refreshSettings(ctx)
		case <-diskTicker.C:
			m.sampleDisk(ctx)
		}
	}
}

func (m *SystemOverloadMonitor) sampleAll(ctx context.Context) {
	m.sampleCPUAndMem(ctx)
	m.sampleDisk(ctx)
}

// sampleCPUAndMem 高频采样 CPU 与内存（5s）。磁盘采样开销大放在 sampleDisk 里 30s 跑一次。
func (m *SystemOverloadMonitor) sampleCPUAndMem(_ context.Context) {
	prev := m.GetStatus()
	next := *prev
	next.UpdatedAt = time.Now()

	if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
		next.CPUUsage = percents[0]
	}
	if memInfo, err := mem.VirtualMemory(); err == nil {
		next.MemoryUsage = memInfo.UsedPercent
	}
	m.latest.Store(&next)
}

func (m *SystemOverloadMonitor) sampleDisk(_ context.Context) {
	prev := m.GetStatus()
	next := *prev
	next.UpdatedAt = time.Now()

	if usage, err := disk.Usage(m.diskPath); err == nil {
		next.DiskUsage = usage.UsedPercent
	}
	m.latest.Store(&next)
}

// defaultDiskMonitorPath 选择监控的磁盘挂载点。Linux 用 /，Windows 用 C:\
func defaultDiskMonitorPath() string {
	if runtime.GOOS == "windows" {
		return `C:\`
	}
	return "/"
}

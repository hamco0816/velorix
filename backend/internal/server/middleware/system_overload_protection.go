package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// overloadLogLastNanos 控制过载日志节流（最多每 overloadLogInterval 一条），避免雪崩时日志风暴
var overloadLogLastNanos atomic.Int64

const overloadLogInterval = 5 * time.Second

// SystemOverloadProtection 在 LLM 转发路径上做实时门控：
// CPU/内存/磁盘 任一项超过配置阈值时直接 503，避免雪崩。
//
// 设计要点：
//   - settings 与 status 都是 atomic 读取，无 DB 调用、无锁开销
//   - settings 在 SystemOverloadMonitor 后台 5s 刷新一次进 atomic 缓存，
//     避免系统过载时中间件本身回 DB 给数据库反向加压（GPT round 15 #3）
//   - 默认 fail-open：监控未启用 / settings 未加载 / 监控未注入 都放行
//   - 错误响应通过 writeError（per 协议格式）返回，避免暴露内部信息
//
// settingService 仍传入是为了向后兼容，未来如有"配置保存即推送 invalidate"需求时可在此处 hook。
func SystemOverloadProtection(monitor *service.SystemOverloadMonitor, settingService *service.SettingService, writeError GatewayErrorWriter) gin.HandlerFunc {
	_ = settingService
	return func(c *gin.Context) {
		if monitor == nil {
			c.Next()
			return
		}
		settings := monitor.Settings()
		if settings == nil || !settings.Enabled {
			c.Next()
			return
		}
		overloaded, dim, current, threshold := monitor.CheckOverload(settings)
		if !overloaded {
			c.Next()
			return
		}

		now := time.Now().UnixNano()
		last := overloadLogLastNanos.Load()
		if now-last > overloadLogInterval.Nanoseconds() && overloadLogLastNanos.CompareAndSwap(last, now) {
			slog.Warn("system overload: rejecting request",
				"dimension", dim,
				"current", fmt.Sprintf("%.1f", current),
				"threshold", fmt.Sprintf("%.0f", threshold),
				"path", c.Request.URL.Path,
			)
		}

		c.Header("Retry-After", "10")
		writeError(c, http.StatusServiceUnavailable, "service temporarily overloaded; please retry shortly")
		c.Abort()
	}
}

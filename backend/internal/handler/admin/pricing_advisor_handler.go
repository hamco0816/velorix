package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PricingAdvisorHandler 提供"订阅定价助手"接口：
// - 按 (platform, tier) 聚合 5h/7d 滚动窗口成本统计
// - 按 (platform, tier) 返回 30 天日均成本趋势
type PricingAdvisorHandler struct {
	svc *service.PricingAdvisorService
}

// NewPricingAdvisorHandler 构造函数。
func NewPricingAdvisorHandler(svc *service.PricingAdvisorService) *PricingAdvisorHandler {
	return &PricingAdvisorHandler{svc: svc}
}

// TierStats GET /api/v1/admin/pricing-advisor/tier-stats?days=30&platform=openai
func (h *PricingAdvisorHandler) TierStats(c *gin.Context) {
	params := h.parseParams(c)
	stats, err := h.svc.GetTierStats(c.Request.Context(), params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": stats, "days_window": params.DaysWindow})
}

// TierTrend GET /api/v1/admin/pricing-advisor/tier-trend?days=30&platform=openai
func (h *PricingAdvisorHandler) TierTrend(c *gin.Context) {
	params := h.parseParams(c)
	trend, err := h.svc.GetTierTrend(c.Request.Context(), params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": trend, "days_window": params.DaysWindow})
}

func (h *PricingAdvisorHandler) parseParams(c *gin.Context) service.PricingAdvisorParams {
	days := 30
	if v := c.Query("days"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}
	return service.PricingAdvisorParams{
		DaysWindow: days,
		Platform:   strings.TrimSpace(c.Query("platform")),
	}
}

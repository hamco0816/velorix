package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// DesktopHandler 桌面客户端公开接口（下载页用，无需鉴权）。
type DesktopHandler struct {
	desktopReleaseService *service.DesktopReleaseService
}

func NewDesktopHandler(desktopReleaseService *service.DesktopReleaseService) *DesktopHandler {
	return &DesktopHandler{desktopReleaseService: desktopReleaseService}
}

// GetLatest 返回当前对外的最新版本信息（含安装包下载相对地址）。
// GET /api/v1/desktop/latest?channel=stable
func (h *DesktopHandler) GetLatest(c *gin.Context) {
	channel := c.DefaultQuery("channel", "stable")
	rel, err := h.desktopReleaseService.GetLatest(c.Request.Context(), channel)
	if err != nil {
		// 还没有发布任何版本：返回 available=false，下载页展示"敬请期待"
		response.Success(c, gin.H{"available": false})
		return
	}
	response.Success(c, gin.H{
		"available":   true,
		"version":     rel.Version,
		"channel":     rel.Channel,
		"notes":       rel.Notes,
		"mandatory":   rel.Mandatory,
		"file_size":   rel.FileSize,
		"setup_url":   "/desktop/updates/" + rel.SetupFile,
		"released_at": rel.CreatedAt,
	})
}

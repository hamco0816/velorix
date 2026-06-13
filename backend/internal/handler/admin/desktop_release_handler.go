package admin

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	maxDesktopSetupSize    = 300 << 20 // 安装包上限 300MB
	maxDesktopYmlSize      = 2 << 20   // latest.yml 上限 2MB
	maxDesktopBlockmapSize = 64 << 20  // blockmap 上限 64MB
)

var errDesktopFileTooLarge = errors.New("文件超过大小上限")

// DesktopReleaseHandler 桌面客户端版本发布（管理端）。
type DesktopReleaseHandler struct {
	desktopReleaseService *service.DesktopReleaseService
}

func NewDesktopReleaseHandler(desktopReleaseService *service.DesktopReleaseService) *DesktopReleaseHandler {
	return &DesktopReleaseHandler{desktopReleaseService: desktopReleaseService}
}

// List 版本列表
// GET /api/v1/admin/desktop-releases
func (h *DesktopReleaseHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	filters := service.DesktopReleaseListFilters{
		Channel: strings.TrimSpace(c.Query("channel")),
		Status:  strings.TrimSpace(c.Query("status")),
		Search:  strings.TrimSpace(c.Query("search")),
	}
	items, paginationResult, err := h.desktopReleaseService.List(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.DesktopRelease, 0, len(items))
	for i := range items {
		out = append(out, *dto.DesktopReleaseFromService(&items[i]))
	}
	response.Paginated(c, out, paginationResult.Total, page, pageSize)
}

// Create 上传并发布新版本（multipart：setup + latest_yml + 可选 blockmap + 表单字段）
// POST /api/v1/admin/desktop-releases
func (h *DesktopReleaseHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	setupName, setupData, err := readDesktopMultipartFile(c, "setup", maxDesktopSetupSize)
	if err != nil {
		response.BadRequest(c, "安装包读取失败："+err.Error())
		return
	}
	_, ymlData, err := readDesktopMultipartFile(c, "latest_yml", maxDesktopYmlSize)
	if err != nil {
		response.BadRequest(c, "latest.yml 读取失败："+err.Error())
		return
	}

	input := &service.CreateDesktopReleaseInput{
		Version:       strings.TrimSpace(c.PostForm("version")),
		Channel:       strings.TrimSpace(c.PostForm("channel")),
		Mandatory:     c.PostForm("mandatory") == "true",
		Notes:         c.PostForm("notes"),
		SetupFilename: setupName,
		SetupData:     setupData,
		LatestYml:     string(ymlData),
		ActorID:       &subject.UserID,
	}

	// blockmap 可选
	if blockmapName, blockmapData, e := readDesktopMultipartFile(c, "blockmap", maxDesktopBlockmapSize); e == nil {
		input.BlockmapFilename = blockmapName
		input.BlockmapData = blockmapData
	}

	created, err := h.desktopReleaseService.Create(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.DesktopReleaseFromService(created))
}

// Rollback 把某历史版本重新设为当前对外版本
// POST /api/v1/admin/desktop-releases/:id/rollback
func (h *DesktopReleaseHandler) Rollback(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid release ID")
		return
	}
	rel, err := h.desktopReleaseService.Rollback(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.DesktopReleaseFromService(rel))
}

// Delete 删除一个非当前版本（连同磁盘安装包）
// DELETE /api/v1/admin/desktop-releases/:id
func (h *DesktopReleaseHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid release ID")
		return
	}
	if err := h.desktopReleaseService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "deleted"})
}

// readDesktopMultipartFile 读取 multipart 文件字段到内存（带大小上限）。
func readDesktopMultipartFile(c *gin.Context, field string, maxSize int64) (string, []byte, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return "", nil, err
	}
	if fh.Size > maxSize {
		return "", nil, errDesktopFileTooLarge
	}
	f, err := fh.Open()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(io.LimitReader(f, maxSize+1))
	if err != nil {
		return "", nil, err
	}
	return fh.Filename, data, nil
}

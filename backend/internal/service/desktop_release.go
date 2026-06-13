package service

import (
	"context"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// 桌面客户端发布状态。
const (
	DesktopReleaseStatusActive     = "active"     // 当前对外提供更新
	DesktopReleaseStatusArchived   = "archived"   // 历史版本
	DesktopReleaseStatusRolledback = "rolledback" // 被回滚（曾经激活，现已被替换）
)

const (
	DesktopChannelStable = "stable"
)

var (
	ErrDesktopReleaseNotFound = infraerrors.NotFound("DESKTOP_RELEASE_NOT_FOUND", "桌面版本不存在")
	ErrDesktopReleaseExists   = infraerrors.Conflict(
		"DESKTOP_RELEASE_EXISTS",
		"该版本号在此通道下已存在",
	)
	ErrDesktopReleaseNilInput       = infraerrors.BadRequest("DESKTOP_RELEASE_INPUT_REQUIRED", "发布参数不能为空")
	ErrDesktopReleaseInvalidVersion = infraerrors.BadRequest("DESKTOP_RELEASE_VERSION_INVALID", "版本号不合法")
	ErrDesktopReleaseSetupRequired  = infraerrors.BadRequest("DESKTOP_RELEASE_SETUP_REQUIRED", "缺少安装包文件")
	ErrDesktopReleaseLatestRequired = infraerrors.BadRequest("DESKTOP_RELEASE_LATEST_REQUIRED", "缺少 latest.yml 文件")
	ErrDesktopReleaseActiveDeletion = infraerrors.BadRequest(
		"DESKTOP_RELEASE_ACTIVE_DELETION",
		"当前对外版本不能删除，请先回滚或发布新版本",
	)
)

// DesktopRelease 桌面客户端版本发布记录（service 领域模型，无复杂领域行为，不走 domain 包）。
type DesktopRelease struct {
	ID           int64
	Version      string
	Channel      string
	Mandatory    bool
	Notes        string
	SetupFile    string
	BlockmapFile string
	LatestYml    string
	FileSize     int64
	Status       string
	CreatedBy    *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DesktopReleaseListFilters 版本列表过滤条件。
type DesktopReleaseListFilters struct {
	Channel string
	Status  string
	Search  string
}

// DesktopReleaseRepository 桌面版本持久化接口。
type DesktopReleaseRepository interface {
	Create(ctx context.Context, r *DesktopRelease) error
	GetByID(ctx context.Context, id int64) (*DesktopRelease, error)
	GetByVersionChannel(ctx context.Context, version, channel string) (*DesktopRelease, error)
	GetActiveByChannel(ctx context.Context, channel string) (*DesktopRelease, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	// ArchiveActiveByChannel 把指定通道下当前 active 的记录（exceptID 除外）置为 archived。
	ArchiveActiveByChannel(ctx context.Context, channel string, exceptID int64) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params pagination.PaginationParams, filters DesktopReleaseListFilters) ([]DesktopRelease, *pagination.PaginationResult, error)
}

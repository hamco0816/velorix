package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// DesktopReleaseService 桌面客户端版本发布：元数据入库 + 安装包落盘 +
// 维护更新目录根的 latest.yml / release.json（供 electron-updater 拉取）。
type DesktopReleaseService struct {
	repo       DesktopReleaseRepository
	updatesDir string
}

func NewDesktopReleaseService(repo DesktopReleaseRepository, cfg *config.Config) *DesktopReleaseService {
	return &DesktopReleaseService{
		repo:       repo,
		updatesDir: cfg.Desktop.UpdatesDir,
	}
}

// CreateDesktopReleaseInput 上传新版本入参（文件内容已读入内存）。
type CreateDesktopReleaseInput struct {
	Version          string
	Channel          string
	Mandatory        bool
	Notes            string
	SetupFilename    string
	SetupData        []byte
	BlockmapFilename string
	BlockmapData     []byte
	LatestYml        string
	ActorID          *int64
}

// releaseManifest = 更新目录根的 release.json 结构，客户端据此判定强制更新与展示更新说明。
type releaseManifest struct {
	Version   string `json:"version"`
	Mandatory bool   `json:"mandatory"`
	Notes     string `json:"notes"`
}

func (s *DesktopReleaseService) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters DesktopReleaseListFilters,
) ([]DesktopRelease, *pagination.PaginationResult, error) {
	return s.repo.List(ctx, params, filters)
}

// GetLatest 返回指定通道当前对外的版本（供公开下载页用）。
func (s *DesktopReleaseService) GetLatest(ctx context.Context, channel string) (*DesktopRelease, error) {
	if strings.TrimSpace(channel) == "" {
		channel = DesktopChannelStable
	}
	return s.repo.GetActiveByChannel(ctx, channel)
}

// Create 上传并发布一个新版本：落盘 → 入库(active) → 旧 active 归档 → 同步 latest.yml/release.json。
func (s *DesktopReleaseService) Create(ctx context.Context, input *CreateDesktopReleaseInput) (*DesktopRelease, error) {
	if input == nil {
		return nil, ErrDesktopReleaseNilInput
	}
	version := strings.TrimSpace(input.Version)
	if version == "" || len(version) > 50 || strings.ContainsAny(version, "/\\") {
		return nil, ErrDesktopReleaseInvalidVersion
	}
	channel := strings.TrimSpace(input.Channel)
	if channel == "" {
		channel = DesktopChannelStable
	}
	if len(input.SetupData) == 0 {
		return nil, ErrDesktopReleaseSetupRequired
	}
	if strings.TrimSpace(input.LatestYml) == "" {
		return nil, ErrDesktopReleaseLatestRequired
	}

	// 同版本号 + 通道唯一
	if existing, err := s.repo.GetByVersionChannel(ctx, version, channel); err == nil && existing != nil {
		return nil, ErrDesktopReleaseExists
	}

	setupName := sanitizeReleaseFilename(input.SetupFilename)
	if setupName == "" {
		setupName = fmt.Sprintf("Velorix-%s-setup.exe", version)
	}
	blockmapName := ""
	if len(input.BlockmapData) > 0 {
		blockmapName = sanitizeReleaseFilename(input.BlockmapFilename)
		if blockmapName == "" {
			blockmapName = setupName + ".blockmap"
		}
	}

	// 落盘安装包 / blockmap 到更新目录根（文件名含版本号，永久保留供回滚）
	if err := os.MkdirAll(s.updatesDir, 0o755); err != nil {
		return nil, fmt.Errorf("create updates dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.updatesDir, setupName), input.SetupData, 0o644); err != nil {
		return nil, fmt.Errorf("write setup file: %w", err)
	}
	if blockmapName != "" {
		if err := os.WriteFile(filepath.Join(s.updatesDir, blockmapName), input.BlockmapData, 0o644); err != nil {
			return nil, fmt.Errorf("write blockmap file: %w", err)
		}
	}

	rel := &DesktopRelease{
		Version:      version,
		Channel:      channel,
		Mandatory:    input.Mandatory,
		Notes:        strings.TrimSpace(input.Notes),
		SetupFile:    setupName,
		BlockmapFile: blockmapName,
		LatestYml:    input.LatestYml,
		FileSize:     int64(len(input.SetupData)),
		Status:       DesktopReleaseStatusActive,
		CreatedBy:    input.ActorID,
	}
	if err := s.repo.Create(ctx, rel); err != nil {
		return nil, err
	}

	// 旧 active 归档，再把根 latest.yml/release.json 指向新版本
	if err := s.repo.ArchiveActiveByChannel(ctx, channel, rel.ID); err != nil {
		return nil, fmt.Errorf("archive previous active: %w", err)
	}
	if err := s.syncActiveFiles(rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// Rollback 把某个历史版本重新设为当前对外版本（其安装包仍在磁盘）。
func (s *DesktopReleaseService) Rollback(ctx context.Context, id int64) (*DesktopRelease, error) {
	rel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 安装包文件必须还在，否则回滚后客户端会下载 404
	if _, statErr := os.Stat(filepath.Join(s.updatesDir, rel.SetupFile)); statErr != nil {
		return nil, ErrDesktopReleaseSetupRequired
	}
	if err := s.repo.UpdateStatus(ctx, rel.ID, DesktopReleaseStatusActive); err != nil {
		return nil, err
	}
	rel.Status = DesktopReleaseStatusActive
	if err := s.repo.ArchiveActiveByChannel(ctx, rel.Channel, rel.ID); err != nil {
		return nil, fmt.Errorf("archive previous active: %w", err)
	}
	if err := s.syncActiveFiles(rel); err != nil {
		return nil, err
	}
	return rel, nil
}

// Delete 删除一个非当前版本（连同磁盘安装包）。当前 active 不允许删除。
func (s *DesktopReleaseService) Delete(ctx context.Context, id int64) error {
	rel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rel.Status == DesktopReleaseStatusActive {
		return ErrDesktopReleaseActiveDeletion
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete desktop release: %w", err)
	}
	// 删磁盘文件（失败不阻断，记录的元数据已删）
	if rel.SetupFile != "" {
		_ = os.Remove(filepath.Join(s.updatesDir, rel.SetupFile))
	}
	if rel.BlockmapFile != "" {
		_ = os.Remove(filepath.Join(s.updatesDir, rel.BlockmapFile))
	}
	return nil
}

// syncActiveFiles 把更新目录根的 latest.yml / release.json 指向给定版本。
func (s *DesktopReleaseService) syncActiveFiles(rel *DesktopRelease) error {
	if err := os.MkdirAll(s.updatesDir, 0o755); err != nil {
		return fmt.Errorf("create updates dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.updatesDir, "latest.yml"), []byte(rel.LatestYml), 0o644); err != nil {
		return fmt.Errorf("write latest.yml: %w", err)
	}
	manifest := releaseManifest{Version: rel.Version, Mandatory: rel.Mandatory, Notes: rel.Notes}
	data, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal release.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.updatesDir, "release.json"), data, 0o644); err != nil {
		return fmt.Errorf("write release.json: %w", err)
	}
	return nil
}

// sanitizeReleaseFilename 只取文件名 base，剔除路径分隔，防目录穿越。
func sanitizeReleaseFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	if name == "." || name == ".." {
		return ""
	}
	return name
}

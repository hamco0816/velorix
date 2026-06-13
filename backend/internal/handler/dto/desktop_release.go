package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// DesktopRelease 桌面客户端版本（管理端展示用，不含 latest.yml 全文以减小体积）。
type DesktopRelease struct {
	ID           int64     `json:"id"`
	Version      string    `json:"version"`
	Channel      string    `json:"channel"`
	Mandatory    bool      `json:"mandatory"`
	Notes        string    `json:"notes"`
	SetupFile    string    `json:"setup_file"`
	BlockmapFile string    `json:"blockmap_file"`
	FileSize     int64     `json:"file_size"`
	Status       string    `json:"status"`
	CreatedBy    *int64    `json:"created_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func DesktopReleaseFromService(r *service.DesktopRelease) *DesktopRelease {
	if r == nil {
		return nil
	}
	return &DesktopRelease{
		ID:           r.ID,
		Version:      r.Version,
		Channel:      r.Channel,
		Mandatory:    r.Mandatory,
		Notes:        r.Notes,
		SetupFile:    r.SetupFile,
		BlockmapFile: r.BlockmapFile,
		FileSize:     r.FileSize,
		Status:       r.Status,
		CreatedBy:    r.CreatedBy,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

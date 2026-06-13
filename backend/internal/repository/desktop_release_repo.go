package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/desktoprelease"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type desktopReleaseRepository struct {
	client *dbent.Client
}

func NewDesktopReleaseRepository(client *dbent.Client) service.DesktopReleaseRepository {
	return &desktopReleaseRepository{client: client}
}

func (r *desktopReleaseRepository) Create(ctx context.Context, rel *service.DesktopRelease) error {
	client := clientFromContext(ctx, r.client)
	builder := client.DesktopRelease.Create().
		SetVersion(rel.Version).
		SetChannel(rel.Channel).
		SetMandatory(rel.Mandatory).
		SetNotes(rel.Notes).
		SetSetupFile(rel.SetupFile).
		SetBlockmapFile(rel.BlockmapFile).
		SetLatestYml(rel.LatestYml).
		SetFileSize(rel.FileSize).
		SetStatus(rel.Status)
	if rel.CreatedBy != nil {
		builder.SetCreatedBy(*rel.CreatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrDesktopReleaseExists)
	}
	rel.ID = created.ID
	rel.CreatedAt = created.CreatedAt
	rel.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *desktopReleaseRepository) GetByID(ctx context.Context, id int64) (*service.DesktopRelease, error) {
	m, err := r.client.DesktopRelease.Query().
		Where(desktoprelease.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrDesktopReleaseNotFound, nil)
	}
	return desktopReleaseEntityToService(m), nil
}

func (r *desktopReleaseRepository) GetByVersionChannel(ctx context.Context, version, channel string) (*service.DesktopRelease, error) {
	m, err := r.client.DesktopRelease.Query().
		Where(desktoprelease.VersionEQ(version), desktoprelease.ChannelEQ(channel)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrDesktopReleaseNotFound, nil)
	}
	return desktopReleaseEntityToService(m), nil
}

func (r *desktopReleaseRepository) GetActiveByChannel(ctx context.Context, channel string) (*service.DesktopRelease, error) {
	m, err := r.client.DesktopRelease.Query().
		Where(desktoprelease.ChannelEQ(channel), desktoprelease.StatusEQ(service.DesktopReleaseStatusActive)).
		Order(dbent.Desc(desktoprelease.FieldID)).
		First(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrDesktopReleaseNotFound, nil)
	}
	return desktopReleaseEntityToService(m), nil
}

func (r *desktopReleaseRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.DesktopRelease.UpdateOneID(id).SetStatus(status).Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrDesktopReleaseNotFound, nil)
	}
	return nil
}

func (r *desktopReleaseRepository) ArchiveActiveByChannel(ctx context.Context, channel string, exceptID int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.DesktopRelease.Update().
		Where(
			desktoprelease.ChannelEQ(channel),
			desktoprelease.StatusEQ(service.DesktopReleaseStatusActive),
			desktoprelease.IDNEQ(exceptID),
		).
		SetStatus(service.DesktopReleaseStatusRolledback).
		Save(ctx)
	return err
}

func (r *desktopReleaseRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.DesktopRelease.Delete().Where(desktoprelease.IDEQ(id)).Exec(ctx)
	return err
}

func (r *desktopReleaseRepository) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.DesktopReleaseListFilters,
) ([]service.DesktopRelease, *pagination.PaginationResult, error) {
	q := r.client.DesktopRelease.Query()
	if filters.Channel != "" {
		q = q.Where(desktoprelease.ChannelEQ(filters.Channel))
	}
	if filters.Status != "" {
		q = q.Where(desktoprelease.StatusEQ(filters.Status))
	}
	if filters.Search != "" {
		q = q.Where(
			desktoprelease.Or(
				desktoprelease.VersionContainsFold(filters.Search),
				desktoprelease.NotesContainsFold(filters.Search),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	items, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(desktoprelease.FieldID)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := make([]service.DesktopRelease, 0, len(items))
	for i := range items {
		if s := desktopReleaseEntityToService(items[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func desktopReleaseEntityToService(m *dbent.DesktopRelease) *service.DesktopRelease {
	if m == nil {
		return nil
	}
	out := &service.DesktopRelease{
		ID:           m.ID,
		Version:      m.Version,
		Channel:      m.Channel,
		Mandatory:    m.Mandatory,
		Notes:        m.Notes,
		SetupFile:    m.SetupFile,
		BlockmapFile: m.BlockmapFile,
		LatestYml:    m.LatestYml,
		FileSize:     m.FileSize,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.CreatedBy != nil {
		out.CreatedBy = m.CreatedBy
	}
	return out
}

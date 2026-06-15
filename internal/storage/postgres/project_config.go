package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProjectNotFound = errors.New("project not found")
var ErrProjectConfigNotFound = errors.New("project config not found")

type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	if dsn == "" {
		return nil, errors.New("postgres dsn is empty")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Store{pool: pool, q: db.New(pool)}, nil
}

func NewProjectConfigStore(ctx context.Context, dsn string) (*Store, error) {
	return NewStore(ctx, dsn)
}

func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

func (s *Store) ListProjectConfig(ctx context.Context, projectID string) ([]domain.ProjectConfig, error) {
	rows, err := s.q.ListProjectConfig(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("query project config: %w", err)
	}

	items := make([]domain.ProjectConfig, 0)
	for _, row := range rows {
		items = append(items, domain.ProjectConfig{Key: row.Key, Value: row.Value})
	}
	return items, nil
}

func (s *Store) GetProjectConfig(ctx context.Context, projectID string, key string) (domain.ProjectConfig, error) {
	row, err := s.q.GetProjectConfig(ctx, db.GetProjectConfigParams{ProjectID: projectID, Key: key})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectConfigNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("get project config: %w", err)
	}
	return domain.ProjectConfig{Key: row.Key, Value: row.Value}, nil
}

func (s *Store) SetProjectConfig(ctx context.Context, projectID string, key string, value string) (domain.ProjectConfig, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("begin project config set tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readProjectConfigTx(ctx, tx, projectID, key)
	if err != nil && !errors.Is(err, ErrProjectConfigNotFound) {
		return domain.ProjectConfig{}, err
	}

	row, err := q.SetProjectConfig(ctx, db.SetProjectConfigParams{ProjectID: projectID, Key: key, Value: value})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("set project config: %w", err)
	}
	item := domain.ProjectConfig{Key: row.Key, Value: row.Value}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "set",
		EntityType:  "config",
		EntityID:    key,
		BeforeState: before,
		AfterState:  item,
	}); err != nil {
		return domain.ProjectConfig{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("commit project config set: %w", err)
	}

	return item, nil
}

func (s *Store) DeleteProjectConfig(ctx context.Context, projectID string, key string) (domain.ProjectConfig, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("begin project config delete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readProjectConfigTx(ctx, tx, projectID, key)
	if err != nil {
		if errors.Is(err, ErrProjectConfigNotFound) {
			return domain.ProjectConfig{}, ErrProjectConfigNotFound
		}
		return domain.ProjectConfig{}, err
	}

	row, err := q.DeleteProjectConfig(ctx, db.DeleteProjectConfigParams{ProjectID: projectID, Key: key})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectConfigNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("delete project config: %w", err)
	}
	item := domain.ProjectConfig{Key: row.Key, Value: row.Value}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "unset",
		EntityType:  "config",
		EntityID:    key,
		BeforeState: before,
		AfterState:  nil,
	}); err != nil {
		return domain.ProjectConfig{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.ProjectConfig{}, fmt.Errorf("commit project config delete: %w", err)
	}

	return item, nil
}

func (s *Store) readProjectConfigTx(ctx context.Context, tx pgx.Tx, projectID string, key string) (domain.ProjectConfig, error) {
	row, err := db.New(tx).GetProjectConfig(ctx, db.GetProjectConfigParams{ProjectID: projectID, Key: key})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectConfigNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("get project config: %w", err)
	}
	return domain.ProjectConfig{Key: row.Key, Value: row.Value}, nil
}

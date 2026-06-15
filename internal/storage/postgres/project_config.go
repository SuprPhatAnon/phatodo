package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProjectNotFound = errors.New("project not found")
var ErrProjectConfigNotFound = errors.New("project config not found")

type Store struct {
	pool *pgxpool.Pool
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

	return &Store{pool: pool}, nil
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
	rows, err := s.pool.Query(ctx, `
		SELECT key, value
		FROM project_config
		WHERE project_id = $1
		ORDER BY key
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query project config: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ProjectConfig, 0)
	for rows.Next() {
		var item domain.ProjectConfig
		if err := rows.Scan(&item.Key, &item.Value); err != nil {
			return nil, fmt.Errorf("scan project config: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project config: %w", err)
	}

	return items, nil
}

func (s *Store) GetProjectConfig(ctx context.Context, projectID string, key string) (domain.ProjectConfig, error) {
	var item domain.ProjectConfig
	err := s.pool.QueryRow(ctx, `
		SELECT key, value
		FROM project_config
		WHERE project_id = $1 AND key = $2
	`, projectID, key).Scan(&item.Key, &item.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectConfigNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("get project config: %w", err)
	}
	return item, nil
}

func (s *Store) SetProjectConfig(ctx context.Context, projectID string, key string, value string) (domain.ProjectConfig, error) {
	var item domain.ProjectConfig
	err := s.pool.QueryRow(ctx, `
		WITH project_row AS (
			SELECT workspace_id
			FROM projects
			WHERE id = $1
		), upserted AS (
			INSERT INTO project_config (
				workspace_id, project_id, key, value
			)
			SELECT workspace_id, $1, $2, $3
			FROM project_row
			ON CONFLICT (project_id, key) DO UPDATE
			SET value = EXCLUDED.value,
				updated_at = now()
			RETURNING key, value
		)
		SELECT key, value FROM upserted
	`, projectID, key, value).Scan(&item.Key, &item.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("set project config: %w", err)
	}

	return item, nil
}

func (s *Store) DeleteProjectConfig(ctx context.Context, projectID string, key string) (domain.ProjectConfig, error) {
	var item domain.ProjectConfig
	err := s.pool.QueryRow(ctx, `
		DELETE FROM project_config
		WHERE project_id = $1 AND key = $2
		RETURNING key, value
	`, projectID, key).Scan(&item.Key, &item.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProjectConfig{}, ErrProjectConfigNotFound
		}
		return domain.ProjectConfig{}, fmt.Errorf("delete project config: %w", err)
	}
	return item, nil
}

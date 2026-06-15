package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

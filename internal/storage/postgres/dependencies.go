package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5"
)

var ErrDuplicateDependency = errors.New("dependency already exists")
var ErrDependencyCycle = errors.New("dependency cycle detected")
var ErrDependencyNotFound = errors.New("dependency not found")

func (s *Store) ListDependencies(ctx context.Context, projectID string, taskID string) (domain.DependencyListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.DependencyListResponse{}, err
	}
	if _, err := s.readTaskDetail(ctx, projectID, taskID); err != nil {
		return domain.DependencyListResponse{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, workspace_id, project_id, task_id, depends_on_id, created_at
		FROM dependencies
		WHERE project_id = $1 AND task_id = $2
		ORDER BY created_at ASC, id ASC
	`, projectID, taskID)
	if err != nil {
		return domain.DependencyListResponse{}, fmt.Errorf("query dependencies: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Dependency, 0)
	for rows.Next() {
		item, err := scanDependency(rows)
		if err != nil {
			return domain.DependencyListResponse{}, fmt.Errorf("scan dependency: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.DependencyListResponse{}, fmt.Errorf("iterate dependencies: %w", err)
	}

	return domain.DependencyListResponse{
		ProjectID: projectID,
		TaskID:    taskID,
		Items:     items,
	}, nil
}

func (s *Store) AddDependency(ctx context.Context, projectID string, taskID string, dependsOnID string, actorUserID string) (domain.Dependency, error) {
	_ = actorUserID
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("begin dependency add tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var workspaceID string
	if err := tx.QueryRow(ctx, `SELECT workspace_id FROM projects WHERE id = $1`, projectID).Scan(&workspaceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dependency{}, ErrProjectNotFound
		}
		return domain.Dependency{}, fmt.Errorf("lookup project: %w", err)
	}

	if taskID == dependsOnID {
		return domain.Dependency{}, ErrDependencyCycle
	}

	if _, err := s.loadTaskEdge(ctx, tx, projectID, taskID); err != nil {
		return domain.Dependency{}, err
	}
	if _, err := s.loadTaskEdge(ctx, tx, projectID, dependsOnID); err != nil {
		return domain.Dependency{}, err
	}

	var duplicateID string
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM dependencies
		WHERE project_id = $1 AND task_id = $2 AND depends_on_id = $3
	`, projectID, taskID, dependsOnID).Scan(&duplicateID)
	if err == nil {
		return domain.Dependency{}, ErrDuplicateDependency
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Dependency{}, fmt.Errorf("check duplicate dependency: %w", err)
	}

	var cycleID string
	err = tx.QueryRow(ctx, `
		WITH RECURSIVE chain AS (
			SELECT task_id, depends_on_id
			FROM dependencies
			WHERE project_id = $1 AND task_id = $2
			UNION ALL
			SELECT d.task_id, d.depends_on_id
			FROM dependencies d
			JOIN chain c ON d.project_id = $1 AND d.task_id = c.depends_on_id
		)
		SELECT task_id
		FROM chain
		WHERE depends_on_id = $3
		LIMIT 1
	`, projectID, dependsOnID, taskID).Scan(&cycleID)
	if err == nil {
		return domain.Dependency{}, ErrDependencyCycle
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Dependency{}, fmt.Errorf("check dependency cycle: %w", err)
	}

	dependencyID, err := randomID("dep")
	if err != nil {
		return domain.Dependency{}, err
	}

	var dependency domain.Dependency
	err = tx.QueryRow(ctx, `
		INSERT INTO dependencies (
			id, workspace_id, project_id, task_id, depends_on_id
		) VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING id, workspace_id, project_id, task_id, depends_on_id, created_at
	`, dependencyID, workspaceID, projectID, taskID, dependsOnID).Scan(
		&dependency.ID,
		&dependency.WorkspaceID,
		&dependency.ProjectID,
		&dependency.TaskID,
		&dependency.DependsOnID,
		&dependency.CreatedAt,
	)
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("insert dependency: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Dependency{}, fmt.Errorf("commit dependency add: %w", err)
	}

	return dependency, nil
}

func (s *Store) RemoveDependency(ctx context.Context, projectID string, taskID string, dependsOnID string, actorUserID string) (domain.Dependency, error) {
	_ = actorUserID
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Dependency{}, err
	}

	row := s.pool.QueryRow(ctx, `
		DELETE FROM dependencies
		WHERE project_id = $1 AND task_id = $2 AND depends_on_id = $3
		RETURNING id, workspace_id, project_id, task_id, depends_on_id, created_at
	`, projectID, taskID, dependsOnID)
	dependency, err := scanDependency(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dependency{}, ErrDependencyNotFound
		}
		return domain.Dependency{}, fmt.Errorf("delete dependency: %w", err)
	}
	return dependency, nil
}

func (s *Store) loadTaskEdge(ctx context.Context, tx pgx.Tx, projectID string, taskID string) (domain.TaskDetail, error) {
	row := tx.QueryRow(ctx, taskDetailQuery+` WHERE project_id = $1 AND id = $2`, projectID, taskID)
	task, err := scanTaskDetail(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("load dependency task: %w", err)
	}
	return task, nil
}

func scanDependency(row interface {
	Scan(dest ...any) error
}) (domain.Dependency, error) {
	var dep domain.Dependency
	var workspaceID sql.NullString
	if err := row.Scan(&dep.ID, &workspaceID, &dep.ProjectID, &dep.TaskID, &dep.DependsOnID, &dep.CreatedAt); err != nil {
		return domain.Dependency{}, err
	}
	if workspaceID.Valid {
		dep.WorkspaceID = workspaceID.String
	}
	return dep, nil
}

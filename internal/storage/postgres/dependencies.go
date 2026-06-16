package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
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

	rows, err := s.q.ListDependencies(ctx, db.ListDependenciesParams{ProjectID: projectID, TaskID: taskID})
	if err != nil {
		return domain.DependencyListResponse{}, fmt.Errorf("query dependencies: %w", err)
	}

	items := make([]domain.Dependency, 0)
	for _, row := range rows {
		items = append(items, dependencyFromSQLC(row))
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

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
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

	if _, err := q.GetDependency(ctx, db.GetDependencyParams{ProjectID: projectID, TaskID: taskID, DependsOnID: dependsOnID}); err == nil {
		return domain.Dependency{}, ErrDuplicateDependency
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Dependency{}, fmt.Errorf("check duplicate dependency: %w", err)
	}

	hasCycle, err := s.dependencyWouldCycle(ctx, q, projectID, taskID, dependsOnID)
	if err != nil {
		return domain.Dependency{}, err
	}
	if hasCycle {
		return domain.Dependency{}, ErrDependencyCycle
	}

	dependencyID, err := randomID("dep")
	if err != nil {
		return domain.Dependency{}, err
	}

	dependency, err := q.CreateDependency(ctx, db.CreateDependencyParams{
		ID:          dependencyID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		TaskID:      taskID,
		DependsOnID: dependsOnID,
	})
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("insert dependency: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "add",
		EntityType:  "dependency",
		EntityID:    dependency.ID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		AfterState:  dependency,
	}); err != nil {
		return domain.Dependency{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Dependency{}, fmt.Errorf("commit dependency add: %w", err)
	}

	return dependencyFromSQLC(dependency), nil
}

func (s *Store) RemoveDependency(ctx context.Context, projectID string, taskID string, dependsOnID string, actorUserID string) (domain.Dependency, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Dependency{}, fmt.Errorf("begin dependency remove tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Dependency{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dependency{}, ErrProjectNotFound
		}
		return domain.Dependency{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readDependencyTx(ctx, tx, projectID, taskID, dependsOnID)
	if err != nil {
		return domain.Dependency{}, err
	}

	dependency, err := q.DeleteDependency(ctx, db.DeleteDependencyParams{
		ProjectID:   projectID,
		TaskID:      taskID,
		DependsOnID: dependsOnID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dependency{}, ErrDependencyNotFound
		}
		return domain.Dependency{}, fmt.Errorf("delete dependency: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "remove",
		EntityType:  "dependency",
		EntityID:    before.ID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  nil,
	}); err != nil {
		return domain.Dependency{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Dependency{}, fmt.Errorf("commit dependency remove: %w", err)
	}

	return dependencyFromSQLC(dependency), nil
}

func (s *Store) loadTaskEdge(ctx context.Context, tx pgx.Tx, projectID string, taskID string) (domain.TaskDetail, error) {
	task, err := s.q.WithTx(tx).GetTaskDetail(ctx, db.GetTaskDetailParams{ProjectID: projectID, TaskID: taskID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("load dependency task: %w", err)
	}
	return taskDetailFromGetTaskDetailRow(task)
}

func (s *Store) readDependencyTx(ctx context.Context, tx pgx.Tx, projectID string, taskID string, dependsOnID string) (domain.Dependency, error) {
	dep, err := s.q.WithTx(tx).GetDependency(ctx, db.GetDependencyParams{
		ProjectID:   projectID,
		TaskID:      taskID,
		DependsOnID: dependsOnID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dependency{}, ErrDependencyNotFound
		}
		return domain.Dependency{}, fmt.Errorf("load dependency: %w", err)
	}
	return dependencyFromSQLC(dep), nil
}

func (s *Store) dependencyWouldCycle(ctx context.Context, q *db.Queries, projectID string, taskID string, dependsOnID string) (bool, error) {
	seen := map[string]bool{dependsOnID: true}
	queue := []string{dependsOnID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		rows, err := q.ListDependencies(ctx, db.ListDependenciesParams{ProjectID: projectID, TaskID: current})
		if err != nil {
			return false, fmt.Errorf("query dependency chain: %w", err)
		}
		for _, row := range rows {
			if row.DependsOnID == taskID {
				return true, nil
			}
			if !seen[row.DependsOnID] {
				seen[row.DependsOnID] = true
				queue = append(queue, row.DependsOnID)
			}
		}
	}
	return false, nil
}

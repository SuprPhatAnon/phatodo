package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrLockNotFound = errors.New("lock not found")
var ErrLockConflict = errors.New("active lock already exists")
var ErrInvalidLockEntityType = errors.New("invalid lock entity type")

func (s *Store) ListLocks(ctx context.Context, projectID string, entityTypes []string, entityID string, active bool) (domain.LockListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.LockListResponse{}, err
	}

	rows, err := s.q.ListWorkItemLocks(ctx, db.ListWorkItemLocksParams{
		ProjectID:   projectID,
		EntityTypes: entityTypes,
		EntityID:    entityID,
		Active:      active,
	})
	if err != nil {
		return domain.LockListResponse{}, fmt.Errorf("query locks: %w", err)
	}

	items := make([]domain.WorkItemLock, 0, len(rows))
	for _, row := range rows {
		items = append(items, lockFromSQLC(row))
	}

	return domain.LockListResponse{
		ProjectID: projectID,
		Items:     items,
	}, nil
}

func (s *Store) AcquireLock(ctx context.Context, projectID string, req domain.LockAcquireRequest, actorUserID string) (domain.WorkItemLock, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.WorkItemLock{}, fmt.Errorf("begin lock acquire tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.WorkItemLock{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkItemLock{}, ErrProjectNotFound
		}
		return domain.WorkItemLock{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	entityType, err := normalizeLockEntityType(req.EntityType)
	if err != nil {
		return domain.WorkItemLock{}, err
	}
	if strings.TrimSpace(req.EntityID) == "" {
		return domain.WorkItemLock{}, fmt.Errorf("entity id is required")
	}

	if err := s.ensureLockTargetExists(ctx, tx, projectID, entityType, req.EntityID); err != nil {
		return domain.WorkItemLock{}, err
	}

	expiresAt, err := lockExpiresAt(req)
	if err != nil {
		return domain.WorkItemLock{}, err
	}

	_, err = q.GetActiveWorkItemLock(ctx, db.GetActiveWorkItemLockParams{
		ProjectID:  projectID,
		EntityType: entityType,
		EntityID:   req.EntityID,
	})
	if err == nil {
		return domain.WorkItemLock{}, ErrLockConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.WorkItemLock{}, fmt.Errorf("check active lock: %w", err)
	}

	if err := q.ReleaseExpiredWorkItemLock(ctx, db.ReleaseExpiredWorkItemLockParams{
		ProjectID:  projectID,
		EntityType: entityType,
		EntityID:   req.EntityID,
	}); err != nil {
		return domain.WorkItemLock{}, fmt.Errorf("release expired lock: %w", err)
	}

	counter, err := q.NextLockCounter(ctx, projectID)
	if err != nil {
		return domain.WorkItemLock{}, fmt.Errorf("allocate lock counter: %w", err)
	}

	lockID := fmt.Sprintf("LOCK-%d", counter)
	row, err := q.CreateWorkItemLock(ctx, db.CreateWorkItemLockParams{
		ID:          lockID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		EntityType:  entityType,
		EntityID:    req.EntityID,
		LockedBy:    actorUserID,
		Reason:      req.Reason,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.WorkItemLock{}, ErrLockConflict
		}
		return domain.WorkItemLock{}, fmt.Errorf("insert lock: %w", err)
	}

	lock := lockFromSQLC(row)
	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "acquire",
		EntityType:  entityType,
		EntityID:    req.EntityID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		AfterState:  lock,
		Metadata: map[string]any{
			"lock_id": lockID,
		},
	}); err != nil {
		return domain.WorkItemLock{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.WorkItemLock{}, fmt.Errorf("commit lock acquire: %w", err)
	}

	return lock, nil
}

func (s *Store) ReleaseLock(ctx context.Context, projectID string, lockID string, actorUserID string) (domain.WorkItemLock, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.WorkItemLock{}, fmt.Errorf("begin lock release tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.WorkItemLock{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkItemLock{}, ErrProjectNotFound
		}
		return domain.WorkItemLock{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readLockTx(ctx, tx, projectID, lockID)
	if err != nil {
		return domain.WorkItemLock{}, err
	}
	if !before.ReleasedAt.IsZero() {
		return before, nil
	}

	row, err := q.ReleaseWorkItemLock(ctx, db.ReleaseWorkItemLockParams{
		ProjectID: projectID,
		LockID:    lockID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkItemLock{}, ErrLockNotFound
		}
		return domain.WorkItemLock{}, fmt.Errorf("release lock: %w", err)
	}

	after := lockFromSQLC(row)
	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "release",
		EntityType:  before.EntityType,
		EntityID:    before.EntityID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  after,
		Metadata: map[string]any{
			"lock_id": lockID,
		},
	}); err != nil {
		return domain.WorkItemLock{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.WorkItemLock{}, fmt.Errorf("commit lock release: %w", err)
	}

	return after, nil
}

func (s *Store) releaseEntityLocksTx(ctx context.Context, tx pgx.Tx, projectID string, entityType string, entityID string, actorUserID string) error {
	q := s.q.WithTx(tx)
	rows, err := q.ListWorkItemLocks(ctx, db.ListWorkItemLocksParams{
		ProjectID:   projectID,
		EntityTypes: []string{entityType},
		EntityID:    entityID,
		Active:      true,
	})
	if err != nil {
		return fmt.Errorf("query entity locks: %w", err)
	}
	for _, row := range rows {
		before := lockFromSQLC(row)
		afterRow, err := q.ReleaseWorkItemLock(ctx, db.ReleaseWorkItemLockParams{
			ProjectID: projectID,
			LockID:    row.ID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return fmt.Errorf("release entity lock: %w", err)
		}
		after := lockFromSQLC(afterRow)
		if err := s.recordEventTx(ctx, tx, auditEvent{
			WorkspaceID: before.WorkspaceID,
			ProjectID:   projectID,
			Action:      "release",
			EntityType:  entityType,
			EntityID:    entityID,
			ActorUserID: actorUserID,
			ActorLabel:  actorUserID,
			BeforeState: before,
			AfterState:  after,
			Metadata: map[string]any{
				"lock_id": row.ID,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) readLockTx(ctx context.Context, tx pgx.Tx, projectID string, lockID string) (domain.WorkItemLock, error) {
	row, err := s.q.WithTx(tx).GetWorkItemLock(ctx, db.GetWorkItemLockParams{
		ProjectID: projectID,
		LockID:    lockID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkItemLock{}, ErrLockNotFound
		}
		return domain.WorkItemLock{}, fmt.Errorf("load lock: %w", err)
	}
	return lockFromSQLC(row), nil
}

func (s *Store) ensureLockTargetExists(ctx context.Context, tx pgx.Tx, projectID string, entityType string, entityID string) error {
	q := s.q.WithTx(tx)
	switch entityType {
	case "epic":
		if _, err := q.GetEpicID(ctx, db.GetEpicIDParams{ProjectID: projectID, EpicID: entityID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrEpicNotFound
			}
			return fmt.Errorf("lookup epic for lock: %w", err)
		}
	case "task", "subtask":
		row, err := q.GetTaskCreateParentInfo(ctx, db.GetTaskCreateParentInfoParams{ProjectID: projectID, TaskID: entityID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTaskNotFound
			}
			return fmt.Errorf("lookup task for lock: %w", err)
		}
		if entityType == "task" && row.ParentTaskID != nil {
			return ErrInvalidLockEntityType
		}
		if entityType == "subtask" && row.ParentTaskID == nil {
			return ErrInvalidLockEntityType
		}
	default:
		return ErrInvalidLockEntityType
	}
	return nil
}

func lockExpiresAt(req domain.LockAcquireRequest) (time.Time, error) {
	if req.ExpiresAt != nil {
		if req.ExpiresAt.IsZero() {
			return time.Time{}, fmt.Errorf("expires_at is invalid")
		}
		return req.ExpiresAt.UTC(), nil
	}
	ttl := strings.TrimSpace(req.TTL)
	if ttl == "" {
		ttl = "1h"
	}
	d, err := time.ParseDuration(ttl)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid ttl: %w", err)
	}
	if d <= 0 {
		return time.Time{}, fmt.Errorf("ttl must be greater than zero")
	}
	return time.Now().UTC().Add(d), nil
}

func normalizeLockEntityType(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "epic", "task", "subtask":
		return strings.TrimSpace(value), nil
	default:
		return "", ErrInvalidLockEntityType
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}

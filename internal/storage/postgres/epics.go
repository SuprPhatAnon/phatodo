package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateEpic(ctx context.Context, projectID string, req domain.EpicCreateRequest, actorUserID string) (domain.Epic, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Epic{}, fmt.Errorf("begin epic create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrProjectNotFound
		}
		return domain.Epic{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	counter, err := q.NextEpicCounter(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrProjectNotFound
		}
		return domain.Epic{}, fmt.Errorf("allocate epic counter: %w", err)
	}

	epicID := fmt.Sprintf("EPIC-%d", counter)
	priority := domain.PriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}
	criteria := req.AcceptanceCriteria
	if criteria == nil {
		criteria = []string{}
	}
	criteriaJSON, err := json.Marshal(criteria)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("marshal acceptance criteria: %w", err)
	}

	row, err := q.CreateEpic(ctx, db.CreateEpicParams{
		ID:                 epicID,
		WorkspaceID:        workspaceID,
		ProjectID:          projectID,
		AssignedTo:         req.AssignedTo,
		CreatedBy:          actorUserID,
		Title:              req.Title,
		Description:        req.Description,
		Priority:           int32(priority),
		AcceptanceCriteria: criteriaJSON,
	})
	if err != nil {
		return domain.Epic{}, fmt.Errorf("insert epic: %w", err)
	}

	epic, err := epicFromSQLC(row)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("decode epic: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "create",
		EntityType:  "epic",
		EntityID:    epicID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		AfterState:  epic,
	}); err != nil {
		return domain.Epic{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Epic{}, fmt.Errorf("commit epic create: %w", err)
	}

	return epic, nil
}

func (s *Store) ListEpics(ctx context.Context, projectID string, status string, limit int) (domain.EpicListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.EpicListResponse{}, err
	}

	rows, err := s.q.ListEpics(ctx, db.ListEpicsParams{ProjectID: projectID, Status: status})
	if err != nil {
		return domain.EpicListResponse{}, fmt.Errorf("query epics: %w", err)
	}

	items := make([]domain.Epic, 0, len(rows))
	for _, row := range rows {
		item, err := epicFromSQLC(row)
		if err != nil {
			return domain.EpicListResponse{}, fmt.Errorf("decode epic: %w", err)
		}
		items = append(items, item)
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return domain.EpicListResponse{ProjectID: projectID, Items: items}, nil
}

func (s *Store) GetEpic(ctx context.Context, projectID string, epicID string) (domain.Epic, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Epic{}, err
	}
	return s.readEpic(ctx, projectID, epicID)
}

func (s *Store) UpdateEpic(ctx context.Context, projectID string, epicID string, req domain.EpicUpdateRequest, actorUserID string) (domain.Epic, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Epic{}, fmt.Errorf("begin epic update tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Epic{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrProjectNotFound
		}
		return domain.Epic{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readEpicTx(ctx, tx, projectID, epicID)
	if err != nil {
		return domain.Epic{}, err
	}

	var title *string
	if req.Title != nil {
		title = req.Title
	}
	var description *string
	if req.Description != nil {
		description = req.Description
	}
	var priority *int32
	if req.Priority != nil {
		p := int32(*req.Priority)
		priority = &p
	}
	var status *string
	if req.Status != nil {
		value := string(*req.Status)
		status = &value
	}
	var assignedTo *string
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}
	var criteria []byte
	if req.AcceptanceCriteria != nil {
		criteriaJSON, err := json.Marshal(*req.AcceptanceCriteria)
		if err != nil {
			return domain.Epic{}, fmt.Errorf("marshal acceptance criteria: %w", err)
		}
		criteria = criteriaJSON
	}
	var summary *string
	if req.CompletionSummary != nil {
		summary = req.CompletionSummary
	}
	var evidence []byte
	if req.CompletionEvidence != nil {
		evidenceJSON, err := json.Marshal(*req.CompletionEvidence)
		if err != nil {
			return domain.Epic{}, fmt.Errorf("marshal completion evidence: %w", err)
		}
		evidence = evidenceJSON
	}

	updated, err := q.UpdateEpic(ctx, db.UpdateEpicParams{
		UpdatedBy:          &actorUserID,
		Title:              title,
		Description:        description,
		Priority:           priority,
		Status:             status,
		AssignedTo:         assignedTo,
		AcceptanceCriteria: criteria,
		CompletionSummary:  summary,
		CompletionEvidence: evidence,
		ProjectID:          projectID,
		EpicID:             epicID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrEpicNotFound
		}
		return domain.Epic{}, fmt.Errorf("update epic: %w", err)
	}

	after, err := epicFromSQLC(updated)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("decode updated epic: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "update",
		EntityType:  "epic",
		EntityID:    epicID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  after,
	}); err != nil {
		return domain.Epic{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Epic{}, fmt.Errorf("commit epic update: %w", err)
	}

	return after, nil
}

func (s *Store) CompleteEpic(ctx context.Context, projectID string, epicID string, actorUserID string) (domain.Epic, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Epic{}, fmt.Errorf("begin epic complete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Epic{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrProjectNotFound
		}
		return domain.Epic{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readEpicTx(ctx, tx, projectID, epicID)
	if err != nil {
		return domain.Epic{}, err
	}

	rows, err := q.ListTasksUnified(ctx, db.ListTasksUnifiedParams{ProjectID: projectID, Status: ""})
	if err != nil {
		return domain.Epic{}, fmt.Errorf("list epic tasks: %w", err)
	}
	for _, row := range rows {
		if row.EpicID == nil || *row.EpicID != epicID {
			continue
		}
		if _, err := s.archiveTaskTx(ctx, tx, projectID, row.ID, actorUserID); err != nil {
			return domain.Epic{}, err
		}
	}

	updated, err := q.UpdateEpic(ctx, db.UpdateEpicParams{
		UpdatedBy: &actorUserID,
		Status:    strPtr(string(domain.StatusCompleted)),
		ProjectID: projectID,
		EpicID:    epicID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrEpicNotFound
		}
		return domain.Epic{}, fmt.Errorf("complete epic: %w", err)
	}

	after, err := epicFromSQLC(updated)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("decode completed epic: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "complete",
		EntityType:  "epic",
		EntityID:    epicID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  after,
	}); err != nil {
		return domain.Epic{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Epic{}, fmt.Errorf("commit epic complete: %w", err)
	}

	return after, nil
}

func (s *Store) DeleteEpic(ctx context.Context, projectID string, epicID string, actorUserID string) (domain.Epic, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Epic{}, fmt.Errorf("begin epic delete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Epic{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrProjectNotFound
		}
		return domain.Epic{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readEpicTx(ctx, tx, projectID, epicID)
	if err != nil {
		return domain.Epic{}, err
	}

	if err := s.releaseEntityLocksTx(ctx, tx, projectID, "epic", epicID, actorUserID); err != nil {
		return domain.Epic{}, err
	}

	deleted, err := q.DeleteEpic(ctx, db.DeleteEpicParams{ProjectID: projectID, EpicID: epicID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrEpicNotFound
		}
		return domain.Epic{}, fmt.Errorf("delete epic: %w", err)
	}

	after, err := epicFromSQLC(deleted)
	if err != nil {
		return domain.Epic{}, fmt.Errorf("decode deleted epic: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "delete",
		EntityType:  "epic",
		EntityID:    epicID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  nil,
	}); err != nil {
		return domain.Epic{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Epic{}, fmt.Errorf("commit epic delete: %w", err)
	}

	return after, nil
}

func (s *Store) readEpic(ctx context.Context, projectID string, epicID string) (domain.Epic, error) {
	epic, err := s.q.GetEpic(ctx, db.GetEpicParams{ProjectID: projectID, EpicID: epicID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrEpicNotFound
		}
		return domain.Epic{}, fmt.Errorf("load epic: %w", err)
	}
	return epicFromSQLC(epic)
}

func (s *Store) readEpicTx(ctx context.Context, tx pgx.Tx, projectID string, epicID string) (domain.Epic, error) {
	epic, err := s.q.WithTx(tx).GetEpic(ctx, db.GetEpicParams{ProjectID: projectID, EpicID: epicID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, ErrEpicNotFound
		}
		return domain.Epic{}, fmt.Errorf("load epic: %w", err)
	}
	return epicFromSQLC(epic)
}

func (s *Store) archiveTaskTx(ctx context.Context, tx pgx.Tx, projectID string, taskID string, actorUserID string) (domain.TaskDetail, error) {
	before, err := s.readTaskDetailTx(ctx, tx, projectID, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	q := s.q.WithTx(tx)
	after, err := q.UpdateTask(ctx, db.UpdateTaskParams{
		UpdatedBy: &actorUserID,
		Status:    strPtr(string(domain.StatusArchived)),
		ProjectID: projectID,
		TaskID:    taskID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("archive task: %w", err)
	}

	afterTask, err := taskDetailFromUpdateTaskRow(after)
	if err != nil {
		return domain.TaskDetail{}, fmt.Errorf("decode archived task: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: before.WorkspaceID,
		ProjectID:   projectID,
		Action:      "archive",
		EntityType:  taskEntityType(before.ParentTaskID),
		EntityID:    taskID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  afterTask,
	}); err != nil {
		return domain.TaskDetail{}, err
	}

	return afterTask, nil
}

func strPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

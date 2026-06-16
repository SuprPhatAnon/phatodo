package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

var ErrEpicNotFound = errors.New("epic not found")
var ErrAssignedUserNotFound = errors.New("assigned user not found")
var ErrInvalidIssuePrefix = errors.New("issue prefix is invalid")
var ErrInvalidTaskKind = errors.New("invalid task kind")
var ErrTaskKindRequiresRootCause = errors.New("root cause analysis is required for bug kind")
var ErrTaskNotFound = errors.New("task not found")

func (s *Store) CreateTask(ctx context.Context, projectID string, req domain.TaskCreateRequest, actorUserID string) (domain.TaskCreateResponse, error) {
	kind := req.Kind
	if kind == "" {
		kind = domain.TaskKindTask
	}
	if !isAllowedTaskKind(kind) {
		return domain.TaskCreateResponse{}, ErrInvalidTaskKind
	}

	rootCause := strings.TrimSpace(req.RootCauseAnalysis)
	if kind == domain.TaskKindBug && rootCause == "" {
		return domain.TaskCreateResponse{}, ErrTaskKindRequiresRootCause
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("begin task create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskCreateResponse{}, ErrProjectNotFound
		}
		return domain.TaskCreateResponse{}, fmt.Errorf("lookup project: %w", err)
	}

	if req.ParentTaskID != "" {
		parent, err := q.GetTaskCreateParentInfo(ctx, db.GetTaskCreateParentInfoParams{ProjectID: projectID, TaskID: req.ParentTaskID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.TaskCreateResponse{}, ErrTaskNotFound
			}
			return domain.TaskCreateResponse{}, fmt.Errorf("lookup parent task: %w", err)
		}
		if parent.ParentTaskID != nil {
			return domain.TaskCreateResponse{}, ErrTaskNotFound
		}
		if req.IssuePrefix == "" {
			prefixParts := strings.SplitN(parent.ID, "-", 2)
			req.IssuePrefix = prefixParts[0]
		}
		if req.EpicID == "" && parent.EpicID != nil {
			req.EpicID = *parent.EpicID
		}
	}

	if req.EpicID != "" {
		if _, err := q.GetEpicID(ctx, db.GetEpicIDParams{ProjectID: projectID, EpicID: req.EpicID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.TaskCreateResponse{}, ErrEpicNotFound
			}
			return domain.TaskCreateResponse{}, fmt.Errorf("lookup epic: %w", err)
		}
	}

	if req.AssignedTo != "" {
		if _, err := q.GetUserID(ctx, req.AssignedTo); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.TaskCreateResponse{}, ErrAssignedUserNotFound
			}
			return domain.TaskCreateResponse{}, fmt.Errorf("lookup assigned user: %w", err)
		}
	}

	prefix, err := normalizeIssuePrefix(req.IssuePrefix)
	if err != nil {
		return domain.TaskCreateResponse{}, err
	}

	counter, err := q.NextTaskCounter(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskCreateResponse{}, ErrProjectNotFound
		}
		return domain.TaskCreateResponse{}, fmt.Errorf("allocate task counter: %w", err)
	}

	taskID := fmt.Sprintf("%s-%d", prefix, counter)
	priority := domain.PriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	acceptanceCriteria := req.AcceptanceCriteria
	if acceptanceCriteria == nil {
		acceptanceCriteria = []string{}
	}
	acceptanceCriteriaJSON, err := json.Marshal(acceptanceCriteria)
	if err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("marshal acceptance criteria: %w", err)
	}

	row, err := q.CreateTask(ctx, db.CreateTaskParams{
		ID:                 taskID,
		WorkspaceID:        workspaceID,
		ProjectID:          projectID,
		EpicID:             req.EpicID,
		ParentTaskID:       req.ParentTaskID,
		AssignedTo:         req.AssignedTo,
		CreatedBy:          actorUserID,
		Kind:               string(kind),
		Title:              req.Title,
		Description:        req.Description,
		RootCauseAnalysis:  rootCause,
		Priority:           int32(priority),
		Tags:               tags,
		AcceptanceCriteria: acceptanceCriteriaJSON,
	})
	if err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("insert task: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "create",
		EntityType:  taskEntityType(req.ParentTaskID),
		EntityID:    taskID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		AfterState: map[string]any{
			"id":                  row.ID,
			"issue_prefix":        prefix,
			"title":               row.Title,
			"kind":                row.Kind,
			"priority":            row.Priority,
			"status":              row.Status,
			"root_cause_analysis": row.RootCauseAnalysis,
			"project_id":          row.ProjectID,
			"workspace_id":        row.WorkspaceID,
			"epic_id":             req.EpicID,
			"parent_task_id":      req.ParentTaskID,
			"tags":                tags,
		},
	}); err != nil {
		return domain.TaskCreateResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("commit task create: %w", err)
	}

	return domain.TaskCreateResponse{
		ID:          row.ID,
		IssuePrefix: prefix,
		Title:       row.Title,
		Kind:        domain.TaskKind(row.Kind),
		Status:      domain.Status(row.Status),
		Priority:    domain.Priority(row.Priority),
		RootCause:   row.RootCauseAnalysis,
		ProjectID:   row.ProjectID,
		WorkspaceID: row.WorkspaceID,
	}, nil
}

func (s *Store) ListTasks(ctx context.Context, projectID string, status string, epicID string, limit int) (domain.TaskListResponse, error) {
	return s.listTasks(ctx, projectID, status, epicID, "", limit)
}

func (s *Store) ListSubtasks(ctx context.Context, projectID string, parentTaskID string, limit int) (domain.TaskListResponse, error) {
	if _, err := s.readTaskDetail(ctx, projectID, parentTaskID); err != nil {
		return domain.TaskListResponse{}, err
	}
	return s.listTasks(ctx, projectID, "", "", parentTaskID, limit)
}

func (s *Store) listTasks(ctx context.Context, projectID string, status string, epicID string, parentTaskID string, limit int) (domain.TaskListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.TaskListResponse{}, err
	}

	items := make([]domain.TaskListItem, 0)
	if parentTaskID == "" {
		rows, err := s.q.ListTasks(ctx, db.ListTasksParams{ProjectID: projectID, Status: status, EpicID: epicID})
		if err != nil {
			return domain.TaskListResponse{}, fmt.Errorf("query tasks: %w", err)
		}
		for _, row := range rows {
			items = append(items, taskListItemFromSQLC(row))
		}
	} else {
		rows, err := s.q.ListSubtasks(ctx, db.ListSubtasksParams{ProjectID: projectID, ParentTaskID: &parentTaskID, Status: status})
		if err != nil {
			return domain.TaskListResponse{}, fmt.Errorf("query subtasks: %w", err)
		}
		for _, row := range rows {
			item := taskListItemFromSQLC(db.ListTasksRow{
				ID:           row.ID,
				Title:        row.Title,
				Status:       row.Status,
				Priority:     row.Priority,
				Kind:         row.Kind,
				EpicID:       row.EpicID,
				ParentTaskID: row.ParentTaskID,
				Tags:         row.Tags,
				CreatedAt:    row.CreatedAt,
				UpdatedAt:    row.UpdatedAt,
			})
			if epicID != "" && item.EpicID != epicID {
				continue
			}
			items = append(items, item)
		}
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return domain.TaskListResponse{
		ProjectID: projectID,
		Items:     items,
	}, nil
}

func (s *Store) GetTask(ctx context.Context, projectID string, taskID string) (domain.TaskDetail, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.TaskDetail{}, err
	}
	return s.readTaskDetail(ctx, projectID, taskID)
}

func (s *Store) UpdateTask(ctx context.Context, projectID string, taskID string, req domain.TaskUpdateRequest, actorUserID string) (domain.TaskDetail, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.TaskDetail{}, fmt.Errorf("begin task update tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.TaskDetail{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrProjectNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readTaskDetailTx(ctx, tx, projectID, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	resultKind := before.Kind
	if req.Kind != nil {
		if !isAllowedTaskKind(*req.Kind) {
			return domain.TaskDetail{}, ErrInvalidTaskKind
		}
		resultKind = *req.Kind
	}

	resultRootCause := before.RootCauseAnalysis
	if req.RootCauseAnalysis != nil {
		resultRootCause = strings.TrimSpace(*req.RootCauseAnalysis)
	}
	if resultKind == domain.TaskKindBug && resultRootCause == "" {
		return domain.TaskDetail{}, ErrTaskKindRequiresRootCause
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
	var tags []string
	if req.Tags != nil {
		tags = *req.Tags
	}
	var kind *string
	if req.Kind != nil {
		value := string(*req.Kind)
		kind = &value
	}
	var rootCause *string
	if req.RootCauseAnalysis != nil {
		value := strings.TrimSpace(*req.RootCauseAnalysis)
		rootCause = &value
	}
	var epicID *string
	if req.EpicID != nil {
		epicID = req.EpicID
	}
	var assignedTo *string
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}
	var acceptanceCriteria []byte
	if req.AcceptanceCriteria != nil {
		criteriaJSON, err := json.Marshal(*req.AcceptanceCriteria)
		if err != nil {
			return domain.TaskDetail{}, fmt.Errorf("marshal acceptance criteria: %w", err)
		}
		acceptanceCriteria = criteriaJSON
	}
	var completionSummary *string
	if req.CompletionSummary != nil {
		completionSummary = req.CompletionSummary
	}
	var completionEvidence []byte
	if req.CompletionEvidence != nil {
		evidenceJSON, err := json.Marshal(*req.CompletionEvidence)
		if err != nil {
			return domain.TaskDetail{}, fmt.Errorf("marshal completion evidence: %w", err)
		}
		completionEvidence = evidenceJSON
	}

	updated, err := q.UpdateTask(ctx, db.UpdateTaskParams{
		UpdatedBy:          &actorUserID,
		Title:              title,
		Description:        description,
		Priority:           priority,
		Status:             status,
		Tags:               tags,
		Kind:               kind,
		RootCauseAnalysis:  rootCause,
		ClearEpic:          req.NoEpic,
		EpicID:             epicID,
		AssignedTo:         assignedTo,
		AcceptanceCriteria: acceptanceCriteria,
		CompletionSummary:  completionSummary,
		CompletionEvidence: completionEvidence,
		ProjectID:          projectID,
		TaskID:             taskID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("update task: %w", err)
	}

	after, err := taskDetailFromSQLC(updated)
	if err != nil {
		return domain.TaskDetail{}, fmt.Errorf("decode updated task: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "update",
		EntityType:  taskEntityType(before.ParentTaskID),
		EntityID:    taskID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  after,
	}); err != nil {
		return domain.TaskDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.TaskDetail{}, fmt.Errorf("commit task update: %w", err)
	}

	return after, nil
}

func (s *Store) DeleteTask(ctx context.Context, projectID string, taskID string, actorUserID string) (domain.TaskDetail, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.TaskDetail{}, fmt.Errorf("begin task delete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.TaskDetail{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrProjectNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readTaskDetailTx(ctx, tx, projectID, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	if err := s.releaseEntityLocksTx(ctx, tx, projectID, taskEntityType(before.ParentTaskID), taskID, actorUserID); err != nil {
		return domain.TaskDetail{}, err
	}

	taskRow, err := q.DeleteTask(ctx, db.DeleteTaskParams{ProjectID: projectID, TaskID: taskID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("delete task: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "delete",
		EntityType:  taskEntityType(before.ParentTaskID),
		EntityID:    taskID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  nil,
	}); err != nil {
		return domain.TaskDetail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.TaskDetail{}, fmt.Errorf("commit task delete: %w", err)
	}

	return taskDetailFromSQLC(taskRow)
}

func (s *Store) ListReadyTasks(ctx context.Context, projectID string, epicID string, limit int) (domain.ReadyListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.ReadyListResponse{}, err
	}

	rows, err := s.q.ListReadyTasks(ctx, db.ListReadyTasksParams{ProjectID: projectID, EpicID: epicID})
	if err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("query ready tasks: %w", err)
	}

	items := make([]domain.ReadyListItem, 0)
	readyIDs := make([]string, 0)
	for _, row := range rows {
		item := readyListItemFromSQLC(row)
		items = append(items, item)
		readyIDs = append(readyIDs, item.ID)
	}

	if limit > 0 && len(items) > limit {
		items = items[:limit]
		readyIDs = readyIDs[:limit]
	}

	if len(readyIDs) == 0 {
		return domain.ReadyListResponse{
			ProjectID: projectID,
			Items:     items,
		}, nil
	}

	blockedRows, err := s.q.ListReadyDependents(ctx, db.ListReadyDependentsParams{ProjectID: projectID, ReadyIds: readyIDs, EpicID: epicID})
	if err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("query ready dependents: %w", err)
	}

	unblocks := make(map[string][]domain.TaskListItem)
	for _, row := range blockedRows {
		unblocks[row.DependsOnID] = append(unblocks[row.DependsOnID], readyDependentFromSQLC(row))
	}

	for i := range items {
		if len(unblocks[items[i].ID]) > 0 {
			items[i].Unblocks = unblocks[items[i].ID]
		}
	}

	return domain.ReadyListResponse{
		ProjectID: projectID,
		Items:     items,
	}, nil
}

func (s *Store) ensureProjectExists(ctx context.Context, projectID string) error {
	if _, err := s.q.GetProjectWorkspaceID(ctx, projectID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("lookup project: %w", err)
	}
	return nil
}

func (s *Store) readTaskDetail(ctx context.Context, projectID string, taskID string) (domain.TaskDetail, error) {
	task, err := s.q.GetTaskDetail(ctx, db.GetTaskDetailParams{ProjectID: projectID, TaskID: taskID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("load task detail: %w", err)
	}
	return taskDetailFromSQLC(task)
}

func (s *Store) readTaskDetailTx(ctx context.Context, tx pgx.Tx, projectID string, taskID string) (domain.TaskDetail, error) {
	task, err := s.q.WithTx(tx).GetTaskDetail(ctx, db.GetTaskDetailParams{ProjectID: projectID, TaskID: taskID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("load task detail: %w", err)
	}
	return taskDetailFromSQLC(task)
}

func isAllowedTaskKind(value domain.TaskKind) bool {
	switch value {
	case domain.TaskKindTask, domain.TaskKindBug, domain.TaskKindFeature, domain.TaskKindChore, domain.TaskKindSpike:
		return true
	default:
		return false
	}
}

func taskEntityType(parentTaskID string) string {
	if parentTaskID != "" {
		return "subtask"
	}
	return "task"
}

func normalizeIssuePrefix(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", ErrInvalidIssuePrefix
	}

	var builder strings.Builder
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	prefix := builder.String()
	if prefix == "" {
		return "", ErrInvalidIssuePrefix
	}
	return prefix, nil
}

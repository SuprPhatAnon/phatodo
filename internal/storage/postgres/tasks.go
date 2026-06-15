package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5"
)

var ErrEpicNotFound = errors.New("epic not found")
var ErrAssignedUserNotFound = errors.New("assigned user not found")
var ErrInvalidIssuePrefix = errors.New("issue prefix is invalid")
var ErrTaskNotFound = errors.New("task not found")

func (s *Store) CreateTask(ctx context.Context, projectID string, req domain.TaskCreateRequest, actorUserID string) (domain.TaskCreateResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("begin task create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var workspaceID string
	if err := tx.QueryRow(ctx, `SELECT workspace_id FROM projects WHERE id = $1`, projectID).Scan(&workspaceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskCreateResponse{}, ErrProjectNotFound
		}
		return domain.TaskCreateResponse{}, fmt.Errorf("lookup project: %w", err)
	}

	if req.ParentTaskID != "" {
		var parentID string
		var parentEpicID sql.NullString
		var parentParentID sql.NullString
		if err := tx.QueryRow(ctx, `
			SELECT id, epic_id, parent_task_id
			FROM tasks
			WHERE project_id = $1 AND id = $2
		`, projectID, req.ParentTaskID).Scan(&parentID, &parentEpicID, &parentParentID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.TaskCreateResponse{}, ErrTaskNotFound
			}
			return domain.TaskCreateResponse{}, fmt.Errorf("lookup parent task: %w", err)
		}
		if parentParentID.Valid {
			return domain.TaskCreateResponse{}, ErrTaskNotFound
		}
		if req.IssuePrefix == "" {
			prefixParts := strings.SplitN(parentID, "-", 2)
			req.IssuePrefix = prefixParts[0]
		}
		if req.EpicID == "" && parentEpicID.Valid {
			req.EpicID = parentEpicID.String
		}
	}

	if req.EpicID != "" {
		var epicID string
		if err := tx.QueryRow(ctx, `
			SELECT id
			FROM epics
			WHERE project_id = $1 AND id = $2
		`, projectID, req.EpicID).Scan(&epicID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.TaskCreateResponse{}, ErrEpicNotFound
			}
			return domain.TaskCreateResponse{}, fmt.Errorf("lookup epic: %w", err)
		}
	}

	if req.AssignedTo != "" {
		var assignedTo string
		if err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1`, req.AssignedTo).Scan(&assignedTo); err != nil {
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

	var counter int64
	if err := tx.QueryRow(ctx, `
		WITH project_row AS (
			SELECT workspace_id
			FROM projects
			WHERE id = $1
		), upserted AS (
			INSERT INTO id_counters (workspace_id, project_id, entity_type, counter)
			SELECT workspace_id, $1, 'task', 1
			FROM project_row
			ON CONFLICT (project_id, entity_type) DO UPDATE
			SET counter = id_counters.counter + 1
			RETURNING counter
		)
		SELECT counter FROM upserted
	`, projectID).Scan(&counter); err != nil {
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

	_, err = tx.Exec(ctx, `
		INSERT INTO tasks (
			id, workspace_id, project_id, epic_id, parent_task_id, assigned_to,
			created_by, updated_by, title, description, priority,
			status, tags, acceptance_criteria
		) VALUES (
			$1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''),
			NULLIF($7, ''), NULL, $8, NULLIF($9, ''), $10,
			'todo', $11, $12::jsonb
		)
	`, taskID, workspaceID, projectID, req.EpicID, req.ParentTaskID, req.AssignedTo, actorUserID, req.Title, req.Description, priority, tags, string(acceptanceCriteriaJSON))
	if err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("insert task: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.TaskCreateResponse{}, fmt.Errorf("commit task create: %w", err)
	}

	return domain.TaskCreateResponse{
		ID:          taskID,
		IssuePrefix: prefix,
		Title:       req.Title,
		Status:      domain.StatusTodo,
		Priority:    priority,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
	}, nil
}

func (s *Store) ListTasks(ctx context.Context, projectID string, status string, epicID string) (domain.TaskListResponse, error) {
	return s.listTasks(ctx, projectID, status, epicID, "")
}

func (s *Store) ListSubtasks(ctx context.Context, projectID string, parentTaskID string) (domain.TaskListResponse, error) {
	if _, err := s.readTaskDetail(ctx, projectID, parentTaskID); err != nil {
		return domain.TaskListResponse{}, err
	}
	return s.listTasks(ctx, projectID, "", "", parentTaskID)
}

func (s *Store) listTasks(ctx context.Context, projectID string, status string, epicID string, parentTaskID string) (domain.TaskListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.TaskListResponse{}, err
	}

	query := strings.Builder{}
	args := []any{projectID}
	query.WriteString(`
		SELECT id, title, status, priority, epic_id, parent_task_id, tags
		FROM tasks
		WHERE project_id = $1`)
	if parentTaskID == "" {
		query.WriteString(` AND parent_task_id IS NULL`)
	} else {
		args = append(args, parentTaskID)
		fmt.Fprintf(&query, " AND parent_task_id = $%d", len(args))
	}

	if status != "" {
		args = append(args, status)
		fmt.Fprintf(&query, " AND status = $%d", len(args))
	}
	if epicID != "" {
		args = append(args, epicID)
		fmt.Fprintf(&query, " AND epic_id = $%d", len(args))
	}

	query.WriteString(`
		ORDER BY priority ASC, created_at ASC, id ASC
	`)

	rows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return domain.TaskListResponse{}, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	items := make([]domain.TaskListItem, 0)
	for rows.Next() {
		var item domain.TaskListItem
		var statusValue string
		var epicValue sql.NullString
		var parentValue sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &statusValue, &item.Priority, &epicValue, &parentValue, &item.Tags); err != nil {
			return domain.TaskListResponse{}, fmt.Errorf("scan task list item: %w", err)
		}
		item.Status = domain.Status(statusValue)
		if epicValue.Valid {
			item.EpicID = epicValue.String
		}
		if parentValue.Valid {
			item.ParentTaskID = parentValue.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.TaskListResponse{}, fmt.Errorf("iterate tasks: %w", err)
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

	updates := []string{"updated_by = $1", "updated_at = now()"}
	args := []any{actorUserID}
	if req.Title != nil {
		args = append(args, *req.Title)
		updates = append(updates, fmt.Sprintf("title = $%d", len(args)))
	}
	if req.Description != nil {
		args = append(args, *req.Description)
		updates = append(updates, fmt.Sprintf("description = $%d", len(args)))
	}
	if req.Priority != nil {
		args = append(args, *req.Priority)
		updates = append(updates, fmt.Sprintf("priority = $%d", len(args)))
	}
	if req.Status != nil {
		args = append(args, string(*req.Status))
		updates = append(updates, fmt.Sprintf("status = $%d", len(args)))
		if *req.Status == domain.StatusCompleted {
			updates = append(updates, "completed_by = COALESCE(NULLIF($1, ''), completed_by)")
			updates = append(updates, "completed_at = COALESCE(completed_at, now())")
		}
	}
	if req.NoEpic {
		updates = append(updates, "epic_id = NULL")
	} else if req.EpicID != nil {
		args = append(args, *req.EpicID)
		updates = append(updates, fmt.Sprintf("epic_id = NULLIF($%d, '')", len(args)))
	}
	if req.AssignedTo != nil {
		args = append(args, *req.AssignedTo)
		updates = append(updates, fmt.Sprintf("assigned_to = NULLIF($%d, '')", len(args)))
	}
	if req.Tags != nil {
		args = append(args, *req.Tags)
		updates = append(updates, fmt.Sprintf("tags = $%d", len(args)))
	}
	if req.AcceptanceCriteria != nil {
		criteriaJSON, err := json.Marshal(*req.AcceptanceCriteria)
		if err != nil {
			return domain.TaskDetail{}, fmt.Errorf("marshal acceptance criteria: %w", err)
		}
		args = append(args, string(criteriaJSON))
		updates = append(updates, fmt.Sprintf("acceptance_criteria = $%d::jsonb", len(args)))
	}
	if req.CompletionSummary != nil {
		args = append(args, *req.CompletionSummary)
		updates = append(updates, fmt.Sprintf("completion_summary = NULLIF($%d, '')", len(args)))
	}
	if req.CompletionEvidence != nil {
		evidenceJSON, err := json.Marshal(*req.CompletionEvidence)
		if err != nil {
			return domain.TaskDetail{}, fmt.Errorf("marshal completion evidence: %w", err)
		}
		args = append(args, string(evidenceJSON))
		updates = append(updates, fmt.Sprintf("completion_evidence = $%d::jsonb", len(args)))
	}

	query := fmt.Sprintf(`
		UPDATE tasks
		SET %s
		WHERE project_id = $%d AND id = $%d
		RETURNING %s
	`, strings.Join(updates, ", "), len(args)+1, len(args)+2, taskDetailSelectList)
	args = append(args, projectID, taskID)

	row := tx.QueryRow(ctx, query, args...)
	updated, err := scanTaskDetail(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("update task: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.TaskDetail{}, fmt.Errorf("commit task update: %w", err)
	}

	return updated, nil
}

func (s *Store) DeleteTask(ctx context.Context, projectID string, taskID string, actorUserID string) (domain.TaskDetail, error) {
	_ = actorUserID
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.TaskDetail{}, err
	}

	row := s.pool.QueryRow(ctx, `
		DELETE FROM tasks
		WHERE project_id = $1 AND id = $2
		RETURNING `+taskDetailSelectList, projectID, taskID)
	task, err := scanTaskDetail(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("delete task: %w", err)
	}

	return task, nil
}

func (s *Store) ListReadyTasks(ctx context.Context, projectID string, epicID string) (domain.ReadyListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.ReadyListResponse{}, err
	}

	readyQuery := strings.Builder{}
	readyArgs := []any{projectID}
	readyQuery.WriteString(`
		SELECT id, title, status, priority, epic_id, parent_task_id, tags
		FROM tasks t
		WHERE t.project_id = $1
		  AND t.parent_task_id IS NULL
		  AND t.status = 'todo'
		  AND NOT EXISTS (
			SELECT 1
			FROM dependencies d
			JOIN tasks dep ON dep.project_id = d.project_id AND dep.id = d.depends_on_id
			WHERE d.project_id = t.project_id
			  AND d.task_id = t.id
			  AND dep.status NOT IN ('completed', 'wont_fix', 'archived')
		  )`)
	if epicID != "" {
		readyArgs = append(readyArgs, epicID)
		fmt.Fprintf(&readyQuery, " AND t.epic_id = $%d", len(readyArgs))
	}
	readyQuery.WriteString(` ORDER BY t.priority ASC, t.created_at ASC, t.id ASC`)

	rows, err := s.pool.Query(ctx, readyQuery.String(), readyArgs...)
	if err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("query ready tasks: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ReadyListItem, 0)
	readyIDs := make([]string, 0)
	for rows.Next() {
		var item domain.ReadyListItem
		var statusValue string
		var epicValue sql.NullString
		var parentValue sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &statusValue, &item.Priority, &epicValue, &parentValue, &item.Tags); err != nil {
			return domain.ReadyListResponse{}, fmt.Errorf("scan ready task: %w", err)
		}
		item.Status = domain.Status(statusValue)
		if epicValue.Valid {
			item.EpicID = epicValue.String
		}
		if parentValue.Valid {
			item.ParentTaskID = parentValue.String
		}
		items = append(items, item)
		readyIDs = append(readyIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("iterate ready tasks: %w", err)
	}

	if len(readyIDs) == 0 {
		return domain.ReadyListResponse{
			ProjectID: projectID,
			Items:     items,
		}, nil
	}

	type blockedRow struct {
		readyID      string
		item         domain.TaskListItem
		status       string
		epicID       sql.NullString
		parentTaskID sql.NullString
	}

	query := strings.Builder{}
	args := []any{projectID, readyIDs}
	query.WriteString(`
		SELECT d.depends_on_id, t.id, t.title, t.status, t.priority, t.epic_id, t.parent_task_id, t.tags
		FROM dependencies d
		JOIN tasks t ON t.project_id = d.project_id AND t.id = d.task_id
		WHERE d.project_id = $1
		  AND d.depends_on_id = ANY($2::text[])
		  AND t.parent_task_id IS NULL
		  AND t.status = 'todo'
		  AND NOT EXISTS (
			SELECT 1
			FROM dependencies d2
			JOIN tasks dep2 ON dep2.project_id = d2.project_id AND dep2.id = d2.depends_on_id
			WHERE d2.project_id = t.project_id
			  AND d2.task_id = t.id
			  AND d2.depends_on_id <> d.depends_on_id
			  AND dep2.status NOT IN ('completed', 'wont_fix', 'archived')
		  )`)
	if epicID != "" {
		args = append(args, epicID)
		fmt.Fprintf(&query, " AND t.epic_id = $%d", len(args))
	}
	query.WriteString(`
		ORDER BY t.priority ASC, t.created_at ASC, t.id ASC`)

	blockedRows, err := s.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("query ready dependents: %w", err)
	}
	defer blockedRows.Close()

	unblocks := make(map[string][]domain.TaskListItem)
	for blockedRows.Next() {
		var row blockedRow
		if err := blockedRows.Scan(&row.readyID, &row.item.ID, &row.item.Title, &row.status, &row.item.Priority, &row.epicID, &row.parentTaskID, &row.item.Tags); err != nil {
			return domain.ReadyListResponse{}, fmt.Errorf("scan ready dependent: %w", err)
		}
		row.item.Status = domain.Status(row.status)
		if row.epicID.Valid {
			row.item.EpicID = row.epicID.String
		}
		if row.parentTaskID.Valid {
			row.item.ParentTaskID = row.parentTaskID.String
		}
		unblocks[row.readyID] = append(unblocks[row.readyID], row.item)
	}
	if err := blockedRows.Err(); err != nil {
		return domain.ReadyListResponse{}, fmt.Errorf("iterate ready dependents: %w", err)
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

const taskDetailSelectList = `
	id, workspace_id, project_id, epic_id, parent_task_id, assigned_to,
	created_by, updated_by, completed_by, title, description, priority,
	status, tags, acceptance_criteria, completion_evidence, completion_summary,
	completed_at, created_at, updated_at`

const taskDetailQuery = `
	SELECT ` + taskDetailSelectList + `
	FROM tasks`

func scanTaskDetail(row interface {
	Scan(dest ...any) error
}) (domain.TaskDetail, error) {
	var task domain.TaskDetail
	var epicID sql.NullString
	var parentTaskID sql.NullString
	var assignedTo sql.NullString
	var createdBy sql.NullString
	var updatedBy sql.NullString
	var completedBy sql.NullString
	var description sql.NullString
	var completionSummary sql.NullString
	var completedAt sql.NullTime
	var acceptanceCriteriaRaw []byte
	var completionEvidenceRaw []byte
	if err := row.Scan(
		&task.ID,
		&task.WorkspaceID,
		&task.ProjectID,
		&epicID,
		&parentTaskID,
		&assignedTo,
		&createdBy,
		&updatedBy,
		&completedBy,
		&task.Title,
		&description,
		&task.Priority,
		&task.Status,
		&task.Tags,
		&acceptanceCriteriaRaw,
		&completionEvidenceRaw,
		&completionSummary,
		&completedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return domain.TaskDetail{}, err
	}
	if epicID.Valid {
		task.EpicID = epicID.String
	}
	if parentTaskID.Valid {
		task.ParentTaskID = parentTaskID.String
	}
	if assignedTo.Valid {
		task.AssignedTo = assignedTo.String
	}
	if createdBy.Valid {
		task.CreatedBy = createdBy.String
	}
	if updatedBy.Valid {
		task.UpdatedBy = updatedBy.String
	}
	if completedBy.Valid {
		task.CompletedBy = completedBy.String
	}
	if description.Valid {
		task.Description = description.String
	}
	if completionSummary.Valid {
		task.CompletionSummary = completionSummary.String
	}
	if completedAt.Valid {
		task.CompletedAt = completedAt.Time
	}
	if len(acceptanceCriteriaRaw) > 0 {
		if err := json.Unmarshal(acceptanceCriteriaRaw, &task.AcceptanceCriteria); err != nil {
			return domain.TaskDetail{}, fmt.Errorf("decode acceptance criteria: %w", err)
		}
	}
	if len(completionEvidenceRaw) > 0 {
		if err := json.Unmarshal(completionEvidenceRaw, &task.CompletionEvidence); err != nil {
			return domain.TaskDetail{}, fmt.Errorf("decode completion evidence: %w", err)
		}
	}
	return task, nil
}

func (s *Store) ensureProjectExists(ctx context.Context, projectID string) error {
	var workspaceID string
	if err := s.pool.QueryRow(ctx, `SELECT workspace_id FROM projects WHERE id = $1`, projectID).Scan(&workspaceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("lookup project: %w", err)
	}
	return nil
}

func (s *Store) readTaskDetail(ctx context.Context, projectID string, taskID string) (domain.TaskDetail, error) {
	row := s.pool.QueryRow(ctx, taskDetailQuery+` WHERE project_id = $1 AND id = $2`, projectID, taskID)
	task, err := scanTaskDetail(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TaskDetail{}, ErrTaskNotFound
		}
		return domain.TaskDetail{}, fmt.Errorf("load task detail: %w", err)
	}
	return task, nil
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

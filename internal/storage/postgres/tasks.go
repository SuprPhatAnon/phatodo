package postgres

import (
	"context"
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
			id, workspace_id, project_id, epic_id, assigned_to,
			created_by, updated_by, title, description, priority,
			status, tags, acceptance_criteria
		) VALUES (
			$1, $2, $3, NULLIF($4, ''), NULLIF($5, ''),
			NULLIF($6, ''), NULL, $7, NULLIF($8, ''), $9,
			'todo', $10, $11::jsonb
		)
	`, taskID, workspaceID, projectID, req.EpicID, req.AssignedTo, actorUserID, req.Title, req.Description, priority, tags, string(acceptanceCriteriaJSON))
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

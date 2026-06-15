package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5"
)

var ErrCommentNotFound = errors.New("comment not found")

func (s *Store) ListComments(ctx context.Context, projectID string, taskID string) (domain.CommentListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.CommentListResponse{}, err
	}
	if _, err := s.readTaskDetail(ctx, projectID, taskID); err != nil {
		return domain.CommentListResponse{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at
		FROM comments
		WHERE project_id = $1 AND task_id = $2
		ORDER BY created_at ASC, id ASC
	`, projectID, taskID)
	if err != nil {
		return domain.CommentListResponse{}, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Comment, 0)
	for rows.Next() {
		item, err := scanComment(rows)
		if err != nil {
			return domain.CommentListResponse{}, fmt.Errorf("scan comment: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.CommentListResponse{}, fmt.Errorf("iterate comments: %w", err)
	}

	return domain.CommentListResponse{
		ProjectID: projectID,
		TaskID:    taskID,
		Items:     items,
	}, nil
}

func (s *Store) CreateComment(ctx context.Context, projectID string, taskID string, req domain.CommentCreateRequest, actorUserID string) (domain.Comment, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Comment{}, fmt.Errorf("begin comment create tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var workspaceID string
	if err := tx.QueryRow(ctx, `SELECT workspace_id FROM projects WHERE id = $1`, projectID).Scan(&workspaceID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrProjectNotFound
		}
		return domain.Comment{}, fmt.Errorf("lookup project: %w", err)
	}

	var existingTaskID string
	if err := tx.QueryRow(ctx, `SELECT id FROM tasks WHERE project_id = $1 AND id = $2`, projectID, taskID).Scan(&existingTaskID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrTaskNotFound
		}
		return domain.Comment{}, fmt.Errorf("lookup task: %w", err)
	}

	commentID, err := randomID("cmt")
	if err != nil {
		return domain.Comment{}, err
	}

	var comment domain.Comment
	err = tx.QueryRow(ctx, `
		INSERT INTO comments (
			id, workspace_id, project_id, task_id, author_user_id, author, kind, content
		) VALUES (
			$1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8
		)
		RETURNING `+commentSelectList,
		commentID, workspaceID, projectID, taskID, actorUserID, req.Author, req.Kind, req.Content,
	).Scan(
		&comment.ID,
		&comment.WorkspaceID,
		&comment.ProjectID,
		&comment.TaskID,
		&comment.AuthorUserID,
		&comment.Author,
		&comment.Kind,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("insert comment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, fmt.Errorf("commit comment create: %w", err)
	}

	return comment, nil
}

func (s *Store) UpdateComment(ctx context.Context, projectID string, commentID string, req domain.CommentUpdateRequest, actorUserID string) (domain.Comment, error) {
	_ = actorUserID
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Comment{}, err
	}

	row := s.pool.QueryRow(ctx, `
		UPDATE comments
		SET content = $1,
			updated_at = now()
		WHERE project_id = $2 AND id = $3
		RETURNING `+commentSelectList, req.Content, projectID, commentID)
	comment, err := scanComment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrCommentNotFound
		}
		return domain.Comment{}, fmt.Errorf("update comment: %w", err)
	}
	return comment, nil
}

func (s *Store) DeleteComment(ctx context.Context, projectID string, commentID string, actorUserID string) (domain.Comment, error) {
	_ = actorUserID
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Comment{}, err
	}

	row := s.pool.QueryRow(ctx, `
		DELETE FROM comments
		WHERE project_id = $1 AND id = $2
		RETURNING `+commentSelectList, projectID, commentID)
	comment, err := scanComment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrCommentNotFound
		}
		return domain.Comment{}, fmt.Errorf("delete comment: %w", err)
	}
	return comment, nil
}

const commentSelectList = `
	id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at`

func scanComment(row interface {
	Scan(dest ...any) error
}) (domain.Comment, error) {
	var comment domain.Comment
	var authorUserID sql.NullString
	if err := row.Scan(
		&comment.ID,
		&comment.WorkspaceID,
		&comment.ProjectID,
		&comment.TaskID,
		&authorUserID,
		&comment.Author,
		&comment.Kind,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	); err != nil {
		return domain.Comment{}, err
	}
	if authorUserID.Valid {
		comment.AuthorUserID = authorUserID.String
	}
	return comment, nil
}

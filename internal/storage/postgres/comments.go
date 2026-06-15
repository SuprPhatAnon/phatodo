package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
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

	rows, err := s.q.ListComments(ctx, db.ListCommentsParams{ProjectID: projectID, TaskID: taskID})
	if err != nil {
		return domain.CommentListResponse{}, fmt.Errorf("query comments: %w", err)
	}

	items := make([]domain.Comment, 0)
	for _, row := range rows {
		items = append(items, commentFromSQLC(row))
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
	q := s.q.WithTx(tx)

	workspaceID, err = q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrProjectNotFound
		}
		return domain.Comment{}, fmt.Errorf("lookup project: %w", err)
	}

	if _, err := q.GetTaskDetail(ctx, db.GetTaskDetailParams{ProjectID: projectID, TaskID: taskID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrTaskNotFound
		}
		return domain.Comment{}, fmt.Errorf("lookup task: %w", err)
	}

	commentID, err := randomID("cmt")
	if err != nil {
		return domain.Comment{}, err
	}

	commentRow, err := q.CreateComment(ctx, db.CreateCommentParams{
		ID:           commentID,
		WorkspaceID:  workspaceID,
		ProjectID:    projectID,
		TaskID:       taskID,
		AuthorUserID: actorUserID,
		Author:       req.Author,
		Kind:         req.Kind,
		Content:      req.Content,
	})
	if err != nil {
		return domain.Comment{}, fmt.Errorf("insert comment: %w", err)
	}
	comment := commentFromSQLC(commentRow)

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "create",
		EntityType:  "comment",
		EntityID:    comment.ID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		AfterState:  comment,
	}); err != nil {
		return domain.Comment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, fmt.Errorf("commit comment create: %w", err)
	}

	return comment, nil
}

func (s *Store) UpdateComment(ctx context.Context, projectID string, commentID string, req domain.CommentUpdateRequest, actorUserID string) (domain.Comment, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Comment{}, fmt.Errorf("begin comment update tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Comment{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrProjectNotFound
		}
		return domain.Comment{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readCommentTx(ctx, tx, projectID, commentID)
	if err != nil {
		return domain.Comment{}, err
	}

	commentRow, err := q.UpdateComment(ctx, db.UpdateCommentParams{
		Content:   req.Content,
		ProjectID: projectID,
		CommentID: commentID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrCommentNotFound
		}
		return domain.Comment{}, fmt.Errorf("update comment: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "update",
		EntityType:  "comment",
		EntityID:    commentID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  commentRow,
	}); err != nil {
		return domain.Comment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, fmt.Errorf("commit comment update: %w", err)
	}

	return commentFromSQLC(commentRow), nil
}

func (s *Store) DeleteComment(ctx context.Context, projectID string, commentID string, actorUserID string) (domain.Comment, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Comment{}, fmt.Errorf("begin comment delete tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.Comment{}, err
	}

	q := s.q.WithTx(tx)

	workspaceID, err := q.GetProjectWorkspaceID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrProjectNotFound
		}
		return domain.Comment{}, fmt.Errorf("lookup project workspace: %w", err)
	}

	before, err := s.readCommentTx(ctx, tx, projectID, commentID)
	if err != nil {
		return domain.Comment{}, err
	}

	commentRow, err := q.DeleteComment(ctx, db.DeleteCommentParams{ProjectID: projectID, CommentID: commentID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrCommentNotFound
		}
		return domain.Comment{}, fmt.Errorf("delete comment: %w", err)
	}

	if err := s.recordEventTx(ctx, tx, auditEvent{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Action:      "delete",
		EntityType:  "comment",
		EntityID:    commentID,
		ActorUserID: actorUserID,
		ActorLabel:  actorUserID,
		BeforeState: before,
		AfterState:  nil,
	}); err != nil {
		return domain.Comment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, fmt.Errorf("commit comment delete: %w", err)
	}

	return commentFromSQLC(commentRow), nil
}

func (s *Store) readCommentTx(ctx context.Context, tx pgx.Tx, projectID string, commentID string) (domain.Comment, error) {
	comment, err := s.q.WithTx(tx).GetComment(ctx, db.GetCommentParams{ProjectID: projectID, CommentID: commentID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, ErrCommentNotFound
		}
		return domain.Comment{}, fmt.Errorf("load comment: %w", err)
	}
	return commentFromSQLC(comment), nil
}

func commentFromSQLC(row db.Comment) domain.Comment {
	comment := domain.Comment{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		ProjectID:   row.ProjectID,
		TaskID:      row.TaskID,
		Author:      row.Author,
		Kind:        row.Kind,
		Content:     row.Content,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.AuthorUserID != nil {
		comment.AuthorUserID = *row.AuthorUserID
	}
	return comment
}

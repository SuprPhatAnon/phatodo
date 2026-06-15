-- name: ListComments :many
SELECT id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at
FROM comments
WHERE project_id = sqlc.arg(project_id) AND task_id = sqlc.arg(task_id)
ORDER BY created_at ASC, id ASC;

-- name: GetComment :one
SELECT id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at
FROM comments
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(comment_id)
LIMIT 1;

-- name: CreateComment :one
INSERT INTO comments (
	id, workspace_id, project_id, task_id, author_user_id, author, kind, content
) VALUES (
	sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(project_id), sqlc.arg(task_id), NULLIF(sqlc.arg(author_user_id), ''),
	sqlc.arg(author), sqlc.arg(kind), sqlc.arg(content)
)
RETURNING id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at;

-- name: UpdateComment :one
UPDATE comments
SET content = sqlc.arg(content),
	updated_at = now()
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(comment_id)
RETURNING id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at;

-- name: DeleteComment :one
DELETE FROM comments
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(comment_id)
RETURNING id, workspace_id, project_id, task_id, author_user_id, author, kind, content, created_at, updated_at;


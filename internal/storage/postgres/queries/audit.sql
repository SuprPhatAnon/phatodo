-- name: InsertEvent :exec
INSERT INTO events (
	workspace_id, project_id, action, entity_type, entity_id,
	actor_user_id, actor_label, before_state, after_state, metadata
) VALUES (
	sqlc.arg(workspace_id), sqlc.arg(project_id), sqlc.arg(action), sqlc.arg(entity_type), sqlc.arg(entity_id),
	NULLIF(sqlc.arg(actor_user_id), ''), NULLIF(sqlc.arg(actor_label), ''),
	sqlc.narg(before_state)::jsonb, sqlc.narg(after_state)::jsonb, COALESCE(sqlc.narg(metadata)::jsonb, '{}'::jsonb)
);

-- name: SearchEpics :many
SELECT id, title, COALESCE(description, ''), status, priority, created_at, updated_at
FROM epics
WHERE project_id = sqlc.arg(project_id)
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
  AND (
	LOWER(id) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(title) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(COALESCE(description, '')) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
  )
ORDER BY created_at DESC, id ASC;

-- name: SearchTasks :many
SELECT id, title, COALESCE(description, ''), status, priority, epic_id, parent_task_id, created_at, updated_at
FROM tasks
WHERE project_id = sqlc.arg(project_id)
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
  AND (
	LOWER(id) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(title) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(COALESCE(description, '')) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
  )
ORDER BY created_at DESC, id ASC;

-- name: SearchComments :many
SELECT id, author, kind, content, created_at, updated_at, task_id
FROM comments
WHERE project_id = sqlc.arg(project_id)
  AND (
	LOWER(id) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(author) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(kind) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
	OR LOWER(content) LIKE '%' || LOWER(sqlc.arg(query)) || '%'
  )
ORDER BY created_at DESC, id ASC;

-- name: ListTasksUnified :many
SELECT id, title, COALESCE(description, ''), status, priority, epic_id, parent_task_id, tags, created_at, updated_at
FROM tasks
WHERE project_id = sqlc.arg(project_id)
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
ORDER BY created_at ASC, id ASC;

-- name: HistoryEvents :many
SELECT id, workspace_id, project_id, action, entity_type, entity_id, actor_user_id, actor_label, before_state, after_state, metadata, created_at
FROM events
WHERE project_id = sqlc.arg(project_id)
  AND (sqlc.arg(entity_id) = '' OR entity_id = sqlc.arg(entity_id))
  AND (sqlc.arg(entity_type) = '' OR entity_type = sqlc.arg(entity_type))
  AND (sqlc.arg(action) = '' OR action = sqlc.arg(action))
  AND (sqlc.arg(since) IS NULL OR created_at >= sqlc.arg(since))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(query_limit);

-- name: NextLockCounter :one
WITH project_row AS (
	SELECT workspace_id
	FROM projects
	WHERE id = sqlc.arg(project_id)
), upserted AS (
	INSERT INTO id_counters (workspace_id, project_id, entity_type, counter)
	SELECT workspace_id, sqlc.arg(project_id), 'lock', 1
	FROM project_row
	ON CONFLICT (project_id, entity_type) DO UPDATE
	SET counter = id_counters.counter + 1
	RETURNING counter
)
SELECT counter FROM upserted;

-- name: CreateWorkItemLock :one
INSERT INTO work_item_locks (
	id, workspace_id, project_id, entity_type, entity_id,
	locked_by, reason, expires_at
) VALUES (
	sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(project_id), sqlc.arg(entity_type), sqlc.arg(entity_id),
	sqlc.arg(locked_by), NULLIF(sqlc.arg(reason), ''), sqlc.arg(expires_at)
)
RETURNING id, workspace_id, project_id, entity_type, entity_id, locked_by, reason, leased_at, expires_at, released_at;

-- name: GetWorkItemLock :one
SELECT id, workspace_id, project_id, entity_type, entity_id, locked_by, reason, leased_at, expires_at, released_at
FROM work_item_locks
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(lock_id)
LIMIT 1;

-- name: GetActiveWorkItemLock :one
SELECT id, workspace_id, project_id, entity_type, entity_id, locked_by, reason, leased_at, expires_at, released_at
FROM work_item_locks
WHERE project_id = sqlc.arg(project_id)
  AND entity_type = sqlc.arg(entity_type)
  AND entity_id = sqlc.arg(entity_id)
  AND released_at IS NULL
  AND expires_at > now()
LIMIT 1;

-- name: ReleaseExpiredWorkItemLock :exec
UPDATE work_item_locks
SET released_at = now()
WHERE project_id = sqlc.arg(project_id)
  AND entity_type = sqlc.arg(entity_type)
  AND entity_id = sqlc.arg(entity_id)
  AND released_at IS NULL
  AND expires_at <= now();

-- name: ReleaseWorkItemLock :one
UPDATE work_item_locks
SET released_at = COALESCE(released_at, now())
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(lock_id)
RETURNING id, workspace_id, project_id, entity_type, entity_id, locked_by, reason, leased_at, expires_at, released_at;

-- name: ListWorkItemLocks :many
SELECT id, workspace_id, project_id, entity_type, entity_id, locked_by, reason, leased_at, expires_at, released_at
FROM work_item_locks
WHERE project_id = sqlc.arg(project_id)
  AND (COALESCE(array_length(sqlc.arg(entity_types)::text[], 1), 0) = 0 OR entity_type = ANY(sqlc.arg(entity_types)::text[]))
  AND (sqlc.arg(entity_id) = '' OR entity_id = sqlc.arg(entity_id))
  AND (NOT sqlc.arg(active)::bool OR (released_at IS NULL AND expires_at > now()))
ORDER BY leased_at DESC, id DESC;

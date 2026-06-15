-- name: ListProjectConfig :many
SELECT key, value
FROM project_config
WHERE project_id = sqlc.arg(project_id)
ORDER BY key;

-- name: GetProjectConfig :one
SELECT key, value
FROM project_config
WHERE project_id = sqlc.arg(project_id) AND key = sqlc.arg(key);

-- name: SetProjectConfig :one
WITH project_row AS (
	SELECT workspace_id
	FROM projects
	WHERE id = sqlc.arg(project_id)
), upserted AS (
	INSERT INTO project_config (
		workspace_id, project_id, key, value
	)
	SELECT workspace_id, sqlc.arg(project_id), sqlc.arg(key), sqlc.arg(value)
	FROM project_row
	ON CONFLICT (project_id, key) DO UPDATE
	SET value = EXCLUDED.value,
		updated_at = now()
	RETURNING key, value
)
SELECT key, value FROM upserted;

-- name: DeleteProjectConfig :one
DELETE FROM project_config
WHERE project_id = sqlc.arg(project_id) AND key = sqlc.arg(key)
RETURNING key, value;


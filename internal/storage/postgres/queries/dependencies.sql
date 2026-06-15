-- name: ListDependencies :many
SELECT id, workspace_id, project_id, task_id, depends_on_id, created_at
FROM dependencies
WHERE project_id = sqlc.arg(project_id) AND task_id = sqlc.arg(task_id)
ORDER BY created_at ASC, id ASC;

-- name: GetDependency :one
SELECT id, workspace_id, project_id, task_id, depends_on_id, created_at
FROM dependencies
WHERE project_id = sqlc.arg(project_id) AND task_id = sqlc.arg(task_id) AND depends_on_id = sqlc.arg(depends_on_id)
LIMIT 1;

-- name: CreateDependency :one
INSERT INTO dependencies (
	id, workspace_id, project_id, task_id, depends_on_id
) VALUES (
	sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(project_id), sqlc.arg(task_id), sqlc.arg(depends_on_id)
)
RETURNING id, workspace_id, project_id, task_id, depends_on_id, created_at;

-- name: DeleteDependency :one
DELETE FROM dependencies
WHERE project_id = sqlc.arg(project_id) AND task_id = sqlc.arg(task_id) AND depends_on_id = sqlc.arg(depends_on_id)
RETURNING id, workspace_id, project_id, task_id, depends_on_id, created_at;

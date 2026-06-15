-- name: NextEpicCounter :one
WITH project_row AS (
	SELECT workspace_id
	FROM projects
	WHERE id = sqlc.arg(project_id)
), upserted AS (
	INSERT INTO id_counters (workspace_id, project_id, entity_type, counter)
	SELECT workspace_id, sqlc.arg(project_id), 'epic', 1
	FROM project_row
	ON CONFLICT (project_id, entity_type) DO UPDATE
	SET counter = id_counters.counter + 1
	RETURNING counter
)
SELECT counter FROM upserted;

-- name: CreateEpic :one
INSERT INTO epics (
	id, workspace_id, project_id, assigned_to, created_by, updated_by,
	title, description, priority, status, acceptance_criteria
) VALUES (
	sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(project_id),
	NULLIF(sqlc.arg(assigned_to), ''), NULLIF(sqlc.arg(created_by), ''), NULL,
	sqlc.arg(title), NULLIF(sqlc.arg(description), ''), sqlc.arg(priority), 'todo', sqlc.arg(acceptance_criteria)::jsonb
)
RETURNING id, workspace_id, project_id, assigned_to, created_by, updated_by, completed_by,
	title, description, status, priority, acceptance_criteria, completion_evidence,
	completion_summary, completed_at, created_at, updated_at;

-- name: ListEpics :many
SELECT id, workspace_id, project_id, assigned_to, created_by, updated_by, completed_by,
	title, description, status, priority, acceptance_criteria, completion_evidence,
	completion_summary, completed_at, created_at, updated_at
FROM epics
WHERE project_id = sqlc.arg(project_id)
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
ORDER BY priority ASC, created_at ASC, id ASC;

-- name: GetEpic :one
SELECT id, workspace_id, project_id, assigned_to, created_by, updated_by, completed_by,
	title, description, status, priority, acceptance_criteria, completion_evidence,
	completion_summary, completed_at, created_at, updated_at
FROM epics
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(epic_id)
LIMIT 1;

-- name: UpdateEpic :one
UPDATE epics
SET updated_by = sqlc.arg(updated_by),
	updated_at = now(),
	title = COALESCE(sqlc.narg(title), title),
	description = COALESCE(sqlc.narg(description), description),
	priority = COALESCE(sqlc.narg(priority), priority),
	status = COALESCE(sqlc.narg(status), status),
	assigned_to = COALESCE(sqlc.narg(assigned_to), assigned_to),
	acceptance_criteria = COALESCE(sqlc.narg(acceptance_criteria)::jsonb, acceptance_criteria),
	completion_summary = COALESCE(sqlc.narg(completion_summary), completion_summary),
	completion_evidence = COALESCE(sqlc.narg(completion_evidence)::jsonb, completion_evidence),
	completed_by = CASE
		WHEN COALESCE(sqlc.narg(status), '') = 'completed' THEN COALESCE(NULLIF(sqlc.arg(updated_by), ''), completed_by)
		ELSE completed_by
	END,
	completed_at = CASE
		WHEN COALESCE(sqlc.narg(status), '') = 'completed' THEN COALESCE(completed_at, now())
		ELSE completed_at
	END
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(epic_id)
RETURNING id, workspace_id, project_id, assigned_to, created_by, updated_by, completed_by,
	title, description, status, priority, acceptance_criteria, completion_evidence,
	completion_summary, completed_at, created_at, updated_at;

-- name: DeleteEpic :one
DELETE FROM epics
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(epic_id)
RETURNING id, workspace_id, project_id, assigned_to, created_by, updated_by, completed_by,
	title, description, status, priority, acceptance_criteria, completion_evidence,
	completion_summary, completed_at, created_at, updated_at;

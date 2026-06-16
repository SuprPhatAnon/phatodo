-- name: GetProjectWorkspaceID :one
SELECT workspace_id
FROM projects
WHERE id = sqlc.arg(project_id)
LIMIT 1;

-- name: GetTaskCreateParentInfo :one
SELECT id, epic_id, parent_task_id
FROM tasks
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(task_id)
LIMIT 1;

-- name: GetEpicID :one
SELECT id
FROM epics
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(epic_id)
LIMIT 1;

-- name: GetUserID :one
SELECT id
FROM users
WHERE id = sqlc.arg(user_id)
LIMIT 1;

-- name: NextTaskCounter :one
WITH project_row AS (
	SELECT workspace_id
	FROM projects
	WHERE id = sqlc.arg(project_id)
), upserted AS (
	INSERT INTO id_counters (workspace_id, project_id, entity_type, counter)
	SELECT workspace_id, sqlc.arg(project_id), 'task', 1
	FROM project_row
	ON CONFLICT (project_id, entity_type) DO UPDATE
	SET counter = id_counters.counter + 1
	RETURNING counter
)
SELECT counter FROM upserted;

-- name: CreateTask :one
INSERT INTO tasks (
	id, workspace_id, project_id, epic_id, parent_task_id, assigned_to,
	created_by, updated_by, title, description, kind, root_cause_analysis, priority,
	status, tags, acceptance_criteria
) VALUES (
	sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(project_id), NULLIF(sqlc.arg(epic_id), ''),
	NULLIF(sqlc.arg(parent_task_id), ''), NULLIF(sqlc.arg(assigned_to), ''),
	NULLIF(sqlc.arg(created_by), ''), NULL, sqlc.arg(title), NULLIF(sqlc.arg(description), ''), sqlc.arg(kind), sqlc.arg(root_cause_analysis), sqlc.arg(priority),
	'todo', sqlc.arg(tags), sqlc.arg(acceptance_criteria)::jsonb
)
RETURNING id, title, status, priority, project_id, workspace_id, kind, root_cause_analysis;

-- name: ListTasks :many
SELECT id, title, status, priority, kind, epic_id, parent_task_id, tags, created_at, updated_at
FROM tasks
WHERE project_id = sqlc.arg(project_id)
  AND parent_task_id IS NULL
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
  AND (sqlc.arg(epic_id) = '' OR epic_id = sqlc.arg(epic_id))
ORDER BY priority ASC, created_at ASC, id ASC;

-- name: ListSubtasks :many
SELECT id, title, status, priority, kind, epic_id, parent_task_id, tags, created_at, updated_at
FROM tasks
WHERE project_id = sqlc.arg(project_id)
  AND parent_task_id = sqlc.arg(parent_task_id)
  AND (sqlc.arg(status) = '' OR status = sqlc.arg(status))
ORDER BY priority ASC, created_at ASC, id ASC;

-- name: GetTaskDetail :one
SELECT id, workspace_id, project_id, epic_id, parent_task_id, assigned_to,
	created_by, updated_by, completed_by, title, description, kind, root_cause_analysis, priority,
	status, tags, acceptance_criteria, completion_evidence, completion_summary,
	completed_at, created_at, updated_at
FROM tasks
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(task_id);

-- name: UpdateTask :one
UPDATE tasks
SET updated_by = sqlc.arg(updated_by),
	updated_at = now(),
	title = COALESCE(sqlc.narg(title), title),
	description = COALESCE(sqlc.narg(description), description),
	kind = COALESCE(sqlc.narg(kind), kind),
	root_cause_analysis = COALESCE(sqlc.narg(root_cause_analysis), root_cause_analysis),
	priority = COALESCE(sqlc.narg(priority), priority),
	status = COALESCE(sqlc.narg(status), status),
	tags = COALESCE(sqlc.narg(tags), tags),
	epic_id = CASE WHEN sqlc.arg(clear_epic)::bool THEN NULL ELSE COALESCE(sqlc.narg(epic_id), epic_id) END,
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
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(task_id)
RETURNING id, workspace_id, project_id, epic_id, parent_task_id, assigned_to,
	created_by, updated_by, completed_by, title, description, priority,
	status, tags, acceptance_criteria, completion_evidence, completion_summary,
	completed_at, created_at, updated_at;

-- name: DeleteTask :one
DELETE FROM tasks
WHERE project_id = sqlc.arg(project_id) AND id = sqlc.arg(task_id)
RETURNING id, workspace_id, project_id, epic_id, parent_task_id, kind, root_cause_analysis, assigned_to,
	created_by, updated_by, completed_by, title, description, priority,
	status, tags, acceptance_criteria, completion_evidence, completion_summary,
	completed_at, created_at, updated_at;

-- name: ListReadyTasks :many
SELECT id, title, description, status, priority, epic_id, parent_task_id, tags
FROM tasks t
WHERE t.project_id = sqlc.arg(project_id)
  AND t.parent_task_id IS NULL
  AND t.status = 'todo'
  AND (sqlc.arg(epic_id) = '' OR t.epic_id = sqlc.arg(epic_id))
  AND NOT EXISTS (
	SELECT 1
	FROM dependencies d
	JOIN tasks dep ON dep.project_id = d.project_id AND dep.id = d.depends_on_id
	WHERE d.project_id = t.project_id
	  AND d.task_id = t.id
	  AND dep.status NOT IN ('completed', 'wont_fix', 'archived')
  )
ORDER BY t.priority ASC, t.created_at ASC, t.id ASC;

-- name: ListReadyDependents :many
SELECT d.depends_on_id, t.id, t.title, t.status, t.priority, t.epic_id, t.parent_task_id, t.tags
FROM dependencies d
JOIN tasks t ON t.project_id = d.project_id AND t.id = d.task_id
WHERE d.project_id = sqlc.arg(project_id)
  AND d.depends_on_id = ANY(sqlc.arg(ready_ids)::text[])
  AND t.parent_task_id IS NULL
  AND t.status = 'todo'
  AND (sqlc.arg(epic_id) = '' OR t.epic_id = sqlc.arg(epic_id))
  AND NOT EXISTS (
	SELECT 1
	FROM dependencies d2
	JOIN tasks dep ON dep.project_id = d2.project_id AND dep.id = d2.depends_on_id
	WHERE d2.project_id = t.project_id
	  AND d2.task_id = t.id
	  AND dep.status NOT IN ('completed', 'wont_fix', 'archived')
  )
ORDER BY t.priority ASC, t.created_at ASC, t.id ASC;

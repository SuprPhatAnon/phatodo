# Command Contracts

This document defines the request and response contract for the `ptodo` command.

The same contract applies to `ptodo`.

## Shared Conventions

- All API calls are project-scoped unless explicitly noted.
- All authenticated requests send `X-Phatodo-Access-Key` and `X-Phatodo-Access-Secret`, except the bootstrap routes described below.
- `--toon` is the compact machine-friendly TOON output mode.
- Subtasks are stored in the `tasks` table, so subtask routes point at task IDs.

## Shared Field Model

### Work item fields

These fields may appear on epics, tasks, and subtasks depending on the command:

- `title`
- `description`
- `status`
- `priority`
- `assigned_to`
- `created_by`
- `updated_by`
- `completed_by`
- `acceptance_criteria`
- `completion_evidence`
- `completion_summary`
- `completed_at`

### Task-only fields

- `epic_id`
- `parent_task_id`
- `tags`

### Comment fields

- `author`
- `author_user_id`
- `kind`
- `content`

Allowed comment kinds:

- `comment`
- `analysis`
- `summary`
- `checkpoint`
- `handoff`

### Lock fields

- `entity_type`
- `entity_id`
- `reason`
- `leased_at`
- `expires_at`
- `released_at`

## Setup

### `ptodo admin init`

Purpose:
- create the first admin user only

Required input:
- `-u <username>`
- `--url <api-server-url>`

Prompts:
- password twice for confirmation

Contract:
- create the first admin user only
- fail if any admin user already exists
- create the admin access credentials for that admin user record
- do not create a workspace
- do not create a project
- do not write local CLI config
- do not create project config

### `ptodo admin bootstrap`

Purpose:
- provision the workspace, project, project-scoped CLI identity, and access row

Required input:
- `-u <username>`
- `--url <api-server-url>`

Prompts:
- admin password

Optional input:
- `--project <name>` or legacy `--project-name <name>`

Contract:
- authenticate as the existing admin user
- create the workspace and project if they do not already exist
- create the project-scoped CLI identity used by the local config
- create the matching `user_project_access` row for that identity
- write `.phatodo/config.json`
- fail if the target project already exists

Local file fields:
- `api_url`
- `workspace_id`
- `project_id`
- `access_key`
- `access_secret`

### `ptodo wipe -y`

Purpose:
- remove local client configuration and optional cache only

## Epics

### `ptodo epic create`

Required input:
- `-t <title>`

Optional input:
- `-d <description>`
- `-p <priority>`
- `-a <user-id>`
- `--criteria-json '["..."]'`

Server request:
- `POST /api/v1/projects/{projectID}/epics`

Expected body fields:
- `title`
- `description`
- `priority`
- `assigned_to`
- `acceptance_criteria`

### `ptodo epic list`

Optional filters:
- `--status`
- `--limit <n>` defaults to `20`

Server request:
- `GET /api/v1/projects/{projectID}/epics`

### `ptodo epic show`

Required input:
- `<epic-id>`

Server request:
- `GET /api/v1/projects/{projectID}/epics/{epicID}`

### `ptodo epic update`

Required input:
- `<epic-id>`

Optional input:
- `-t <title>`
- `-d <description>`
- `-p <priority>`
- `-s <status>`
- `-a <user-id>`
- `--criteria-json '["..."]'`
- `--summary <completion-summary>`
- `--evidence-json '["..."]'`

Server request:
- `PATCH /api/v1/projects/{projectID}/epics/{epicID}`

Expected body fields:
- any mutable epic field

### `ptodo epic complete`

Required input:
- `<epic-id>`

Server request:
- `POST /api/v1/projects/{projectID}/epics/{epicID}/complete`

Contract:
- marks the epic completed
- archives or closes child tasks according to validation rules
- records completion metadata and history

### `ptodo epic delete`

Required input:
- `<epic-id>`

Server request:
- `DELETE /api/v1/projects/{projectID}/epics/{epicID}`

## Tasks

### `ptodo task create`

Required input:
- `-t <title>`
- `--prefix <prefix>` or legacy `--issue-prefix <prefix>`

Optional input:
- `-d <description>`
- `-p <priority>`
- `-e <epic-id>`
- `--tags "a,b"`
- `-a <user-id>`
- `--criteria-json '["..."]'`

Server request:
- `POST /api/v1/projects/{projectID}/tasks`

Expected body fields:
- `title`
- `issue_prefix`
- `description`
- `priority`
- `epic_id`
- `tags`
- `assigned_to`
- `acceptance_criteria`

Response:
- the created task row, including generated task ID and the normalized issue prefix used to create it

### `ptodo task list`

Optional filters:
- `--status <status>`
- `--epic <epic-id>`
- `--limit <n>` defaults to `20`

Server request:
- `GET /api/v1/projects/{projectID}/tasks`

Contract:
- return top-level tasks for the project by default
- apply `--status` as an exact status filter when provided
- apply `--epic` as an exact epic filter when provided
- do not include subtasks unless a subtask-specific command is used

Response:
- a project-scoped task list with the current task records
- each item includes the task ID, title, status, priority, and any epic or parent task relationship that is present

### `ptodo task show`

Required input:
- `<task-id>`

Server request:
- `GET /api/v1/projects/{projectID}/tasks/{taskID}`

### `ptodo task update`

Required input:
- `<task-id>`

Optional input:
- `-t <title>`
- `-d <description>`
- `-p <priority>`
- `-s <status>`
- `--tags "a,b"`
- `-e <epic-id>`
- `--no-epic`
- `-a <user-id>`
- `--criteria-json '["..."]'`
- `--summary <completion-summary>`
- `--evidence-json '["..."]'`

Server request:
- `PATCH /api/v1/projects/{projectID}/tasks/{taskID}`

### `ptodo task delete`

Required input:
- `<task-id>`

Server request:
- `DELETE /api/v1/projects/{projectID}/tasks/{taskID}`

## Subtasks

### `ptodo subtask create`

Required input:
- `<task-id>`
- `-t <title>`

Optional input:
- `-d <description>`
- `-p <priority>`
- `-a <user-id>`
- `--criteria-json '["..."]'`

Server request:
- `POST /api/v1/projects/{projectID}/tasks/{taskID}/subtasks`

### `ptodo subtask list`

Required input:
- `<task-id>`

Optional filters:
- `--limit <n>` defaults to `20`

Server request:
- `GET /api/v1/projects/{projectID}/tasks/{taskID}/subtasks`

### `ptodo subtask update`

Required input:
- `<subtask-id>`

Optional input:
- `-t <title>`
- `-d <description>`
- `-p <priority>`
- `-s <status>`
- `-a <user-id>`
- `--criteria-json '["..."]'`
- `--summary <completion-summary>`
- `--evidence-json '["..."]'`

Server request:
- `PATCH /api/v1/projects/{projectID}/tasks/{subtaskID}`

### `ptodo subtask delete`

Required input:
- `<subtask-id>`

Server request:
- `DELETE /api/v1/projects/{projectID}/tasks/{subtaskID}`

## Comments

### `ptodo comment add`

Required input:
- `<task-id>`
- `-a <author>`
- `-c <content>`

Optional input:
- `-k <kind>`

Server request:
- `POST /api/v1/projects/{projectID}/tasks/{taskID}/comments`

Expected body fields:
- `author`
- `content`
- `kind`

### `ptodo comment list`

Required input:
- `<task-id>`

Server request:
- `GET /api/v1/projects/{projectID}/tasks/{taskID}/comments`

### `ptodo comment update`

Required input:
- `<comment-id>`
- `-c <new-content>`

Server request:
- `PATCH /api/v1/projects/{projectID}/comments/{commentID}`

### `ptodo comment delete`

Required input:
- `<comment-id>`

Server request:
- `DELETE /api/v1/projects/{projectID}/comments/{commentID}`

## Dependencies

### `ptodo dep add`

Required input:
- `<task-id>`
- `<depends-on-id>`

Server request:
- `POST /api/v1/projects/{projectID}/tasks/{taskID}/dependencies`

Expected body fields:
- `depends_on_id`

Response:
- the created dependency edge

### `ptodo dep remove`

Required input:
- `<task-id>`
- `<depends-on-id>`

Server request:
- `DELETE /api/v1/projects/{projectID}/tasks/{taskID}/dependencies/{dependsOnID}`

Response:
- the removed dependency edge

### `ptodo dep list`

Required input:
- `<task-id>`

Server request:
- `GET /api/v1/projects/{projectID}/tasks/{taskID}/dependencies`

Response:
- a dependency edge list for the task

## Workflow

### `ptodo ready`

Optional input:
- `--epic <epic-id>`

Server request:
- `GET /api/v1/projects/{projectID}/ready`

Contract:
- return top-level tasks only
- return only tasks in `todo` status
- sort by priority, then creation order, then ID
- include a `dependents` list for any top-level `todo` tasks that would become ready if the item completed
- if `--epic` is provided, scope both the ready list and the dependents hints to that epic

Response:
- a project-scoped ready list with task rows and nested `dependents` hints

## Config

### `ptodo config list`

Server request:
- `GET /api/v1/projects/{projectID}/config`

Response:
- list of `{key, value}` rows

### `ptodo config get`

Required input:
- `<key>`

Server request:
- `GET /api/v1/projects/{projectID}/config/{key}`

Response:
- the `{key, value}` row for the requested project config entry

### `ptodo config set`

Required input:
- `<key>`
- `<value>`

Server request:
- `PUT /api/v1/projects/{projectID}/config/{key}`

Expected body fields:
- `value`

Response:
- the stored `{key, value}` row for the project config entry

### `ptodo config unset`

Required input:
- `<key>`

Server request:
- `DELETE /api/v1/projects/{projectID}/config/{key}`

Response:
- the removed `{key, value}` row for the deleted project config entry

## Search, History, List

### `ptodo search`

Required input:
- `<query>`

Optional input:
- `--type`
- `--status`
- `--limit`

Server request:
- `GET /api/v1/projects/{projectID}/search`

### `ptodo history`

Optional input:
- `--entity`
- `--type`
- `--action`
- `--since`
- `--limit`

Server request:
- `GET /api/v1/projects/{projectID}/history`

### `ptodo list`

Optional input:
- `--type`
- `--status`
- `--priority`
- `--sort`
- `--limit`

Server request:
- `GET /api/v1/projects/{projectID}/list`

## Locks

These commands expose the `work_item_locks` table directly.

### `ptodo lock acquire`

Required input:
- `<entity-type>`
- `<entity-id>`

Optional input:
- `--reason`
- `--expires`

Server request:
- `POST /api/v1/projects/{projectID}/locks`

Expected body fields:
- `entity_type`
- `entity_id`
- `reason`
- `expires_at` or `ttl`
- returns the acquired lock row with server-generated `lock_id`, `leased_at`, and `expires_at`

### `ptodo lock release`

Required input:
- `<lock-id>`

Server request:
- `DELETE /api/v1/projects/{projectID}/locks/{lockID}`
- returns the released lock row; releasing an already released lock should be a no-op

### `ptodo lock list`

Optional filters:
- `--type`
- `--entity`
- `--active`

Server request:
- `GET /api/v1/projects/{projectID}/locks`
- returns a lock list with optional filtering by entity type, entity ID, and active status

## Output Contract

- list commands return multiple records
- show/get commands return a single record
- `--toon` should serialize responses as TOON objects or arrays with stable field order
- write commands should return the created or updated resource plus any server-generated metadata

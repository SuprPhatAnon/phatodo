# Dataflow

This document maps the exact CLI-to-API-to-database path `ptodo` uses today and the path we intend to preserve as more commands are implemented.

## Source Files

Current wiring lives in these paths:

- CLI entrypoint: `cmd/ptodo/main.go`
- CLI command routing: `internal/cli/commands.go`
- CLI HTTP client: `internal/cli/api.go`
- Local config file layout: `internal/config/local.go`
- Server entrypoint: `cmd/phatodo-server/main.go`
- Server route registration: `internal/server/app.go`
- Server request handlers: `internal/server/handlers.go`
- Server auth middleware: `internal/server/auth.go`
- Server config repository: `internal/storage/postgres/project_config.go`
- Shared model types: `internal/domain/types.go`
- Database schema: `migrations/0001_initial.sql`

## Local Config Path

The CLI writes and reads a single local config file:

- Directory: `<repo>/.phatodo/`
- File: `<repo>/.phatodo/config.json`

The JSON fields are:

- `api_url`
- `workspace_id`
- `project_id`
- `access_key`
- `access_secret`

Default values are defined in `internal/config/local.go`.

`workspace_id` is stored locally for future workspace-scoped flows and for the broader server tenancy model. The current `config list` request path uses `project_id` directly.

## Environment Variables

The server entrypoint reads:

- `PHATODO_ADDR`
- `PHATODO_DATABASE_URL`

The Makefile also pins build/runtime helpers:

- `GO`
- `GOPATH`
- `GOPRIVATE`
- `GOCACHE`
- `GOMODCACHE`

## Current End-to-End Flow: `config list`, `config get`, `config set`, and `config unset`

The first wired command paths are the config subcommands.

### 1. Command entry

- `cmd/ptodo/main.go` calls `cli.Run`.
- `internal/cli/commands.go` recognizes `config list`, `config get`, `config set`, and `config unset`.

### 2. Local config load

- `internal/cli/commands.go` reads `<repo>/.phatodo/config.json` via `internal/config.ReadLocal`.
- The CLI uses `api_url`, `project_id`, `access_key`, and `access_secret` from that file.

### 3. HTTP request build

- `internal/cli/api.go` creates the request.
- Method: `GET`
- URL: `{{api_url}}/api/v1/projects/{projectID}/config`
- Headers:
  - `X-Phatodo-Access-Key: <access_key>`
  - `X-Phatodo-Access-Secret: <access_secret>`

### 4. Server route handling

- `cmd/phatodo-server/main.go` constructs `server.Config`.
- If `PHATODO_DATABASE_URL` is set, it creates a `ProjectConfigStore` with `internal/storage/postgres.NewProjectConfigStore`.
- `internal/server/app.go` routes `GET /api/v1/projects/{projectID}/config` to `listProjectConfig`.
- `internal/server/auth.go` requires the access key and access secret headers before the handler runs.

### 5. Repository query

- `internal/server/handlers.go` calls `ProjectConfigReader.ListProjectConfig`.
- `internal/storage/postgres/project_config.go` queries the `project_config` table:

```sql
SELECT key, value
FROM project_config
WHERE project_id = $1
ORDER BY key
```

### 6. Response shape

The server responds with:

```json
{
  "project_id": "default",
  "items": [
    { "key": "theme", "value": "dark" }
  ]
}
```

The CLI prints that as:

```text
- theme: dark
```

For `ptodo config set`, the server responds with the stored `{key, value}` row and the CLI prints the same `- key: value` line for confirmation.

For `ptodo config get`, the server responds with the requested `{key, value}` row and the CLI prints `- key: value`.

For `ptodo config unset`, the server responds with the removed `{key, value}` row and the CLI prints `- key: value` for confirmation.

## Current End-to-End Flow: `task create`

### 1. Command entry

- `cmd/ptodo/main.go` calls `cli.Run`.
- `internal/cli/commands.go` recognizes `task create`.

### 2. Local config load

- `internal/cli/commands.go` reads `<repo>/.phatodo/config.json` via `internal/config.ReadLocal`.
- The CLI uses `api_url`, `project_id`, `access_key`, and `access_secret` from that file.

### 3. HTTP request build

- `internal/cli/api.go` creates the request.
- Method: `POST`
- URL: `{{api_url}}/api/v1/projects/{projectID}/tasks`
- Body fields:
  - `title`
  - `issue_prefix`
  - `description`
  - `priority`
  - `epic_id`
  - `tags`
  - `assigned_to`
  - `acceptance_criteria`
- Headers:
  - `X-Phatodo-Access-Key: <access_key>`
  - `X-Phatodo-Access-Secret: <access_secret>`

### 4. Server route handling

- `cmd/phatodo-server/main.go` constructs `server.Config`.
- If `PHATODO_DATABASE_URL` is set, it creates the Postgres store and wires it as the task creator.
- `internal/server/app.go` routes `POST /api/v1/projects/{projectID}/tasks` to `createTask`.
- `internal/server/auth.go` requires the access key and access secret headers before the handler runs.

### 5. Repository query

- `internal/server/handlers.go` calls `TaskCreator.CreateTask`.
- `internal/storage/postgres/tasks.go` allocates a per-project task counter in `id_counters`, generates the task ID from the supplied issue prefix, and inserts the row into `tasks`.

### 6. Response shape

The server responds with:

```json
{
  "id": "ABC-1",
  "issue_prefix": "ABC",
  "title": "Write docs",
  "status": "todo",
  "priority": 2,
  "project_id": "default",
  "workspace_id": "default"
}
```

The CLI prints that as:

```text
- id: ABC-1
  issue_prefix: ABC
  title: "Write docs"
  status: todo
  priority: 2
  project_id: default
  workspace_id: default
```

## Current End-to-End Flow: `task list`

### 1. Command entry

- `cmd/ptodo/main.go` calls `cli.Run`.
- `internal/cli/commands.go` recognizes `task list`.

### 2. Local config load

- `internal/cli/commands.go` reads `<repo>/.phatodo/config.json` via `internal/config.ReadLocal`.
- The CLI uses `api_url`, `project_id`, `access_key`, and `access_secret` from that file.

### 3. HTTP request build

- `internal/cli/api.go` creates the request.
- Method: `GET`
- URL: `{{api_url}}/api/v1/projects/{projectID}/tasks`
- Query params:
  - `status` if a status filter is provided
  - `epic` if an epic filter is provided
- Headers:
  - `X-Phatodo-Access-Key: <access_key>`
  - `X-Phatodo-Access-Secret: <access_secret>`

### 4. Server route handling

- `cmd/phatodo-server/main.go` constructs `server.Config`.
- If `PHATODO_DATABASE_URL` is set, it creates the Postgres store and wires it as the task lister.
- `internal/server/app.go` routes `GET /api/v1/projects/{projectID}/tasks` to `listTasks`.
- `internal/server/auth.go` requires the access key and access secret headers before the handler runs.

### 5. Repository query

- `internal/server/handlers.go` calls `TaskLister.ListTasks`.
- `internal/storage/postgres/tasks.go` validates the project, queries the `tasks` table for top-level rows, applies optional status and epic filters, and orders by priority, creation time, and ID.

### 6. Response shape

The server responds with:

```json
{
  "project_id": "default",
  "items": [
    {
      "id": "ABC-1",
      "title": "Write docs",
      "status": "in_progress",
      "priority": 2,
      "epic_id": "epic-1"
    }
  ]
}
```

The CLI prints each item on its own line as:

```text
tasks[1]:
  - id: ABC-1
    title: "Write docs"
    description: ""
    priority: 2
    status: in_progress
    epicId: epic-1
```

## Current End-to-End Flow: `ready`

### 1. Command entry

- `cmd/ptodo/main.go` calls `cli.Run`.
- `internal/cli/commands.go` recognizes `ready`.

### 2. Local config load

- `internal/cli/commands.go` reads `<repo>/.phatodo/config.json` via `internal/config.ReadLocal`.
- The CLI uses `api_url`, `project_id`, `access_key`, and `access_secret` from that file.

### 3. HTTP request build

- `internal/cli/api.go` creates the request.
- Method: `GET`
- URL: `{{api_url}}/api/v1/projects/{projectID}/ready`
- Query params:
  - `epic` if an epic filter is provided
- Headers:
  - `X-Phatodo-Access-Key: <access_key>`
  - `X-Phatodo-Access-Secret: <access_secret>`

### 4. Server route handling

- `cmd/phatodo-server/main.go` constructs `server.Config`.
- If `PHATODO_DATABASE_URL` is set, it creates the Postgres store and wires it as the ready lister.
- `internal/server/app.go` routes `GET /api/v1/projects/{projectID}/ready` to `listReadyTasks`.
- `internal/server/auth.go` requires the access key and access secret headers before the handler runs.

### 5. Repository query

- `internal/server/handlers.go` calls `ReadyLister.ListReadyTasks`.
- `internal/storage/postgres/tasks.go` finds top-level `todo` tasks with satisfied dependencies, then finds top-level `todo` tasks that would become ready if each row completed.

### 6. Response shape

The server responds with:

```json
{
  "project_id": "default",
  "items": [
    {
      "id": "CORE-1",
      "title": "Health endpoints and k8s liveness/readiness probes for worker services",
      "status": "todo",
      "priority": 1,
      "epic_id": "epic-1",
      "tags": ["infra", "api"],
      "unblocks": [
        {
          "id": "CORE-5",
          "title": "Two-tier automated database backups via SCP to NAS and GCS",
          "status": "todo",
          "priority": 1,
          "epic_id": "epic-1",
          "tags": ["infra"]
        }
      ]
    }
  ]
}
```

The CLI prints ready work as a TOON array of objects, with nested `tags` fields and a compact `dependents` array when present.

## Bootstrap Flow

The bootstrap path is split into two steps:

1. `ptodo admin init`
   - CLI prompts for the new admin password twice.
   - CLI sends the admin username and API server URL to `POST /api/v1/admin/init`.
   - The server creates the first admin user only.

2. `ptodo admin bootstrap`
   - CLI prompts for the admin password.
   - CLI sends the admin username and API server URL to `POST /api/v1/admin/bootstrap`.
   - The server authenticates the admin, creates the bootstrap workspace/project and the project-scoped CLI identity, creates the matching `user_project_access` row, and returns the local config payload.
   - CLI writes the returned `api_url`, `workspace_id`, `project_id`, `access_key`, and `access_secret` into `<repo>/.phatodo/config.json`.
   - CLI prints the returned `workspace_id`, `project_id`, `access_key`, `access_secret`, and `config_path` as TOON fields.

These bootstrap routes are the exception to the normal access-key header flow documented above.

This bootstrap flow must fail if the target project already exists.

## Planned End-State Flow

The same pattern should hold for future commands:

1. CLI command parses a Trekker-compatible shape.
2. CLI loads `<repo>/.phatodo/config.json`.
3. CLI sends a request to `/api/v1/projects/{projectID}/...`.
4. Server authenticates the caller.
5. Server validates project access and lifecycle rules.
6. Server calls a Postgres repository.
7. Repository reads or writes `migrations/0001_initial.sql` tables.
8. Server returns structured JSON.
9. CLI renders compact terminal output.

## Database Paths In Play

The current flow uses these tables directly:

- `users`
- `user_project_access`
- `projects`
- `project_config`

The next layers will add:

- `epics`
- remaining task lifecycle operations beyond create/list
- `comments`
- `dependencies`
- `work_item_locks`
- `search_index`

## What Is Not Wired Yet

These pieces are still planned, not implemented:

- `epic` command backends
- `lock` command backends
- project creation and access-management flows
- lock management

## Canonical References

- Command shape: `docs/trekker_reference.txt`
- API surface: `docs/API.md`
- Layer responsibilities: `docs/ARCHITECTURE.md`
- Implementation sequence: `docs/IMPLEMENTATION_PLAN.md`
- Table layout: `docs/DATABASE_SCHEMA.md`

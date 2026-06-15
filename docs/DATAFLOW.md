# Dataflow

This document maps the exact CLI-to-API-to-database path phatodo uses today and the path we intend to preserve as more commands are implemented.

## Source Files

Current wiring lives in these paths:

- CLI entrypoints: `cmd/phatodo/main.go`, `cmd/ptd/main.go`, `cmd/ptodo/main.go`
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

## Current End-to-End Flow: `config list`

The first wired command path is `phatodo config list`, `ptd config list`, or `ptodo config list`.

### 1. Command entry

- `cmd/phatodo/main.go`, `cmd/ptd/main.go`, and `cmd/ptodo/main.go` call `cli.Run`.
- `internal/cli/commands.go` recognizes `config list`.

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
    { "key": "issue_prefix", "value": "ABC" }
  ]
}
```

The CLI prints that as:

```text
issue_prefix=ABC
```

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
- `tasks`
- `comments`
- `dependencies`
- `events`
- `work_item_locks`
- `search_index`

## What Is Not Wired Yet

These pieces are still planned, not implemented:

- `config get`, `config set`, and `config unset`
- `epic`, `task`, `subtask`, `comment`, `dep`, `search`, `history`, and `list` command backends
- project creation and access-management flows
- write-path audit history
- lock management

## Canonical References

- Command shape: `docs/trekker_reference.txt`
- API surface: `docs/API.md`
- Layer responsibilities: `docs/ARCHITECTURE.md`
- Implementation sequence: `docs/IMPLEMENTATION_PLAN.md`
- Table layout: `docs/DATABASE_SCHEMA.md`

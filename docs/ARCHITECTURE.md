# Architecture

Phatodo has two deliverables:

1. Trekker-compatible command-line executables named `phatodo` and `ptd`.
2. A server-side API and dashboard backed by Postgres.

## Components

- `cmd/phatodo` is the primary CLI entrypoint. It should preserve the command structure documented in `docs/trekker_reference.txt`.
- `cmd/ptd` is the short CLI alias entrypoint.
- `cmd/phatodo-server` is the API and dashboard server entrypoint.
- `internal/cli` owns command parsing, output modes such as `--toon`, and API client wiring.
- `internal/server` owns HTTP routing for JSON APIs and dashboard delivery.
- `internal/domain` contains shared entity names, statuses, priorities, and validation rules.
- `internal/storage/postgres` will contain Postgres repositories.
- `migrations/` contains database schema migrations.
- `web/dashboard` is reserved for dashboard source assets.

## Runtime Model

The CLI is a client. It should read local configuration for workspace ID, project ID, API URL, and auth token, then call the central API. The server is the state authority for tasks, comments, dependencies, locks, search, history, and project configuration.

Postgres stores canonical state and audit history. Local `.phatodo` data should be limited to client configuration and optional cache data.

The database model is documented in `docs/DATABASE_SCHEMA.md`. It follows Trekker's SQLite schema while adding a `workspaces` layer above individual projects and accountability fields for ownership, completion evidence, audit history, and time-bound work locks.

Authentication is server-side. Users authenticate with an access key and access secret for CLI/API calls, with optional username/password credentials for dashboard login. Admin users can access all projects; regular users are limited to one project through `user_project_access`.

## API Scope

The API should expose resources for epics, tasks, subtasks, comments, dependencies, config, search, history, and unified list views. Server-side validation should enforce lifecycle rules, including requiring a summary comment before task completion and checking completion evidence against explicit acceptance criteria.

The non-dashboard API scaffold is documented in `docs/API.md`. It uses `/api/v1` routes, access-key headers, and project-scoped resource paths.

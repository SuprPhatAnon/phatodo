# Architecture

Phatodo has two deliverables:

1. Trekker-compatible command-line executable named `ptodo`.
2. A server-side API and dashboard backed by Postgres.

## Components

- `cmd/ptodo` is the CLI entrypoint. It should preserve the command structure documented in `docs/trekker_reference.txt`.
- `cmd/phatodo-server` is the API and dashboard server entrypoint.
- `internal/cli` owns command parsing, output modes such as `--toon`, TOON serialization, and API client wiring.
- `internal/server` owns HTTP routing for JSON APIs and dashboard delivery.
- `internal/domain` contains shared entity names, statuses, priorities, and validation rules.
- `internal/storage/postgres` will contain Postgres repositories.
- `migrations/` contains database schema migrations.
- `web/dashboard` is reserved for dashboard source assets.

## Runtime Model

The CLI is a client. It should read local configuration for workspace ID, project ID, API URL, access key, and access secret, then call the central API. The server is the state authority for tasks, comments, dependencies, locks, search, history, and project configuration.

Postgres stores canonical state and audit history. Local `.phatodo` data should be limited to client configuration and optional cache data.

The database model is documented in `docs/DATABASE_SCHEMA.md`. It follows Trekker's SQLite schema while adding a `workspaces` layer above individual projects and accountability fields for ownership, completion evidence, audit history, and time-bound work locks.

All Postgres access should be defined in `sqlc` query files under `internal/storage/postgres/queries/` and regenerated through `make sqlc`; handwritten SQL in Go is not the intended steady-state.

Authentication is server-side. Users authenticate with an access key and access secret for CLI/API calls, with optional username/password credentials for dashboard login. Admin users can access all projects; regular users are limited to one project through `user_project_access`.

## API Scope

The API should expose resources for epics, tasks, subtasks, comments, dependencies, config, search, history, and unified list views. Server-side validation should enforce lifecycle rules, including requiring a summary comment before task completion and checking completion evidence against explicit acceptance criteria.

The non-dashboard API scaffold is documented in `docs/API.md`. It uses `/api/v1` routes, access-key headers, and project-scoped resource paths.

The build order and intended end state are documented in `docs/IMPLEMENTATION_PLAN.md`.

The concrete CLI-to-API-to-database request path is documented in `docs/DATAFLOW.md`.

The command-to-endpoint-to-table matrix is documented in `docs/COMMAND_MAP.md`.

The bootstrap, contract, validation, auth, error, operations, import, and output docs are:

- `docs/BOOTSTRAP.md`
- `docs/COMMAND_CONTRACTS.md`
- `docs/VALIDATION_RULES.md`
- `docs/AUTHORIZATION.md`
- `docs/ERRORS.md`
- `docs/OPERATIONS.md`
- `docs/IMPORT.md`
- `docs/CLI_OUTPUT.md`

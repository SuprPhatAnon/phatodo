# Architecture

`trakkr` has two deliverables:

1. A Trekker-compatible command-line executable named `trakkr`.
2. A server-side API and dashboard backed by Postgres.

## Components

- `cmd/trakkr` is the CLI entrypoint. It should preserve the command structure documented in `docs/trekker_reference.txt`.
- `cmd/trakkr-server` is the API and dashboard server entrypoint.
- `internal/cli` owns command parsing, output modes such as `--toon`, and API client wiring.
- `internal/server` owns HTTP routing for JSON APIs and dashboard delivery.
- `internal/domain` contains shared entity names, statuses, priorities, and validation rules.
- `internal/storage/postgres` will contain Postgres repositories.
- `migrations/` contains database schema migrations.
- `web/dashboard` is reserved for dashboard source assets.

## Runtime Model

The CLI is a client. It should read local configuration for project ID, API URL, and auth token, then call the central API. The server is the state authority for tasks, comments, dependencies, locks, search, history, and project configuration.

Postgres stores canonical state and audit history. Local `.trakkr` data should be limited to client configuration and optional cache data.

## API Scope

The API should expose resources for epics, tasks, subtasks, comments, dependencies, config, search, history, and unified list views. Server-side validation should enforce lifecycle rules, including requiring a summary comment before task completion.

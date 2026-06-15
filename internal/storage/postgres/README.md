# Postgres Storage

This package owns the Postgres-backed repository implementations for epics, tasks, subtasks, comments, dependencies, project config, search, and history.

Use `PHATODO_DATABASE_URL` for the connection string. Keep SQL migrations in `migrations/` and keep storage methods aligned with the domain terms in `internal/domain`.

Database access must go through `sqlc`:

- put query definitions in `internal/storage/postgres/queries/`
- run `make sqlc` to regenerate the Go query package
- prefer generated query methods over handwritten SQL in Go

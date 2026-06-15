# Postgres Storage

This package will own the Postgres-backed repository implementations for epics, tasks, subtasks, comments, dependencies, project config, search, and history.

Use `PHATODO_DATABASE_URL` for the connection string. Keep SQL migrations in `migrations/` and keep storage methods aligned with the domain terms in `internal/domain`.

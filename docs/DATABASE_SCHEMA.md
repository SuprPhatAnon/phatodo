# Database Schema

The Postgres schema in `migrations/0001_initial.sql` is based on the schema created by `trekker init` in `.trekker/trekker.db`, with one important expansion: a server can host multiple workspaces, and each workspace can contain multiple individual projects.

## Scope Model

- `workspaces` are the top-level tenancy boundary for a hosted server.
- `projects` represent individual tracked codebases or work areas inside a workspace.
- `users` are global server identities. `admin` users can access every workspace and project.
- `user_project_access` locks a regular `user` to one project.
- Project-owned tables carry both `workspace_id` and `project_id` so API queries can be scoped explicitly.
- Work items carry explicit accountability metadata, including assignment, creator, updater, completion owner, completion timestamps, acceptance criteria, completion evidence, and time-bound locks.

The CLI should store only client configuration locally, such as API URL, workspace ID, project ID, access key, and access secret. Canonical task data belongs in Postgres.

## Authentication & Access

Every user has an `access_key` and `access_secret_hash` for CLI/API authentication. Store only the secret hash in Postgres; show the raw secret only once when generated.

`username` and `password_hash` are optional for dashboard login. Password rows require a username, but API-only users can omit both fields.

Authorization rules should be enforced by the API:

- `role = 'admin'` can see and manage all projects.
- `role = 'user'` must have exactly one `user_project_access` row and can only see that project.

## Trekker Table Mapping

- `users` and `user_project_access` are phatodo additions for centralized auth.
- `projects` maps to Trekker projects, now scoped under `workspaces`.
- `project_config` maps Trekker config keys, now scoped per project instead of globally.
- `id_counters` maps Trekker ID counters, now scoped per project.
- `epics` maps directly to Trekker epics.
- `tasks` maps Trekker tasks and subtasks. A row with `parent_task_id IS NULL` is a task; a row with `parent_task_id` is a subtask. Tasks and epics now also carry accountability metadata for ownership, completion, and validation.
- `work_item_locks` tracks time-bound leases on tasks, subtasks, and epics.
- `comments` remains task-scoped, matching Trekker's current comment model, but adds a comment kind and optional user identity for checkpoint and summary tracking.
- `dependencies` matches Trekker task dependencies and prevents duplicates per project.
- `events` replaces Trekker's local event log with project-scoped audit history and stores actor, before-state, after-state, and metadata details.
- `search_index` replaces SQLite FTS5 with a Postgres `tsvector` search table.

## Compatibility Notes

IDs remain `TEXT` so imported Trekker IDs can be preserved. Status and priority constraints follow `docs/trekker_reference.txt`: priorities are `0` through `5`; epics support `todo`, `in_progress`, `completed`, and `archived`; tasks also support `wont_fix`.

Subtasks are stored in `tasks`, not a separate table, because that is how Trekker models them today. This keeps migration/import simpler and preserves unified task search, history, dependency, and comment behavior.

Completion-related fields are intentionally explicit so the server can enforce accountability: tasks and epics can record who owns work, who completed it, what evidence was used, and which summary or checkpoint comment was left behind.

The implementation sequence for wiring the CLI, API, and database together is documented in `docs/IMPLEMENTATION_PLAN.md`.

The schema-backed CLI extensions that expose these fields are listed in `docs/CLI_COMPATIBILITY.md`, and their endpoint/table mapping is documented in `docs/COMMAND_MAP.md`.

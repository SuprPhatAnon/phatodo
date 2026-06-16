# API

The non-dashboard API is rooted at `/api/v1`. `/api/health` is public. Most `/api/v1` routes require access-key authentication, but `POST /api/v1/admin/init` and `POST /api/v1/admin/bootstrap` are bootstrap routes and use the bootstrap credential flow instead.

## Authentication

Send CLI/API credentials with each request:

```http
X-Phatodo-Access-Key: <access-key>
X-Phatodo-Access-Secret: <access-secret>
```

The current scaffold validates that both headers are present. The storage-backed implementation will verify the key and secret hash against `users`, resolve the user role, and enforce project access through `user_project_access`.

## Route Shape

Project-scoped resources use this prefix:

```text
/api/v1/projects/{projectID}
```

Initial route groups:

- `POST /api/v1/admin/init`
- `POST /api/v1/admin/bootstrap`
- `GET|POST /api/v1/projects`
- `GET|PATCH|DELETE /api/v1/projects/{projectID}`
- `/epics` for epic list/create/show/update/complete/delete
- `/tasks` for task list/create/show/update/delete; `GET /tasks` returns top-level tasks and supports `status` and `epic` filters, while task create accepts `issue_prefix` so the server can generate the task ID from the command input
- `/tasks/{taskID}/subtasks` for subtask list/create, using the parent task to scope child rows and derive the child issue prefix
- `/tasks/{taskID}/comments` and `/comments/{commentID}`
- `/tasks/{taskID}/dependencies`
- `/config`, `/search`, `/history`, `/list`, and `/ready`
- `/locks` for work-item lease acquire/release/list

Most handlers now return structured JSON from the Postgres-backed store. The config routes, admin bootstrap routes, epic routes, task create/list/show/update/delete routes, subtask create/list routes, comment routes, dependency routes, lock routes, and the ready/search/history/list routes are wired end to end. This lets the CLI and tests stabilize around URL shape while any remaining command surface is filled in.

Resource payloads will include accountability fields from the schema, including assignment, creator, updater, completion owner, planned files, changed files, acceptance criteria, completion evidence, completion timestamps, task kind, root-cause analysis, comment kind, and audit metadata.

The config route returns a project-scoped list of `{key, value}` entries for `config list`, the ready route returns top-level todo tasks ordered by priority plus any tasks they would unblock, and the lock route family exposes time-bound leases on epics, tasks, and subtasks.

The broader route and data rollout plan is described in `docs/IMPLEMENTATION_PLAN.md`.

The concrete request/handler/storage path for this route is documented in `docs/DATAFLOW.md`.

The full command-to-endpoint-to-table map is documented in `docs/COMMAND_MAP.md`.

The bootstrap, contract, validation, auth, error, operations, import, and output docs provide the remaining no-dashboard operating rules.

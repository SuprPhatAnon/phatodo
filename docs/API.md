# API

The non-dashboard API is rooted at `/api/v1`. `/api/health` is public; all `/api/v1` routes require access-key authentication.

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

- `GET|POST /api/v1/projects`
- `GET|PATCH|DELETE /api/v1/projects/{projectID}`
- `/epics` for epic list/create/show/update/complete/delete
- `/tasks` for task list/create/show/update/delete
- `/tasks/{taskID}/subtasks` for subtask list/create
- `/tasks/{taskID}/comments` and `/comments/{commentID}`
- `/tasks/{taskID}/dependencies`
- `/config`, `/search`, `/history`, and `/list`

Handlers currently return structured `501 not_implemented` responses. This lets the CLI and tests stabilize around URL shape before Postgres repositories are wired in.

Resource payloads will include accountability fields from the schema, including assignment, creator, updater, completion owner, acceptance criteria, completion evidence, completion timestamps, comment kind, and audit metadata.

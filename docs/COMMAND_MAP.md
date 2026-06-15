# Command Map

This document maps the `ptodo` command family to its API endpoints and the database tables they primarily read or write.

## Implementation Checklist

- [x] `ptodo admin init`
- [x] `ptodo admin bootstrap`
- [x] `ptodo config list`
- [x] `ptodo config get`
- [x] `ptodo config set`
- [x] `ptodo config unset`
- [ ] `ptodo epic create`
- [ ] `ptodo epic list`
- [ ] `ptodo epic show`
- [ ] `ptodo epic update`
- [ ] `ptodo epic complete`
- [ ] `ptodo epic delete`
- [x] `ptodo task create`
- [x] `ptodo task list`
- [x] `ptodo task show`
- [x] `ptodo task update`
- [x] `ptodo task delete`
- [x] `ptodo subtask create`
- [x] `ptodo subtask list`
- [x] `ptodo subtask update`
- [x] `ptodo subtask delete`
- [x] `ptodo comment add`
- [x] `ptodo comment list`
- [x] `ptodo comment update`
- [x] `ptodo comment delete`
- [x] `ptodo dep add`
- [x] `ptodo dep remove`
- [x] `ptodo dep list`
- [x] `ptodo ready`
- [ ] `ptodo lock acquire`
- [ ] `ptodo lock release`
- [ ] `ptodo lock list`
- [x] `ptodo search`
- [x] `ptodo history`
- [x] `ptodo list`

## Conventions

- `primary tables` are the rows the command directly acts on.
- `adjunct tables` are updated as part of the same operation for audit history, search indexing, or cascaded cleanup.
- Subtasks are stored in `tasks`, so subtask commands reuse task endpoints with a subtask ID.

## Setup

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo admin init` | `POST /api/v1/admin/init` | `users` | Creates the first admin user only; fails if any admin user already exists. |
| `ptodo admin bootstrap` | `POST /api/v1/admin/bootstrap` | `workspaces`, `projects`, `users`, `user_project_access` | Creates the bootstrap workspace/project, the project-scoped CLI identity, and writes `<repo>/.phatodo/config.json`. Fails if the target project already exists. |
| `ptodo wipe -y` | none | none | Removes local client state only. |

## Epics

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo epic create` | `POST /api/v1/projects/{projectID}/epics` | `epics` | `events`, `search_index`, `id_counters` if ID generation is server-side. |
| `ptodo epic list` | `GET /api/v1/projects/{projectID}/epics` | `epics` | May read `search_index` for filtering/sorting. |
| `ptodo epic show` | `GET /api/v1/projects/{projectID}/epics/{epicID}` | `epics` | None. |
| `ptodo epic update` | `PATCH /api/v1/projects/{projectID}/epics/{epicID}` | `epics` | `events`, `search_index`. |
| `ptodo epic complete` | `POST /api/v1/projects/{projectID}/epics/{epicID}/complete` | `epics`, `tasks` | `events`, `search_index`; archives child tasks as part of the operation. |
| `ptodo epic delete` | `DELETE /api/v1/projects/{projectID}/epics/{epicID}` | `epics` | `tasks` if epic links are cleared, `events`, `search_index`. |

## Tasks

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo task create` | `POST /api/v1/projects/{projectID}/tasks` | `tasks` | `events`, `search_index`, `id_counters` if ID generation is server-side; `issue_prefix` must be supplied in the create payload. |
| `ptodo task list` | `GET /api/v1/projects/{projectID}/tasks` | `tasks` | None. Returns top-level tasks unless a subtask-specific command is used. |
| `ptodo task show` | `GET /api/v1/projects/{projectID}/tasks/{taskID}` | `tasks` | None. |
| `ptodo task update` | `PATCH /api/v1/projects/{projectID}/tasks/{taskID}` | `tasks` | `events`, `search_index`. |
| `ptodo task delete` | `DELETE /api/v1/projects/{projectID}/tasks/{taskID}` | `tasks` | `comments`, `dependencies`, `events`, `search_index`, and any active `work_item_locks` cleanup. |

## Subtasks

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo subtask create` | `POST /api/v1/projects/{projectID}/tasks/{taskID}/subtasks` | `tasks` | `events`, `search_index`, `id_counters` if ID generation is server-side; the server derives the subtask issue prefix from the parent task. |
| `ptodo subtask list` | `GET /api/v1/projects/{projectID}/tasks/{taskID}/subtasks` | `tasks` | None. Returns direct child subtasks for the parent task. |
| `ptodo subtask update` | `PATCH /api/v1/projects/{projectID}/tasks/{subtaskID}` | `tasks` | `events`, `search_index`. |
| `ptodo subtask delete` | `DELETE /api/v1/projects/{projectID}/tasks/{subtaskID}` | `tasks` | `comments`, `dependencies`, `events`, `search_index`, and any active `work_item_locks` cleanup. |

## Comments

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo comment add` | `POST /api/v1/projects/{projectID}/tasks/{taskID}/comments` | `comments` | `events`, `search_index`. |
| `ptodo comment list` | `GET /api/v1/projects/{projectID}/tasks/{taskID}/comments` | `comments` | None. |
| `ptodo comment update` | `PATCH /api/v1/projects/{projectID}/comments/{commentID}` | `comments` | `events`, `search_index`. |
| `ptodo comment delete` | `DELETE /api/v1/projects/{projectID}/comments/{commentID}` | `comments` | `events`, `search_index`. |

## Dependencies

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo dep add` | `POST /api/v1/projects/{projectID}/tasks/{taskID}/dependencies` | `dependencies` | `events`. |
| `ptodo dep remove` | `DELETE /api/v1/projects/{projectID}/tasks/{taskID}/dependencies/{dependsOnID}` | `dependencies` | `events`. |
| `ptodo dep list` | `GET /api/v1/projects/{projectID}/tasks/{taskID}/dependencies` | `dependencies` | None. Returns direct dependency edges for the task. |

## Workflow

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo ready` | `GET /api/v1/projects/{projectID}/ready` | `tasks`, `dependencies` | Returns top-level `todo` tasks ordered by priority and includes any top-level tasks that would become ready if the row completed. May read dependency state from `tasks` for status checks. |

## Config

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo config list` | `GET /api/v1/projects/{projectID}/config` | `project_config` | None. |
| `ptodo config get` | `GET /api/v1/projects/{projectID}/config/{key}` | `project_config` | None. |
| `ptodo config set` | `PUT /api/v1/projects/{projectID}/config/{key}` | `project_config` | `events` for audit history. |
| `ptodo config unset` | `DELETE /api/v1/projects/{projectID}/config/{key}` | `project_config` | `events` for audit history. |

## Search, History, List

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo search` | `GET /api/v1/projects/{projectID}/search` | `search_index` | Joins or lookups against `epics`, `tasks`, `comments`, and `dependencies` may be needed to render full results. |
| `ptodo history` | `GET /api/v1/projects/{projectID}/history` | `events` | May join `users` for actor display. |
| `ptodo list` | `GET /api/v1/projects/{projectID}/list` | `epics`, `tasks` | May use `search_index` for unified filtering and `comments` for summaries. |

## Locks

These commands are schema-driven additions so the `work_item_locks` table is reachable from the CLI.

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `ptodo lock acquire` | `POST /api/v1/projects/{projectID}/locks` | `work_item_locks` | `events`. |
| `ptodo lock release` | `DELETE /api/v1/projects/{projectID}/locks/{lockID}` | `work_item_locks` | `events`. |
| `ptodo lock list` | `GET /api/v1/projects/{projectID}/locks` | `work_item_locks` | None. |

## Schema Notes

- `users` is accessed for auth and actor resolution, not through the Trekker command surface directly.
- `user_project_access` is enforced by the server.
- `projects` is an administrative/server-side concept, not a Trekker command today.
- `events` should receive an immutable row for meaningful writes.
- `search_index` should be kept in sync on any write that affects search results.

# Command Map

This document maps the phatodo command family (`phatodo`, `ptd`, and `ptodo`) to its API endpoints and the database tables they primarily read or write.

## Implementation Checklist

- [x] `phatodo config list`
- [ ] `phatodo config get`
- [ ] `phatodo config set`
- [ ] `phatodo config unset`
- [ ] `phatodo epic create`
- [ ] `phatodo epic list`
- [ ] `phatodo epic show`
- [ ] `phatodo epic update`
- [ ] `phatodo epic complete`
- [ ] `phatodo epic delete`
- [ ] `phatodo task create`
- [ ] `phatodo task list`
- [ ] `phatodo task show`
- [ ] `phatodo task update`
- [ ] `phatodo task delete`
- [ ] `phatodo subtask create`
- [ ] `phatodo subtask list`
- [ ] `phatodo subtask update`
- [ ] `phatodo subtask delete`
- [ ] `phatodo comment add`
- [ ] `phatodo comment list`
- [ ] `phatodo comment update`
- [ ] `phatodo comment delete`
- [ ] `phatodo dep add`
- [ ] `phatodo dep remove`
- [ ] `phatodo dep list`
- [ ] `phatodo lock acquire`
- [ ] `phatodo lock release`
- [ ] `phatodo lock list`
- [ ] `phatodo search`
- [ ] `phatodo history`
- [ ] `phatodo list`

## Conventions

- `primary tables` are the rows the command directly acts on.
- `adjunct tables` are updated as part of the same operation for audit history, search indexing, or cascaded cleanup.
- Subtasks are stored in `tasks`, so subtask commands reuse task endpoints with a subtask ID.

## Setup

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo init` | none | none | Writes `<repo>/.phatodo/config.json` locally. |
| `phatodo wipe -y` | none | none | Removes local client state only. |

## Epics

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo epic create` | `POST /api/v1/projects/{projectID}/epics` | `epics` | `events`, `search_index`, `id_counters` if ID generation is server-side. |
| `phatodo epic list` | `GET /api/v1/projects/{projectID}/epics` | `epics` | May read `search_index` for filtering/sorting. |
| `phatodo epic show` | `GET /api/v1/projects/{projectID}/epics/{epicID}` | `epics` | None. |
| `phatodo epic update` | `PATCH /api/v1/projects/{projectID}/epics/{epicID}` | `epics` | `events`, `search_index`. |
| `phatodo epic complete` | `POST /api/v1/projects/{projectID}/epics/{epicID}/complete` | `epics`, `tasks` | `events`, `search_index`; archives child tasks as part of the operation. |
| `phatodo epic delete` | `DELETE /api/v1/projects/{projectID}/epics/{epicID}` | `epics` | `tasks` if epic links are cleared, `events`, `search_index`. |

## Tasks

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo task create` | `POST /api/v1/projects/{projectID}/tasks` | `tasks` | `events`, `search_index`, `id_counters` if ID generation is server-side. |
| `phatodo task list` | `GET /api/v1/projects/{projectID}/tasks` | `tasks` | May read `search_index` for filters. |
| `phatodo task show` | `GET /api/v1/projects/{projectID}/tasks/{taskID}` | `tasks` | None. |
| `phatodo task update` | `PATCH /api/v1/projects/{projectID}/tasks/{taskID}` | `tasks` | `events`, `search_index`. |
| `phatodo task delete` | `DELETE /api/v1/projects/{projectID}/tasks/{taskID}` | `tasks` | `comments`, `dependencies`, `events`, `search_index`, and any active `work_item_locks` cleanup. |

## Subtasks

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo subtask create` | `POST /api/v1/projects/{projectID}/tasks/{taskID}/subtasks` | `tasks` | `events`, `search_index`, `id_counters` if ID generation is server-side. |
| `phatodo subtask list` | `GET /api/v1/projects/{projectID}/tasks/{taskID}/subtasks` | `tasks` | None. |
| `phatodo subtask update` | `PATCH /api/v1/projects/{projectID}/tasks/{subtaskID}` | `tasks` | `events`, `search_index`. |
| `phatodo subtask delete` | `DELETE /api/v1/projects/{projectID}/tasks/{subtaskID}` | `tasks` | `comments`, `dependencies`, `events`, `search_index`, and any active `work_item_locks` cleanup. |

## Comments

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo comment add` | `POST /api/v1/projects/{projectID}/tasks/{taskID}/comments` | `comments` | `events`, `search_index`. |
| `phatodo comment list` | `GET /api/v1/projects/{projectID}/tasks/{taskID}/comments` | `comments` | None. |
| `phatodo comment update` | `PATCH /api/v1/projects/{projectID}/comments/{commentID}` | `comments` | `events`, `search_index`. |
| `phatodo comment delete` | `DELETE /api/v1/projects/{projectID}/comments/{commentID}` | `comments` | `events`, `search_index`. |

## Dependencies

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo dep add` | `POST /api/v1/projects/{projectID}/tasks/{taskID}/dependencies` | `dependencies` | `events`. |
| `phatodo dep remove` | `DELETE /api/v1/projects/{projectID}/tasks/{taskID}/dependencies/{dependsOnID}` | `dependencies` | `events`. |
| `phatodo dep list` | `GET /api/v1/projects/{projectID}/tasks/{taskID}/dependencies` | `dependencies` | None. |

## Config

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo config list` | `GET /api/v1/projects/{projectID}/config` | `project_config` | None. |
| `phatodo config get` | `GET /api/v1/projects/{projectID}/config/{key}` | `project_config` | None. |
| `phatodo config set` | `PUT /api/v1/projects/{projectID}/config/{key}` | `project_config` | `events` for audit history. |
| `phatodo config unset` | `DELETE /api/v1/projects/{projectID}/config/{key}` | `project_config` | `events` for audit history. |

## Search, History, List

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo search` | `GET /api/v1/projects/{projectID}/search` | `search_index` | Joins or lookups against `epics`, `tasks`, `comments`, and `dependencies` may be needed to render full results. |
| `phatodo history` | `GET /api/v1/projects/{projectID}/history` | `events` | May join `users` for actor display. |
| `phatodo list` | `GET /api/v1/projects/{projectID}/list` | `epics`, `tasks` | May use `search_index` for unified filtering and `comments` for summaries. |

## Locks

These commands are schema-driven additions so the `work_item_locks` table is reachable from the CLI.

| Command | API endpoint | Primary tables | Adjunct tables / notes |
| --- | --- | --- | --- |
| `phatodo lock acquire` | `POST /api/v1/projects/{projectID}/locks` | `work_item_locks` | `events`. |
| `phatodo lock release` | `DELETE /api/v1/projects/{projectID}/locks/{lockID}` | `work_item_locks` | `events`. |
| `phatodo lock list` | `GET /api/v1/projects/{projectID}/locks` | `work_item_locks` | None. |

## Schema Notes

- `users` is accessed for auth and actor resolution, not through the Trekker command surface directly.
- `user_project_access` is enforced by the server.
- `projects` is an administrative/server-side concept, not a Trekker command today.
- `events` should receive an immutable row for meaningful writes.
- `search_index` should be kept in sync on any write that affects search results.

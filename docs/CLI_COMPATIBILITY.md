# CLI Compatibility

The `ptodo` executable should keep the command shape preserved in `docs/trekker_reference.txt` while moving storage to the server.

## Required Command Groups

- Setup: `admin init`, `admin bootstrap`, `wipe -y`
- Epics: `epic create`, `epic list`, `epic show`, `epic update`, `epic complete`, `epic delete`
- Tasks: `task create`, `task list`, `task show`, `task update`, `task delete`
- Subtasks: `subtask create`, `subtask list`, `subtask update`, `subtask delete`
- Comments: `comment add`, `comment list`, `comment update`, `comment delete`
- Dependencies: `dep add`, `dep remove`, `dep list`
- Workflow: `ready`
- Config: `config list`, `config get`, `config set`, `config unset`
- Query: `search`, `history`, `list`

## Behavior

Support `--toon` for compact agent output. Preserve status values, priority values, and task workflow rules from the Trekker reference. Prefer additive enhancements over breaking command changes.

`config list` is the first server-backed command and should print the project configuration returned from `/api/v1/projects/{projectID}/config`.
`ready` should print top-level todo tasks that are currently unblocked, ordered by priority, with optional `--epic` scoping and inline `unblocks` hints for tasks that become available next.

## Schema-Driven Extensions

The Trekker-compatible command surface is extended to cover schema-backed fields that are not explicit in the original reference:

- `-a/--assigned-to` for task and epic ownership
- `--criteria-json` for acceptance criteria
- `--summary` and `--evidence-json` for completion metadata
- `-k/--kind` for comment type
- `lock acquire`, `lock release`, and `lock list` for `work_item_locks`

## Client Configuration

`ptodo admin bootstrap` should write local project configuration and the project-scoped CLI identity:

- workspace ID
- project ID
- API base URL
- access key
- access secret
- optional cache metadata

The local file lives at `.phatodo/config.json` and must not be committed. It should be created with mode `0600` because it contains the access secret.

Canonical task data belongs on the server.

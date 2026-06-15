# CLI Compatibility

The `phatodo` and `ptd` executables should keep the Trekker command shape from `docs/trekker_reference.txt` while moving storage to the server.

## Required Command Groups

- Setup: `init`, `wipe -y`
- Epics: `epic create`, `epic list`, `epic show`, `epic update`, `epic complete`, `epic delete`
- Tasks: `task create`, `task list`, `task show`, `task update`, `task delete`
- Subtasks: `subtask create`, `subtask list`, `subtask update`, `subtask delete`
- Comments: `comment add`, `comment list`, `comment update`, `comment delete`
- Dependencies: `dep add`, `dep remove`, `dep list`
- Config: `config list`, `config get`, `config set`, `config unset`
- Query: `search`, `history`, `list`

## Behavior

Support `--toon` for compact agent output. Preserve status values, priority values, and task workflow rules from the Trekker reference. Prefer additive enhancements over breaking command changes.

## Client Configuration

`phatodo init` and `ptd init` should write local project configuration only:

- workspace ID
- project ID
- API base URL
- access key
- access secret
- optional cache metadata

The local file lives at `.phatodo/config.json` and must not be committed. It should be created with mode `0600` because it contains the access secret.

Canonical task data belongs on the server.

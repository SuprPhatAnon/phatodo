# Phatodo Quickstart

Task tracker for AI agents. Data stored on the server; local `.phatodo/config.json` stores API URL and credentials only.

## Setup
ptodo admin init             # Create the first admin user
ptodo admin bootstrap        # Provision the workspace, project, and local CLI identity
ptodo config list            # Confirm access to the bootstrapped project

## Core Rules
1. Start by checking ready work: `ptodo --toon ready`
2. Set status to `in_progress` when starting, `completed` when done
3. Add a summary comment before marking work complete
4. Use `--toon` for token-efficient output
5. Keep descriptions as plain text; use `planned_files`, `changed_files`, `acceptance_criteria`, and `completion_evidence` for ordered lists of requirements and proof
6. Leave checkpoint comments when handing off context

## Commands

### Epics
ptodo epic create -t "Title" [-d "desc"] [-p 0-5] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]']
ptodo epic list [--status <status>] [--limit <n>]
ptodo epic show <epic-id>
ptodo epic update <epic-id> [-t "Title"] [-d "desc"] [-p 0-5] [-s <status>] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]'] [--summary "completion summary"] [--evidence-json '["link-or-note"]']
ptodo epic complete <epic-id>
ptodo epic delete <epic-id>

### Tasks
ptodo task create -t "Title" --prefix <prefix> (or --issue-prefix <prefix>) [-d "desc"] [-p 0-5] [-e <epic-id>] [--kind task|bug|feature|chore|spike] [--root-cause-analysis "why"] [--planned-files-json '["file.go"]'] [--tags "a,b"] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]']
ptodo task list [--status <status>] [--epic <epic-id>] [--limit <n>]
ptodo task show <task-id>
ptodo task update <task-id> [-t "Title"] [-d "desc"] [-p 0-5] [-s <status>] [-k task|bug|feature|chore|spike] [--root-cause-analysis "why"] [--changed-files-json '["file.go"]'] [--tags "a,b"] [-e <epic-id>] [--no-epic] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]'] [--summary "completion summary"] [--evidence-json '["link-or-note"]']
ptodo task delete <task-id>

### Updating Criteria and Evidence

Use `--criteria-json` while work is still open to record what completion must satisfy.
Use `--summary`, `--changed-files-json`, and `--evidence-json` when the work is finished and you need to capture proof.

Good usage:
ptodo task update ABC-1 --criteria-json '["docs written","tests passing"]'
ptodo task update ABC-1 -s completed --summary "Implemented in internal/cli" --changed-files-json '["internal/cli/commands.go"]' --evidence-json '["go test ./internal/cli","docs/QUICKSTART.md updated"]'

Bad usage:
ptodo task update ABC-1 -s completed --summary "Done" --evidence-json "fixed it"
ptodo task update ABC-1 -s completed --summary "Done" --evidence-json '["tested"]'
ptodo task update ABC-1 --description '["docs written","tests passing"]'

### Subtasks
ptodo subtask create <task-id> -t "Title" [-d "desc"] [-p 0-5] [--kind task|bug|feature|chore|spike] [--root-cause-analysis "why"] [--planned-files-json '["file.go"]'] [-a <user-id>] [--criteria-json '["criterion 1","criterion 2"]']
ptodo subtask list <task-id> [--limit <n>]
ptodo subtask update <subtask-id> [-t "Title"] [-d "desc"] [-p 0-5] [-s <status>] [-a <user-id>] [--changed-files-json '["file.go"]'] [--criteria-json '["criterion 1","criterion 2"]'] [--summary "completion summary"] [--evidence-json '["link-or-note"]']
ptodo subtask delete <subtask-id>

### Comments
ptodo comment add <task-id> -a "agent" -c "content" [-k comment|analysis|summary|checkpoint|handoff]
ptodo comment list <task-id>
ptodo comment update <comment-id> -c "new content"
ptodo comment delete <comment-id>

### Dependencies
ptodo dep add <task-id> <depends-on-id>
ptodo dep remove <task-id> <depends-on-id>
ptodo dep list <task-id>

### Locks
ptodo lock acquire <entity-type> <entity-id> [--reason "why"] [--expires 1h]
ptodo lock release <lock-id>
ptodo lock list [--type epic,task,subtask] [--entity <entity-id>] [--active]

### Workflow
ptodo ready [--epic <epic-id>] [--limit <n>]

### Config
ptodo config list
ptodo config get <key>
ptodo config set <key> <value>
ptodo config unset <key>

### Query
ptodo search "query" [--type epic,task,subtask,comment] [--status <status>] [--limit <n>]
ptodo history [--entity <entity-id>] [--type task] [--action create,update,delete] [--since <date>] [--limit <n>]
ptodo list [--type epic,task,subtask] [--status <status>] [--priority 0,1] [--sort priority:asc,created:desc] [--limit <n>]

## Status Values
Tasks: todo, in_progress, completed, wont_fix, archived
Epics: todo, in_progress, completed, archived

## Priority Scale
0=critical, 1=high, 2=medium (default), 3=low, 4=backlog, 5=someday

## Agent Workflow
1. Run `ptodo --toon ready` or `ptodo --toon list` to pick up work.
2. Update the task to `in_progress`.
3. Work the task and leave checkpoint comments if context changes.
4. Add a summary comment before completion.
5. Complete tasks with `ptodo task update <task-id> -s completed --summary "..." --changed-files-json '["file.go"]' --evidence-json '["verification"]'`.

## Before Context Reset
ptodo comment add <task-id> -a "agent" -c "Checkpoint: done A,B. Next: C. Files: x.ts, y.ts"

## Writing Effective Descriptions

Good descriptions help future agents continue your work:

### Epic descriptions should include:
- Goal and success criteria
- High-level implementation approach
- Key files or modules affected

### Task descriptions should include:
- What needs to be done
- Implementation steps
- Files to create or modify
- Acceptance criteria

### Example:
Bad: "Add authentication"
Good: "Implement JWT auth for API.
- Add /auth/login and /auth/logout endpoints
- Create middleware in src/middleware/auth.ts
- Use bcrypt for password hashing
- Protect /api/users and /api/tasks"

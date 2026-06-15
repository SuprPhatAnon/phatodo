# Validation Rules

This document defines the server-side rules that keep the base ticket system consistent.

## Status Rules

### Tasks

Allowed statuses:

- `todo`
- `in_progress`
- `completed`
- `wont_fix`
- `archived`

Recommended lifecycle:

1. `todo`
2. `in_progress`
3. `completed`

Allowed terminal exceptions:

- `wont_fix`
- `archived`

### Epics

Allowed statuses:

- `todo`
- `in_progress`
- `completed`
- `archived`

Recommended lifecycle:

1. `todo`
2. `in_progress`
3. `completed`

## Completion Rules

### Tasks and subtasks

Before a task or subtask can be marked `completed`:

- the item must have a summary comment or completion summary
- the item must satisfy its acceptance criteria if any are present
- the server must persist completion metadata

Completion metadata includes:

- `completed_by`
- `completed_at`
- `completion_summary`
- `completion_evidence`

### Epics

Before an epic can be marked `completed`:

- child tasks must be in a terminal or compatible state
- the epic must satisfy its own acceptance criteria if present
- the server must archive or finalize child tasks according to the completion flow

## Acceptance Criteria

Acceptance criteria are stored as an ordered list.

Rules:

- empty criteria are allowed for simple items
- if criteria are present, completion must not ignore them
- the server should reject completion if evidence does not cover the declared criteria

## Completion Evidence

Completion evidence is an ordered list of items such as:

- links
- notes
- verification steps
- checkpoint references

Rules:

- evidence should be non-empty when the completion requires external proof
- evidence should be stored immutably as part of the completion event

## Comment Rules

Allowed comment kinds:

- `comment`
- `analysis`
- `summary`
- `checkpoint`
- `handoff`

Rules:

- `summary` comments should precede completion
- `checkpoint` comments should be used when context is handed off
- `analysis` comments are for working notes
- `handoff` comments should describe what remains and the next step

## Dependency Rules

Rules:

- a task cannot depend on itself
- duplicate dependencies per project are forbidden
- dependency cycles should be rejected
- a task with unsatisfied dependencies should not be completed

## Ready Rules

Rules:

- `ptodo ready` should show only top-level tasks in `todo` status
- a ready task must have all dependencies satisfied
- `--epic` should scope both the ready list and any dependents hints to that epic
- the ready view should stay ordered by priority first so agents can pick the most urgent unblocked work quickly

## Lock Rules

Rules:

- only one active lock is allowed per entity
- a lock may expire automatically
- expired or released locks should not block new work
- lock release should be idempotent
- the server should reject competing active locks on the same entity

## Project Config Rules

Rules:

- config keys are unique per project
- config updates should be recorded in audit history

## Search and History Rules

Rules:

- search indexes should be updated whenever text-bearing records change
- history events should be immutable
- history should record actor, action, entity, before state, after state, and timestamp

## Auth and Access Rules

Rules:

- requests missing the access headers must fail
- users without project access must not see or modify project data
- admins can access all projects

## Bootstrap Rules

Rules:

- `ptodo admin init` must fail if any admin user already exists
- `ptodo admin init` must create the first admin user only
- `ptodo admin bootstrap` must authenticate against the admin user created by `ptodo admin init`
- `ptodo admin bootstrap` must fail if the target project already exists
- `ptodo admin bootstrap` must create the workspace/project, project-scoped CLI identity, and matching `user_project_access` row needed for the local config
- a server with no admin user or no bootstrap project is not ready for CLI use
- the CLI should not guess bootstrap state

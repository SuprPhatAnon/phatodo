# CLI Output

This document defines the user-facing output contract for the `ptodo` command.

## Modes

### Default mode

Default mode should be human-readable and suitable for direct terminal use.
The examples in this document use the compact TOON rendering for brevity unless explicitly noted otherwise.

### `--toon`

`--toon` is the compact, machine-friendly TOON mode.
Use [TOON Reference](TOON_SPEC.md) as the source of truth for syntax.

Rules:

- keep fields stable
- avoid decorative text
- prefer TOON objects or arrays over ad hoc prose
- avoid multiline prose unless it is content from the record itself

## Output Principles

- list commands should emit one record per line or a compact structured block
- show/get commands should emit a single record
- writes should echo the affected item and its server-generated identifiers
- errors should go to stderr

## Suggested Record Shapes

### Config

`ptodo config list`:

```text
- theme: dark
```

`ptodo config set theme dark`:

```text
- theme: dark
```

`ptodo config get theme`:

```text
- theme: dark
```

`ptodo config unset theme`:

```text
- theme: dark
```

### Tasks and epics

Render:

- ID
- title
- status
- priority
- assigned owner
- parent or epic relation

`ptodo task create -t "Write docs" --prefix ABC`:

```text
- id: ABC-1
  issue_prefix: ABC
  title: "Write docs"
```

`ptodo task list --status in_progress --epic epic-1 --limit 20`:

```text
tasks[1]:
  - id: ABC-1
    title: "Write docs"
    description: ""
    priority: 2
    status: in_progress
    epicId: epic-1
```

`ptodo epic list --status in_progress --limit 20`:

```text
epics[1]:
  - id: EPIC-1
    title: "Track auth"
    description: "Add end-to-end auth checks across the CLI and API."
    priority: 0
    status: in_progress
    assignedTo: bryan
    acceptanceCriteria[2]:
      - "CLI commands require project config"
      - "API routes reject missing credentials"
```

`ptodo epic show EPIC-1`:

```text
- id: EPIC-1
  title: "Track auth"
  description: "Add end-to-end auth checks across the CLI and API."
  priority: 0
  status: in_progress
  assignedTo: bryan
  acceptanceCriteria[2]:
    - "CLI commands require project config"
    - "API routes reject missing credentials"
```

`ptodo subtask list ABC-1 --limit 20`:

```text
subtasks[1]:
  - id: ABC-2
    title: "Write docs"
    description: ""
    priority: 2
    status: todo
    parentTaskId: ABC-1
```

`ptodo ready --epic epic-1`:

```text
ready[1]:
  - id: CORE-1
    title: "Health endpoints and k8s liveness/readiness probes for worker services"
    description: "Future work:..."
    priority: 1
    status: todo
    epicId: epic-1
    tags: "infra,api"
    createdAt: "2026-06-09T02:13:02Z"
    updatedAt: "2026-06-13T15:01:51Z"
    dependents[1]{id,title,status,priority}:
      - CORE-5,"Two-tier automated database backups via SCP to NAS and GCS",todo,1
```

### Comments

Render:

- comment ID
- author
- kind
- content
- timestamp

`ptodo comment add <task-id> -a agent -c "Done" -k summary`:

```text
- id: cmt-1
  author: agent
  kind: summary
  content: Done
```

`ptodo comment list <task-id>`:

```text
comments[1]:
  - id: cmt-1
    author: agent
    kind: analysis
    content: "Working notes"
```

### Dependencies

Render:

- task ID
- depends-on ID

`ptodo dep add ABC-1 ABC-2`:

```text
- id: dep-1
  taskId: ABC-1
  dependsOnId: ABC-2
```

`ptodo dep list ABC-1`:

```text
dependencies[1]:
  - id: dep-1
    taskId: ABC-1
    dependsOnId: ABC-2
```

### Locks

Render:

- lock ID
- entity type
- entity ID
- owner
- expiration

`ptodo lock acquire task ABC-1 --reason "editing" --expires 30m`:

```text
- id: LOCK-1
  entityType: task
  entityId: ABC-1
  lockedBy: user-1
  reason: editing
  leasedAt: "2026-06-15T12:00:00Z"
  expiresAt: "2026-06-15T12:30:00Z"
```

`ptodo lock list --type epic,task --entity EPIC-1 --active`:

```text
locks[1]:
  - id: LOCK-2
    entityType: epic
    entityId: EPIC-1
    lockedBy: user-1
    reason: planning
    leasedAt: "2026-06-15T12:00:00Z"
    expiresAt: "2026-06-15T13:00:00Z"
```

### Search, History, List

`ptodo search "auth bug" --type task --status todo --limit 10`:

```text
search[1]:
  - id: ABC-1
    entityType: task
    title: "Fix auth bug"
    status: todo
    priority: 1
```

`ptodo history --entity ABC-1 --type task --action update --since 2025-01-01 --limit 5`:

```text
history[1]:
  - id: 42
    entityType: task
    entityId: ABC-1
    action: update
    actorLabel: alice
```

`ptodo list --type epic,task --status todo --priority 0,1 --sort priority:asc,created:desc --limit 2`:

```text
list[1]:
  - id: EPIC-1
    entityType: epic
    title: "Track auth"
    status: todo
    priority: 0
```

### Locks

Render:

- lock ID
- entity type
- entity ID
- owner
- expiration

## `--toon` Output Contract

For implementation simplicity, `--toon` should prefer the same fields every time:

- IDs first
- then names or titles
- then status and priority
- then ownership and timestamps

The goal is to keep the output easy for agents to parse and compare across runs.
When a response is a list of homogeneous objects, use a TOON array with an explicit item count and field list.

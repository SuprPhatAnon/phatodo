# CLI Output

This document defines the user-facing output contract for the `ptodo` command.

## Modes

### Default mode

Default mode should be human-readable and suitable for direct terminal use.

### `--toon`

`--toon` is the compact, machine-friendly mode.

Rules:

- keep fields stable
- avoid decorative text
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
theme=dark
```

`ptodo config set theme dark`:

```text
theme=dark
```

`ptodo config get theme`:

```text
theme=dark
```

`ptodo config unset theme`:

```text
theme=dark
```

### Tasks and epics

Render:

- ID
- title
- status
- priority
- assigned owner
- parent or epic relation

`ptodo task create -t "Write docs" --issue-prefix ABC`:

```text
id=ABC-1
issue_prefix=ABC
title=Write docs
```

### Comments

Render:

- comment ID
- author
- kind
- content
- timestamp

### Dependencies

Render:

- task ID
- depends-on ID

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

# Implementation Plan

This document describes the intended end state for the CLI, API, and database layers and the order we expect to build them.

## End State

Phatodo should work as a three-layer system:

1. `phatodo`, `ptd`, and `ptodo` act as thin clients.
2. The server owns all canonical task data and validation.
3. Postgres stores projects, tasks, epics, comments, dependencies, config, history, locks, and search indexes.

Local `.phatodo/` data should stay limited to client config and optional cache data.

## Layer Responsibilities

### CLI

- Parse Trekker-compatible commands.
- Read `.phatodo/config.json`.
- Send authenticated requests to the server.
- Render JSON-backed responses as compact terminal output.
- Preserve `--toon` as the agent-friendly output mode.

### API

- Expose Trekker-compatible route shapes under `/api/v1`.
- Enforce authentication and project access.
- Validate state transitions and completion rules.
- Return structured JSON for command output, history, and errors.

### Database

- Store the source of truth for projects and work items.
- Enforce relational integrity, ownership, and audit history.
- Support project-scoped config, search, and history queries.
- Keep IDs stable so Trekker imports remain possible.

## Build Phases

### Phase 1

- Wire CLI config loading and request plumbing.
- Implement the first read path end to end.
- Keep server behavior simple and explicit.

### Phase 2

- Fill out read APIs for list/show/search/history.
- Add repository implementations for projects, epics, tasks, comments, and dependencies.
- Make output formatting consistent across commands.

### Phase 3

- Implement write paths and lifecycle validation.
- Record immutable audit events for all meaningful changes.
- Add lock handling for edit and completion flows.

### Phase 4

- Complete import/export and migration support.
- Add dashboard views backed by the same API and database state.
- Harden authorization and operational tooling.

## Current Anchors

- The CLI command shape is documented in `docs/trekker_reference.txt`.
- The command-to-endpoint-to-table map is documented in `docs/COMMAND_MAP.md`.
- The server route shape is documented in `docs/API.md`.
- The schema model is documented in `docs/DATABASE_SCHEMA.md`.
- The end-to-end request and storage path is documented in `docs/DATAFLOW.md`.
- The current CLI-to-server slice starts with `config list`.

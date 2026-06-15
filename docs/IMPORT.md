# Import

This document describes how Trekker-style local data should be migrated into phatodo.

## Purpose

The import path exists so existing local task data can be moved into the central Postgres-backed system without losing IDs or workflow history.

## Source Data

The source format is the local Trekker SQLite database:

- `.trekker/trekker.db`

## Target Data

The target is the phatodo Postgres schema:

- `workspaces`
- `projects`
- `users`
- `user_project_access`
- `project_config`
- `id_counters`
- `epics`
- `tasks`
- `comments`
- `dependencies`
- `events`
- `work_item_locks`
- `search_index`

## Import Scope

The import should preserve:

- IDs
- titles
- descriptions
- statuses
- priorities
- tags
- comments
- dependencies
- config values
- any available audit history

## Import Mapping

### Project structure

- one local Trekker project becomes one phatodo project
- the target project must belong to a workspace
- the imported project should preserve the original issue prefix if present

### Tasks and subtasks

- Trekker tasks map to `tasks`
- Trekker subtasks also map to `tasks`
- parent-child relationships should be preserved through `parent_task_id`

### Epics

- Trekker epics map to `epics`
- task-to-epic relationships should be retained

### Comments

- Trekker comments map to `comments`
- comment kind should default to `comment` unless the source data is richer

### Dependencies

- Trekker dependencies map to `dependencies`
- duplicate edges should be rejected

### Config

- Trekker config values map to `project_config`
- `issue_prefix` should be imported explicitly

### History

- if source history exists, preserve it as `events`
- if not, create synthetic import events so the server history is not empty

## Import Contract

The importer should:

- create or reuse a workspace
- create or reuse a project
- preserve original IDs whenever possible
- report conflicts instead of silently renaming records
- reject imports that would corrupt dependencies or identity mappings

## Authentication Data

Local Trekker credentials should not be imported as-is into server auth storage unless explicitly converted.

The importer should treat auth records separately from ticket data.

## Recommended Import Tool

The implementation should provide one of:

- a dedicated import command
- a migration utility
- a server-side import endpoint

The docs should treat the import path as a first-class operational feature, even before it is implemented.


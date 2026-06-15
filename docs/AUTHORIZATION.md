# Authorization

This document defines who can do what in phatodo.

## Authentication Model

Every API request uses:

- `X-Phatodo-Access-Key`
- `X-Phatodo-Access-Secret`

The server should resolve the caller to a user record and a role.

## Roles

### `admin`

Admins can:

- access all projects
- create and manage project data
- bootstrap or manage server-wide resources
- see all audit history for their scope

### `user`

Regular users can:

- access only the project granted to them
- read and modify project data within that project
- manage their own task and comment workflow if permitted by project rules

## Project Access

The `user_project_access` table is the project-level gate.

Rules:

- one regular user maps to one project access row
- a user without a project access row should not be able to act on project data
- the server must reject cross-project requests even if the user knows the IDs

## Workspace Scope

The workspace is the server tenancy boundary.

Rules:

- workspace IDs are local config values and server-side tenancy values
- project access should never bypass workspace ownership
- the server should scope queries by both workspace and project where possible

## Command Authorization Matrix

### Read commands

Usually allowed for the scoped project:

- `config list`
- `config get`
- `epic list`
- `epic show`
- `task list`
- `task show`
- `subtask list`
- `comment list`
- `dep list`
- `search`
- `history`
- `list`
- `lock list`

### Write commands

Usually restricted by project membership and role:

- `config set`
- `config unset`
- `epic create`
- `epic update`
- `epic complete`
- `epic delete`
- `task create`
- `task update`
- `task delete`
- `subtask create`
- `subtask update`
- `subtask delete`
- `comment add`
- `comment update`
- `comment delete`
- `dep add`
- `dep remove`
- `lock acquire`
- `lock release`

## Ownership Rules

Ownership fields are distinct from authorization.

- `assigned_to` means the person expected to do the work
- `created_by` means the person who created the record
- `updated_by` means the person who last changed it
- `completed_by` means the person who finished it

These fields should never be used as a substitute for access control.

## Bootstrap Authorization

Bootstrap is a privileged action.

Rules:

- `ptodo admin init` must only be allowed before any admin user exists
- `ptodo admin bootstrap` must require valid admin authentication
- the system should not let an ordinary project user create the first admin account, workspace, project, project-scoped CLI identity, or project access row

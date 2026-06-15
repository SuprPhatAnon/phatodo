# Bootstrap

This document defines the minimum server state required for a no-dashboard phatodo installation to be usable from the CLI.

## Goal

A fresh install should be able to:

1. Start `phatodo-server`.
2. Create the first admin identity with `ptodo admin init`.
3. Bootstrap the first project with `ptodo admin bootstrap`.
4. Write a local `.phatodo/config.json`.
5. Run `ptodo config list`.
6. Run the first task and epic commands without a dashboard.

## Bootstrap Phases

Phatodo bootstrap has two distinct phases:

1. `ptodo admin init` creates the first admin user only.
2. `ptodo admin bootstrap` uses that admin user to provision the initial project and local CLI config.

These operations are intentionally separate so the system can enforce a clean first-admin boundary and a clean project-provisioning boundary.

## Admin Init

`ptodo admin init` is the only command that may create the first admin user.

Required behavior:

- accept the admin username and API server URL as parameters
- prompt for the password twice
- fail if any admin user already exists
- create the admin identity only
- do not create a workspace
- do not create a project
- do not write local CLI config
- do not create project config

The command is the identity bootstrap step and should be treated as one-time only.

## Admin Bootstrap

`ptodo admin bootstrap` provisions the project-side state needed for the repository.

Required behavior:

- accept the admin username and API server URL as parameters
- authenticate as the admin user created by `ptodo admin init`
- create the workspace and project needed for the current repository if they do not already exist
- create the project-scoped access identity needed by the CLI
- create the matching `user_project_access` row for that identity
- write the local `.phatodo/config.json`
- fail if the target project already exists

The command is the project bootstrap step and should be treated as a one-time provisioning path for a given project.

## Minimum Bootstrap Data

The server needs one initial record set before normal CLI usage is possible.

### `workspaces`

Required fields:

- `id`
- `name`
- `slug`

This is the top-level tenancy container.

### `projects`

Required fields:

- `id`
- `workspace_id`
- `name`

This is the project the CLI will point at in `.phatodo/config.json`.

### `users`

Required fields for the first admin created by `ptodo admin init`:

- `id`
- `display_name`
- `role = admin`
- `access_key`
- `access_secret_hash`
- `username`
- `password_hash`

Required fields for the project-scoped CLI identity created by `ptodo admin bootstrap`:

- `id`
- `display_name`
- `role = user`
- `access_key`
- `access_secret_hash`

Required associated access row:

- `user_project_access` linking the user to the bootstrap project

## Recommended Bootstrap Sequence

The preferred order is:

1. Apply the schema migrations.
2. Run `ptodo admin init` to create the first admin user.
3. Run `ptodo admin bootstrap` to create the project-scoped CLI identity and project config.
4. Verify access with `ptodo config list`.

## Initial Local Config

The local config file should point at the bootstrap project and the project-scoped CLI identity:

- `api_url`
- `workspace_id`
- `project_id`
- `access_key`
- `access_secret`

Example intent:

- `api_url` should point at the running `phatodo-server`
- `workspace_id` should match the bootstrapped workspace
- `project_id` should match the bootstrapped project
- `access_key` and `access_secret` should match the project-scoped CLI identity returned by bootstrap

## What Must Exist Before Normal Use

Before any user can run the base ticket system, all of the following must be true:

- the database exists
- the migrations are applied
- at least one admin user exists
- the bootstrap project exists
- the project-scoped CLI identity exists
- the matching `user_project_access` row exists
- the local CLI config points at the project
- the CLI can authenticate with the server

## Dashboard-Free Startup Contract

If the dashboard is not running, the CLI must still be able to:

- list project config
- create and update tasks
- add comments
- manage dependencies
- query history and search results

The bootstrap path exists solely to make that possible from a clean install.

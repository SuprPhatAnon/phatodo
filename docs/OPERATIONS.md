# Operations

This document describes how to run the base ticket system without the dashboard.

## Prerequisites

- Go toolchain installed
- Postgres available
- schema migrations available in `migrations/`
- a first admin user created with `ptodo admin init`
- a project provisioned with `ptodo admin bootstrap`

## Local Development

### Start the server

Use the configured Makefile target:

```sh
make run-server
```

Required environment:

- `PHATODO_DATABASE_URL`
- optionally `PHATODO_ADDR`

### Run the CLI

Use the configured Makefile target:

```sh
make run-ptodo
```

To install the CLI into `$(GOPATH)/bin` for use from your `PATH`, run:

```sh
make install
```

## Initial Bring-Up

Recommended order:

1. Start Postgres.
2. Apply the migrations.
3. Run `ptodo admin init` to create the first admin user.
4. Run `ptodo admin bootstrap` to provision the bootstrap project and local config.
5. Verify `ptodo config list`.
6. Verify `ptodo task list`.

## Configuration Files

### Local CLI config

Path:

- `<repo>/.phatodo/config.json`

Fields:

- `api_url`
- `workspace_id`
- `project_id`
- `access_key`
- `access_secret`

### Server environment

Required:

- `PHATODO_DATABASE_URL`

Optional:

- `PHATODO_ADDR`

## Deployment Modes

### Compose

Use:

```sh
make compose-up
```

This starts the server and Postgres together for local validation.

### k3s

Use:

```sh
make deploy-k3s
```

The docs assume the image has already been built and pushed.

## Validation Checklist

Before calling the system ready:

- `ptodo config list` works
- `ptodo task list` works
- auth failures return `401`
- invalid project access returns `403`
- missing data returns clear `404` or `422` responses
- audit events are created for writes

## No-Dashboard Operational Contract

The system is considered usable without the dashboard only when:

- every base ticket command works from the CLI
- the CLI can bootstrap local config through `ptodo admin bootstrap`
- the server can authenticate and scope a project
- the server persists data in Postgres
- errors are actionable from terminal output alone

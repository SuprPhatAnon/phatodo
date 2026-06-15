# Phatodo

`Phatodo` is planned as a Trekker-compatible task tracker with one command-line entrypoint:

- `ptodo`: command-line executable that preserves the command structure in `docs/trekker_reference.txt`.
- `phatodo-server`: an API server and dashboard backed by Postgres.

## Current Scaffold

- `cmd/ptodo` contains the CLI entrypoint.
- `cmd/phatodo-server` contains the server entrypoint.
- `internal/cli` contains Trekker-compatible command registration.
- `internal/server` contains HTTP server routing.
- `internal/domain` contains shared task-tracking domain types.
- `internal/storage/postgres` is reserved for Postgres repositories.
- `migrations` contains database schema migrations.
- `web/dashboard` is reserved for dashboard assets.

## Development

```sh
make build
make test
make run-ptodo
make run-server
make compose-up
```

Run `ptodo init` from a project checkout to create `.phatodo/config.json` with local API URL, workspace ID, project ID, access key, and access secret settings. The `.phatodo/` directory is ignored by Git.

The server listens on `:8080` by default. Set `PHATODO_ADDR` to override the address and `PHATODO_DATABASE_URL` when wiring Postgres storage.

See `docs/IMPLEMENTATION_PLAN.md` for the CLI-to-API-to-database end state and build phases.

See `docs/DEPLOYMENT.md` for Docker, Compose, and k3s deployment notes.

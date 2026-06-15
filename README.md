trakkr - centralized AI task tracking

`trakkr` is planned as a Trekker-compatible task tracker with two deliverables:

- `trakkr`: a command-line executable that preserves the command structure in `docs/trekker_reference.txt`.
- `trakkr-server`: an API server and dashboard backed by Postgres.

## Current Scaffold

- `cmd/trakkr` contains the CLI entrypoint.
- `cmd/trakkr-server` contains the server entrypoint.
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
make run-cli
make run-server
make compose-up
```

The server listens on `:8080` by default. Set `TRAKKR_ADDR` to override the address and `TRAKKR_DATABASE_URL` when wiring Postgres storage.

See `docs/DEPLOYMENT.md` for Docker, Compose, and k3s deployment notes.

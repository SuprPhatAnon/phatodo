# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go scaffold for `trakkr`, a centralized Trekker-compatible task tracker.

- `cmd/trakkr` contains the CLI executable.
- `cmd/trakkr-server` contains the API and dashboard server executable.
- `internal/cli` owns command parsing and Trekker-compatible command shape.
- `internal/server` owns HTTP routing and dashboard delivery.
- `internal/domain` contains shared statuses, priorities, and task entities.
- `internal/storage/postgres` is reserved for Postgres repositories.
- `migrations/` contains SQL schema migrations.
- `web/dashboard` is reserved for dashboard source assets.
- `docs/` contains architecture notes and the original Trekker command reference.

## Build, Test, and Development Commands

Use the standard Go toolchain:

- `go run ./cmd/trakkr --help` shows the CLI command scaffold.
- `go run ./cmd/trakkr-server` starts the server on `:8080`.
- `go test ./...` runs all Go tests.
- `go build ./cmd/trakkr ./cmd/trakkr-server` builds both executables.

## Coding Style & Naming Conventions

Run `gofmt` on Go files. Keep package names short and domain-focused, such as `cli`, `server`, `domain`, and `postgres`. Use names from `docs/trekker_reference.txt`: `epic`, `task`, `subtask`, `comment`, `dependency`, `history`, and `priority`.

Keep Markdown concise and actionable. Use fenced code blocks for command examples and bullets for short lists. Prefer ASCII unless a referenced source already uses other characters.

## Testing Guidelines

Place Go tests next to the package under test and name files `*_test.go`. Prioritize tests for CLI compatibility, task lifecycle rules, dependency validation, audit history, search behavior, and Postgres repository behavior. Name tests after observable behavior, for example `TestTaskCannotCompleteWithoutSummary`.

## Commit & Pull Request Guidelines

The current Git history only contains `first commit`, so no project-specific commit convention is established. Use short, imperative commit subjects such as `Add Trekker reference notes` or `Define task API model`.

Pull requests should include a brief summary, the reason for the change, any commands run, and screenshots or terminal output when user-facing CLI behavior changes. Link related issues or Trekker tasks when available.

## Agent-Specific Instructions

For agent workflows, use `docs/trekker_reference.txt` as the behavioral reference. Preserve the documented Trekker command shape unless a design note explicitly changes it. The server is the source of truth; local `.trakkr` data should be limited to client config and optional cache data.

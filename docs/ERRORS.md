# Errors

This document defines the error contract for the CLI and API.

## API Error Shape

API failures should return JSON like:

```json
{
  "error": "error_code",
  "message": "human readable explanation"
}
```

## Common Error Classes

### `400 Bad Request`

Use for:

- malformed input
- missing required fields
- invalid enum values
- invalid ID formats

### `401 Unauthorized`

Use for:

- missing access key
- missing access secret
- invalid credentials

### `403 Forbidden`

Use for:

- authenticated user lacks access to the project
- user tries to act outside their workspace/project scope

### `404 Not Found`

Use for:

- missing project
- missing epic/task/comment/dependency/lock/config key

### `409 Conflict`

Use for:

- duplicate config keys
- duplicate dependencies
- status transitions that violate lifecycle rules
- completion attempted while dependencies are unresolved
- lock contention

### `422 Unprocessable Entity`

Use for:

- valid JSON with invalid business semantics
- completion rules not satisfied
- acceptance criteria or evidence mismatch

### `500 Internal Server Error`

Use for:

- unexpected repository errors
- unexpected server state

### `503 Service Unavailable`

Use for:

- storage not configured
- server started without required backend resources
- bootstrap not completed

## CLI Exit Codes

Recommended exit codes:

- `0` success
- `1` operational failure
- `2` usage error or invalid command shape

## CLI Error Output

The CLI should:

- print a concise human-readable message to stderr
- avoid dumping raw stack traces by default
- preserve enough detail to diagnose the failing command

## `--toon` Error Output

When `--toon` is active:

- the CLI should still return the same exit code
- the error text should remain compact and parseable
- the CLI should not add decorative formatting

## Error Codes to Standardize

The server should prefer stable machine-readable error codes such as:

- `missing_credentials`
- `unauthorized`
- `forbidden`
- `not_found`
- `conflict`
- `validation_failed`
- `bootstrap_required`
- `storage_unavailable`
- `not_implemented`

## Error Handling Rules

- commands should fail fast on invalid input
- the server should validate before writing
- the CLI should not guess on error recovery
- errors that block the base ticket workflow should be documented and actionable


#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-pre-commit}"
COMMIT_MSG_FILE="${2:-${COMMIT_MSG_FILE:-}}"

DB_CHANGE_PATTERN='(^migrations/)|(\.sql$)|(\.sql\.go$)|(^|/)sqlc\.ya?ml$|(^internal/storage/postgres/)'
AUTH_TRAILER_PATTERN='^DB-Change-Authorized:[[:space:]]*yes[[:space:]]*$'

staged_files() {
  if [[ -n "${DB_GUARD_FILE_LIST:-}" ]]; then
    cat "$DB_GUARD_FILE_LIST"
    return
  fi

  git diff --cached --name-only --diff-filter=ACMRT
}

db_files() {
  staged_files | grep -E "$DB_CHANGE_PATTERN" || true
}

print_db_files() {
  local files="$1"

  echo "$files" | sed 's/^/  - /' >&2
}

require_no_locked_db_change() {
  local files="$1"

  if [[ -z "$files" ]]; then
    return 0
  fi

  if [[ "${ALLOW_DB_CHANGE:-0}" == "1" ]]; then
    echo "DB/SQL guard: ALLOW_DB_CHANGE=1 set; allowing approved DB change."
    return 0
  fi

  cat >&2 <<'EOF'
DB/SQL changes are locked by default.

Staged DB-related files:
EOF
  print_db_files "$files"
  cat >&2 <<'EOF'

Approved DB changes require:
  - explicit task-specific user authorization
  - an active Trekker task
  - a new migration for schema changes
  - sqlc query and generated-code updates when applicable
  - affected tests and documentation updates
  - verification evidence

Re-run the commit with ALLOW_DB_CHANGE=1 only after those requirements are met.
EOF
  return 1
}

require_authorization_trailer() {
  local files="$1"
  local msg_file="$COMMIT_MSG_FILE"

  if [[ -z "$files" ]]; then
    return 0
  fi

  if [[ -z "$msg_file" ]]; then
    echo "commit-msg DB/SQL guard requires a commit message file." >&2
    return 1
  fi

  if grep -Eq "$AUTH_TRAILER_PATTERN" "$msg_file"; then
    return 0
  fi

  cat >&2 <<'EOF'
This commit touches DB/SQL files and must record explicit authorization.

DB-related files:
EOF
  print_db_files "$files"
  cat >&2 <<'EOF'

Add this exact trailer to the commit message after confirming approval:

DB-Change-Authorized: yes

This trailer records authorization; it does not replace the required Trekker task,
new migration, sqlc regeneration, tests, documentation, or verification evidence.
EOF
  return 1
}

FILES="$(db_files)"

case "$MODE" in
  pre-commit)
    require_no_locked_db_change "$FILES"
    ;;
  commit-msg)
    require_authorization_trailer "$FILES"
    ;;
  *)
    echo "unknown DB guard mode: $MODE" >&2
    exit 2
    ;;
esac

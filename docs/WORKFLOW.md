# Workflow Guardrails

This document collects the repo-level rules that are most useful while working in phatodo.

## Track Work In Trekker

- Start from `ptodo --toon ready` or `ptodo --toon list` before picking up work.
- Move the chosen task to `in_progress` before coding.
- Leave a checkpoint or handoff comment whenever work is paused or interrupted.
- Complete tasks with `ptodo task update <task-id> -s completed --summary "..." --changed-files-json '["file.go"]' --evidence-json '["verification"]'` only after the work is implemented, validated, and documented.

## Task Hygiene

- Keep task titles specific and actionable.
- Use the correct epic and the right issue prefix for new work.
- Prefer lower numeric priorities for blockers and correctness work.
- Use `checkpoint` comments when handing context to the next agent.
- Use `handoff` comments when the next step is clear but unfinished.

## Database Rules

These are the important guardrails for database work in this repo:

- Any database shape change requires explicit, task-specific user authorization before implementation.
- Any `INSERT`, `UPDATE`, `UPSERT`, or `DELETE` query requires explicit, task-specific user authorization before implementation.
- DB changes should go through the repo's `sqlc` flow, not ad hoc handwritten SQL in Go.
- Schema changes must be reflected in the migration, the `sqlc` query files, the regenerated Go code, the storage layer, the server handlers, and the docs that describe the affected behavior.
- Do not assume a task ticket is enough authorization for DB or SQL changes.
- If the requested DB change is broader than the authorization, stop and ask for the narrower approval.

## Completion Discipline

- Do not mark a task done unless the acceptance criteria are verified.
- Prefer concise summary comments that state what changed and what remains.
- If work is blocked, record the blocker clearly instead of leaving the state ambiguous.

-- name: CountAdminUsers :one
SELECT COUNT(*)::bigint
FROM users
WHERE role = 'admin';

-- name: LockUsersTable :exec
LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE;

-- name: InsertAdminUser :exec
INSERT INTO users (
	id, display_name, role, access_key, access_secret_hash,
	username, password_hash
) VALUES (
	sqlc.arg(id), sqlc.arg(display_name), 'admin', sqlc.arg(access_key), sqlc.arg(access_secret_hash),
	sqlc.arg(username), sqlc.arg(password_hash)
);

-- name: LookupAdminByUsername :one
SELECT id, display_name, role, access_key, access_secret_hash, username, password_hash
FROM users
WHERE username = sqlc.arg(username) AND role = 'admin'
LIMIT 1;

-- name: LookupUserByAccessKey :one
SELECT id, display_name, role, access_key, access_secret_hash, username, password_hash, disabled_at, last_seen_at, created_at, updated_at
FROM users
WHERE access_key = sqlc.arg(access_key)
LIMIT 1;

-- name: CreateWorkspace :one
INSERT INTO workspaces (id, name, slug)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(slug))
ON CONFLICT (slug) DO NOTHING
RETURNING id;

-- name: GetWorkspaceIDBySlug :one
SELECT id
FROM workspaces
WHERE slug = sqlc.arg(slug)
LIMIT 1;

-- name: GetProjectIDByWorkspaceAndName :one
SELECT id
FROM projects
WHERE workspace_id = sqlc.arg(workspace_id) AND name = sqlc.arg(name)
LIMIT 1;

-- name: CreateProject :one
INSERT INTO projects (id, workspace_id, name)
VALUES (sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(name))
ON CONFLICT (workspace_id, name) DO NOTHING
RETURNING id;

-- name: InsertBootstrapUser :exec
INSERT INTO users (
	id, display_name, role, access_key, access_secret_hash
) VALUES (
	sqlc.arg(id), sqlc.arg(display_name), 'user', sqlc.arg(access_key), sqlc.arg(access_secret_hash)
);

-- name: InsertUserProjectAccess :exec
INSERT INTO user_project_access (
	user_id, workspace_id, project_id
) VALUES (
	sqlc.arg(user_id), sqlc.arg(workspace_id), sqlc.arg(project_id)
);

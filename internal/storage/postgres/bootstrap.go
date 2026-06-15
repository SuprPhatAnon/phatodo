package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAdminAlreadyExists      = errors.New("admin user already exists")
	ErrProjectAlreadyExists    = errors.New("project already exists")
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
)

func (s *Store) InitAdmin(ctx context.Context, req domain.AdminInitRequest) (domain.AdminInitResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("begin admin init tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("lock users table: %w", err)
	}

	var adminCount int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount); err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("count admin users: %w", err)
	}
	if adminCount > 0 {
		return domain.AdminInitResponse{}, ErrAdminAlreadyExists
	}

	userID, err := randomID("usr")
	if err != nil {
		return domain.AdminInitResponse{}, err
	}
	accessKey, err := randomID("key")
	if err != nil {
		return domain.AdminInitResponse{}, err
	}
	accessSecret, accessSecretHash, err := newSecretAndHash()
	if err != nil {
		return domain.AdminInitResponse{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("hash admin password: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (
			id, display_name, role, access_key, access_secret_hash,
			username, password_hash
		) VALUES ($1, $2, 'admin', $3, $4, $5, $6)
	`, userID, req.Username, accessKey, accessSecretHash, req.Username, string(passwordHash)); err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("insert admin user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("commit admin init: %w", err)
	}

	return domain.AdminInitResponse{
		UserID:       userID,
		Username:     req.Username,
		AccessKey:    accessKey,
		AccessSecret: accessSecret,
	}, nil
}

func (s *Store) BootstrapProject(ctx context.Context, req domain.AdminBootstrapRequest) (domain.AdminBootstrapResponse, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.AdminBootstrapResponse{}, fmt.Errorf("begin admin bootstrap tx: %w", err)
	}
	defer tx.Rollback(ctx)

	admin, err := s.lookupAdmin(ctx, tx, req.Username)
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return domain.AdminBootstrapResponse{}, ErrInvalidAdminCredentials
	}

	workspaceID, err := s.ensureWorkspace(ctx, tx, req.WorkspaceName)
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	projectID, err := s.ensureProject(ctx, tx, workspaceID, req.ProjectName)
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}

	cliUserID, err := randomID("usr")
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	cliAccessKey, err := randomID("key")
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	cliAccessSecret, cliAccessSecretHash, err := newSecretAndHash()
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}

	displayName := req.ProjectName + " CLI"
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (
			id, display_name, role, access_key, access_secret_hash
		) VALUES ($1, $2, 'user', $3, $4)
	`, cliUserID, displayName, cliAccessKey, cliAccessSecretHash); err != nil {
		return domain.AdminBootstrapResponse{}, fmt.Errorf("insert bootstrap cli user: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_project_access (
			user_id, workspace_id, project_id
		) VALUES ($1, $2, $3)
	`, cliUserID, workspaceID, projectID); err != nil {
		return domain.AdminBootstrapResponse{}, fmt.Errorf("insert project access: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.AdminBootstrapResponse{}, fmt.Errorf("commit admin bootstrap: %w", err)
	}

	return domain.AdminBootstrapResponse{
		WorkspaceID:  workspaceID,
		ProjectID:    projectID,
		AccessKey:    cliAccessKey,
		AccessSecret: cliAccessSecret,
	}, nil
}

func (s *Store) lookupAdmin(ctx context.Context, tx pgx.Tx, username string) (domain.User, error) {
	var user domain.User
	err := tx.QueryRow(ctx, `
		SELECT id, display_name, role, access_key, access_secret_hash, username, password_hash
		FROM users
		WHERE username = $1 AND role = 'admin'
		LIMIT 1
	`, username).Scan(
		&user.ID,
		&user.DisplayName,
		&user.Role,
		&user.AccessKey,
		&user.AccessSecretHash,
		&user.Username,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrInvalidAdminCredentials
		}
		return domain.User{}, fmt.Errorf("lookup admin: %w", err)
	}
	return user, nil
}

func (s *Store) ensureWorkspace(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	slug := slugify(name)
	if slug == "" {
		slug = "phatodo"
	}

	id, err := randomID("ws")
	if err != nil {
		return "", err
	}

	var workspaceID string
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (id, name, slug)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO NOTHING
		RETURNING id
	`, id, name, slug).Scan(&workspaceID)
	if err == nil {
		return workspaceID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("insert workspace: %w", err)
	}

	if err := tx.QueryRow(ctx, `SELECT id FROM workspaces WHERE slug = $1`, slug).Scan(&workspaceID); err != nil {
		return "", fmt.Errorf("select workspace: %w", err)
	}
	return workspaceID, nil
}

func (s *Store) ensureProject(ctx context.Context, tx pgx.Tx, workspaceID string, name string) (string, error) {
	var existing string
	if err := tx.QueryRow(ctx, `
		SELECT id
		FROM projects
		WHERE workspace_id = $1 AND name = $2
	`, workspaceID, name).Scan(&existing); err == nil {
		return "", ErrProjectAlreadyExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("check project existence: %w", err)
	}

	id, err := randomID("prj")
	if err != nil {
		return "", err
	}

	var projectID string
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (id, workspace_id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, name) DO NOTHING
		RETURNING id
	`, id, workspaceID, name).Scan(&projectID)
	if err == nil {
		return projectID, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrProjectAlreadyExists
	}
	return "", fmt.Errorf("insert project: %w", err)
}

func randomID(prefix string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(buf), nil
}

func newSecretAndHash() (string, string, error) {
	secret, err := randomID("sec")
	if err != nil {
		return "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hash access secret: %w", err)
	}
	return secret, string(hash), nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	builder.Grow(len(value))
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '_' || r == '-':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

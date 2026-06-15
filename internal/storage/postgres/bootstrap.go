package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
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

	q := s.q.WithTx(tx)

	if err := q.LockUsersTable(ctx); err != nil {
		return domain.AdminInitResponse{}, fmt.Errorf("lock users table: %w", err)
	}

	adminCount, err := q.CountAdminUsers(ctx)
	if err != nil {
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

	username := req.Username
	passwordHashValue := string(passwordHash)
	if err := q.InsertAdminUser(ctx, db.InsertAdminUserParams{
		ID:               userID,
		DisplayName:      req.Username,
		AccessKey:        accessKey,
		AccessSecretHash: accessSecretHash,
		Username:         &username,
		PasswordHash:     &passwordHashValue,
	}); err != nil {
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

	q := s.q.WithTx(tx)

	admin, err := s.lookupAdmin(ctx, q, req.Username)
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return domain.AdminBootstrapResponse{}, ErrInvalidAdminCredentials
	}

	workspaceID, err := s.ensureWorkspace(ctx, q, req.WorkspaceName)
	if err != nil {
		return domain.AdminBootstrapResponse{}, err
	}
	projectID, err := s.ensureProject(ctx, q, workspaceID, req.ProjectName)
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
	if err := q.InsertBootstrapUser(ctx, db.InsertBootstrapUserParams{
		ID:               cliUserID,
		DisplayName:      displayName,
		AccessKey:        cliAccessKey,
		AccessSecretHash: cliAccessSecretHash,
	}); err != nil {
		return domain.AdminBootstrapResponse{}, fmt.Errorf("insert bootstrap cli user: %w", err)
	}

	if err := q.InsertUserProjectAccess(ctx, db.InsertUserProjectAccessParams{
		UserID:      cliUserID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
	}); err != nil {
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

func (s *Store) lookupAdmin(ctx context.Context, q *db.Queries, username string) (domain.User, error) {
	row, err := q.LookupAdminByUsername(ctx, &username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrInvalidAdminCredentials
		}
		return domain.User{}, fmt.Errorf("lookup admin: %w", err)
	}
	user := domain.User{
		ID:               row.ID,
		DisplayName:      row.DisplayName,
		Role:             domain.UserRole(row.Role),
		AccessKey:        row.AccessKey,
		AccessSecretHash: row.AccessSecretHash,
	}
	if row.Username != nil {
		user.Username = *row.Username
	}
	if row.PasswordHash != nil {
		user.PasswordHash = *row.PasswordHash
	}
	return user, nil
}

func (s *Store) ensureWorkspace(ctx context.Context, q *db.Queries, name string) (string, error) {
	slug := slugify(name)
	if slug == "" {
		slug = "phatodo"
	}

	id, err := randomID("ws")
	if err != nil {
		return "", err
	}

	workspaceID, err := q.CreateWorkspace(ctx, db.CreateWorkspaceParams{ID: id, Name: name, Slug: slug})
	if err == nil {
		return workspaceID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("insert workspace: %w", err)
	}

	workspaceID, err = q.GetWorkspaceIDBySlug(ctx, slug)
	if err != nil {
		return "", fmt.Errorf("select workspace: %w", err)
	}
	return workspaceID, nil
}

func (s *Store) ensureProject(ctx context.Context, q *db.Queries, workspaceID string, name string) (string, error) {
	if _, err := q.GetProjectIDByWorkspaceAndName(ctx, db.GetProjectIDByWorkspaceAndNameParams{WorkspaceID: workspaceID, Name: name}); err == nil {
		return "", ErrProjectAlreadyExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("check project existence: %w", err)
	}

	id, err := randomID("prj")
	if err != nil {
		return "", err
	}

	projectID, err := q.CreateProject(ctx, db.CreateProjectParams{ID: id, WorkspaceID: workspaceID, Name: name})
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

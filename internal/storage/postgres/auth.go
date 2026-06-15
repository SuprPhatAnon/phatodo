package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidAccessCredentials = errors.New("invalid access credentials")
var ErrUserNotFound = errors.New("user not found")

func (s *Store) ResolveAPIPrincipal(ctx context.Context, accessKey string, accessSecret string) (domain.User, error) {
	row, err := s.q.LookupUserByAccessKey(ctx, accessKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrInvalidAccessCredentials
		}
		return domain.User{}, fmt.Errorf("lookup user by access key: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.AccessSecretHash), []byte(accessSecret)); err != nil {
		return domain.User{}, ErrInvalidAccessCredentials
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
	if row.DisabledAt.Valid {
		user.DisabledAt = &row.DisabledAt.Time
	}
	if row.LastSeenAt.Valid {
		user.LastSeenAt = &row.LastSeenAt.Time
	}
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return user, nil
}

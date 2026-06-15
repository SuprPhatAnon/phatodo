package server

import (
	"context"
	"net/http"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
)

const (
	AccessKeyHeader    = "X-Phatodo-Access-Key"
	AccessSecretHeader = "X-Phatodo-Access-Secret"
	UserIDHeader       = "X-Phatodo-User-ID"
)

type principal struct {
	UserID    string
	Role      domain.UserRole
	AccessKey string
}

type principalContextKey struct{}

func (a *app) withAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessKey := r.Header.Get(AccessKeyHeader)
		accessSecret := r.Header.Get(AccessSecretHeader)
		if accessKey == "" || accessSecret == "" {
			respondError(w, http.StatusUnauthorized, "missing_credentials", "API requests require access key and access secret headers")
			return
		}

		userID := r.Header.Get(UserIDHeader)
		if userID == "" {
			userID = accessKey
		}

		ctx := context.WithValue(r.Context(), principalContextKey{}, principal{
			UserID:    userID,
			Role:      domain.UserRoleUser,
			AccessKey: accessKey,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func principalFromContext(ctx context.Context) (principal, bool) {
	value, ok := ctx.Value(principalContextKey{}).(principal)
	return value, ok
}

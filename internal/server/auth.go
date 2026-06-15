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
	UserID      string
	DisplayName string
	Role        domain.UserRole
	AccessKey   string
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

		principal := principal{
			AccessKey: accessKey,
			Role:      domain.UserRoleUser,
		}
		if a.config.AuthResolver != nil {
			user, err := a.config.AuthResolver.ResolveAPIPrincipal(r.Context(), accessKey, accessSecret)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "invalid_credentials", "access key or access secret is invalid")
				return
			}
			principal.UserID = user.ID
			principal.DisplayName = user.DisplayName
			principal.Role = user.Role
			principal.AccessKey = user.AccessKey
		} else {
			userID := r.Header.Get(UserIDHeader)
			if userID == "" {
				userID = accessKey
			}
			principal.UserID = userID
		}

		ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func principalFromContext(ctx context.Context) (principal, bool) {
	value, ok := ctx.Value(principalContextKey{}).(principal)
	return value, ok
}

package auth

import (
	"context"
	"net/http"

	"github.com/sokol/system-control/internal/pkg/httputil"
)

type contextKey string

const userContextKey contextKey = "user"

func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			user, err := svc.ValidateSession(cookie.Value)
			if err != nil {
				httputil.Error(w, http.StatusUnauthorized, "session expired")
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}

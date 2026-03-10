package audit

import (
	"net/http"
	"strings"

	"github.com/sokol/system-control/internal/auth"
)

// Middleware logs all mutating API requests (POST, PUT, DELETE) to the audit log.
func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only log mutating requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth endpoints from audit
			if strings.HasPrefix(r.URL.Path, "/api/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			// Wrap response to capture status
			rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			// Only log successful mutations
			if rw.status >= 200 && rw.status < 300 {
				user := auth.UserFromContext(r.Context())
				username := "unknown"
				var userID *int64
				if user != nil {
					username = user.Username
					userID = &user.ID
				}

				action, entityType := parseAction(r.Method, r.URL.Path)
				svc.Log(userID, username, action, entityType, nil, nil, r.RemoteAddr)
			}
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// parseAction extracts action and entity type from method + path.
// e.g. POST /api/nat-groups -> "nat_group.create", "nat_group"
// e.g. DELETE /api/nginx-domains/5 -> "nginx_domain.delete", "nginx_domain"
// e.g. POST /api/nat-groups/3/enable -> "nat_group.enable", "nat_group"
func parseAction(method, path string) (action, entityType string) {
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return "unknown", "unknown"
	}

	// Convert "nat-groups" -> "nat_group"
	entityType = strings.TrimSuffix(strings.ReplaceAll(parts[0], "-", "_"), "s")

	switch {
	case len(parts) >= 3:
		// e.g. nat-groups/3/enable
		action = entityType + "." + parts[2]
	case method == http.MethodPost:
		action = entityType + ".create"
	case method == http.MethodPut:
		action = entityType + ".update"
	case method == http.MethodDelete:
		action = entityType + ".delete"
	default:
		action = entityType + "." + strings.ToLower(method)
	}

	return action, entityType
}

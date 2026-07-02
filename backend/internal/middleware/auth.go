package middleware

import (
	"context"
	"net/http"
	"strings"

	"expense-tracker/backend/internal/auth"
	"expense-tracker/backend/internal/models"
)

type ctxKey string

const (
	ctxUserID ctxKey = "userID"
	ctxRole   ctxKey = "role"
)

// AuthError writes a JSON error; assigned by main to avoid an import cycle.
var writeError = func(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}

// RequireAuth validates the Bearer access token and injects the user id/role.
func RequireAuth(tm *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := tm.ParseAccessToken(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, claims.Subject)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin must be chained after RequireAuth; it rejects non-admins.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RoleFromContext(r.Context()) != models.RoleAdmin {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserIDFromContext returns the authenticated user id, or "" if unauthenticated.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxUserID).(string); ok {
		return v
	}
	return ""
}

// RoleFromContext returns the authenticated user's role.
func RoleFromContext(ctx context.Context) models.Role {
	if v, ok := ctx.Value(ctxRole).(models.Role); ok {
		return v
	}
	return ""
}

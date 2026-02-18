package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

// Middleware wraps handlers with authentication
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/health" || r.URL.Path == "/api/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Try Authorization header
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if apiKey == "" {
			respondError(w, http.StatusUnauthorized, "Missing API key")
			return
		}

		// Verify API key
		user, err := a.Verify(apiKey)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		// Check permissions based on method
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			// Write operations
			if !user.CanWrite() {
				respondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}
		} else {
			// Read operations
			if !user.CanRead() {
				respondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin checks if user is admin
func (a *Auth) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil || !user.IsAdmin() {
			respondError(w, http.StatusForbidden, "Admin access required")
			return
		}
		next(w, r)
	}
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"success":false,"error":"` + message + `"}`))
}
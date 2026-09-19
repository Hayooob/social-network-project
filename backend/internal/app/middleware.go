package app

import (
	"context"
	"net/http"

	"social-network/internal/models"
)

type contextKey string

const userContextKey contextKey = "currentUser"

// attaches the current user (if any) to the request context
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			// No session cookie -> continue as anonymous
			next.ServeHTTP(w, r)
			return
		}

		token := cookie.Value

		// Use your app-level auth helper (which is in the same package app)
		user, err := GetUserBySessionToken(r.Context(), s.DB, token)
		if err != nil || user == nil {
			// Invalid/expired session → treat as anonymous
			next.ServeHTTP(w, r)
			return
		}

		// Attach user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CurrentUser returns the User from context, or nil
func CurrentUser(ctx context.Context) *models.User {
	val := ctx.Value(userContextKey)
	if val == nil {
		return nil
	}
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}

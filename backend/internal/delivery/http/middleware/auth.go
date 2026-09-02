package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CORS preflight
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Public routes - no auth required
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/auth/") || path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Protected routes - auth required
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]

		userID := extractUserIDFromToken(token)
		if userID == uuid.Nil {
			http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

func extractUserIDFromToken(token string) uuid.UUID {
	id, err := uuid.Parse(token)
	if err != nil {
		return uuid.Nil
	}
	return id
}

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

		// TODO: Implement proper JWT validation
		// For now, extract user ID from token (simplified)
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
	// Simplified token extraction
	// In production, use proper JWT library to validate and extract claims
	parts := strings.Split(token, "-")
	if len(parts) >= 2 {
		if id, err := uuid.Parse(parts[1]); err == nil {
			return id
		}
	}
	return uuid.Nil
}

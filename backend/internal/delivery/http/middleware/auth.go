package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/PresensiGo/backend/internal/auth"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

var jwtService *auth.JWTService

func InitJWT(secret string, expireHour int) {
	jwtService = auth.NewJWTService(secret, expireHour)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path
		if strings.HasPrefix(path, "/api/auth/") || path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

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

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

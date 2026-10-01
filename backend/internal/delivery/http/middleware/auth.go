package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/PresensiGo/backend/internal/auth"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

var jwtService *auth.JWTService

func InitJWT(secret string, expireHour int) {
	jwtService = auth.NewJWTService(secret, expireHour)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		
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
			log.Printf("[%s] Authentication failed: missing authorization header for %s %s", requestID, r.Method, path)
			http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[%s] Authentication failed: invalid authorization format for %s %s", requestID, r.Method, path)
			http.Error(w, `{"error": "invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			log.Printf("[%s] Authentication failed: invalid or expired token for %s %s", requestID, r.Method, path)
			http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

func GetRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(RoleKey).(string); ok {
		return role
	}
	return "employee" // default role
}

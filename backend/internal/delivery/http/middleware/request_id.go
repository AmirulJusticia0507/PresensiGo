package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKey string

const RequestIDKey requestIDKey = "requestID"

// RequestIDMiddleware generates a unique request ID for each request
// If X-Request-ID header is provided, it uses that; otherwise generates a UUID
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "req_" + uuid.New().String()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

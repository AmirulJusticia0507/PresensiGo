package middleware

import (
	"fmt"
	"net/http"
	"time"

	redisc "github.com/PresensiGo/backend/internal/cache"
)

func RateLimitMiddleware(redisClient *redisc.Client, maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("ratelimit:%s", r.RemoteAddr)

			count, err := redisClient.Incr(r.Context(), key)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				redisClient.Expire(r.Context(), key, window)
			}

			if count > int64(maxRequests) {
				http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
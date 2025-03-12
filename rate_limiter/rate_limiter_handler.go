package rate_limiter

import (
	"net/http"
)

func RateLimitedHandler(rl *RateLimiterConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr // Identify clients by IP

		if !rl.AllowRequest(clientIP) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next(w, r) // Proceed with the original handler
	}
}

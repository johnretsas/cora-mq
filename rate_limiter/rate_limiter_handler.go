package rate_limiter

import (
	"fmt"
	"net"
	"net/http"
)

// ExtractClientIP extracts the client IP from the RemoteAddr
func ExtractClientIP(r *http.Request) string {
	// RemoteAddr is in the format IP:Port, so split by colon
	ip := r.RemoteAddr
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		// Handle error if SplitHostPort fails
		fmt.Println("Error extracting client IP:", err)
		return ""
	}
	return host
}

func RateLimitedHandler(rl *RateLimiterConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := ExtractClientIP(r)

		if !rl.AllowRequest(clientIP, false) {
			fmt.Println("Rate limit exceeded for IP:", clientIP)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next(w, r) // Proceed with the original handler
	}
}

package api

import (
	"fmt"
	"go-queue-service/queue_server"
	"go-queue-service/rate_limiter"
	"net/http"
	"strings"
)

type RouteConfig struct {
	Path        string
	Handler     http.HandlerFunc
	Method      string
	RateLimiter *rate_limiter.RateLimiterConfig
	Description string
}

func GetAPIConfig(rateLimiter *rate_limiter.RateLimiterConfig, server *queue_server.QueueServer) []RouteConfig {
	return []RouteConfig{
		{
			Path:        "/health",
			Handler:     server.HealthCheckHandler,
			Method:      "GET",
			RateLimiter: nil, // Health check might not need rate-limiting
			Description: "Health check endpoint",
		},
		{
			Path:        "/sizeOfQueue",
			Handler:     server.SizeOfQueueHandler,
			Method:      "GET",
			RateLimiter: rateLimiter,
			Description: "Returns the current size of the queue",
		},
		{
			Path:        "/createQueue",
			Handler:     server.CreateQueueHandler,
			Method:      "POST",
			RateLimiter: rateLimiter,
			Description: "Creates a new queue",
		},
		{
			Path:        "/enqueue",
			Handler:     server.EnqueueHandler,
			Method:      "POST",
			RateLimiter: rateLimiter,
			Description: "Enqueues an item to the queue",
		},
		{
			Path:        "/enqueue/batch",
			Handler:     server.EnqueueBatchHandler,
			Method:      "POST",
			RateLimiter: rateLimiter,
			Description: "Enqueues multiple items to the queue in a batch",
		},
		{
			Path:        "/dequeue",
			Handler:     server.DequeueHandler,
			Method:      "POST",
			RateLimiter: rateLimiter,
			Description: "Dequeues an item from the queue",
		},
		{
			Path:        "/acknowledge",
			Handler:     server.AcknowledgeHandler,
			Method:      "POST",
			RateLimiter: rateLimiter,
			Description: "Acknowledges a dequeued item",
		},
		{
			Path:        "/scan",
			Handler:     server.ScanHandler,
			Method:      "GET",
			RateLimiter: rateLimiter,
			Description: "Scans the queue and returns at most 100 items",
		},
	}
}

func SetupRoutes(config []RouteConfig) *http.ServeMux {
	mux := http.NewServeMux()

	const colWidth = 30
	totalWidth := colWidth*2 + 2
	fmt.Printf("%-*s %-*s\n", colWidth, "Route Path", colWidth, "Description")
	fmt.Println(strings.Repeat("-", totalWidth))

	for _, route := range config {
		fmt.Printf("%-*s %-*s\n", colWidth, route.Path, colWidth, route.Description)

		// Build the handler chain
		handler := route.Handler

		// Apply rate limiting if configured
		if route.RateLimiter != nil {
			handler = rate_limiter.RateLimitedHandler(route.RateLimiter, handler)
		}

		// Wrap with method enforcement
		handler = enforceMethod(route.Method, handler)

		// Register on our explicit mux
		mux.HandleFunc(route.Path, handler)
	}

	return mux
}

// Helper function for method enforcement
func enforceMethod(expectedMethod string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if expectedMethod != "" && !strings.EqualFold(r.Method, expectedMethod) {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

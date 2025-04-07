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
			RateLimiter: nil, // Health check might not need rate-limiting
			Description: "Health check endpoint",
		},
		{
			Path:        "/sizeOfQueue",
			Handler:     server.SizeOfQueueHandler,
			RateLimiter: rateLimiter,
			Description: "Returns the current size of the queue",
		},
		{
			Path:        "/createQueue",
			Handler:     server.CreateQueueHandler,
			RateLimiter: rateLimiter,
			Description: "Creates a new queue",
		},
		{
			Path:        "/enqueue",
			Handler:     server.EnqueueHandler,
			RateLimiter: rateLimiter,
			Description: "Enqueues an item to the queue",
		},
		{
			Path:        "/enqueue/batch",
			Handler:     server.EnqueueBatchHandler,
			RateLimiter: rateLimiter,
			Description: "Enqueues multiple items to the queue in a batch",
		},
		{
			Path:        "/dequeue",
			Handler:     server.DequeueHandler,
			RateLimiter: rateLimiter,
			Description: "Dequeues an item from the queue",
		},
		{
			Path:        "/acknowledge",
			Handler:     server.AcknowledgeHandler,
			RateLimiter: rateLimiter,
			Description: "Acknowledges a dequeued item",
		},
		{
			Path:        "/scan",
			Handler:     server.ScanHandler,
			RateLimiter: rateLimiter,
			Description: "Scans the queue and returns at most 100 items",
		},
	}
}

func SetupRoutes(config []RouteConfig) {
	// Define a column width
	const colWidth = 30
	totalWidth := colWidth*2 + 2

	// Print the header
	fmt.Printf("%-*s %-*s\n", colWidth, "Route Path", colWidth, "Description")
	fmt.Println(strings.Repeat("-", totalWidth))

	// Loop through each route and print its path and description in aligned columns
	for _, route := range config {
		// Use fmt.Printf with width formatting to align the output
		fmt.Printf("%-*s %-*s\n", colWidth, route.Path, colWidth, route.Description)
	}
}

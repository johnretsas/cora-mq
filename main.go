package main

import (
	"fmt"
	"go-queue-service/queue_server"
	"go-queue-service/rate_limiter"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("CORA Queue Service - Version 1.0\t|")
	fmt.Println("========================================|")
	// Read env variable CORA_NUMBER_OF_WORKERS:
	workersEnv := os.Getenv("CORA_NUMBER_OF_WORKERS")
	workers := 40 // default number of workers
	if workersEnv != "" {
		var err error
		workers, err = strconv.Atoi(workersEnv)
		if err != nil {
			fmt.Println("Invalid value for CORA_NUMBER_OF_WORKERS, using default:", workers)
		}
	}

	numOfWorkersMsg := fmt.Sprintln("Number of workers: ", workers)
	logger := log.New(os.Stdout, "QueueServer - "+numOfWorkersMsg, log.LstdFlags)

	// Create a new queue server
	server := queue_server.NewQueueServer(logger, workers)

	// Setting up rate limiter
	rateLimiter := rate_limiter.NewRateLimiterConfig(rate.Limit(100), 200)
	// Set up health check endpoint
	http.HandleFunc("/health", server.HealthCheckHandler)

	// Queue endpoints. Handlers create a request and send it to the request channel for processing
	http.HandleFunc("/createQueue", rate_limiter.RateLimitedHandler(rateLimiter, server.CreateQueueHandler))
	http.HandleFunc("/enqueue", rate_limiter.RateLimitedHandler(rateLimiter, server.EnqueueHandler))
	http.HandleFunc("/enqueue/batch", rate_limiter.RateLimitedHandler(rateLimiter, server.EnqueueBatchHandler))
	http.HandleFunc("/dequeue", rate_limiter.RateLimitedHandler(rateLimiter, server.DequeueHandler))
	http.HandleFunc("/acknowledge", rate_limiter.RateLimitedHandler(rateLimiter, server.AcknowledgeHandler))
	http.HandleFunc("/scan", rate_limiter.RateLimitedHandler(rateLimiter, server.ScanHandler))

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	serverConfig := &http.Server{
		Addr:         "localhost:" + port,
		Handler:      nil, // Use default handler (http.HandleFunc)
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second, // Optional: Limit idle time for connections
	}

	// Start the server with custom configuration
	fmt.Println("Starting server on port:", port)
	if err := serverConfig.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

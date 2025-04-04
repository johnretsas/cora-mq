package main

import (
	"fmt"
	"go-queue-service/api"
	"go-queue-service/queue_server"
	"go-queue-service/rate_limiter"
	utils "go-queue-service/utils/banner"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	utils.PrintStartupBanner()
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

	numOfWorkersMsg := fmt.Sprint("Number of workers: ", workers)
	logger := log.New(os.Stdout, "[QueueServer - "+numOfWorkersMsg+"]", log.LstdFlags)

	// Create a new queue server
	server := queue_server.NewQueueServer(logger, workers)

	// Setting up rate limiter
	rateLimiter := rate_limiter.NewRateLimiterConfig(rate.Limit(100), 500, logger)

	// Set up the api routes
	apiConfig := api.GetAPIConfig(rateLimiter, server)
	api.SetupRoutes(apiConfig)

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
	fmt.Printf("\033[1;34mStarting server on port: %s\033[0m\n", port)
	fmt.Printf("\033[1;34mNumber of workers: %d\033[0m\n", workers)
	if err := serverConfig.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

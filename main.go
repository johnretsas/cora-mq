package main

import (
	"fmt"
	"go-queue-service/api"
	"go-queue-service/queue_server"
	"go-queue-service/rate_limiter"
	utils "go-queue-service/utils/banner"
	"go-queue-service/utils/logger"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	utils.PrintStartupBanner()

	// Read env variable CORA_NUMBER_OF_WORKERS
	workersEnv := os.Getenv("CORA_NUMBER_OF_WORKERS")
	workers := 40 // default number of workers
	if workersEnv != "" {
		var err error
		workers, err = strconv.Atoi(workersEnv)
		if err != nil {
			fmt.Println("Invalid value for CORA_NUMBER_OF_WORKERS, using default:", workers)
		}
	}

	// Create new structured logger for queue server
	queueLogger := logger.New("QueueServer", logger.INFO).
		WithColors(true).
		WithDevelopmentMode(false)

	queueLogger.Info("initializing server - %s", logger.WithField("workers", workers))

	// Create a new queue server
	server := queue_server.NewQueueServer(queueLogger, workers)

	// Setting up rate limiter (still uses standard logger for now)
	stdLogger := log.New(os.Stdout, "[RateLimiter] ", log.LstdFlags)
	rateLimiter := rate_limiter.NewRateLimiterConfig(rate.Limit(5000), 10000, stdLogger)

	// Set up the api routes
	apiConfig := api.GetAPIConfig(rateLimiter, server)
	queueServerAPI_mux := api.SetupRoutes(apiConfig)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	httpServerConfig := &http.Server{
		Addr:         "127.0.0.1:" + port,
		Handler:      queueServerAPI_mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second, // Optional: Limit idle time for connections
	}

	// Start the server with custom configuration
	queueLogger.Info("starting HTTP server - %s",
		logger.WithFields(map[string]interface{}{
			"port":    port,
			"workers": workers,
			"address": httpServerConfig.Addr,
		}))

	if err := httpServerConfig.ListenAndServe(); err != nil {
		queueLogger.Error("server failed - %s", err.Error())
	}
}

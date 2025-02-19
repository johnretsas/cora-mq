package main

import (
	"fmt"
	"go-queue-service/queue_server"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	fmt.Println("CORA Queue Service - Version 1.0")
	fmt.Println("---------------------------------")
	// Read env variable CORA_NUMBER_OF_WORKERS:
	workersEnv := os.Getenv("CORA_NUMBER_OF_WORKERS")
	workers := 3 // default number of workers
	if workersEnv != "" {
		var err error
		workers, err = strconv.Atoi(workersEnv)
		if err != nil {
			fmt.Println("Invalid value for CORA_NUMBER_OF_WORKERS, using default:", workers)
		}
	}

	numOfWorkersMsg := fmt.Sprintln("Number of workers: ", workers)
	logger := log.New(os.Stdout, "QueueServer - "+numOfWorkersMsg, log.LstdFlags)
	server := queue_server.NewQueueServer(logger, workers)

	// Set up health check endpoint
	http.HandleFunc("/health", server.HealthCheckHandler)

	// Queue endpoints. Handlers create a request and send it to the request channel for processing
	http.HandleFunc("/createQueue", server.CreateQueueHandler)
	http.HandleFunc("/enqueue", server.EnqueueHandler)
	http.HandleFunc("/enqueue/batch", server.EnqueueBatchHandler)
	http.HandleFunc("/dequeue", server.DequeueHandler)
	http.HandleFunc("/acknowledge", server.AcknowledgeHandler)
	http.HandleFunc("/scan", server.ScanHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Starting server on port: ", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

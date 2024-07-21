package main

import (
	"fmt"
	"go-queue-service/queue_server"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Hello, from the server!")
	logger := log.New(os.Stdout, "QueueServer: ", log.LstdFlags)
	server := queue_server.NewQueueServer(logger)

	// Set up health check endpoint
	http.HandleFunc("/health", server.HealthCheckHandler)

	// Queue endpoints
	http.HandleFunc("/create_queue", server.CreateQueueHandler)
	http.HandleFunc("/enqueue", server.EnqueueHandler)
	http.HandleFunc("/dequeue", server.DequeueHandler)
	http.HandleFunc("/acknowledge", server.AcknowledgeHandler)
	http.HandleFunc("/scan", server.ScanHandler)

	fmt.Println("Starting server on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

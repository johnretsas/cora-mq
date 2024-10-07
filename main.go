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
	http.HandleFunc("/createQueue", server.CreateQueueHandler)
	http.HandleFunc("/enqueue", server.EnqueueHandler)
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

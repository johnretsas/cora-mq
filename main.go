package main

import (
	"fmt"
	"go-queue-service/queue_server"
	"net/http"
)

func main() {
	fmt.Println("Hello, from the server!")
	server := queue_server.NewQueueServer()

	http.HandleFunc("/enqueue", server.EnqueueHandler)
	http.HandleFunc("/dequeue", server.DequeueHandler)

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

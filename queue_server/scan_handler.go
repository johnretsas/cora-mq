package queue_server

import (
	"encoding/json"
	"fmt"
	"go-queue-service/queue"
	"log"
	"net/http"
)

func (queueServer *QueueServer) ScanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: "Method not allowed",
		})
		return
	}
	// Assuming you have a queue instance named 'q'
	var requestBody struct {
		QueueName string `json:"queueName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})

		return
	}

	if requestBody.QueueName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: "Queue name is required as queueName in the request body",
		})

		return
	}

	basicQueueItems, deadLetterQueueItems, err := queueServer.Scan(requestBody.QueueName)
	fmt.Println("Items scanned successfully", basicQueueItems, deadLetterQueueItems)
	if err != nil {
		log.Println("Error scanning the queue:", err)
		// You can handle the error here, e.g., return an appropriate HTTP response
		json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}

	// Handle the successful scan here, e.g., return an appropriate HTTP response

	// Or you can return the items as a JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Message              string            `json:"message"`
		QueueItems           []queue.QueueItem `json:"items"`
		DeadLetterQueueItems []queue.QueueItem `json:"deadLetterQueueItems"`
	}{
		Message:              "Items scanned successfully",
		QueueItems:           basicQueueItems,
		DeadLetterQueueItems: deadLetterQueueItems,
	})
}

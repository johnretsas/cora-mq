package queue_server

import (
	"encoding/json"
	"fmt"
	"go-queue-service/queue"
	"log"
	"net/http"
	"io"
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
		if err == io.EOF || err.Error() == "EOF" {
			errorMsg := struct {
				Error string `json:"error"`
			}{
				Error: "Missing body",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorMsg)
			return
		}

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

	basicQueueItems, deadLetterQueueItems, sizeOfQueue, err := queueServer.Scan(requestBody.QueueName)
	fmt.Println("Items scanned successfully")
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
		SizeOfQueue          int               `json:"sizeOfQueue"`
	}{
		Message:              "Items scanned successfully. Fetched at most 100 items",
		QueueItems:           basicQueueItems,
		DeadLetterQueueItems: deadLetterQueueItems,
		SizeOfQueue:          sizeOfQueue,
	})
}

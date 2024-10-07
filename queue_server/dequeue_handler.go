package queue_server

import (
	"encoding/json"
	"fmt"
	"go-queue-service/queue"
	"net/http"
)

func (queueServer *QueueServer) DequeueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Method Not Allowed",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var requestBody struct {
		QueueName string `json:"queueName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.QueueName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Missing queueName", http.StatusBadRequest)
	}

	item, err := queueServer.Dequeue(requestBody.QueueName)

	fmt.Printf("Item: %v\n", item)
	fmt.Printf("Error: %v\n", err)

	if err != nil {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	dequeueMsg := struct {
		Message string          `json:"message"`
		Item    queue.QueueItem `json:"item"`
	}{
		Message: "Item dequeued",
		Item:    item,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dequeueMsg)
}

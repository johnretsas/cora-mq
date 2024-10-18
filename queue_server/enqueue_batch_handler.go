package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
)

func (server *QueueServer) EnqueueBatchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if the request method is POST
	if r.Method != http.MethodPost {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Method not allowed",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var requestBody struct {
		QueueName string            `json:"queueName"`
		Items     []queue.QueueItem `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.QueueName == "" {
		http.Error(w, "Missing queueName", http.StatusBadRequest)
		return
	}

	if len(requestBody.Items) == 0 {
		http.Error(w, "Missing items", http.StatusBadRequest)
		return
	}

	items, err := server.EnqueueBatch(requestBody.QueueName, requestBody.Items)

	if err != nil {
		errorMsg := struct {
			Error     string            `json:"error"`
			Items     []queue.QueueItem `json:"items"`
			QueueName string            `json:"queueName"`
		}{
			Error:     err.Error(),
			Items:     items,
			QueueName: requestBody.QueueName,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	w.WriteHeader(http.StatusCreated)
	enqueueMsg := struct {
		Message string            `json:"message"`
		Items   []queue.QueueItem `json:"items"`
	}{
		Message: "Items enqueued",
		Items:   items,
	}

	json.NewEncoder(w).Encode(enqueueMsg)
}

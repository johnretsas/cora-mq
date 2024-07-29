package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
)

func (queueServer *QueueServer) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var requestBody struct {
		QueueName string          `json:"queueName"`
		Item      queue.QueueItem `json:"item"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.QueueName == "" {
		http.Error(w, "Missing queueName", http.StatusBadRequest)
		return
	}

	item, err := queueServer.Enqueue(requestBody.QueueName, requestBody.Item)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enqueueMsg := struct {
		Message string `json:"message"`
		Id      string `json:"id"`
	}{
		Message: "Item enqueued",
		Id:      item.ID,
	}

	json.NewEncoder(w).Encode(enqueueMsg)
}

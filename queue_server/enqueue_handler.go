package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
)

func (queueServer *QueueServer) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
		QueueName string          `json:"queueName"`
		Item      queue.QueueItem `json:"item"`
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

	item, err := queueServer.Enqueue(requestBody.QueueName, requestBody.Item)

	if err != nil {
		errorMsg := struct {
			Error     string `json:"error"`
			Item      string `json:"item"`
			QueueName string `json:"queueName"`
		}{
			Error:     err.Error(),
			Item:      item.ID,
			QueueName: requestBody.QueueName,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

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

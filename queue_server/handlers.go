package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
)

func (queueServer *QueueServer) CreateQueueHandler(w http.ResponseWriter, r *http.Request) {
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

	var queueName struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&queueName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueServer.CreateQueue(queueName.Name)
	w.WriteHeader(http.StatusCreated)
}

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.QueueName == "" {
		http.Error(w, "Missing queueName", http.StatusBadRequest)
		return
	}

	queueServer.Enqueue(requestBody.QueueName, requestBody.Item)
	w.WriteHeader(http.StatusCreated)
	enqueueMsg := struct {
		Message string `json:"message"`
		Id      string `json:"id"`
	}{
		Message: "Item enqueued",
		Id:      requestBody.Item.ID,
	}

	json.NewEncoder(w).Encode(enqueueMsg)
}

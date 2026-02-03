package queue_server

import (
	"encoding/json"
	"go-queue-service/utils/logger"
	"net/http"
)

func (qs *QueueServer) SizeOfQueueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Query().Get("queueName")
	if queueName == "" {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Missing queueName parameter",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	qs.mu.Lock()
	defer qs.mu.Unlock()
	queue, exists := qs.queues[queueName]

	if !exists {
		qs.logger.Error("queue not found - %s", logger.WithField("queue", queueName))
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Queue not found",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	size := queue.Len()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		QueueName string `json:"queueName"`
		Size      int    `json:"size"`
	}{
		QueueName: queueName,
		Size:      size,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		qs.logger.Error("failed to encode response - %s", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

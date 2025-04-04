package queue_server

import (
	"encoding/json"
	"net/http"
)

func (qs *QueueServer) SizeOfQueueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Query().Get("queueName")
	if queueName == "" {
		http.Error(w, "Missing queueName parameter", http.StatusBadRequest)
		return
	}

	qs.mu.Lock()
	defer qs.mu.Unlock()
	queue, exists := qs.queues[queueName]

	if !exists {
		qs.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		http.Error(w, "Queue not found", http.StatusNotFound)
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
		qs.logger.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

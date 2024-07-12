package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
)

type QueueServer struct {
	queue *queue.Queue
}

func NewQueueServer() *QueueServer {
	return &QueueServer{
		queue: queue.NewQueue(),
	}
}

func (queueServer *QueueServer) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	var item queue.QueueItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueServer.queue.Enqueue(item)
	w.WriteHeader(http.StatusCreated)
}

func (queueServer *QueueServer) DequeueHandler(w http.ResponseWriter, r *http.Request) {
	item, err := queueServer.queue.Dequeue()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

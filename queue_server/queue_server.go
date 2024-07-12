package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"log"
	"net/http"
	"sync"
)

type QueueServer struct {
	queues map[string]*queue.Queue
	logger *log.Logger
	mu     sync.Mutex
}

func NewQueueServer(logger *log.Logger) *QueueServer {
	return &QueueServer{
		queues: make(map[string]*queue.Queue),
		logger: logger,
	}
}

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

// func (queueServer *QueueServer) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		errorMsg := struct {
// 			Error string `json:"error"`
// 		}{
// 			Error: "Method not allowed",
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		json.NewEncoder(w).Encode(errorMsg)
// 		return
// 	}

// 	var item queue.QueueItem
// 	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	queueServer.queue.Enqueue(item)
// 	w.WriteHeader(http.StatusCreated)
// }

// func (queueServer *QueueServer) DequeueHandler(w http.ResponseWriter, r *http.Request) {
// 	item, err := queueServer.queue.Dequeue()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	if err := json.NewEncoder(w).Encode(item); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

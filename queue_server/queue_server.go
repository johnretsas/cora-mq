package queue_server

import (
	"go-queue-service/queue"
	"log"
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

func (queueServer *QueueServer) CreateQueue(queueName string) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	if _, exists := queueServer.queues[queueName]; exists {
		queueServer.logger.Printf("Queue with name '%s' already exists\n", queueName)
		return
	}

	queueServer.logger.Printf("Creating queue: %s\n", queueName)
	queueServer.queues[queueName] = queue.NewQueue()
}

func (queueServer *QueueServer) Enqueue(queueName string, item queue.QueueItem) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return
	}

	q.Enqueue(item)
}

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

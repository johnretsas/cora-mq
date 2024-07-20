package queue_server

import (
	"fmt"
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

func (queueServer *QueueServer) Enqueue(queueName string, item queue.QueueItem) (queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	q.Enqueue(item)
	return item, nil
}

func (queueServer *QueueServer) Dequeue(queueName string) (queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	item, err := q.Dequeue()
	if err != nil {
		queueServer.logger.Printf("Error with dequeueing from queue: '%s'\n", queueName)
		return queue.QueueItem{}, err
	}

	return *item, nil
}

func (queueServer *QueueServer) Acknowledge(queueName string, id string) error {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return fmt.Errorf("queue '%s' does not exist", queueName)
	}

	err := q.Acknowledge(id)
	if err != nil {
		queueServer.logger.Printf("Error acknowledging message with id: '%s'\n", id)
		return err
	}

	return nil
}

func (queueServer *QueueServer) Scan(queueName string) ([]queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return []queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	items := q.Scan()
	return items, nil
}

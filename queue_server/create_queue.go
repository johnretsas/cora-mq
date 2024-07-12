package queue_server

import (
	"go-queue-service/queue"
)

func (queueServer *QueueServer) CreateQueue(queueName string) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	if _, ok := queueServer.queues[queueName]; ok {
		queueServer.logger.Printf("Queue with name '%s' already exists\n", queueName)
		return
	}

	queueServer.logger.Printf("Creating queue: %s\n", queueName)
	queueServer.queues[queueName] = queue.NewQueue()
}

package queue_server

import (
	"fmt"
	"go-queue-service/queue"
)

func (queueServer *QueueServer) ProcessCreateQueueRequest(req Request) {
	queueServer.logger.Println("Handling CreateQueueRequest")
	queueName := req.QueueName
	// Handle CreateQueueRequest
	if _, exists := queueServer.queues[req.QueueName]; exists {
		// Queue already exists
		queueServer.logger.Printf("Queue with name '%s' already exists\n", queueName)
		msg := CreateQueueResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' already exists", queueName)},
			QueueName:    req.QueueName,
		}
		req.ResponseCh <- msg
	} else {
		// Create new queue
		queueServer.logger.Printf("Creating queue: %s\n", req.QueueName)
		queueServer.queues[req.QueueName] = queue.NewQueue()

		msg := CreateQueueResponse{
			BaseResponse: BaseResponse{Message: "Queue created successfully"},
			QueueName:    req.QueueName,
		}
		req.ResponseCh <- msg
	}
}

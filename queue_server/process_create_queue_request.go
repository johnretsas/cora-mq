package queue_server

import "go-queue-service/queue"

func (queueServer *QueueServer) ProcessCreateQueueRequest(req Request) {
	queueServer.logger.Println("Handling CreateQueueRequest")

	// Handle CreateQueueRequest
	if _, exists := queueServer.queues[req.QueueName]; exists {
		// Queue already exists
		queueServer.logger.Printf("Queue with name '%s' already exists\n", req.QueueName)
		msg := CreateQueueResponse{
			BaseResponse: BaseResponse{Message: "Queue already exists"},
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

package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"go-queue-service/utils/logger"
)

func (queueServer *QueueServer) ProcessCreateQueueRequest(req Request) {
	queueName := req.QueueName
	queueConfig := req.QueueConfig

	queueServer.logger.Debug("processing create queue request - %s", logger.WithField("queue", queueName))

	// Handle CreateQueueRequest
	if _, exists := queueServer.queues[req.QueueName]; exists {
		// Queue already exists
		queueServer.logger.Warn("queue already exists - %s", logger.WithField("queue", queueName))
		msg := CreateQueueResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' already exists", queueName)},
			QueueName:    req.QueueName,
			QueueConfig:  queueConfig,
		}
		req.ResponseCh <- msg
	} else {
		// Create new queue
		queueServer.logger.Info("creating queue - %s", logger.WithField("queue", queueName))

		queueServer.queues[req.QueueName] = queue.NewQueue(queueConfig, queueName)

		msg := CreateQueueResponse{
			BaseResponse: BaseResponse{Message: "Queue created successfully"},
			QueueName:    req.QueueName,
		}
		req.ResponseCh <- msg
	}
}

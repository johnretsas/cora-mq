package queue_server

import (
	"fmt"
	"go-queue-service/utils/logger"
)

func (queueServer *QueueServer) ProcessEnqueueBatchRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]

	queueServer.logger.Debug("processing batch enqueue request - %s",
		logger.WithFields(map[string]interface{}{
			"queue": queueName,
			"count": len(req.Items),
		}))

	if !exists {
		queueServer.logger.Error("queue not found - %s", logger.WithField("queue", queueName))

		msg := EnqueueBatchResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
			Items:        req.Items,
		}
		req.ResponseCh <- msg
	} else {
		items := req.Items
		q.EnqueueBatch(items)

		queueServer.logger.Info("enqueued batch - %s",
			logger.WithFields(map[string]interface{}{
				"queue": queueName,
				"count": len(items),
			}))

		msg := EnqueueBatchResponse{
			BaseResponse: BaseResponse{Message: "Items enqueued successfully"},
			QueueName:    queueName,
			Items:        items,
		}

		req.ResponseCh <- msg
	}
}

// TODO: ENABLE LONG POLLING FOR BATCH ENQUEUE

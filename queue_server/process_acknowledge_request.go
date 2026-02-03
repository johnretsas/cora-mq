package queue_server

import (
	"fmt"
	"go-queue-service/utils/logger"
)

func (queueServer *QueueServer) ProcessAcknowledgeRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]

	queueServer.logger.Debug("processing acknowledge request - %s",
		logger.WithFields(map[string]interface{}{
			"queue":  queueName,
			"itemId": req.Item.ID,
		}))

	if !exists {
		queueServer.logger.Error("queue not found - %s", logger.WithField("queue", queueName))

		msg := AcknowledgeResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
		}

		req.ResponseCh <- msg
	} else {
		err := q.Acknowledge(req.Item.ID)
		if err != nil {
			queueServer.logger.Error("acknowledge failed - %s",
				logger.WithFields(map[string]interface{}{
					"queue":  queueName,
					"itemId": req.Item.ID,
					"error":  err.Error(),
				}))

			msg := AcknowledgeResponse{
				BaseResponse: BaseResponse{Error: err},
			}

			req.ResponseCh <- msg
		} else {
			queueServer.logger.Info("acknowledged item - %s",
				logger.WithFields(map[string]interface{}{
					"queue":  queueName,
					"itemId": req.Item.ID,
				}))

			msg := AcknowledgeResponse{
				BaseResponse: BaseResponse{Message: "Message acknowledged successfully"},
				QueueName:    queueName,
				ID:           req.Item.ID,
			}

			req.ResponseCh <- msg
		}
	}
}

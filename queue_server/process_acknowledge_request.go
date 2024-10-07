package queue_server

import (
	"fmt"
)

func (queueServer *QueueServer) ProcessAcknowledgeRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		// return an error

		msg := AcknowledgeResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
		}

		req.ResponseCh <- msg
	} else {
		queueServer.logger.Printf("Attempt to acknowledge item with id: %s\n", req.Item.ID)
		err := q.Acknowledge(req.Item.ID)
		if err != nil {
			queueServer.logger.Printf("Error acknowledging message with id: '%s'\n", req.Item.ID)
			// return an error
			msg := AcknowledgeResponse{
				BaseResponse: BaseResponse{Error: err},
			}

			req.ResponseCh <- msg
		} else {
			msg := AcknowledgeResponse{
				BaseResponse: BaseResponse{Message: "Message acknowledged successfully"},
				QueueName:    queueName,
				ID:           req.Item.ID,
			}

			req.ResponseCh <- msg
		}
	}
}

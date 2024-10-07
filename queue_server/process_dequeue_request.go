package queue_server

import (
	"fmt"
)

func (queueServer *QueueServer) ProcessDequeueRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		// return an error

		msg := DequeueResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
		}

		req.ResponseCh <- msg

	} else {
		queueServer.logger.Printf("Dequeueing item to queue: %s\n", queueName)
		item, err := q.Dequeue()
		if err != nil {
			queueServer.logger.Printf("Error with dequeueing from queue: '%s'\n", queueName)
			// return an error
			msg := DequeueResponse{
				BaseResponse: BaseResponse{Error: err},
			}

			req.ResponseCh <- msg
		} else {
			msg := DequeueResponse{
				BaseResponse: BaseResponse{Message: "Item dequeued successfully"},
				QueueName:    queueName,
				Item:         *item,
			}

			req.ResponseCh <- msg
		}

	}

}

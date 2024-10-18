package queue_server

import "fmt"

func (queueServer *QueueServer) ProcessEnqueueBatchRequest(req Request) {
	queueServer.logger.Println("Handling EnqueueBatchRequest")
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		// return an error
		msg := EnqueueBatchResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
			Items:        req.Items,
		}
		req.ResponseCh <- msg

	} else {
		queueServer.logger.Printf("Enqueueing items to queue: %s\n", queueName)
		items := req.Items
		q.EnqueueBatch(items)

		msg := EnqueueBatchResponse{
			BaseResponse: BaseResponse{Message: "Items enqueued successfully"},
			QueueName:    queueName,
			Items:        items,
		}

		req.ResponseCh <- msg
	}
}

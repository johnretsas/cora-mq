package queue_server

import "fmt"

func (queueServer *QueueServer) ProcessEnqueueRequest(req Request) {
	queueServer.logger.Println("Handling EnqueueRequest")
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		// return an error
		msg := EnqueueResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
			Item:         req.Item,
		}
		req.ResponseCh <- msg

	} else {
		queueServer.logger.Printf("Enqueueing item to queue: %s\n", queueName)
		item := req.Item
		q.Enqueue(item)

		msg := EnqueueResponse{
			BaseResponse: BaseResponse{Message: "Item enqueued successfully"},
			QueueName:    queueName,
			Item:         item,
		}

		req.ResponseCh <- msg
	}
}

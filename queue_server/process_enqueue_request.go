package queue_server

func (queueServer *QueueServer) ProcessEnqueueRequest(req Request) {
	queueServer.logger.Println("Handling EnqueueRequest")
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		msg := EnqueueQueueResponse{
			BaseResponse: BaseResponse{Message: "Queue does not exist"},
			QueueName:    queueName,
			Item:         req.Item,
		}
		req.ResponseCh <- msg

	} else {
		queueServer.logger.Printf("Enqueueing item to queue: %s\n", queueName)
		item := req.Item
		q.Enqueue(item)

		msg := EnqueueQueueResponse{
			BaseResponse: BaseResponse{Message: "Item enqueued successfully"},
			QueueName:    queueName,
			Item:         item,
		}

		req.ResponseCh <- msg
	}

}

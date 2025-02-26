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
		item := req.Item
		// Add the item to the queue
		q.Enqueue(item)
		// Create the response message for the Enqueue action
		msg := EnqueueResponse{
			BaseResponse: BaseResponse{Message: "Item enqueued successfully"},
			QueueName:    queueName,
			Item:         item,
		}

		// Since we have a new item we should check if there are any clients waiting for an item
		// If there are clients waiting, we should send the item to the first client in the waiting list
		queueServer.mu.Lock()
		defer queueServer.mu.Unlock()
		if len(queueServer.waitingListClients[queueName]) > 0 {
			queueServer.logger.Printf("Found clients waiting for queue: '%s'\n", queueName)
			// Get the first client in the waiting list
			clientCh := queueServer.waitingListClients[queueName][0]
			// Send the item to the client. The client is waiting in the select statement in the ProcessDequeueRequest method
			// The client can grab the item from the channel and use it to make a DequeueResponse
			clientCh <- &item
			// Close the channel
			close(clientCh)
			// Remove the client from the waiting list
			queueServer.waitingListClients[queueName] = queueServer.waitingListClients[queueName][1:]
		}

		req.ResponseCh <- msg
	}
}

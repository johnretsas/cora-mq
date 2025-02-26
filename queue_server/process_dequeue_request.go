package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"time"
)

func (queueServer *QueueServer) ProcessDequeueRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		// return an error
		queueServer.sendDequeueError(req, fmt.Errorf("queue '%s' does not exist", queueName))
	} else {
		queueServer.logger.Printf("Dequeueing item to queue: %s\n", queueName)
		// if the queue is empty, add the client to the waiting list and wait for an item to be enqueued
		// else dequeue the item and send it to the client
		if q.Len() == 0 {
			queueServer.handleLongPolling(req, queueName)
		} else {
			queueServer.processImmediateDequeue(req, q, queueName)
		}
	}

}

func (queueServer *QueueServer) processImmediateDequeue(req Request, q *queue.Queue, queueName string) {
	item, err := q.Dequeue()
	if err != nil {
		queueServer.sendDequeueError(req, err)
	} else {
		queueServer.sendDequeueSuccess(req, item, queueName)
	}
}

func (queueServer *QueueServer) handleLongPolling(req Request, queueName string) {
	queueServer.logger.Printf("Queue with name '%s' is empty, client is long polling...\n", queueName)

	// create a channel to send the item to the client
	clientCh := make(chan *queue.QueueItem, 1)

	// add the client to the waiting list - lock the mutex so that the waiting list can be updated safely
	queueServer.mu.Lock()
	// append the client to the waiting list
	queueServer.waitingListClients[queueName] = append(queueServer.waitingListClients[queueName], clientCh)
	// unlock the mutex so that other clients can be added to the waiting list
	queueServer.mu.Unlock()

	// wait for an item to be enqueued or a timeout - the first one to happen will be sent to the response channel.
	select {
	case item := <-clientCh:
		queueServer.sendDequeueSuccess(req, item, queueName)
	case <-time.After(30 * time.Second):
		queueServer.sendDequeueError(req, fmt.Errorf("timeout waiting for item in queue '%s'", queueName))
	}
}

func (queueServer *QueueServer) sendDequeueSuccess(req Request, item *queue.QueueItem, queueName string) {
	queueServer.logger.Printf("Item successfully dequeued from queue '%s'\n", queueName)
	msg := DequeueResponse{
		BaseResponse: BaseResponse{Message: "Item dequeued successfully"},
		QueueName:    queueName,
		Item:         *item,
	}

	req.ResponseCh <- msg
}

func (queueServer *QueueServer) sendDequeueError(req Request, err error) {
	queueServer.logger.Printf("Error dequeuing item: %s\n", err)
	msg := DequeueResponse{
		BaseResponse: BaseResponse{Error: err},
	}

	req.ResponseCh <- msg
}

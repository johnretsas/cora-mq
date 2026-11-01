package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"go-queue-service/utils/logger"
	"slices"
	"time"
)

func (queueServer *QueueServer) ProcessDequeueRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.getQueue(queueName)

	queueServer.logger.Debug("processing dequeue request - %s", logger.WithField("queue", queueName))

	if !exists {
		queueServer.logger.Error("queue not found - %s", logger.WithField("queue", queueName))
		queueServer.sendDequeueError(req, fmt.Errorf("queue '%s' does not exist", queueName))
	} else {
		// if the queue is empty, add the client to the waiting list and wait for an item to be enqueued
		// else dequeue the item and send it to the client
		// lock the mutex so that the queue can be accessed safely
		q.Lock()
		if q.Len() == 0 || q.LenOfVisibleItems() == 0 {
			q.Unlock()
			queueServer.handleLongPolling(req, queueName)
		} else {
			q.Unlock()
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
	queueServer.logger.Info("queue empty, long polling - %s", logger.WithField("queue", queueName))

	// create a channel to send the item to the client
	clientCh := make(chan *queue.QueueItem, 1)

	// add the client to the waiting list - lock the mutex so that the waiting list can be updated safely
	queueServer.mu.Lock()
	// append the client to the waiting list
	queueServer.waitingListClients[queueName] = append(queueServer.waitingListClients[queueName], clientCh)
	// unlock the mutex so that other clients can be added to the waiting list
	queueServer.mu.Unlock()

	// wait for an item to be enqueued or a timeout - the first one to happen will be sent to the response channel.
	go func() {
		select {
		case item := <-clientCh:
			queueServer.sendDequeueSuccess(req, item, queueName)
		case <-time.After(30 * time.Second):
			queueServer.mu.Lock()

			// find clientCh in the waiting list and remove
			idx := slices.Index(queueServer.waitingListClients[queueName], clientCh)

			if idx >= 0 {
				queueServer.waitingListClients[queueName] = slices.Delete(queueServer.waitingListClients[queueName], idx, idx+1)
				queueServer.mu.Unlock()

				queueServer.sendDequeueLongPollingTimeout(req, queueName)
			} else {
				// if we didn't find the clientCh is because a process sent a msg to clientCh and
				// the client was removed in the time between we go into this case and actually
				// check. This means that we should send dequeue success. The item is sent inside the clientCh
				// channel, so we have access to it
				queueServer.mu.Unlock()
				queueServer.sendDequeueSuccess(req, <-clientCh, queueName)
			}
		}
	}()
}

func (queueServer *QueueServer) sendDequeueSuccess(req Request, item *queue.QueueItem, queueName string) {
	queueServer.logger.Info("dequeued item - %s",
		logger.WithFields(map[string]interface{}{
			"queue":  queueName,
			"itemId": item.ID,
		}))

	msg := DequeueResponse{
		BaseResponse: BaseResponse{Message: "Item dequeued successfully"},
		QueueName:    queueName,
		Item:         *item,
	}

	req.ResponseCh <- msg
}

func (queueServer *QueueServer) sendDequeueLongPollingTimeout(req Request, queueName string) {
	queueServer.logger.Info("long polling timeout - %s", logger.WithFields(map[string]interface{}{
		"queue": queueName,
	}))

	msg := DequeueResponse{
		BaseResponse: BaseResponse{Error: ErrLongPollTimeout},
		QueueName:    queueName,
		Item:         queue.QueueItem{},
	}

	req.ResponseCh <- msg
}

func (queueServer *QueueServer) sendDequeueError(req Request, err error) {
	queueServer.logger.Error("dequeue failed - %s", err.Error())
	msg := DequeueResponse{
		BaseResponse: BaseResponse{Error: err},
	}

	req.ResponseCh <- msg
}

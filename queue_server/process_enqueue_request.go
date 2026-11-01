package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"go-queue-service/utils/logger"
)

func (queueServer *QueueServer) ProcessEnqueueRequest(req Request) {
	queueName := req.QueueName
	q, exists := queueServer.getQueue(queueName)

	queueServer.logger.Debug("processing enqueue request - %s", logger.WithField("queue", queueName))

	if !exists {
		queueServer.logger.Error("queue not found - %s", logger.WithField("queue", queueName))
		msg := EnqueueResponse{
			BaseResponse: BaseResponse{Error: fmt.Errorf("queue '%s' does not exist", queueName)},
			Item:         req.Item,
		}
		req.ResponseCh <- msg
	} else {
		item := req.Item
		queueServer.logger.Info("enqueued item - %s",
			logger.WithFields(map[string]interface{}{
				"queue":  queueName,
				"itemId": item.ID,
			}))

		// Add the item to the queue
		q.Enqueue(item)

		// Create the response message for the Enqueue action
		msg := EnqueueResponse{
			BaseResponse: BaseResponse{Message: "Item enqueued successfully"},
			QueueName:    queueName,
			Item:         item,
		}

		// Since we have a new item we should check if there are any clients long polling for an item
		// If there are, we should send the item to the first client in the waiting list
		queueServer.sendItemToOldestWaitingClient(queueName, q)

		req.ResponseCh <- msg
	}
}

func (queueServer *QueueServer) sendItemToOldestWaitingClient(queueName string, q *queue.Queue) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	if len(queueServer.waitingListClients[queueName]) > 0 {
		queueServer.logger.Debug("notifying waiting client - %s",
			logger.WithFields(map[string]interface{}{
				"queue":          queueName,
				"waitingClients": len(queueServer.waitingListClients[queueName]),
			}))

		// Dequeue item so that we trigger all the usual bookeeping.
		item, err := q.Dequeue()
		if err != nil {
			queueServer.logger.Error("Failed to dequeue item and give to long polling client waiting. %s", logger.WithFields(map[string]interface{}{
				"queue":          queueName,
				"waitingClients": len(queueServer.waitingListClients[queueName]),
			}))

			return
		}

		// Get the first client in the waiting list
		clientCh := queueServer.waitingListClients[queueName][0]
		// Send the item to the client. The client is waiting in the select statement in the ProcessDequeueRequest method
		// The client can grab the item from the channel and use it to make a DequeueResponse
		clientCh <- item
		// Close the channel
		close(clientCh)
		// Remove the client from the waiting list
		queueServer.waitingListClients[queueName] = queueServer.waitingListClients[queueName][1:]
	}
}

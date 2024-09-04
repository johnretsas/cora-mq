package queue_server

import (
	"go-queue-service/queue"
)

func (queueServer *QueueServer) processRequests() {
	for request := range queueServer.requestCh {
		// Process the request based on its type
		switch req := request.(type) {
		case Request:
			switch req.Type {
			case CreateQueueRequest:
				queueServer.logger.Println("Handling CreateQueueRequest")
				// Handle CreateQueueRequest
				if _, exists := queueServer.queues[req.QueueName]; exists {
					queueServer.logger.Printf("Queue with name '%s' already exists\n", req.QueueName)
					msg := CreateQueueResponse{
						BaseResponse: BaseResponse{Message: "Queue already exists"},
						QueueName:    req.QueueName,
					}
					req.ResponseCh <- msg
					continue
				}

				queueServer.logger.Printf("Creating queue: %s\n", req.QueueName)
				queueServer.queues[req.QueueName] = queue.NewQueue()

				msg := CreateQueueResponse{
					BaseResponse: BaseResponse{Message: "Queue created successfully"},
					QueueName:    req.QueueName,
				}

				req.ResponseCh <- msg

			case EnqueueRequest:
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
					continue
				}

				queueServer.logger.Printf("Enqueueing item to queue: %s\n", queueName)
				item := req.Item
				q.Enqueue(item)

				msg := EnqueueQueueResponse{
					BaseResponse: BaseResponse{Message: "Item enqueued successfully"},
					QueueName:    queueName,
					Item:         item,
				}

				req.ResponseCh <- msg

			default:
				queueServer.logger.Printf("Unknown request type: %d\n", req.Type)
			}

		default:
			queueServer.logger.Println("Received unexpected request type")
		}
	}
}

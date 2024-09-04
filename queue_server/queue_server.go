package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"log"
	"sync"
)

type QueueServer struct {
	queues    map[string]*queue.Queue
	logger    *log.Logger
	mu        sync.Mutex
	requestCh chan interface{}
}

func NewQueueServer(logger *log.Logger) *QueueServer {
	server := &QueueServer{
		queues:    make(map[string]*queue.Queue),
		logger:    logger,
		requestCh: make(chan interface{}),
	}

	go server.processRequests()

	return server
}

func (queueServer *QueueServer) CreateQueue(queueName string) (string, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	// define a channel to receive the response
	responseCh := make(chan interface{})

	// Create the request
	request := Request{
		Type:       CreateQueueRequest,
		QueueName:  queueName,
		ResponseCh: responseCh, // pass the response channel
	}

	// Send the request to the channel
	queueServer.requestCh <- request

	response := <-responseCh // The response will be received here

	fmt.Printf("response: %v\n", response)
	if res, ok := response.(CreateQueueResponse); ok {
		if res.Error != nil {
			return "", res.Error
		}
		return res.QueueName, nil
	}

	return "", fmt.Errorf("failed to create queue")
}

func (queueServer *QueueServer) Enqueue(queueName string, item queue.QueueItem) (queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	// define a channel to receive the response
	responseCh := make(chan interface{})

	// Create the request
	request := Request{
		Type:       EnqueueRequest,
		QueueName:  queueName,
		Item:       item,
		ResponseCh: responseCh, // pass the response channel
	}

	// Send the request to the channel
	queueServer.requestCh <- request

	response := <-responseCh // The response will be received here

	fmt.Printf("response: %v\n", response)
	if res, ok := response.(EnqueueQueueResponse); ok {
		if res.Error != nil {
			return queue.QueueItem{}, res.Error
		}
		return res.Item, nil
	}

	return queue.QueueItem{}, fmt.Errorf("failed to enqueue item")
	// q, exists := queueServer.queues[queueName]
	// if !exists {
	// 	queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
	// 	return queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	// }

	// q.Enqueue(item)
	// return item, nil
}

func (queueServer *QueueServer) Dequeue(queueName string) (queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	item, err := q.Dequeue()
	if err != nil {
		queueServer.logger.Printf("Error with dequeueing from queue: '%s'\n", queueName)
		return queue.QueueItem{}, err
	}

	return *item, nil
}

func (queueServer *QueueServer) Acknowledge(queueName string, id string) error {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return fmt.Errorf("queue '%s' does not exist", queueName)
	}

	err := q.Acknowledge(id)
	if err != nil {
		queueServer.logger.Printf("Error acknowledging message with id: '%s'\n", id)
		return err
	}

	return nil
}

func (queueServer *QueueServer) Scan(queueName string) ([]queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()

	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return []queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	items := q.Scan()
	return items, nil
}

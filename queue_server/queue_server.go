package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"log"
	"sync"
)

type QueueServer struct {
	queues map[string]*queue.Queue
	logger *log.Logger
	mu     sync.Mutex
	// Channel to receive requests from clients.
	// This channel is listened to by the processRequests
	// goroutine
	requestCh chan interface{}
	// A pool of workers to process requests. Since we are
	// using a goroutine to process requests, we need to
	// limit the number of workers to avoid consuming too
	// many resources
	workerPool chan struct{}
	// waiting list of clients to allow long polling
	waitingListClients map[string][]chan *queue.QueueItem
}

func NewQueueServer(logger *log.Logger, numOfWorkers int) *QueueServer {
	// Use a default number of workers if the number of workers
	// is not provided
	if numOfWorkers <= 0 {
		numOfWorkers = 3
	}

	println("Number of workers: ", numOfWorkers)
	server := &QueueServer{
		queues:             make(map[string]*queue.Queue),
		logger:             logger,
		requestCh:          make(chan interface{}),
		mu:                 sync.Mutex{},
		workerPool:         make(chan struct{}, numOfWorkers),        // Create a pool of 3 workers
		waitingListClients: make(map[string][]chan *queue.QueueItem), // Create a map to store clients waiting for items
	}

	go server.processRequests()

	return server
}

func (queueServer *QueueServer) CreateQueue(queueName string, config queue.QueueConfig) (string, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()
	// define a channel to receive the response
	responseCh := make(chan interface{})

	// Create the request
	request := Request{
		Type:        CreateQueueRequest,
		QueueName:   queueName,
		QueueConfig: config,
		ResponseCh:  responseCh, // pass the response channel
	}

	// Send the request to the channel
	queueServer.requestCh <- request

	response := <-responseCh // The response will be received here

	if res, ok := response.(CreateQueueResponse); ok {
		if res.Error != nil {
			return res.QueueName, res.Error
		}
		return res.QueueName, nil
	}

	return "", fmt.Errorf("failed to create queue")
}

func (queueServer *QueueServer) Enqueue(queueName string, item queue.QueueItem) (queue.QueueItem, error) {
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

	if res, ok := response.(EnqueueResponse); ok {
		if res.Error != nil {
			return res.Item, res.Error
		}
		return res.Item, nil
	}

	return queue.QueueItem{}, fmt.Errorf("failed to enqueue item")
}

func (queueServer *QueueServer) EnqueueBatch(queueName string, items []queue.QueueItem) ([]queue.QueueItem, error) {
	// define a channel to receive the response
	responseCh := make(chan interface{})

	// Create the request
	request := Request{
		Type:       EnqueueBatchRequest,
		QueueName:  queueName,
		Items:      items,
		ResponseCh: responseCh, // pass the response channel
	}

	// Send the request to the channel
	queueServer.requestCh <- request

	response := <-responseCh // The response will be received here

	if res, ok := response.(EnqueueBatchResponse); ok {
		if res.Error != nil {
			return res.Items, res.Error
		}
		return res.Items, nil
	}

	return []queue.QueueItem{}, fmt.Errorf("failed to enqueue items")
}

func (queueServer *QueueServer) Dequeue(queueName string) (queue.QueueItem, error) {
	// define a channel to receive the response
	responseCh := make(chan interface{})

	// Create the request
	request := Request{
		Type:       DequeueRequest,
		QueueName:  queueName,
		ResponseCh: responseCh,
	}

	queueServer.requestCh <- request

	response := <-responseCh

	if res, ok := response.(DequeueResponse); ok {
		if res.Error != nil {
			return queue.QueueItem{}, res.Error
		}
		return res.Item, nil
	}

	return queue.QueueItem{}, fmt.Errorf("failed to dequeue item")
}

func (queueServer *QueueServer) Acknowledge(queueName string, id string) (string, error) {
	// define a channel to receive the response
	responseCh := make(chan interface{})

	// Create the request
	request := Request{
		Type:       AcknowledgeRequest,
		QueueName:  queueName,
		Item:       queue.QueueItem{ID: id},
		ResponseCh: responseCh,
	}

	// Send the request to the server channel for it to be processed
	queueServer.requestCh <- request

	// Wait for the response
	response := <-responseCh

	if res, ok := response.(AcknowledgeResponse); ok {
		if res.Error != nil {
			return res.ID, res.Error
		}
		return res.ID, nil
	}

	return "", fmt.Errorf("failed to acknowledge item")
}

func (queueServer *QueueServer) Scan(queueName string) ([]queue.QueueItem, []queue.QueueItem, error) {
	queueServer.mu.Lock()
	defer queueServer.mu.Unlock()
	q, exists := queueServer.queues[queueName]
	if !exists {
		queueServer.logger.Printf("Queue with name '%s' does not exist\n", queueName)
		return []queue.QueueItem{}, []queue.QueueItem{}, fmt.Errorf("queue '%s' does not exist", queueName)
	}

	basicQueueItems, deadLetterQueueItems := q.Scan()
	return basicQueueItems, deadLetterQueueItems, nil
}

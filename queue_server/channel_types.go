package queue_server

import "go-queue-service/queue"

type RequestType int

const (
	CreateQueueRequest  RequestType = iota
	EnqueueRequest      RequestType = iota
	EnqueueBatchRequest RequestType = iota
	DequeueRequest      RequestType = iota
	AcknowledgeRequest  RequestType = iota
)

type Request struct {
	Type        RequestType       `json:"type"`
	QueueName   string            `json:"queueName"`
	Item        queue.QueueItem   `json:"item"`
	ResponseCh  chan interface{}  `json:"-"`
	Items       []queue.QueueItem `json:"items"`
	QueueConfig queue.QueueConfig `json:"queueConfig"`
}

type BaseResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}

type CreateQueueResponse struct {
	BaseResponse
	QueueName   string            `json:"queueName"`
	QueueConfig queue.QueueConfig `json:"queueConfig"`
}

type EnqueueResponse struct {
	BaseResponse
	QueueName string          `json:"queueName"`
	Item      queue.QueueItem `json:"item"`
}

type EnqueueBatchResponse struct {
	BaseResponse
	QueueName string            `json:"queueName"`
	Items     []queue.QueueItem `json:"items"`
}

type DequeueResponse struct {
	BaseResponse
	QueueName string          `json:"queueName"`
	Item      queue.QueueItem `json:"item"`
}

type AcknowledgeResponse struct {
	BaseResponse
	QueueName string `json:"queueName"`
	ID        string `json:"id"`
}

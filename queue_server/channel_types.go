package queue_server

import "go-queue-service/queue"

type RequestType int

const (
	CreateQueueRequest RequestType = iota
)

type Request struct {
	Type       RequestType
	QueueName  string
	Item       queue.QueueItem
	ResponseCh chan interface{}
}

type BaseResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}

type CreateQueueResponse struct {
	BaseResponse
	QueueName string `json:"queueName"`
}

package queue_server

func (queueServer *QueueServer) processRequests() {
	for request := range queueServer.requestCh {

		// Acquire a worker from the pool
		queueServer.workerPool <- struct{}{}

		// Process the request
		go withAcquiredWorker(queueServer, request)
	}
}

// withAcquiredWorker is a function that processes requests from the request channel
// It is supposed to be run as a goroutine and that a worker from the pool
// has been acquired before calling this function
func withAcquiredWorker(queueServer *QueueServer, request interface{}) {
	defer func() { <-queueServer.workerPool }() // Release the worker back to the pool
	switch req := request.(type) {
	// Process the request based on its type
	case Request:
		switch req.Type {
		case CreateQueueRequest:
			queueServer.ProcessCreateQueueRequest(req)
		case EnqueueRequest:
			queueServer.ProcessEnqueueRequest(req)
		case DequeueRequest:
			queueServer.ProcessDequeueRequest(req)
		case AcknowledgeRequest:
			queueServer.ProcessAcknowledgeRequest(req)
		default:
			queueServer.logger.Printf("Unknown request type: %d\n", req.Type)
		}

	default:
		queueServer.logger.Println("Received unexpected request type")
	}
}

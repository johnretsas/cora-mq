package queue_server

func (queueServer *QueueServer) processRequests() {
	for request := range queueServer.requestCh {
		// Process the request based on its type
		switch req := request.(type) {
		case Request:
			switch req.Type {
			case CreateQueueRequest:
				queueServer.ProcessCreateQueueRequest(req)
			case EnqueueRequest:
				queueServer.ProcessEnqueueRequest(req)
			case DequeueRequest:
				queueServer.ProcessDequeueRequest(req)
			default:
				queueServer.logger.Printf("Unknown request type: %d\n", req.Type)
			}

		default:
			queueServer.logger.Println("Received unexpected request type")
		}
	}
}

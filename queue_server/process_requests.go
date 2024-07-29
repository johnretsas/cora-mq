package queue_server

func (queueServer *QueueServer) processRequests() {
	for {
		// Wait for a request to come in
		request := <-queueServer.requestCh

		switch req := request.(type) {

		default:
			queueServer.logger.Println("Request type: ", req)
		}
	}
}

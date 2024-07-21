package queue_server

import (
	"encoding/json"
	"net/http"
)

func (queueServer *QueueServer) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Server is healthy"))

	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: "Server is healthy",
	})

}

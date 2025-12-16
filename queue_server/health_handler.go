package queue_server

import (
	"encoding/json"
	"net/http"
)

func (queueServer *QueueServer) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: "Server is healthy",
	})

}

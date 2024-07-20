package queue_server

import (
	"encoding/json"
	"net/http"
)

func (queueServer *QueueServer) CreateQueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var queueName struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&queueName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueServer.CreateQueue(queueName.Name)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: "Queue created successfully",
	})
}

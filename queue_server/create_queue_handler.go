package queue_server

import (
	"encoding/json"
	"net/http"
)

func (queueServer *QueueServer) CreateQueueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Method not allowed",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var queueName struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&queueName); err != nil {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
	}

	name, err := queueServer.CreateQueue(queueName.Name)

	if err != nil {
		errorMsg := struct {
			Error     string `json:"error"`
			QueueName string `json:"queueName"`
		}{
			Error:     err.Error(),
			QueueName: name,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(struct {
		Message   string `json:"message"`
		QueueName string `json:"queueName"`
	}{
		Message:   "Queue created successfully",
		QueueName: name,
	})
}

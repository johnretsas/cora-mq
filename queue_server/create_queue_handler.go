package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
	"io"
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

	var requestBody struct {
		Name   string            `json:"name"`
		Config queue.QueueConfig `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		if err == io.EOF || err.Error() == "EOF" {
			errorMsg := struct {
				Error string `json:"error"`
			}{
				Error: "Missing body",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorMsg)
			return
		}

		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	if requestBody.Name == "" {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Missing queue name",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	name, err := queueServer.CreateQueue(requestBody.Name, requestBody.Config)

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

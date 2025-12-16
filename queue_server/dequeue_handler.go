package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
	"io"
)

func (queueServer *QueueServer) DequeueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Method Not Allowed",
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	var requestBody struct {
		QueueName string `json:"queueName"`
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

	if requestBody.QueueName == "" {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Missing queueName",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	item, err := queueServer.Dequeue(requestBody.QueueName)

	if err != nil {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	dequeueMsg := struct {
		Message string          `json:"message"`
		Item    queue.QueueItem `json:"item"`
	}{
		Message: "Item dequeued",
		Item:    item,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dequeueMsg)
}

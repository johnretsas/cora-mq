package queue_server

import (
	"encoding/json"
	"go-queue-service/queue"
	"net/http"
	"io"
)

func (queueServer *QueueServer) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
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
		QueueName string          `json:"queueName"`
		Item      queue.QueueItem `json:"item"`
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

	// if no item is provided, return an error

	if requestBody.Item.ID == "" {
		errorMsg := struct {
			Error string `json:"error"`
		}{
			Error: "Missing item",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	item, err := queueServer.Enqueue(requestBody.QueueName, requestBody.Item)

	if err != nil {
		errorMsg := struct {
			Error     string `json:"error"`
			Item      string `json:"item"`
			QueueName string `json:"queueName"`
		}{
			Error:     err.Error(),
			Item:      item.ID,
			QueueName: requestBody.QueueName,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	w.WriteHeader(http.StatusCreated)
	enqueueMsg := struct {
		Message string `json:"message"`
		Id      string `json:"id"`
	}{
		Message: "Item enqueued",
		Id:      item.ID,
	}

	json.NewEncoder(w).Encode(enqueueMsg)
}

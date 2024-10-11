package queue_server

import (
	"encoding/json"
	"net/http"
)

func (queueServer *QueueServer) AcknowledgeHandler(w http.ResponseWriter, r *http.Request) {
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
		QueueName string `json:"queueName"`
		ID        string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.QueueName == "" || requestBody.ID == "" {
		http.Error(w, "Missing queueName or id", http.StatusBadRequest)
		return
	}

	id, err := queueServer.Acknowledge(requestBody.QueueName, requestBody.ID)

	if err != nil {
		errorMsg := struct {
			Error string `json:"error"`
			ID    string `json:"id"`
		}{
			Error: err.Error(),
			ID:    id,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	ackMsg := struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}{
		Message: "Message acknowledged",
		ID:      id,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ackMsg)
}

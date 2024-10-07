package queue_server

import (
	"encoding/json"
	"net/http"
)

func (queueServer *QueueServer) AcknowledgeHandler(w http.ResponseWriter, r *http.Request) {
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

	var requestBody struct {
		QueueName string `json:"queueName"`
		ID        string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMsg)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	ackMsg := struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}{
		Message: "Message acknowledged",
		ID:      id,
	}

	json.NewEncoder(w).Encode(ackMsg)
}

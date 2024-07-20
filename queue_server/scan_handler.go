package queue_server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (queueServer *QueueServer) ScanHandler(w http.ResponseWriter, r *http.Request) {
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
	// Assuming you have a queue instance named 'q'
	var requestBody struct {
		QueueName string `json:"queueName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.QueueName == "" {
		http.Error(w, "Missing queueName or id", http.StatusBadRequest)
		return
	}

	err := queueServer.Scan(requestBody.QueueName)
	if err != nil {
		log.Println("Error scanning the queue:", err)
		// You can handle the error here, e.g., return an appropriate HTTP response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Handle the successful scan here, e.g., return an appropriate HTTP response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Queue scanned successfully"))
}

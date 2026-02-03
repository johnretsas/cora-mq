package queue_server

import (
	"bytes"
	"encoding/json"
	"go-queue-service/utils/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper to create a test server
func newTestServer() *QueueServer {
	log := logger.New("TestServer", logger.DEBUG).WithDevelopmentMode(true)
	return NewQueueServer(log, 3)
}

// Helper to make a test request
func makeRequest(t *testing.T, handler http.HandlerFunc, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// Helper to check JSON response
func checkJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int, expectedBody map[string]interface{}) {
	t.Helper()

	if rr.Code != expectedStatus {
		t.Errorf("Status code = %d, want %d", rr.Code, expectedStatus)
	}

	if expectedBody != nil {
		var actualBody map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &actualBody); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		for key, expectedValue := range expectedBody {
			if actualValue, ok := actualBody[key]; !ok {
				t.Errorf("Response missing key %q", key)
			} else if actualValue != expectedValue {
				t.Errorf("Response[%q] = %v, want %v", key, actualValue, expectedValue)
			}
		}
	}
}

func TestHealthCheckHandler(t *testing.T) {
	server := newTestServer()
	rr := makeRequest(t, server.HealthCheckHandler, "GET", "/health", nil)

	checkJSONResponse(t, rr, http.StatusOK, map[string]interface{}{
		"message": "Server is healthy",
	})
}

func TestCreateQueueHandler(t *testing.T) {
	t.Run("successful queue creation", func(t *testing.T) {
		server := newTestServer()
		requestBody := map[string]interface{}{"name": "testQueue"}

		rr := makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", requestBody)

		if rr.Code != http.StatusCreated {
			t.Errorf("Status code = %d, want %d", rr.Code, http.StatusCreated)
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["message"] != "Queue created successfully" {
			t.Errorf("Message = %q, want %q", response["message"], "Queue created successfully")
		}
	})

	t.Run("duplicate queue creation", func(t *testing.T) {
		server := newTestServer()
		requestBody := map[string]interface{}{"name": "duplicateQueue"}

		// First creation should succeed
		rr1 := makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", requestBody)
		if rr1.Code != http.StatusCreated {
			t.Errorf("First creation status = %d, want %d", rr1.Code, http.StatusCreated)
		}

		// Second creation should fail
		rr2 := makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", requestBody)
		if rr2.Code != http.StatusBadRequest {
			t.Errorf("Duplicate creation status = %d, want %d", rr2.Code, http.StatusBadRequest)
		}
	})
}

func TestEnqueueHandler(t *testing.T) {
	server := newTestServer()

	// First create a queue
	createBody := map[string]interface{}{"name": "testQueue"}
	makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", createBody)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "successful enqueue",
			requestBody: map[string]interface{}{
				"queueName": "testQueue",
				"item": map[string]interface{}{
					"id":   "item-1",
					"data": "test payload",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "enqueue to non-existent queue",
			requestBody: map[string]interface{}{
				"queueName": "nonExistentQueue",
				"item": map[string]interface{}{
					"id":   "item-2",
					"data": "test payload",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, server.EnqueueHandler, "POST", "/enqueue", tt.requestBody)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Status code = %d, want %d (body: %s)", rr.Code, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestDequeueHandler(t *testing.T) {
	server := newTestServer()

	// Create queue and enqueue an item
	createBody := map[string]interface{}{"name": "testQueue"}
	makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", createBody)

	enqueueBody := map[string]interface{}{
		"queueName": "testQueue",
		"item": map[string]interface{}{
			"id":   "item-1",
			"data": "test payload",
		},
	}
	makeRequest(t, server.EnqueueHandler, "POST", "/enqueue", enqueueBody)

	t.Run("successful dequeue", func(t *testing.T) {
		dequeueBody := map[string]interface{}{"queueName": "testQueue"}
		rr := makeRequest(t, server.DequeueHandler, "POST", "/dequeue", dequeueBody)

		if rr.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", rr.Code, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := response["item"]; !ok {
			t.Error("Response missing 'item' field")
		}
	})

	t.Run("dequeue from empty queue", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping slow test (30s long poll timeout) in short mode")
		}

		// Dequeue again (queue should be empty after first dequeue)
		dequeueBody := map[string]interface{}{"queueName": "testQueue"}
		rr := makeRequest(t, server.DequeueHandler, "POST", "/dequeue", dequeueBody)

		// Dequeue on empty queue triggers long polling, then returns 400 after timeout
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Empty queue status = %d, want %d (long poll timeout)", rr.Code, http.StatusBadRequest)
		}
	})
}

func TestAcknowledgeHandler(t *testing.T) {
	server := newTestServer()

	// Create queue and enqueue an item
	createBody := map[string]interface{}{"name": "testQueue"}
	makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", createBody)

	enqueueBody := map[string]interface{}{
		"queueName": "testQueue",
		"item": map[string]interface{}{
			"id":   "item-1",
			"data": "test payload",
		},
	}
	makeRequest(t, server.EnqueueHandler, "POST", "/enqueue", enqueueBody)

	// Dequeue the item
	dequeueBody := map[string]interface{}{"queueName": "testQueue"}
	makeRequest(t, server.DequeueHandler, "POST", "/dequeue", dequeueBody)

	t.Run("successful acknowledge", func(t *testing.T) {
		ackBody := map[string]interface{}{
			"queueName": "testQueue",
			"id":        "item-1",
		}
		rr := makeRequest(t, server.AcknowledgeHandler, "POST", "/acknowledge", ackBody)

		if rr.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d (body: %s)", rr.Code, http.StatusOK, rr.Body.String())
		}
	})

	t.Run("acknowledge non-existent item", func(t *testing.T) {
		ackBody := map[string]interface{}{
			"queueName": "testQueue",
			"id":        "non-existent-item",
		}
		rr := makeRequest(t, server.AcknowledgeHandler, "POST", "/acknowledge", ackBody)

		// Should return error status
		if rr.Code == http.StatusOK {
			t.Error("Acknowledging non-existent item should fail")
		}
	})
}

func TestSizeOfQueueHandler(t *testing.T) {
	server := newTestServer()

	// Create queue
	createBody := map[string]interface{}{"name": "testQueue"}
	makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", createBody)

	t.Run("empty queue size", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/sizeOfQueue?queueName=testQueue", nil)
		rr := httptest.NewRecorder()
		server.SizeOfQueueHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", rr.Code, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if size, ok := response["size"].(float64); !ok || size != 0 {
			t.Errorf("Size = %v, want 0", response["size"])
		}
	})

	t.Run("queue size after enqueue", func(t *testing.T) {
		// Enqueue an item
		enqueueBody := map[string]interface{}{
			"queueName": "testQueue",
			"item": map[string]interface{}{
				"id":   "item-1",
				"data": "test",
			},
		}
		makeRequest(t, server.EnqueueHandler, "POST", "/enqueue", enqueueBody)

		req, _ := http.NewRequest("GET", "/sizeOfQueue?queueName=testQueue", nil)
		rr := httptest.NewRecorder()
		server.SizeOfQueueHandler(rr, req)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		if size, ok := response["size"].(float64); !ok || size != 1 {
			t.Errorf("Size = %v, want 1", response["size"])
		}
	})
}

func TestEnqueueBatchHandler(t *testing.T) {
	server := newTestServer()

	// Create queue
	createBody := map[string]interface{}{"name": "testQueue"}
	makeRequest(t, server.CreateQueueHandler, "POST", "/createQueue", createBody)

	t.Run("successful batch enqueue", func(t *testing.T) {
		batchBody := map[string]interface{}{
			"queueName": "testQueue",
			"items": []map[string]interface{}{
				{"id": "batch-1", "data": "first"},
				{"id": "batch-2", "data": "second"},
				{"id": "batch-3", "data": "third"},
			},
		}
		rr := makeRequest(t, server.EnqueueBatchHandler, "POST", "/enqueue/batch", batchBody)

		if rr.Code != http.StatusCreated {
			t.Errorf("Status code = %d, want %d (body: %s)", rr.Code, http.StatusCreated, rr.Body.String())
		}

		// Verify queue size
		req, _ := http.NewRequest("GET", "/sizeOfQueue?queueName=testQueue", nil)
		sizeRr := httptest.NewRecorder()
		server.SizeOfQueueHandler(sizeRr, req)

		var sizeResponse map[string]interface{}
		json.Unmarshal(sizeRr.Body.Bytes(), &sizeResponse)

		if size, ok := sizeResponse["size"].(float64); !ok || size != 3 {
			t.Errorf("Size after batch enqueue = %v, want 3", sizeResponse["size"])
		}
	})
}

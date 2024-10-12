package queue_server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

// Color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

func TestCreateQueueHandler(t *testing.T) {
	logger := log.New(os.Stdout, "QueueServer: ", log.LstdFlags)
	server := NewQueueServer(logger, 3)

	reqBody := []byte(`{"name":"testQueue"}`)
	req, err := http.NewRequest("POST", "/createQueue", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.CreateQueueHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("%shandler returned wrong status code: got %v want %v%s", colorRed, status, http.StatusCreated, colorReset)
	}

	t.Logf("%sActual: %q%s", colorYellow, rr.Body.String(), colorReset)
	expected := `{"message":"Queue created successfully"}`
	t.Logf("%sExpected: %q%s", colorGreen, expected, colorReset)

	// Decode the actual response
	var actualResponse map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &actualResponse)
	if err != nil {
		t.Fatalf("%scould not unmarshal response: %v%s", colorRed, err, colorReset)
	}

	// Define the expected response
	expectedResponse := map[string]string{"message": "Queue created successfully", "queueName": "testQueue"}

	// Compare the actual and expected responses
	if !reflect.DeepEqual(actualResponse, expectedResponse) {
		t.Errorf("%shandler returned unexpected body: got %v want %v%s",
			colorRed, actualResponse, expectedResponse, colorReset)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	maxRetries    = 10
	baseBackoffMs = 50
	maxBackoffMs  = 5000
)

// retryWithBackoff executes a function with exponential backoff on 429 errors
func retryWithBackoff(operation func() (*http.Response, error), operationName string) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = operation()

		if err != nil {
			return nil, err
		}

		// If not rate limited, return immediately
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Close the response body before retrying
		resp.Body.Close()

		// If this was the last attempt, return the 429 response
		if attempt == maxRetries {
			return resp, fmt.Errorf("max retries reached for %s", operationName)
		}

		// Calculate exponential backoff with jitter
		backoffMs := math.Min(float64(baseBackoffMs)*math.Pow(2, float64(attempt)), float64(maxBackoffMs))
		backoffDuration := time.Duration(backoffMs) * time.Millisecond

		if attempt == 0 {
			fmt.Printf("⏳ Rate limited, retrying %s with backoff...\n", operationName)
		}

		time.Sleep(backoffDuration)
	}

	return resp, err
}

func createQueueCommon(queueName, baseURL string, client *http.Client) error {
	payload := map[string]interface{}{
		"name": queueName,
		"config": map[string]interface{}{
			"maxRetries":        3,
			"visibilityTimeout": 30,
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := client.Post(baseURL+"/createQueue", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func acknowledgeMessageCommon(itemID, queueName, baseURL string, client *http.Client) error {
	payload := map[string]interface{}{
		"queueName": queueName,
		"id":        itemID,
	}

	body, _ := json.Marshal(payload)
	resp, err := retryWithBackoff(func() (*http.Response, error) {
		return client.Post(baseURL+"/acknowledge", "application/json", bytes.NewBuffer(body))
	}, fmt.Sprintf("ack %s", itemID))

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("acknowledgment failed with status %d", resp.StatusCode)
	}

	return nil
}

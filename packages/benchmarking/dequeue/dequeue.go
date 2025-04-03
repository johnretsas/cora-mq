package dequeue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	numRequests    = 2000                                // Number of messages to send in each batch
	concurrency    = 100                                 // Number of concurrent goroutines for dequeuing/acknowledging
	url            = "http://localhost:8080/enqueue"     // URL for enqueueing messages
	dequeueUrl     = "http://localhost:8080/dequeue"     // URL for dequeuing messages
	acknowledgeUrl = "http://localhost:8080/acknowledge" // URL for acknowledging messages
	duration       = 1 * time.Minute                     // Total runtime duration for simulation
)

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 1000,
		DisableKeepAlives:   false,
	},
}

func TestDequeue() {
	// Track success rate and response times
	var successCount, totalDequeueCount, totalAckCount, totalEnqueueCount int
	var totalDequeueTime, totalAckTime time.Duration
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	createQueueRequest, err := http.NewRequest("POST", "http://localhost:8080/createQueue", bytes.NewBuffer([]byte(`{"name": "testQueue"}`)))
	if err != nil {
		fmt.Println("Error creating queue request:", err)
		return
	}

	// Create a test queue
	resp, err := client.Do(createQueueRequest)
	if err != nil {
		fmt.Println("Error creating queue:", err)
		return
	}

	defer resp.Body.Close()

	// Start time tracking
	startTime := time.Now()

	// Run the simulation for 3-4 minutes
	// for time.Since(startTime) < duration {
	// Send a batch of messages
	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()

			// Prepare message data
			messageData := []byte(fmt.Sprintf(`{
                                        "queueName": "testQueue",
                                        "item": {
                                                "id": "%d",
                                                "payload": "{'name': 'User %d', 'email': 'user%d@example.com'}"
                                        }
                                }`, i, i, i))

			// Send enqueue request
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(messageData))
			req.Header.Set("Content-Type", "application/json")
			// Add a unique request ID to the header
			req.Header.Set("X-Request-ID", strconv.Itoa(i))
			if err != nil {
				fmt.Println("Error creating enqueue request:", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			_, err = SendRequestWithRetry(req, client, i)
			if err != nil {
				fmt.Println("Error sending enqueue request:", err)
				return
			}
			totalEnqueueCount++
		}(i)
	}

	wg.Wait()
	fmt.Printf("Enqueued %d messages\n", numRequests)

	// Dequeue and acknowledge messages concurrently
	for i := 0; i < numRequests; i++ {
		sem <- struct{}{} // Limit concurrency
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Dequeue request
			dequeueReq, err := http.NewRequest("POST", dequeueUrl, bytes.NewBuffer([]byte(`{"queueName": "testQueue"}`)))
			if err != nil {
				fmt.Println("Error creating dequeue request:", err)
				return
			}

			dequeueStart := time.Now()
			fmt.Println("Dequeueing message with ID:", i)
			dequeueResp, err := SendRequestWithRetry(dequeueReq, client, i)
			if err != nil {
				fmt.Println("Error sending dequeue request:", err)
				return
			}

			dequeueDuration := time.Since(dequeueStart)

			bodyBytes, err := io.ReadAll(dequeueResp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}
			defer dequeueResp.Body.Close()

			var responseData struct {
				Item struct {
					ID string `json:"id"`
				} `json:"item"`
			}

			if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
				fmt.Println("Error unmarshaling response body:", err)
				return
			}

			itemID := responseData.Item.ID
			if err != nil || dequeueResp.StatusCode != http.StatusOK {
				fmt.Println("Error or bad response during dequeue:", err, resp.StatusCode, itemID)
				return
			}

			// Update dequeue stats
			totalDequeueCount++
			totalDequeueTime += dequeueDuration

			// Acknowledge request
			acknowledgeReq, err := http.NewRequest("POST", acknowledgeUrl, bytes.NewBuffer([]byte(fmt.Sprintf(`{"queueName": "testQueue", "id": "%s"}`, itemID))))
			if err != nil {
				fmt.Println("Error creating acknowledge request:", err)
				return
			}

			acknowledgeStart := time.Now()
			fmt.Println("Acknowledging message with ID:", itemID)
			itemIDInt, err := strconv.Atoi(itemID)
			if err != nil {
				fmt.Println("Error converting itemID to int:", err)
				return
			}
			resp, err = SendRequestWithRetry(acknowledgeReq, client, itemIDInt)
			acknowledgeDuration := time.Since(acknowledgeStart)

			if err != nil || resp.StatusCode != http.StatusOK {
				fmt.Println("Error or bad response during acknowledge:", err)
				return
			}

			// Update acknowledgment stats
			totalAckCount++
			totalAckTime += acknowledgeDuration

			// Track success
			successCount++
		}(i)
	}

	// Wait for all dequeues and acknowledgments to finish
	wg.Wait()
	// }

	// Calculate and print success rates and average times
	totalTime := time.Since(startTime)

	fmt.Printf("\nSimulation completed in %v\n", totalTime)
	fmt.Printf("Total Enqueue requests processed: %d\n", totalEnqueueCount)
	fmt.Printf("Total dequeued: %d, Total acknowledged: %d\n", totalDequeueCount, totalAckCount)
	fmt.Printf("Success rate of dequeueing items completely: %.2f%%\n", float64(totalDequeueCount)/float64(totalAckCount)*100)

	if totalDequeueCount > 0 {
		avgDequeueTime := totalDequeueTime / time.Duration(totalDequeueCount)
		fmt.Printf("Average dequeue time: %v\n", avgDequeueTime)
	}

	if totalAckCount > 0 {
		avgAckTime := totalAckTime / time.Duration(totalAckCount)
		fmt.Printf("Average acknowledgment time: %v\n", avgAckTime)
	}
}

func SendRequestWithRetry(req *http.Request, client *http.Client, id int) (*http.Response, error) {
	maxRetries := 8
	backoff := time.Second

	// Check if the request body is nil, and handle that case
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
		// Reset the request body for reuse in retries
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // This is important to reset the body
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Do(req)

		if err != nil || (resp != nil && (resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusTooManyRequests)) {
			// If error occurs, retry after backoff
			if err != nil {
				fmt.Println("Error sending request, retrying:", err)
			} else if resp != nil {
				fmt.Printf("Request with url %s and ID %d received status code %d, retrying...\n", req.URL.String(), id, resp.StatusCode)
			}
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			// Reset the request body again after backoff if it's not nil
			if req.Body != nil {
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset body
			}
			continue
		}
		fmt.Printf("Request with url %s and ID %d succeeded.\n", req.URL.String(), id)
		// Return successful response if no error
		return resp, nil
	}

	return nil, fmt.Errorf("failed to send request after %d retries", maxRetries)
}

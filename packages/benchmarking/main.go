package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	numRequests = 1000
	concurrency = 50
	url         = "http://localhost:8080/enqueue"
)

var client = &http.Client{
	Timeout: 10 * time.Second, // Timeout for individual requests
	Transport: &http.Transport{
		MaxIdleConns:        1000,  // Allow more idle connections to be reused
		MaxIdleConnsPerHost: 1000,  // Allow more idle connections per host
		DisableKeepAlives:   false, // Enable keep-alive for connection reuse
	},
}

var requestTimes struct {
	sync.Mutex
	lessThan5ms    int
	lessThan10ms   int
	lessThan50ms   int
	lessThan100ms  int
	lessThan500ms  int
	lessThan1s     int
	lessThan5s     int
	greaterThan5s  int
	failedRequests int
}

func updateRequestTimes(duration time.Duration, err error) {
	if err != nil {
		requestTimes.Lock()
		defer requestTimes.Unlock()
		requestTimes.failedRequests++
		return
	}
	requestTimes.Lock()
	defer requestTimes.Unlock()

	switch {
	case duration < 5*time.Millisecond:
		requestTimes.lessThan5ms++
	case duration < 10*time.Millisecond:
		requestTimes.lessThan10ms++
	case duration < 50*time.Millisecond:
		requestTimes.lessThan50ms++
	case duration < 100*time.Millisecond:
		requestTimes.lessThan100ms++
	case duration < 500*time.Millisecond:
		requestTimes.lessThan500ms++
	case duration < 1*time.Second:
		requestTimes.lessThan1s++
	case duration < 5*time.Second:
		requestTimes.lessThan5s++
	default:
		requestTimes.greaterThan5s++
	}
}

func sendRequestWithRetry(req *http.Request, client *http.Client) (*http.Response, error) {
	maxRetries := 8
	backoff := time.Second

	// Read the body into memory to allow retries with the same content
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	// Reset the request body for reuse in retries
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // This is important to reset the body

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Do(req)

		if err != nil || (resp != nil && resp.StatusCode == http.StatusTooManyRequests) {
			// If error occurs, retry after backoff
			if err != nil {
				fmt.Println("Error sending request, retrying:", err)
			} else if resp != nil {
				fmt.Println("Received status code", resp.StatusCode, "retrying...")
			}
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			// Reset the request body again after backoff
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset body
			continue
		}
		// Return successful response if no error
		return resp, nil
	}

	updateRequestTimes(0, fmt.Errorf("failed to send request after %d retries", maxRetries))

	return nil, fmt.Errorf("failed to send request after %d retries", maxRetries)
}

func main() {
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, concurrency)

	// Create a test queue
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

	startTime := time.Now()
	for i := range numRequests {
		sem <- struct{}{}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			messageData := []byte(`{
				"queueName": "testQueue",
				"item": {
					"id": "2000",
					"payload": "{'name': 'John Retsas', 'email': 'joretsas@gmail.com'}"
				}
			}`)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(messageData))

			if err != nil {
				println("Error creating request:", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			// Add a unique request ID to the header
			req.Header.Set("X-Request-ID", strconv.Itoa(i))
			// Send the request
			requestStartTime := time.Now()
			resp, err := sendRequestWithRetry(req, client)
			// fmt.Println("Request ID:", i+1, "Response Status Code:", resp.StatusCode)
			requestEndTime := time.Since(requestStartTime)
			updateRequestTimes(requestEndTime, nil) // Update request times based on response time

			// Update request times based on response time

			if err != nil {
				fmt.Println("Error sending request:", err.Error())
				return
			}

			// Close the response body
			defer resp.Body.Close()
		}(i)
	}

	wg.Wait()
	elapsedTime := time.Since(startTime)
	fmt.Printf("Total time taken for %d requests: %v\n", numRequests, elapsedTime)
	fmt.Printf("Request times:\n")
	fmt.Printf("Less than 5ms: %d\n", requestTimes.lessThan5ms)
	fmt.Printf("Less than 10ms: %d\n", requestTimes.lessThan10ms)
	fmt.Printf("Less than 50ms: %d\n", requestTimes.lessThan50ms)
	fmt.Printf("Less than 100ms: %d\n", requestTimes.lessThan100ms)
	fmt.Printf("Less than 500ms: %d\n", requestTimes.lessThan500ms)
	fmt.Printf("Less than 1s: %d\n", requestTimes.lessThan1s)
	fmt.Printf("Less than 5s: %d\n", requestTimes.lessThan5s)
	fmt.Printf("Greater than 5s: %d\n", requestTimes.greaterThan5s)
	fmt.Printf("Failed requests: %d\n", requestTimes.failedRequests)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	httpClient *http.Client
)

func init() {
	// Create HTTP client with connection pooling and keep-alive
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}

type BenchmarkStats struct {
	TotalMessages        int
	EnqueueTime          time.Duration
	DequeueTime          time.Duration
	TotalTime            time.Duration
	EnqueueSuccesses     int
	EnqueueFailures      int
	DequeueSuccesses     int
	DequeueFailures      int
	AcknowledgeSuccesses int
	AcknowledgeFailures  int
	MessagesReceived     map[string]bool
	DuplicateMessages    int
	MissingMessages      []string
}

func runSequentialBenchmark(numItems int, baseURL, queueName string, delayMs int) {

	fmt.Println("=== Queue Benchmark Test ===")
	fmt.Printf("Target: %d messages\n", numItems)
	fmt.Printf("Queue: %s\n", queueName)
	fmt.Printf("Server: %s\n\n", baseURL)

	stats := &BenchmarkStats{
		TotalMessages:    numItems,
		MessagesReceived: make(map[string]bool),
	}

	// Create the queue
	fmt.Println("Step 1: Creating queue...")
	if err := createQueueSequential(queueName, baseURL); err != nil {
		fmt.Printf("❌ Failed to create queue: %v\n", err)
		return
	}
	fmt.Println("✅ Queue created successfully")

	// Enqueue messages
	fmt.Printf("Step 2: Enqueuing %d messages...\n", numItems)
	startEnqueue := time.Now()
	sentIDs := enqueueMessages(stats, numItems, queueName, baseURL, delayMs)
	stats.EnqueueTime = time.Since(startEnqueue)
	fmt.Printf("✅ Enqueue complete in %v\n", stats.EnqueueTime)
	fmt.Printf("   Successes: %d, Failures: %d\n\n", stats.EnqueueSuccesses, stats.EnqueueFailures)

	// Small delay to ensure all messages are processed
	time.Sleep(100 * time.Millisecond)

	// Dequeue messages
	fmt.Printf("Step 3: Dequeuing %d messages...\n", numItems)
	startDequeue := time.Now()
	dequeueMessages(stats, numItems, queueName, baseURL, delayMs)
	stats.DequeueTime = time.Since(startDequeue)
	fmt.Printf("✅ Dequeue complete in %v\n", stats.DequeueTime)
	fmt.Printf("   Successes: %d, Failures: %d\n\n", stats.DequeueSuccesses, stats.DequeueFailures)

	// Verify results
	fmt.Println("Step 4: Verifying results...")
	verifySequentialResults(stats, sentIDs)

	// Print summary
	printSequentialSummary(stats)
}

func createQueueSequential(queueName, baseURL string) error {
	return createQueueCommon(queueName, baseURL, httpClient)
}

func enqueueMessages(stats *BenchmarkStats, numItems int, queueName, baseURL string, delayMs int) []string {
	var sentIDs []string
	var mu sync.Mutex

	for i := 0; i < numItems; i++ {
		itemID := fmt.Sprintf("item-%d", i)
		payload := map[string]interface{}{
			"queueName": queueName,
			"item": map[string]interface{}{
				"id":   itemID,
				"data": fmt.Sprintf("Test message number %d", i),
			},
		}

		body, _ := json.Marshal(payload)

		// Use retry with backoff
		resp, err := retryWithBackoff(func() (*http.Response, error) {
			return httpClient.Post(baseURL+"/enqueue", "application/json", bytes.NewBuffer(body))
		}, fmt.Sprintf("enqueue %s", itemID))

		if err != nil {
			mu.Lock()
			stats.EnqueueFailures++
			mu.Unlock()
			fmt.Printf("⚠️  Enqueue failed for %s: %v\n", itemID, err)
			continue
		}

		if resp.StatusCode == http.StatusCreated {
			mu.Lock()
			stats.EnqueueSuccesses++
			sentIDs = append(sentIDs, itemID)
			mu.Unlock()
		} else {
			mu.Lock()
			stats.EnqueueFailures++
			mu.Unlock()
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("⚠️  Enqueue failed for %s: status %d, body: %s\n", itemID, resp.StatusCode, string(bodyBytes))
		}
		resp.Body.Close()

		// Optional delay between requests
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// Progress indicator
		if (i+1)%500 == 0 {
			fmt.Printf("   Progress: %d/%d messages enqueued\n", i+1, numItems)
		}
	}

	return sentIDs
}

func acknowledgeMessageSequential(itemID, queueName, baseURL string) error {
	return acknowledgeMessageCommon(itemID, queueName, baseURL, httpClient)
}

func dequeueMessages(stats *BenchmarkStats, numItems int, queueName, baseURL string, delayMs int) {
	for i := 0; i < numItems; i++ {
		payload := map[string]interface{}{
			"queueName": queueName,
		}

		body, _ := json.Marshal(payload)

		// Use retry with backoff
		resp, err := retryWithBackoff(func() (*http.Response, error) {
			return httpClient.Post(baseURL+"/dequeue", "application/json", bytes.NewBuffer(body))
		}, fmt.Sprintf("dequeue #%d", i))

		if err != nil {
			stats.DequeueFailures++
			fmt.Printf("⚠️  Dequeue failed at iteration %d: %v\n", i, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result struct {
				Item struct {
					ID   string `json:"id"`
					Data string `json:"data"`
				} `json:"item"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				stats.DequeueFailures++
				fmt.Printf("⚠️  Failed to decode response at iteration %d: %v\n", i, err)
			} else {
				stats.DequeueSuccesses++

				// Check for duplicates
				if stats.MessagesReceived[result.Item.ID] {
					stats.DuplicateMessages++
					fmt.Printf("⚠️  Duplicate message detected: %s\n", result.Item.ID)
				} else {
					stats.MessagesReceived[result.Item.ID] = true
				}

				// IMPORTANT: Acknowledge the message so it's not re-delivered
				if err := acknowledgeMessageSequential(result.Item.ID, queueName, baseURL); err != nil {
					stats.AcknowledgeFailures++
				} else {
					stats.AcknowledgeSuccesses++
				}
			}
		} else {
			stats.DequeueFailures++
			bodyBytes, _ := io.ReadAll(resp.Body)
			if i < 10 { // Only print first few errors to avoid spam
				fmt.Printf("⚠️  Dequeue failed at iteration %d: status %d, body: %s\n", i, resp.StatusCode, string(bodyBytes))
			}
		}
		resp.Body.Close()

		// Optional delay between requests
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		// Progress indicator
		if (i+1)%500 == 0 {
			fmt.Printf("   Progress: %d/%d messages dequeued\n", i+1, numItems)
		}
	}
}

func verifySequentialResults(stats *BenchmarkStats, sentIDs []string) {
	// Check for missing messages
	for _, id := range sentIDs {
		if !stats.MessagesReceived[id] {
			stats.MissingMessages = append(stats.MissingMessages, id)
		}
	}

	if len(stats.MissingMessages) == 0 && stats.DuplicateMessages == 0 {
		fmt.Println("✅ All messages accounted for - no duplicates, no missing messages")
	} else {
		if len(stats.MissingMessages) > 0 {
			fmt.Printf("❌ Missing %d messages\n", len(stats.MissingMessages))
			if len(stats.MissingMessages) <= 10 {
				fmt.Printf("   Missing IDs: %v\n", stats.MissingMessages)
			}
		}
		if stats.DuplicateMessages > 0 {
			fmt.Printf("❌ Found %d duplicate messages\n", stats.DuplicateMessages)
		}
	}
	fmt.Println()
}

func printSequentialSummary(stats *BenchmarkStats) {
	stats.TotalTime = stats.EnqueueTime + stats.DequeueTime

	fmt.Println("=== Benchmark Summary ===")
	fmt.Printf("Total Messages:        %d\n", stats.TotalMessages)
	fmt.Printf("Enqueue Time:          %v (%.2f msg/sec)\n",
		stats.EnqueueTime,
		float64(stats.EnqueueSuccesses)/stats.EnqueueTime.Seconds())
	fmt.Printf("Dequeue Time:          %v (%.2f msg/sec)\n",
		stats.DequeueTime,
		float64(stats.DequeueSuccesses)/stats.DequeueTime.Seconds())
	fmt.Printf("Total Time:            %v\n", stats.TotalTime)
	fmt.Printf("\nEnqueue Successes:     %d\n", stats.EnqueueSuccesses)
	fmt.Printf("Enqueue Failures:      %d\n", stats.EnqueueFailures)
	fmt.Printf("Dequeue Successes:     %d\n", stats.DequeueSuccesses)
	fmt.Printf("Dequeue Failures:      %d\n", stats.DequeueFailures)
	fmt.Printf("Acknowledge Successes: %d\n", stats.AcknowledgeSuccesses)
	fmt.Printf("Acknowledge Failures:  %d\n", stats.AcknowledgeFailures)
	fmt.Printf("Duplicate Messages:    %d\n", stats.DuplicateMessages)
	fmt.Printf("Missing Messages:      %d\n", len(stats.MissingMessages))

	// Overall pass/fail
	fmt.Println("\n=== Result ===")
	if stats.EnqueueSuccesses == stats.TotalMessages &&
		stats.DequeueSuccesses == stats.TotalMessages &&
		stats.AcknowledgeSuccesses == stats.TotalMessages &&
		len(stats.MissingMessages) == 0 &&
		stats.DuplicateMessages == 0 {
		fmt.Println("✅ PASS - All tests successful!")
	} else {
		fmt.Println("❌ FAIL - Issues detected")
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	concurrentClient *http.Client
)

type ConcurrentStats struct {
	TotalMessages        int64
	EnqueueSuccesses     int64
	EnqueueFailures      int64
	DequeueSuccesses     int64
	DequeueFailures      int64
	AcknowledgeSuccesses int64
	AcknowledgeFailures  int64
	DuplicateMessages    int64

	EnqueueTime          time.Duration
	DequeueTime          time.Duration

	MessagesReceived     sync.Map
	mu                   sync.Mutex
	MissingMessages      []string
}

func init() {
	concurrentClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}

func runConcurrentBenchmark(numMessages, numClients int, baseURL, queueName string) {

	fmt.Println("=== Concurrent Queue Benchmark Test ===")
	fmt.Printf("Total Messages: %d\n", numMessages)
	fmt.Printf("Concurrent Clients: %d\n", numClients)
	fmt.Printf("Messages per Client: %d\n", numMessages/numClients)
	fmt.Printf("Queue: %s\n", queueName)
	fmt.Printf("Server: %s\n\n", baseURL)

	stats := &ConcurrentStats{
		TotalMessages: int64(numMessages),
	}

	// Create queue
	fmt.Println("Step 1: Creating queue...")
	if err := createQueueConcurrent(queueName, baseURL); err != nil {
		fmt.Printf("❌ Failed to create queue: %v\n", err)
		return
	}
	fmt.Println("✅ Queue created successfully\n")

	// Concurrent enqueue
	fmt.Printf("Step 2: %d clients enqueuing %d messages concurrently...\n", numClients, numMessages)
	startEnqueue := time.Now()
	sentIDs := concurrentEnqueue(stats, numMessages, numClients, queueName, baseURL)
	stats.EnqueueTime = time.Since(startEnqueue)
	fmt.Printf("✅ Enqueue complete in %v\n", stats.EnqueueTime)
	fmt.Printf("   Successes: %d, Failures: %d\n", stats.EnqueueSuccesses, stats.EnqueueFailures)
	fmt.Printf("   Throughput: %.2f msg/sec\n\n", float64(stats.EnqueueSuccesses)/stats.EnqueueTime.Seconds())

	time.Sleep(100 * time.Millisecond)

	// Concurrent dequeue
	fmt.Printf("Step 3: %d clients dequeuing %d messages concurrently...\n", numClients, numMessages)
	startDequeue := time.Now()
	concurrentDequeue(stats, numMessages, numClients, queueName, baseURL)
	stats.DequeueTime = time.Since(startDequeue)
	fmt.Printf("✅ Dequeue complete in %v\n", stats.DequeueTime)
	fmt.Printf("   Successes: %d, Failures: %d\n", stats.DequeueSuccesses, stats.DequeueFailures)
	fmt.Printf("   Acknowledgments: %d successes, %d failures\n", stats.AcknowledgeSuccesses, stats.AcknowledgeFailures)
	fmt.Printf("   Throughput: %.2f msg/sec\n\n", float64(stats.DequeueSuccesses)/stats.DequeueTime.Seconds())

	// Verify
	fmt.Println("Step 4: Verifying results...")
	verifyConcurrentResults(stats, sentIDs)

	printConcurrentSummary(stats, numMessages, numClients)
}

func createQueueConcurrent(queueName, baseURL string) error {
	return createQueueCommon(queueName, baseURL, concurrentClient)
}

func concurrentEnqueue(stats *ConcurrentStats, numMessages, numClients int, queueName, baseURL string) []string {
	var sentIDs []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	messagesPerClient := numMessages / numClients

	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()

			startIdx := cid * messagesPerClient
			endIdx := startIdx + messagesPerClient

			// Last client handles remainder
			if cid == numClients-1 {
				endIdx = numMessages
			}

			for i := startIdx; i < endIdx; i++ {
				itemID := fmt.Sprintf("item-%d", i)
				payload := map[string]interface{}{
					"queueName": queueName,
					"item": map[string]interface{}{
						"id":   itemID,
						"data": fmt.Sprintf("Message %d from client %d", i, cid),
					},
				}

				body, _ := json.Marshal(payload)
				resp, err := retryWithBackoff(func() (*http.Response, error) {
					return concurrentClient.Post(baseURL+"/enqueue", "application/json", bytes.NewBuffer(body))
				}, fmt.Sprintf("enqueue %s", itemID))

				if err != nil {
					atomic.AddInt64(&stats.EnqueueFailures, 1)
					continue
				}

				if resp.StatusCode == http.StatusCreated {
					atomic.AddInt64(&stats.EnqueueSuccesses, 1)
					mu.Lock()
					sentIDs = append(sentIDs, itemID)
					mu.Unlock()
				} else {
					atomic.AddInt64(&stats.EnqueueFailures, 1)
				}
				resp.Body.Close()
			}
		}(clientID)
	}

	wg.Wait()
	return sentIDs
}

func concurrentDequeue(stats *ConcurrentStats, numMessages, numClients int, queueName, baseURL string) {
	var wg sync.WaitGroup

	messagesPerClient := numMessages / numClients

	for clientID := 0; clientID < numClients; clientID++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()

			messagesToDequeue := messagesPerClient
			// Last client handles remainder
			if cid == numClients-1 {
				messagesToDequeue = numMessages - (messagesPerClient * (numClients - 1))
			}

			for i := 0; i < messagesToDequeue; i++ {
				payload := map[string]interface{}{
					"queueName": queueName,
				}

				body, _ := json.Marshal(payload)
				resp, err := retryWithBackoff(func() (*http.Response, error) {
					return concurrentClient.Post(baseURL+"/dequeue", "application/json", bytes.NewBuffer(body))
				}, fmt.Sprintf("dequeue client-%d #%d", cid, i))

				if err != nil {
					atomic.AddInt64(&stats.DequeueFailures, 1)
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
						atomic.AddInt64(&stats.DequeueFailures, 1)
					} else {
						atomic.AddInt64(&stats.DequeueSuccesses, 1)

						// Check for duplicates
						if _, exists := stats.MessagesReceived.LoadOrStore(result.Item.ID, true); exists {
							atomic.AddInt64(&stats.DuplicateMessages, 1)
							fmt.Printf("⚠️  Client %d: Duplicate message detected: %s\n", cid, result.Item.ID)
						}

						// Acknowledge
						if err := acknowledgeMessageConcurrent(result.Item.ID, queueName, baseURL); err != nil {
							atomic.AddInt64(&stats.AcknowledgeFailures, 1)
						} else {
							atomic.AddInt64(&stats.AcknowledgeSuccesses, 1)
						}
					}
				} else {
					atomic.AddInt64(&stats.DequeueFailures, 1)
				}
				resp.Body.Close()
			}
		}(clientID)
	}

	wg.Wait()
}

func acknowledgeMessageConcurrent(itemID, queueName, baseURL string) error {
	return acknowledgeMessageCommon(itemID, queueName, baseURL, concurrentClient)
}

func verifyConcurrentResults(stats *ConcurrentStats, sentIDs []string) {
	for _, id := range sentIDs {
		if _, exists := stats.MessagesReceived.Load(id); !exists {
			stats.mu.Lock()
			stats.MissingMessages = append(stats.MissingMessages, id)
			stats.mu.Unlock()
		}
	}

	if len(stats.MissingMessages) == 0 && stats.DuplicateMessages == 0 {
		fmt.Println("✅ All messages accounted for - no duplicates, no missing messages")
	} else {
		if len(stats.MissingMessages) > 0 {
			fmt.Printf("❌ Missing %d messages\n", len(stats.MissingMessages))
		}
		if stats.DuplicateMessages > 0 {
			fmt.Printf("❌ Found %d duplicate messages\n", stats.DuplicateMessages)
		}
	}
	fmt.Println()
}

func printConcurrentSummary(stats *ConcurrentStats, numMessages, numClients int) {
	totalTime := stats.EnqueueTime + stats.DequeueTime

	fmt.Println("=== Concurrent Benchmark Summary ===")
	fmt.Printf("Total Messages:        %d\n", stats.TotalMessages)
	fmt.Printf("Concurrent Clients:    %d\n", numClients)
	fmt.Printf("Enqueue Time:          %v (%.2f msg/sec)\n",
		stats.EnqueueTime,
		float64(stats.EnqueueSuccesses)/stats.EnqueueTime.Seconds())
	fmt.Printf("Dequeue Time:          %v (%.2f msg/sec)\n",
		stats.DequeueTime,
		float64(stats.DequeueSuccesses)/stats.DequeueTime.Seconds())
	fmt.Printf("Total Time:            %v\n", totalTime)
	fmt.Printf("\nEnqueue Successes:     %d\n", stats.EnqueueSuccesses)
	fmt.Printf("Enqueue Failures:      %d\n", stats.EnqueueFailures)
	fmt.Printf("Dequeue Successes:     %d\n", stats.DequeueSuccesses)
	fmt.Printf("Dequeue Failures:      %d\n", stats.DequeueFailures)
	fmt.Printf("Acknowledge Successes: %d\n", stats.AcknowledgeSuccesses)
	fmt.Printf("Acknowledge Failures:  %d\n", stats.AcknowledgeFailures)
	fmt.Printf("Duplicate Messages:    %d\n", stats.DuplicateMessages)
	fmt.Printf("Missing Messages:      %d\n", len(stats.MissingMessages))

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

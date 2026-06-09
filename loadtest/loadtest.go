// Package loadtest drives a CoraMQ server over HTTP to verify data-integrity
// invariants under concurrency and to measure throughput/latency.
//
// The model is competing consumers: many producers fill one shared queue, then
// many consumers race to drain it. Afterwards we assert four invariants:
//
//	no loss        every enqueued ID was received
//	no duplicates  every ID was received at most once
//	no corruption  each received payload matched what was enqueued for that ID
//	no phantoms    no ID was received that was never enqueued
//
// Corruption is detectable because every payload embeds its own ID (see
// expectedPayload): if the server ever swaps fields between items under a race,
// the received ID and payload will disagree.
package loadtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Config parameterises a run.
type Config struct {
	BaseURL      string        // e.g. http://127.0.0.1:8080
	QueueName    string        // queue to create and exercise
	NumMessages  int           // total messages to enqueue
	NumProducers int           // concurrent enqueue goroutines
	NumConsumers int           // concurrent dequeue+ack goroutines
	Timeout      time.Duration // per-request HTTP timeout
}

func (c *Config) withDefaults() {
	if c.QueueName == "" {
		c.QueueName = "integrity-test"
	}
	if c.NumMessages <= 0 {
		c.NumMessages = 10000
	}
	if c.NumConsumers <= 0 {
		c.NumConsumers = 10
	}
	if c.NumProducers <= 0 {
		c.NumProducers = c.NumConsumers
	}
	if c.Timeout <= 0 {
		// Must exceed the server's 30s long-poll so a legitimate long-poll isn't
		// cut short, but bounded so a wedged request can't hang the run forever.
		c.Timeout = 40 * time.Second
	}
}

// expectedPayload is deterministic and embeds the ID so corruption is detectable.
func expectedPayload(id string) string { return "payload-for-" + id }

// drainClose fully drains then closes a response body. This is the key to
// avoiding ephemeral-port / file-descriptor exhaustion: net/http only returns a
// connection to the keep-alive pool when the body is read to EOF *and* closed.
// Closing without draining discards the connection, forcing a fresh TCP dial per
// request and piling sockets into TIME_WAIT.
func drainClose(resp *http.Response) {
	if resp == nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// newClient builds an HTTP client tuned for high-concurrency localhost load. The
// pool is sized to the worker count so idle connections are reused rather than
// evicted and re-dialed.
func newClient(workers int, timeout time.Duration) *http.Client {
	pool := workers + 16
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        pool,
			MaxIdleConnsPerHost: pool, // >= concurrency or idle conns get evicted -> churn
			MaxConnsPerHost:     pool, // hard cap: never exceed the pool
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: timeout,
	}
}

// Run executes the full produce-then-drain cycle and returns the measured result.
// It returns an error only for setup failures (e.g. the queue can't be created);
// invariant violations are reported in the Result, not as errors.
func Run(ctx context.Context, cfg Config) (*Result, error) {
	cfg.withDefaults()
	workers := cfg.NumConsumers
	if cfg.NumProducers > workers {
		workers = cfg.NumProducers
	}
	client := newClient(workers, cfg.Timeout)

	if err := createQueue(ctx, client, cfg); err != nil {
		return nil, fmt.Errorf("create queue: %w", err)
	}

	res := &Result{NumMessages: cfg.NumMessages}

	expected, encLat, encDur := produce(ctx, client, cfg, res)
	res.EnqueueDuration = encDur
	res.EnqueueLatency = summarise(encLat)
	if encDur > 0 {
		res.EnqueueThroughput = float64(res.EnqueueSuccess) / encDur.Seconds()
	}

	deqLat, ackLat, conDur, received := consume(ctx, client, cfg, res, expected)
	res.ConsumeDuration = conDur
	res.DequeueLatency = summarise(deqLat)
	res.AckLatency = summarise(ackLat)
	if conDur > 0 {
		res.ConsumeThroughput = float64(res.DequeueSuccess) / conDur.Seconds()
	}

	verify(res, expected, received)
	return res, nil
}

func createQueue(ctx context.Context, client *http.Client, cfg Config) error {
	body, _ := json.Marshal(map[string]interface{}{
		"name":   cfg.QueueName,
		"config": map[string]interface{}{"deadLetterQueueRetries": 3},
	})
	resp, err := post(ctx, client, cfg.BaseURL+"/createQueue", body)
	if err != nil {
		return err
	}
	defer drainClose(resp)
	// 201 = created, 400 = already exists; both are fine to proceed.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// produce enqueues cfg.NumMessages messages across cfg.NumProducers goroutines and
// returns the set of IDs that were successfully enqueued (the set we expect back).
func produce(ctx context.Context, client *http.Client, cfg Config, res *Result) (map[string]string, []time.Duration, time.Duration) {
	expected := make(map[string]string, cfg.NumMessages)
	var mu sync.Mutex // guards expected + latency merge
	var allLat []time.Duration

	work := make(chan int, cfg.NumMessages)
	for i := 0; i < cfg.NumMessages; i++ {
		work <- i
	}
	close(work)

	start := time.Now()
	var wg sync.WaitGroup
	for p := 0; p < cfg.NumProducers; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := make([]time.Duration, 0, cfg.NumMessages/cfg.NumProducers+1)
			localExpected := make(map[string]string)
			for i := range work {
				id := fmt.Sprintf("item-%d", i)
				payload := expectedPayload(id)
				body, _ := json.Marshal(map[string]interface{}{
					"queueName": cfg.QueueName,
					"item":      map[string]interface{}{"id": id, "payload": payload},
				})

				t0 := time.Now()
				resp, err := postRetry(ctx, client, cfg.BaseURL+"/enqueue", body)
				local = append(local, time.Since(t0))
				if err != nil {
					atomic.AddInt64(&res.EnqueueFail, 1)
					continue
				}
				if resp.StatusCode == http.StatusCreated {
					atomic.AddInt64(&res.EnqueueSuccess, 1)
					localExpected[id] = payload
				} else {
					atomic.AddInt64(&res.EnqueueFail, 1)
				}
				drainClose(resp)
			}
			mu.Lock()
			allLat = append(allLat, local...)
			for k, v := range localExpected {
				expected[k] = v
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return expected, allLat, time.Since(start)
}

// consume runs cfg.NumConsumers goroutines competing to dequeue+ack until every
// expected message has been received. A shared context is cancelled the instant
// the target is reached so a consumer parked in the server's long-poll returns
// immediately instead of waiting out the timeout.
func consume(ctx context.Context, client *http.Client, cfg Config, res *Result, expected map[string]string) ([]time.Duration, []time.Duration, time.Duration, *sync.Map) {
	target := int64(len(expected))
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dequeueBody, _ := json.Marshal(map[string]string{"queueName": cfg.QueueName})

	// received maps id -> payload. LoadOrStore makes a duplicate delivery
	// observable rather than silently overwritten.
	var received sync.Map
	var receivedCount int64

	var mu sync.Mutex
	var allDeqLat, allAckLat []time.Duration

	start := time.Now()
	var wg sync.WaitGroup
	for c := 0; c < cfg.NumConsumers; c++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var deqLat, ackLat []time.Duration

			for atomic.LoadInt64(&receivedCount) < target {
				t0 := time.Now()
				resp, err := post(cctx, client, cfg.BaseURL+"/dequeue", dequeueBody)
				deqLat = append(deqLat, time.Since(t0))
				if err != nil {
					if cctx.Err() != nil {
						break // target reached; cancelled mid-flight
					}
					atomic.AddInt64(&res.DequeueFail, 1)
					continue
				}

				if resp.StatusCode == http.StatusTooManyRequests {
					drainClose(resp)
					time.Sleep(50 * time.Millisecond)
					continue
				}
				// Non-200 == empty queue / long-poll timeout: nothing to do.
				if resp.StatusCode != http.StatusOK {
					drainClose(resp)
					if atomic.LoadInt64(&receivedCount) >= target {
						break
					}
					continue
				}

				var out struct {
					Item struct {
						ID      string `json:"id"`
						Payload string `json:"payload"`
					} `json:"item"`
				}
				decErr := json.NewDecoder(resp.Body).Decode(&out)
				drainClose(resp)
				if decErr != nil {
					atomic.AddInt64(&res.DequeueFail, 1)
					continue
				}
				atomic.AddInt64(&res.DequeueSuccess, 1)

				if _, dup := received.LoadOrStore(out.Item.ID, out.Item.Payload); dup {
					atomic.AddInt64(&res.Duplicates, 1)
				}

				a0 := time.Now()
				ackErr := acknowledge(cctx, client, cfg, out.Item.ID)
				ackLat = append(ackLat, time.Since(a0))
				if ackErr != nil {
					atomic.AddInt64(&res.AckFail, 1)
				} else {
					atomic.AddInt64(&res.AckSuccess, 1)
				}

				if atomic.AddInt64(&receivedCount, 1) >= target {
					cancel() // wake any consumer parked in a long-poll
					break
				}
			}

			mu.Lock()
			allDeqLat = append(allDeqLat, deqLat...)
			allAckLat = append(allAckLat, ackLat...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	dur := time.Since(start)

	return allDeqLat, allAckLat, dur, &received
}

func acknowledge(ctx context.Context, client *http.Client, cfg Config, id string) error {
	body, _ := json.Marshal(map[string]string{"queueName": cfg.QueueName, "id": id})
	resp, err := postRetry(ctx, client, cfg.BaseURL+"/acknowledge", body)
	if err != nil {
		return err
	}
	defer drainClose(resp)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ack status %d", resp.StatusCode)
	}
	return nil
}

// verify walks the received set against the expected set and fills in the
// invariant counters on res.
func verify(res *Result, expected map[string]string, received *sync.Map) {
	if received == nil {
		return
	}
	for id := range expected {
		if _, ok := received.Load(id); !ok {
			res.Missing = append(res.Missing, id)
		}
	}
	var phantoms, corrupted int64
	received.Range(func(k, v interface{}) bool {
		id := k.(string)
		payload := v.(string)
		want, enqueued := expected[id]
		if !enqueued {
			phantoms++ // received something we never sent
		} else if payload != want {
			corrupted++ // sent it, but the bytes changed
		}
		return true
	})
	res.Phantoms = phantoms
	res.Corrupted = corrupted
}

// RunInterleaved exercises the long-poll delivery path that Run never touches.
// It starts the consumers FIRST so they park as waiters on the empty queue, then
// fires the producers concurrently. Every enqueue then goes through the
// enqueue->waiter handoff (sendItemToOldestWaitingClient), which is where the
// double-delivery / no-in-flight-tracking / stale-waiter bugs live.
//
// Expected signals when those bugs are present: Duplicates > 0 (same item handed
// to a waiter AND left dequeue-able in the heap) and AckFail > 0 (waiter-delivered
// items are never added to inFlight, so their ack is rejected).
func RunInterleaved(ctx context.Context, cfg Config) (*Result, error) {
	cfg.withDefaults()
	workers := cfg.NumConsumers
	if cfg.NumProducers > workers {
		workers = cfg.NumProducers
	}
	client := newClient(workers, cfg.Timeout)

	if err := createQueue(ctx, client, cfg); err != nil {
		return nil, fmt.Errorf("create queue: %w", err)
	}

	res := &Result{NumMessages: cfg.NumMessages}
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		expected   = make(map[string]string, cfg.NumMessages)
		expectedMu sync.Mutex
		enqueued   int64 // successful enqueues
		distinct   int64 // distinct IDs received at least once

		received sync.Map // id -> payload

		latMu                  sync.Mutex
		encLat, deqLat, ackLat []time.Duration
	)

	dequeueBody, _ := json.Marshal(map[string]string{"queueName": cfg.QueueName})

	// --- consumers start first, so they're parked as waiters before any enqueue ---
	consumeStart := time.Now()
	var consumerWg sync.WaitGroup
	for c := 0; c < cfg.NumConsumers; c++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			var localDeq, localAck []time.Duration
			for cctx.Err() == nil {
				t0 := time.Now()
				resp, err := post(cctx, client, cfg.BaseURL+"/dequeue", dequeueBody)
				localDeq = append(localDeq, time.Since(t0))
				if err != nil {
					if cctx.Err() != nil {
						break
					}
					atomic.AddInt64(&res.DequeueFail, 1)
					continue
				}
				if resp.StatusCode == http.StatusTooManyRequests {
					drainClose(resp)
					time.Sleep(50 * time.Millisecond)
					continue
				}
				if resp.StatusCode != http.StatusOK {
					drainClose(resp) // empty / long-poll timeout
					continue
				}
				var out struct {
					Item struct {
						ID      string `json:"id"`
						Payload string `json:"payload"`
					} `json:"item"`
				}
				decErr := json.NewDecoder(resp.Body).Decode(&out)
				drainClose(resp)
				if decErr != nil {
					atomic.AddInt64(&res.DequeueFail, 1)
					continue
				}
				atomic.AddInt64(&res.DequeueSuccess, 1)

				if _, dup := received.LoadOrStore(out.Item.ID, out.Item.Payload); dup {
					atomic.AddInt64(&res.Duplicates, 1)
				} else {
					atomic.AddInt64(&distinct, 1)
				}

				a0 := time.Now()
				ackErr := acknowledge(cctx, client, cfg, out.Item.ID)
				localAck = append(localAck, time.Since(a0))
				if ackErr != nil {
					atomic.AddInt64(&res.AckFail, 1)
				} else {
					atomic.AddInt64(&res.AckSuccess, 1)
				}
			}
			latMu.Lock()
			deqLat = append(deqLat, localDeq...)
			ackLat = append(ackLat, localAck...)
			latMu.Unlock()
		}()
	}

	// Give consumers a moment to register as waiters on the empty queue.
	select {
	case <-time.After(100 * time.Millisecond):
	case <-cctx.Done():
	}

	// --- producers ---
	enqueueStart := time.Now()
	work := make(chan int, cfg.NumMessages)
	for i := 0; i < cfg.NumMessages; i++ {
		work <- i
	}
	close(work)

	var producerWg sync.WaitGroup
	for p := 0; p < cfg.NumProducers; p++ {
		producerWg.Add(1)
		go func() {
			defer producerWg.Done()
			var local []time.Duration
			localExp := make(map[string]string)
			for i := range work {
				id := fmt.Sprintf("item-%d", i)
				payload := expectedPayload(id)
				body, _ := json.Marshal(map[string]interface{}{
					"queueName": cfg.QueueName,
					"item":      map[string]interface{}{"id": id, "payload": payload},
				})
				t0 := time.Now()
				resp, err := postRetry(cctx, client, cfg.BaseURL+"/enqueue", body)
				local = append(local, time.Since(t0))
				if err != nil {
					atomic.AddInt64(&res.EnqueueFail, 1)
					continue
				}
				if resp.StatusCode == http.StatusCreated {
					atomic.AddInt64(&res.EnqueueSuccess, 1)
					atomic.AddInt64(&enqueued, 1)
					localExp[id] = payload
				} else {
					atomic.AddInt64(&res.EnqueueFail, 1)
				}
				drainClose(resp)
			}
			latMu.Lock()
			encLat = append(encLat, local...)
			latMu.Unlock()
			expectedMu.Lock()
			for k, v := range localExp {
				expected[k] = v
			}
			expectedMu.Unlock()
		}()
	}
	producerWg.Wait()
	res.EnqueueDuration = time.Since(enqueueStart)

	// Monitor: stop consumers once every produced message has been received at
	// least once, or after a grace period (then anything still missing is real
	// loss and gets reported as such instead of hanging the run).
	go func() {
		grace := time.NewTimer(45 * time.Second)
		defer grace.Stop()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-cctx.Done():
				return
			case <-grace.C:
				cancel()
				return
			case <-ticker.C:
				if atomic.LoadInt64(&distinct) >= atomic.LoadInt64(&enqueued) {
					cancel()
					return
				}
			}
		}
	}()

	consumerWg.Wait()
	res.ConsumeDuration = time.Since(consumeStart)

	res.EnqueueLatency = summarise(encLat)
	res.DequeueLatency = summarise(deqLat)
	res.AckLatency = summarise(ackLat)
	if res.EnqueueDuration > 0 {
		res.EnqueueThroughput = float64(res.EnqueueSuccess) / res.EnqueueDuration.Seconds()
	}
	if res.ConsumeDuration > 0 {
		res.ConsumeThroughput = float64(res.DequeueSuccess) / res.ConsumeDuration.Seconds()
	}

	verify(res, expected, &received)
	return res, nil
}

// --- small HTTP helpers ---

func post(ctx context.Context, client *http.Client, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

// postRetry retries on HTTP 429 with a short fixed backoff (all localhost clients
// share one rate limiter, so bursts of 429s are expected under load).
func postRetry(ctx context.Context, client *http.Client, url string, body []byte) (*http.Response, error) {
	const maxAttempts = 12
	for attempt := 0; ; attempt++ {
		resp, err := post(ctx, client, url, body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests || attempt >= maxAttempts {
			return resp, nil
		}
		drainClose(resp)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(50*(attempt+1)) * time.Millisecond):
		}
	}
}

package loadtest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// RunSoak drives a *sustained* load for a fixed duration at a fixed target enqueue
// rate, while N consumers continuously drain. Unlike Run/RunInterleaved (which are
// burst tests), this answers the production question: "can it hold this rate
// without latency drift or unbounded backlog growth?"
//
// Every ReportInterval it prints a window line: actual enqueue/dequeue/ack rates,
// dequeue latency percentiles for that window, and the current queue depth. Flat
// depth + stable p99 across windows = consumers keeping up. Climbing depth or
// rising p99 = the next ceiling found.
//
// Soak measures performance, not correctness — it does not track per-message
// invariants (that would need unbounded memory). Use drain/interleaved for those.
func RunSoak(ctx context.Context, cfg Config) (*Result, error) {
	cfg.withDefaults()
	cfg.withSoakDefaults()

	workers := cfg.NumConsumers
	if cfg.NumProducers > workers {
		workers = cfg.NumProducers
	}
	client := newClient(workers, cfg.Timeout)

	if err := createQueue(ctx, client, cfg); err != nil {
		return nil, fmt.Errorf("create queue: %w", err)
	}

	res := &Result{}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	win := &soakWindow{}
	dequeueBody, _ := json.Marshal(map[string]string{"queueName": cfg.QueueName})

	// --- consumers: continuous dequeue+ack until the run ends ---
	var consumerWg sync.WaitGroup
	for c := 0; c < cfg.NumConsumers; c++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			for runCtx.Err() == nil {
				t0 := time.Now()
				resp, err := post(runCtx, client, cfg.BaseURL+"/dequeue", dequeueBody)
				dlat := time.Since(t0)
				if err != nil {
					if runCtx.Err() != nil {
						return
					}
					win.add(kDeq, dlat, false)
					continue
				}
				if resp.StatusCode != http.StatusOK {
					drainClose(resp) // empty / long-poll timeout / 429
					win.add(kDeq, dlat, false)
					continue
				}
				var out struct {
					Item struct {
						ID string `json:"id"`
					} `json:"item"`
				}
				decErr := json.NewDecoder(resp.Body).Decode(&out)
				drainClose(resp)
				if decErr != nil {
					win.add(kDeq, dlat, false)
					continue
				}
				win.add(kDeq, dlat, true)

				a0 := time.Now()
				ackErr := acknowledge(runCtx, client, cfg, out.Item.ID)
				win.add(kAck, time.Since(a0), ackErr == nil)
			}
		}()
	}

	// --- producers: paced to the target rate, stop after Duration ---
	producerCtx, stopProducers := context.WithCancel(runCtx)
	var limiter *rate.Limiter
	if cfg.Rate > 0 {
		burst := int(cfg.Rate / 10)
		if burst < 1 {
			burst = 1
		}
		limiter = rate.NewLimiter(rate.Limit(cfg.Rate), burst)
	}

	var idSeq int64
	var producerWg sync.WaitGroup
	for p := 0; p < cfg.NumProducers; p++ {
		producerWg.Add(1)
		go func() {
			defer producerWg.Done()
			for producerCtx.Err() == nil {
				if limiter != nil {
					if err := limiter.Wait(producerCtx); err != nil {
						return // ctx cancelled
					}
				}
				id := fmt.Sprintf("item-%d", atomic.AddInt64(&idSeq, 1))
				body, _ := json.Marshal(map[string]interface{}{
					"queueName": cfg.QueueName,
					"item":      map[string]interface{}{"id": id, "payload": expectedPayload(id)},
				})
				t0 := time.Now()
				resp, err := post(producerCtx, client, cfg.BaseURL+"/enqueue", body)
				elat := time.Since(t0)
				if err != nil {
					if producerCtx.Err() != nil {
						return
					}
					win.add(kEnq, elat, false)
					continue
				}
				win.add(kEnq, elat, resp.StatusCode == http.StatusCreated)
				drainClose(resp)
			}
		}()
	}

	// Stop producers after the configured duration; consumers get a grace period
	// afterwards so the report shows the backlog draining (or not).
	go func() {
		select {
		case <-time.After(cfg.Duration):
		case <-runCtx.Done():
		}
		stopProducers()
	}()

	// --- reporting loop ---
	fmt.Printf("%-8s %-22s %-14s %-30s %s\n", "t", "enqueue", "dequeue", "dequeue latency", "depth")
	fmt.Println("------------------------------------------------------------------------------------------")

	drainGrace := 2 * cfg.ReportInterval
	overall := time.NewTimer(cfg.Duration + drainGrace)
	defer overall.Stop()
	ticker := time.NewTicker(cfg.ReportInterval)
	defer ticker.Stop()

	start := time.Now()
	var depthPeak, depthFinal int
	for done := false; !done; {
		select {
		case <-ticker.C:
			s := win.drain()
			accumulate(res, s)
			depth := getDepth(runCtx, client, cfg.BaseURL, cfg.QueueName)
			if depth > depthPeak {
				depthPeak = depth
			}
			depthFinal = depth
			printSoakWindow(time.Since(start), s, cfg.ReportInterval, depth)
		case <-overall.C:
			done = true
		case <-ctx.Done():
			done = true
		}
	}

	cancel()
	producerWg.Wait()
	consumerWg.Wait()

	// Fold in any samples from the final partial window.
	accumulate(res, win.drain())

	printSoakSummary(res, cfg, depthPeak, depthFinal)
	return res, nil
}

func (c *Config) withSoakDefaults() {
	if c.Rate < 0 {
		c.Rate = 0
	}
	if c.Rate == 0 {
		c.Rate = 10000
	}
	if c.Duration <= 0 {
		c.Duration = 60 * time.Second
	}
	if c.ReportInterval <= 0 {
		c.ReportInterval = 5 * time.Second
	}
	if c.NumProducers <= 0 {
		c.NumProducers = 20
	}
}

// --- windowed metric collection ---

const (
	kEnq = iota
	kDeq
	kAck
)

type soakWindow struct {
	mu                                             sync.Mutex
	enqLat, deqLat, ackLat                         []time.Duration
	enqOK, enqFail, deqOK, deqFail, ackOK, ackFail int64
}

func (w *soakWindow) add(kind int, d time.Duration, ok bool) {
	w.mu.Lock()
	switch kind {
	case kEnq:
		w.enqLat = append(w.enqLat, d)
		if ok {
			w.enqOK++
		} else {
			w.enqFail++
		}
	case kDeq:
		w.deqLat = append(w.deqLat, d)
		if ok {
			w.deqOK++
		} else {
			w.deqFail++
		}
	case kAck:
		w.ackLat = append(w.ackLat, d)
		if ok {
			w.ackOK++
		} else {
			w.ackFail++
		}
	}
	w.mu.Unlock()
}

type windowSnap struct {
	enqLat, deqLat, ackLat                         []time.Duration
	enqOK, enqFail, deqOK, deqFail, ackOK, ackFail int64
}

func (w *soakWindow) drain() windowSnap {
	w.mu.Lock()
	s := windowSnap{
		enqLat: w.enqLat, deqLat: w.deqLat, ackLat: w.ackLat,
		enqOK: w.enqOK, enqFail: w.enqFail,
		deqOK: w.deqOK, deqFail: w.deqFail,
		ackOK: w.ackOK, ackFail: w.ackFail,
	}
	w.enqLat, w.deqLat, w.ackLat = nil, nil, nil
	w.enqOK, w.enqFail, w.deqOK, w.deqFail, w.ackOK, w.ackFail = 0, 0, 0, 0, 0, 0
	w.mu.Unlock()
	return s
}

func accumulate(res *Result, s windowSnap) {
	res.EnqueueSuccess += s.enqOK
	res.EnqueueFail += s.enqFail
	res.DequeueSuccess += s.deqOK
	res.DequeueFail += s.deqFail
	res.AckSuccess += s.ackOK
	res.AckFail += s.ackFail
}

// --- reporting ---

func printSoakWindow(elapsed time.Duration, s windowSnap, interval time.Duration, depth int) {
	secs := interval.Seconds()
	deq := summarise(s.deqLat)
	depthStr := "n/a"
	if depth >= 0 {
		depthStr = fmt.Sprintf("%d", depth)
	}
	fmt.Printf("%-8s %-22s %-14s p50 %-6s p99 %-7s %s\n",
		elapsed.Round(time.Second).String(),
		fmt.Sprintf("%.0f/s (ok %d, f %d)", float64(s.enqOK)/secs, s.enqOK, s.enqFail),
		fmt.Sprintf("%.0f/s", float64(s.deqOK)/secs),
		deq.P50.Round(time.Microsecond).String(),
		deq.P99.Round(time.Microsecond).String(),
		depthStr,
	)
}

func printSoakSummary(res *Result, cfg Config, depthPeak, depthFinal int) {
	fmt.Println("\n=== Soak Summary ===")
	fmt.Printf("Duration:      %v\n", cfg.Duration)
	fmt.Printf("Target rate:   %.0f msg/sec\n", cfg.Rate)
	actualEnq := float64(res.EnqueueSuccess) / cfg.Duration.Seconds()
	fmt.Printf("Enqueue:       %d ok / %d fail (actual %.0f/s)\n", res.EnqueueSuccess, res.EnqueueFail, actualEnq)
	fmt.Printf("Dequeue:       %d ok / %d fail\n", res.DequeueSuccess, res.DequeueFail)
	fmt.Printf("Ack:           %d ok / %d fail\n", res.AckSuccess, res.AckFail)
	fmt.Printf("Queue depth:   peak %d, final %d\n", depthPeak, depthFinal)

	fmt.Println("\n=== Verdict (speed-wise) ===")
	switch {
	case cfg.Rate > 0 && actualEnq < cfg.Rate*0.9:
		fmt.Printf("⚠️  Could not sustain target rate (%.0f/s vs target %.0f/s) — server is the bottleneck.\n", actualEnq, cfg.Rate)
	case depthFinal > int(cfg.Rate): // backlog larger than ~1s of intake and not draining
		fmt.Printf("⚠️  Backlog did not drain (final depth %d) — consumers can't keep up at this rate; sustained load would grow memory.\n", depthFinal)
	default:
		fmt.Println("✅ Sustained the target rate with a stable, draining backlog at this load.")
	}
	fmt.Println("(Watch the server process's RSS separately — depth is the client-side proxy for backlog growth.)")
}

func getDepth(ctx context.Context, client *http.Client, baseURL, queueName string) int {
	u := baseURL + "/sizeOfQueue?queueName=" + url.QueryEscape(queueName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return -1
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1
	}
	defer drainClose(resp)
	if resp.StatusCode != http.StatusOK {
		return -1
	}
	var out struct {
		Size int `json:"size"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil {
		return -1
	}
	return out.Size
}

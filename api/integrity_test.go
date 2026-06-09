//go:build integration

package api

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"go-queue-service/loadtest"
	"go-queue-service/queue_server"
	"go-queue-service/utils/logger"
)

// TestConcurrentIntegrity spins up the full HTTP stack in-process and drives it
// with the loadtest engine: many producers fill one queue, many consumers race to
// drain it, and we assert nothing was lost, duplicated, corrupted, or invented.
//
// Run with: go test -tags=integration ./api
func TestConcurrentIntegrity(t *testing.T) {
	log := logger.New("IntegrityTest", logger.ERROR).WithDevelopmentMode(false)
	server := queue_server.NewQueueServer(log, 40)

	// Build the real routes, but with no rate limiter so 429s don't muddy the
	// correctness signal (the network CLI exercises the rate-limited path).
	mux := SetupRoutes(GetAPIConfig(nil, server))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	res, err := loadtest.Run(ctx, loadtest.Config{
		BaseURL:      ts.URL,
		QueueName:    "integrity",
		NumMessages:  10000,
		NumConsumers: 50,
	})
	if err != nil {
		t.Fatalf("loadtest run failed: %v", err)
	}

	t.Log("\n" + res.String())

	if n := len(res.Missing); n != 0 {
		t.Errorf("lost %d messages (e.g. %v)", n, firstN(res.Missing, 10))
	}
	if res.Duplicates != 0 {
		t.Errorf("got %d duplicate deliveries", res.Duplicates)
	}
	if res.Corrupted != 0 {
		t.Errorf("got %d corrupted payloads", res.Corrupted)
	}
	if res.Phantoms != 0 {
		t.Errorf("got %d phantom IDs (never enqueued)", res.Phantoms)
	}
	if res.EnqueueFail != 0 {
		t.Errorf("got %d enqueue failures", res.EnqueueFail)
	}
	if res.AckFail != 0 {
		t.Errorf("got %d ack failures", res.AckFail)
	}
}

func firstN(s []string, n int) []string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

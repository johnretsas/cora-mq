// Command benchmark_tool runs CoraMQ's data-integrity + performance load test
// against a *running* server over the network.
//
//	go run ./benchmark_tool -url http://127.0.0.1:8080 -n 10000 -c 50
//
// It exits non-zero if any correctness invariant is violated, so it can gate CI.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go-queue-service/loadtest"
)

func main() {
	mode := flag.String("mode", "drain", "Test mode: 'drain' (produce all, then consume) or 'interleaved' (consumers park as waiters, then producers fire — exercises the long-poll delivery path)")
	url := flag.String("url", "http://127.0.0.1:8080", "Base URL of the running queue server")
	queue := flag.String("queue", "integrity-test", "Queue name to create and exercise")
	n := flag.Int("n", 10000, "Number of messages to enqueue")
	consumers := flag.Int("c", 50, "Number of competing consumers")
	producers := flag.Int("p", 0, "Number of concurrent producers (default: same as -c)")
	timeout := flag.Duration("timeout", 40*time.Second, "Per-request HTTP timeout")
	flag.Parse()

	cfg := loadtest.Config{
		BaseURL:      *url,
		QueueName:    *queue,
		NumMessages:  *n,
		NumConsumers: *consumers,
		NumProducers: *producers,
		Timeout:      *timeout,
	}

	fmt.Println("=== CoraMQ Integrity + Performance Test ===")
	fmt.Printf("Mode:   %s\nServer: %s\nQueue:  %s\nMessages: %d, Consumers: %d\n\n",
		*mode, cfg.BaseURL, cfg.QueueName, cfg.NumMessages, cfg.NumConsumers)

	var (
		res *loadtest.Result
		err error
	)
	switch *mode {
	case "drain":
		res, err = loadtest.Run(context.Background(), cfg)
	case "interleaved":
		res, err = loadtest.RunInterleaved(context.Background(), cfg)
	default:
		fmt.Printf("❌ unknown -mode %q (want 'drain' or 'interleaved')\n", *mode)
		os.Exit(2)
	}
	if err != nil {
		fmt.Printf("❌ run failed: %v\n", err)
		os.Exit(2)
	}

	fmt.Println(res.String())

	if !res.Clean() {
		os.Exit(1)
	}
}

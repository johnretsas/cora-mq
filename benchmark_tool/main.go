package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define all flags
	mode := flag.String("mode", "sequential", "Benchmark mode: 'sequential' or 'concurrent'")

	// Sequential flags
	numItems := flag.Int("n", 4000, "Number of messages (sequential mode)")
	numMessages := flag.Int("messages", 4000, "Number of messages")
	delayMs := flag.Int("delay", 0, "Delay in milliseconds between requests (sequential mode)")

	// Concurrent flags
	numClients := flag.Int("c", 10, "Number of concurrent clients (concurrent mode)")
	numClientsLong := flag.Int("clients", 10, "Number of concurrent clients")

	// Common flags
	baseURL := flag.String("url", "http://127.0.0.1:8080", "Base URL of the queue server")
	queueName := flag.String("queue", "", "Name of the queue (default varies by mode)")

	flag.Parse()

	// Set default queue names if not specified
	if *queueName == "" {
		if *mode == "sequential" {
			*queueName = "benchmark-queue"
		} else {
			*queueName = "concurrent-benchmark-queue"
		}
	}

	switch *mode {
	case "sequential":
		// Use -n if specified, otherwise use -messages
		messages := *numItems
		if flag.Lookup("messages").Value.String() != "4000" {
			messages = *numMessages
		}
		runSequentialBenchmark(messages, *baseURL, *queueName, *delayMs)
	case "concurrent":
		// Use -c if specified, otherwise use -clients
		clients := *numClients
		if flag.Lookup("clients").Value.String() != "10" {
			clients = *numClientsLong
		}
		// Use -n if specified, otherwise use -messages
		messages := *numItems
		if flag.Lookup("messages").Value.String() != "4000" {
			messages = *numMessages
		}
		runConcurrentBenchmark(messages, clients, *baseURL, *queueName)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Usage: go run . -mode [sequential|concurrent] [other flags...]")
		os.Exit(1)
	}
}

.PHONY: test test-unit test-integration test-all test-coverage test-queue test-server build run echo init test-watch

# Pretty test output via gotestsum (go install gotest.tools/gotestsum@latest).
# Override the format, e.g.: make test TEST_FORMAT=pkgname
GOTESTSUM ?= $(shell go env GOPATH)/bin/gotestsum
TEST_FORMAT ?= testdox
GOTEST = $(GOTESTSUM) --format $(TEST_FORMAT) --

echo:
	echo "Hello World"

# Run only fast unit tests (skips integration tests with sleep)
test-unit:
	@echo "Running unit tests (fast)..."
	@$(GOTEST) -short ./...

# Run only integration tests (with sleep/timing)
test-integration:
	@echo "Running integration tests (slow)..."
	@$(GOTEST) -tags=integration ./queue

# Run all tests (unit + integration)
test-all:
	@echo "Running all tests..."
	@$(GOTEST) -tags=integration ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -short -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run only queue package tests
test-queue:
	@echo "Running queue tests..."
	@$(GOTEST) ./queue

# Run only queue_server package tests
test-server:
	@echo "Running queue_server tests..."
	@$(GOTEST) ./queue_server

# Default test command (unit tests only)
test: test-unit

# Re-run unit tests on file save
test-watch:
	@$(GOTESTSUM) --format $(TEST_FORMAT) --watch -- -short ./...

# Build the binary
build:
	mkdir -p ./bin && go build -o ./bin/queue-service .

# Run the server
run:
	go run main.go

# Run the data-integrity + performance load test against a RUNNING server.
# Default 'drain' mode: produce all messages, then drain with competing consumers.
# Override defaults with ARGS, e.g.: make bench ARGS="-n 50000 -c 100 -url http://127.0.0.1:8080"
bench:
	go run ./benchmark_tool $(ARGS)

# Interleaved mode: consumers start first and park as waiters, then producers fire.
# Exercises the long-poll enqueue->waiter delivery path (see todos/long-poll-delivery-bugs.md).
# Override defaults with ARGS, e.g.: make bench-interleaved ARGS="-n 50000 -c 100"
bench-interleaved:
	go run ./benchmark_tool -mode interleaved $(ARGS)

# Soak mode: sustain a fixed enqueue rate for a fixed duration, reporting dequeue
# latency percentiles and queue depth over time. The real production-readiness test:
# flat depth + stable p99 = consumers keeping up; climbing depth = falling behind.
# Override defaults with ARGS, e.g.: make bench-soak ARGS="-rate 20000 -duration 5m -c 100"
bench-soak:
	go run ./benchmark_tool -mode soak $(ARGS)

# Run the in-process integrity test (spins the server up internally, asserts invariants)
test-integrity:
	@echo "Running in-process integrity test..."
	@$(GOTEST) -tags=integration -run TestConcurrentIntegrity ./api

# Start the server, create "testQueue" and enqueue a test message
init:
	./scripts/init.sh

# Clean up build artifacts and test outputs
clean:
	rm -rf ./bin coverage.out coverage.html

# Help command
help:
	@echo "Available commands:"
	@echo "  make test              - Run fast unit tests (default)"
	@echo "  make test-unit         - Run only unit tests (fast)"
	@echo "  make test-integration  - Run only integration tests (slow, requires time)"
	@echo "  make test-all          - Run all tests (unit + integration)"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-queue        - Run only queue package tests"
	@echo "  make test-server       - Run only queue_server package tests"
	@echo "  make test-watch        - Re-run unit tests on file save"
	@echo "  make build             - Build the binary"
	@echo "  make run               - Run the server"
	@echo "  make init              - Start server, create testQueue, send a test message"
	@echo "  make clean             - Clean up build artifacts"

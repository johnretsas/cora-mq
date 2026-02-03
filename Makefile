.PHONY: test test-unit test-integration test-all test-coverage test-queue test-server build run echo

echo:
	echo "Hello World"

# Run only fast unit tests (skips integration tests with sleep)
test-unit:
	@echo "Running unit tests (fast)..."
	@go test -short -v ./...

# Run only integration tests (with sleep/timing)
test-integration:
	@echo "Running integration tests (slow)..."
	@go test -tags=integration -v ./queue

# Run all tests (unit + integration)
test-all:
	@echo "Running all tests..."
	@go test -tags=integration -v ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -short -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run only queue package tests
test-queue:
	@echo "Running queue tests..."
	@go test ./queue -v

# Run only queue_server package tests
test-server:
	@echo "Running queue_server tests..."
	@go test ./queue_server -v

# Default test command (unit tests only)
test: test-unit

# Build the binary
build:
	mkdir -p ./bin && go build -o ./bin/queue-service .

# Run the server
run:
	go run main.go

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
	@echo "  make build             - Build the binary"
	@echo "  make run               - Run the server"
	@echo "  make clean             - Clean up build artifacts"

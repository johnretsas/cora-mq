.PHONY: test run

test-queue:
	go test ./queue -v

test-server:
	go test ./queue_server -v

run:
	go run main.go
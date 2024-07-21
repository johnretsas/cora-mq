.PHONY: test run

test-queue:
	go test ./queue -v

test-server:
	go test ./queue_server -v

build:
	mkdir -p ./bin && go build -o bin/queue_server ./queue_server

run:
	go run main.go
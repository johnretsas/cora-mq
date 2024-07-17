.PHONY: test run

test:
	go test ./queue

run:
	go run main.go
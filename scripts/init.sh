#!/bin/sh
# Starts the queue server, creates "testQueue" and enqueues a test message.
# The server keeps running in the foreground; Ctrl+C stops it.
set -eu

PORT="${PORT:-8080}"
BASE_URL="http://127.0.0.1:${PORT}"
QUEUE_NAME="${QUEUE_NAME:-testQueue}"

PORT="$PORT" go run main.go &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null' INT TERM EXIT

echo "Waiting for server on ${BASE_URL}..."
i=0
until curl -sf "${BASE_URL}/health" >/dev/null; do
	i=$((i + 1))
	if [ "$i" -ge 60 ]; then
		echo "Server did not become healthy in time" >&2
		exit 1
	fi
	sleep 0.5
done

echo "Creating queue ${QUEUE_NAME}..."
curl -sS -X POST "${BASE_URL}/createQueue" \
	-H "Content-Type: application/json" \
	-d "{\"name\": \"${QUEUE_NAME}\"}"
echo

echo "Enqueueing test message..."
curl -sS -X POST "${BASE_URL}/enqueue" \
	-H "Content-Type: application/json" \
	-d "{
		\"queueName\": \"${QUEUE_NAME}\",
		\"item\": {
			\"id\": \"2000\",
			\"payload\": \"{'name': 'John Retsas', 'email': 'joretsas@gmail.com'}\"
		}
	}"
echo

wait "$SERVER_PID"

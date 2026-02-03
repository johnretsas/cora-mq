//go:build integration

package queue

import (
	"testing"
	"time"
)

// TestVisibilityTimeout tests that unacknowledged items become visible again
// after the visibility timeout expires (requires 11 second sleep)
func TestVisibilityTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	q := newTestQueue()

	// Enqueue two items
	q.Enqueue(NewQueueItem("1", "First message", 1))
	q.Enqueue(NewQueueItem("2", "Second message", 1))

	// Dequeue first item
	item1, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item1, &QueueItem{ID: "1", Payload: "First message", Priority: 1})

	// Wait for visibility timeout to expire (default 10 seconds + 1 second buffer)
	t.Log("Waiting 11 seconds for visibility timeout...")
	time.Sleep(11 * time.Second)

	// Dequeue again - should get item 1 again (not acknowledged)
	item1Again, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() after timeout error = %v, want nil", err)
	}
	assertItemEquals(t, item1Again, &QueueItem{ID: "1", Payload: "First message", Priority: 1})

	// Now acknowledge it
	q.Acknowledge("1")

	// Next dequeue should get item 2
	item2, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item2, &QueueItem{ID: "2", Payload: "Second message", Priority: 1})
}

// TestAcknowledgeWithinVisibilityTimeout tests that acknowledged items
// don't reappear after the visibility timeout
func TestAcknowledgeWithinVisibilityTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	q := newTestQueue()

	// Enqueue three items
	q.Enqueue(NewQueueItem("1", "First", 1))
	q.Enqueue(NewQueueItem("2", "Second", 1))
	q.Enqueue(NewQueueItem("3", "Third", 1))

	// Dequeue and acknowledge first two items
	item1, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item1, &QueueItem{ID: "1", Payload: "First", Priority: 1})

	item2, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item2, &QueueItem{ID: "2", Payload: "Second", Priority: 1})

	q.Acknowledge("1")
	q.Acknowledge("2")

	// Wait 6 seconds (less than visibility timeout)
	t.Log("Waiting 6 seconds...")
	time.Sleep(6 * time.Second)

	// Next dequeue should get item 3 (not 1 or 2, since they were acknowledged)
	item3, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item3, &QueueItem{ID: "3", Payload: "Third", Priority: 1})
}

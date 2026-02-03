package queue

import (
	"testing"
)

// Helper function to create a test queue
func newTestQueue() *Queue {
	config := QueueConfig{DeadLetterQueueRetries: 3}
	return NewQueue(config, "TestQueue")
}

// Helper function to assert queue item equality
func assertItemEquals(t *testing.T, got, want *QueueItem) {
	t.Helper()
	if got.ID != want.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, want.ID)
	}
	if got.Payload != want.Payload {
		t.Errorf("Payload mismatch: got %q, want %q", got.Payload, want.Payload)
	}
	if got.Priority != want.Priority {
		t.Errorf("Priority mismatch: got %d, want %d", got.Priority, want.Priority)
	}
}

func TestEnqueueAndSize(t *testing.T) {
	q := newTestQueue()
	q.Enqueue(NewQueueItem("1", "Payload1"))

	if got := q.Size(); got != 1 {
		t.Errorf("Size() = %d, want 1", got)
	}
}

func TestDequeue(t *testing.T) {
	q := newTestQueue()
	q.Enqueue(NewQueueItem("1", "Payload1"))

	item, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}

	want := &QueueItem{ID: "1", Payload: "Payload1", Priority: 1}
	assertItemEquals(t, item, want)
}

func TestDequeueEmptyQueue(t *testing.T) {
	q := newTestQueue()

	_, err := q.Dequeue()
	if err == nil {
		t.Error("Dequeue() on empty queue should return error")
	}
}

func TestDefaultPriority(t *testing.T) {
	q := newTestQueue()
	q.Enqueue(NewQueueItem("1", "Payload1"))

	item, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}

	if item.Priority != 1 {
		t.Errorf("Default priority = %d, want 1", item.Priority)
	}
}

func TestPriorityOrdering(t *testing.T) {
	q := newTestQueue()

	// Enqueue messages with different priorities
	q.Enqueue(NewQueueItem("1", "Low priority 1", 1))
	q.Enqueue(NewQueueItem("2", "Low priority 2", 1))
	q.Enqueue(NewQueueItem("3", "High priority 1", 3))
	q.Enqueue(NewQueueItem("4", "High priority 2", 3))
	q.Enqueue(NewQueueItem("5", "Low priority 3", 1))

	// Expected order: high priorities first (in order), then low priorities (in order)
	expectedOrder := []*QueueItem{
		{ID: "3", Payload: "High priority 1", Priority: 3},
		{ID: "4", Payload: "High priority 2", Priority: 3},
		{ID: "1", Payload: "Low priority 1", Priority: 1},
		{ID: "2", Payload: "Low priority 2", Priority: 1},
		{ID: "5", Payload: "Low priority 3", Priority: 1},
	}

	for i, want := range expectedOrder {
		item, err := q.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue %d: error = %v, want nil", i+1, err)
		}

		q.Acknowledge(item.ID)
		assertItemEquals(t, item, want)
	}

	// Queue should now be empty
	_, err := q.Dequeue()
	if err == nil {
		t.Error("Dequeue() on empty queue should return error")
	}
}

func TestMultipleEnqueueDequeue(t *testing.T) {
	q := newTestQueue()

	// Enqueue multiple items
	items := []struct {
		id       string
		payload  string
		priority int
	}{
		{"1", "First", 1},
		{"2", "Second", 1},
		{"3", "Third", 1},
	}

	for _, item := range items {
		q.Enqueue(NewQueueItem(item.id, item.payload, item.priority))
	}

	// Dequeue and verify order
	for _, expected := range items {
		item, err := q.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue() error = %v, want nil", err)
		}

		q.Acknowledge(item.ID)

		want := &QueueItem{ID: expected.id, Payload: expected.payload, Priority: expected.priority}
		assertItemEquals(t, item, want)
	}
}

func TestAcknowledgeMultipleItems(t *testing.T) {
	q := newTestQueue()

	// Enqueue 3 items
	q.Enqueue(NewQueueItem("1", "Payload 1", 1))
	q.Enqueue(NewQueueItem("2", "Payload 2", 1))
	q.Enqueue(NewQueueItem("3", "Payload 3", 1))

	// Dequeue first two items
	item1, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item1, &QueueItem{ID: "1", Payload: "Payload 1", Priority: 1})

	item2, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item2, &QueueItem{ID: "2", Payload: "Payload 2", Priority: 1})

	// Acknowledge both
	q.Acknowledge("1")
	q.Acknowledge("2")

	// Next dequeue should get item 3 (not 1 or 2 since they were acknowledged)
	item3, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue() error = %v, want nil", err)
	}
	assertItemEquals(t, item3, &QueueItem{ID: "3", Payload: "Payload 3", Priority: 1})
}

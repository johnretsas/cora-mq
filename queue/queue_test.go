package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	q := NewQueue()

	// Test Enqueue
	q.Enqueue(1)
	if q.Size() != 1 {
		t.Errorf("expected size 1, got %d", q.Size())
	}

	// Test Dequeue
	item, err := q.Dequeue()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item != 1 {
		t.Errorf("expected 1, got %v", item)
	}

	// Test Dequeue on empty queue
	_, err = q.Dequeue()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

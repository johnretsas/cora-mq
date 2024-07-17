package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	q := NewQueue()

	// Test Enqueue
	q.Enqueue(QueueItem{ID: "1", Payload: "Payload1"})
	if q.Size() != 1 {
		t.Errorf("expected size 1, got %d", q.Size())
	}

	// Test Dequeue
	item, err := q.Dequeue()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedItem := QueueItem{ID: "1", Payload: "Payload1"}
	// Ignore the 'index' field in the comparison
	if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
		t.Errorf("expected %v, got %v", expectedItem, item)
	}

	// Test Dequeue on empty queue
	_, err = q.Dequeue()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// Testing with q2
	q2 := NewQueue()
	// Using the constructor with default priority
	q2.Enqueue(NewQueueItem("1", "Now the low priorities will commence, in the order they were received"))
	q2.Enqueue(NewQueueItem("2", "Verifying that the priorities work well together", 1))
	q2.Enqueue(NewQueueItem("3", "This is the first high priority", 3))
	q2.Enqueue(NewQueueItem("4", "..then we have this again as high priority.", 3))
	q2.Enqueue(NewQueueItem("5", "and the order of the messages is being conserved", 1))

	item, _ = q2.Dequeue()
	expectedItem = QueueItem{ID: "3", Payload: "This is the first high priority", Priority: 3}
	// Ignore the 'index' field in the comparison
	if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
		t.Errorf("expected %v, got %v", expectedItem, item)
	}

	item, _ = q2.Dequeue()
	expectedItem = QueueItem{ID: "4", Payload: "..then we have this again as high priority.", Priority: 3}
	// Ignore the 'index' field in the comparison
	if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
		t.Errorf("expected %v, got %v", expectedItem, item)
	}

	item, _ = q2.Dequeue()
	expectedItem = QueueItem{ID: "1", Payload: "Now the low priorities will commence, in the order they were received", Priority: 1}
	// Ignore the 'index' field in the comparison
	if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
		t.Errorf("expected %v, got %v", expectedItem, item)
	}

	item, _ = q2.Dequeue()
	expectedItem = QueueItem{ID: "2", Payload: "Verifying that the priorities work well together", Priority: 1}
	// Ignore the 'index' field in the comparison
	if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
		t.Errorf("expected %v, got %v", expectedItem, item)
	}

	item, _ = q2.Dequeue()
	expectedItem = QueueItem{ID: "5", Payload: "and the order of the messages is being conserved", Priority: 1}
	// Ignore the 'index' field in the comparison
	if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
		t.Errorf("expected %v, got %v", expectedItem, item)
	}

	// Test Dequeue on empty queue
	_, err = q2.Dequeue()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

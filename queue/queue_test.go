package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	t.Run("TestEnqueueAndSize", func(t *testing.T) {
		q := NewQueue()
		q.Enqueue(NewQueueItem("1", "Payload1"))
		if q.Size() != 1 {
			t.Errorf("expected size 1, got %d", q.Size())
		}
	})

	t.Run("TestDequeue", func(t *testing.T) {
		q := NewQueue()
		q.Enqueue(NewQueueItem("1", "Payload1"))
		item, err := q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectedItem := QueueItem{ID: "1", Payload: "Payload1", Priority: 1}
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, *item)
		}
	})

	t.Run("TestAcknowledge", func(t *testing.T) {
		queue := NewQueue()
		item1 := QueueItem{ID: "5", Payload: "Some payload"}
		queue.Enqueue(item1)

		item2 := QueueItem{ID: "6", Payload: "Some payload"}
		queue.Enqueue(item2)

		// Dequeue item with id 5
		dequeuedItem, _ := queue.Dequeue()
		fmt.Println(dequeuedItem.ID) // Should print "5"

		// If you don’t acknowledge the item within 5 seconds,
		// and attempt to dequeue again:
		time.Sleep(6 * time.Second)
		dequeuedItem, _ = queue.Dequeue()
		fmt.Println(dequeuedItem.ID) // Should still print "5" if not acknowledged

		// Acknowledge the item
		queue.Acknowledge("5")

		// Dequeue again
		dequeuedItem, _ = queue.Dequeue()
		fmt.Println(dequeuedItem.ID) // Should print "6"
	})

	t.Run("TestMultipleEnqueueDequeue", func(t *testing.T) {
		q := NewQueue()
		q.Enqueue(NewQueueItem("1", "Now the low priorities will commence, in the order they were received"))
		q.Enqueue(NewQueueItem("2", "Verifying that the priorities work well together", 1))
		q.Enqueue(NewQueueItem("3", "This is the first high priority", 3))
		q.Enqueue(NewQueueItem("4", "..then we have this again as high priority.", 3))
		q.Enqueue(NewQueueItem("5", "and the order of the messages is being conserved", 1))

		item, err := q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem := QueueItem{ID: "3", Payload: "This is the first high priority", Priority: 3}
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = QueueItem{ID: "4", Payload: "..then we have this again as high priority.", Priority: 3}
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = QueueItem{ID: "1", Payload: "Now the low priorities will commence, in the order they were received", Priority: 1}
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = QueueItem{ID: "2", Payload: "Verifying that the priorities work well together", Priority: 1}
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = QueueItem{ID: "5", Payload: "and the order of the messages is being conserved", Priority: 1}
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		// Test Dequeue on empty queue
		_, err = q.Dequeue()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

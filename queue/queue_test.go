package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	t.Run("TestEnqueueAndSize", func(t *testing.T) {
		config := QueueConfig{DeadLetterQueueRetries: 3}
		q := NewQueue(config, "TestQueue")
		q.Enqueue(NewQueueItem("1", "Payload1"))
		if q.Size() != 1 {
			t.Errorf("expected size 1, got %d", q.Size())
		}
	})

	t.Run("TestDequeue", func(t *testing.T) {
		config := QueueConfig{DeadLetterQueueRetries: 3}
		q := NewQueue(config, "TestQueue")
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

	t.Run("TestDefaultPriority", func(t *testing.T) {
		config := QueueConfig{DeadLetterQueueRetries: 3}
		q := NewQueue(config, "TestQueue")
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

	// If the item is not acknowledged before the visibility timeout, it becomes available again
	t.Run("TestAcknowledge1", func(t *testing.T) {
		config := QueueConfig{DeadLetterQueueRetries: 3}
		queue := NewQueue(config, "TestQueue")
		item1 := NewQueueItem("5", "Some payload", 1)
		queue.Enqueue(item1)

		item2 := NewQueueItem("6", "Some payload", 1)
		queue.Enqueue(item2)

		// Dequeue item with id 5
		dequeuedItem, _ := queue.Dequeue()
		fmt.Printf("dequeuedItem: %v\n", *dequeuedItem)
		expectedItem1 := NewQueueItem("5", "Some payload", 1)

		if dequeuedItem.ID != expectedItem1.ID || dequeuedItem.Payload != expectedItem1.Payload || dequeuedItem.Priority != expectedItem1.Priority {
			t.Errorf("expected %v, got %v", expectedItem1, dequeuedItem)
		}

		// If you don’t acknowledge the item within 10 seconds,
		// and attempt to dequeue again:
		time.Sleep(11 * time.Second)
		dequeuedItem, _ = queue.Dequeue()

		expectedItem2 := NewQueueItem("5", "Some payload", 1)
		if dequeuedItem.ID != expectedItem2.ID || dequeuedItem.Payload != expectedItem2.Payload || dequeuedItem.Priority != expectedItem2.Priority {
			t.Errorf("expected %v, got %v", expectedItem2, dequeuedItem)
		}

		// Acknowledge the item
		queue.Acknowledge("5")

		// Dequeue again
		dequeuedItem, _ = queue.Dequeue()
		expectedItem3 := NewQueueItem("6", "Some payload", 1)
		if dequeuedItem.ID != expectedItem3.ID || dequeuedItem.Payload != expectedItem3.Payload || dequeuedItem.Priority != expectedItem3.Priority {
			t.Errorf("expected %v, got %v", expectedItem3, dequeuedItem)
		}

	})

	t.Run("TestAcknowledge2", func(t *testing.T) {
		config := QueueConfig{DeadLetterQueueRetries: 3}
		queue := NewQueue(config, "TestQueue")

		item1 := NewQueueItem("5", "Some payload", 1)
		queue.Enqueue(item1)

		item2 := NewQueueItem("6", "Some payload", 1)
		queue.Enqueue(item2)

		item3 := NewQueueItem("7", "Some payload", 1)
		queue.Enqueue(item3)

		res1, err := queue.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectedItem := NewQueueItem("5", "Some payload", 1)
		if res1.ID != expectedItem.ID || res1.Payload != expectedItem.Payload || res1.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, res1)
		}

		res2, err := queue.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectedItem = NewQueueItem("6", "Some payload", 1)
		if res2.ID != expectedItem.ID || res2.Payload != expectedItem.Payload || res2.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, res2)
		}

		// Acknowledge the first item
		queue.Acknowledge("5")
		queue.Acknowledge("6")

		time.Sleep(6 * time.Second)

		// Dequeue again
		res3, err := queue.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectedItem = NewQueueItem("7", "Some payload", 1)
		if res3.ID != expectedItem.ID || res3.Payload != expectedItem.Payload || res3.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, res3)
		}
	})

	t.Run("TestMultipleEnqueueDequeue", func(t *testing.T) {
		config := QueueConfig{DeadLetterQueueRetries: 3}
		q := NewQueue(config, "TestQueue")

		q.Enqueue(NewQueueItem("1", "Now the low priorities will commence, in the order they were received", 1))
		q.Enqueue(NewQueueItem("2", "Verifying that the priorities work well together", 1))
		q.Enqueue(NewQueueItem("3", "This is the first high priority", 3))
		q.Enqueue(NewQueueItem("4", "..then we have this again as high priority.", 3))
		q.Enqueue(NewQueueItem("5", "and the order of the messages is being conserved", 1))

		item, err := q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem := NewQueueItem("3", "This is the first high priority", 3)
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = NewQueueItem("4", "..then we have this again as high priority.", 3)
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = NewQueueItem("1", "Now the low priorities will commence, in the order they were received", 1)
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = NewQueueItem("2", "Verifying that the priorities work well together", 1)
		if item.ID != expectedItem.ID || item.Payload != expectedItem.Payload || item.Priority != expectedItem.Priority {
			t.Errorf("expected %v, got %v", expectedItem, item)
		}

		item, err = q.Dequeue()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		q.Acknowledge(item.ID)
		expectedItem = NewQueueItem("5", "and the order of the messages is being conserved", 1)
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

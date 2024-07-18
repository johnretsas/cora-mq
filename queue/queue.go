package queue

import (
	"fmt"
	"sync"
	"time"
)

type QueueItem struct {
	ID             string
	Payload        string
	Priority       int
	index          int       // The index of the item in the heap
	visibilityTime time.Time // The time at which the item becomes visible
	acknowledged   bool
}

type Queue struct {
	items    []QueueItem
	inFlight []QueueItem
	mu       sync.Mutex
	index    int // The index of the last item in the heap
	ackIndex int // The index of the last acknowledged item

}

func NewQueue() *Queue {
	return &Queue{
		items:    make([]QueueItem, 0),
		mu:       sync.Mutex{},
		inFlight: make([]QueueItem, 0),
		index:    0,
		ackIndex: 0,
	}
}

// NewQueueItem creates a new QueueItem with default priority of 1 if not provided
func NewQueueItem(id string, payload string, priority ...int) QueueItem {
	p := 1 // Default priority
	if len(priority) > 0 {
		p = priority[0]
	}
	return QueueItem{
		ID:       id,
		Payload:  payload,
		Priority: p,
	}
}

func QueueVisibilityScenario() {
	// Enqueue item with id 5
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
}

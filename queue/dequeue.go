package queue

import (
	"container/heap"
	"errors"
	"time"
)

// Dequeue retrieves the highest-priority visible item from the queue
func (q *Queue) Dequeue() (*QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Len() == 0 {
		return nil, errors.New("queue is empty")
	}

	tempList := make([]QueueItem, 0)

	for q.Len() > 0 {
		// Pop an item from the heap
		item := heap.Pop(q).(QueueItem)

		// Skip acknowledged items
		if item.acknowledged {
			continue
		}

		// If item is not visible, add to tempList
		if time.Now().Before(item.visibilityTime) {
			tempList = append(tempList, item)
			continue
		}

		// Item is visible, update visibility timeout and add to in-flight list
		item.visibilityTime = time.Now().Add(5 * time.Second)
		q.inFlight = append(q.inFlight, item)

		// Add the item you've just dequeued to the tempList
		// so that it can be re-added to the heap later
		// if it doesn't get acknowledged
		// it will be available again
		tempList = append(tempList, item)

		// Re-add all temporary items back to the heap
		for _, tempItem := range tempList {
			heap.Push(q, tempItem)
		}

		return &item, nil
	}

	// Re-add the items back to the queue if none were visible
	for _, item := range tempList {
		heap.Push(q, item)
	}

	return nil, errors.New("queue is empty or all items are invisible")
}

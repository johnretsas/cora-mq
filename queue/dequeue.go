package queue

import (
	"container/heap"
	"errors"
	"time"
)

// Dequeue retrieves the highest-priority visible item from the queue
// The first visible item is returned to the client and it is considered in-flight
// As items get dequeued and acknowledged, the dequeue method will remove them from the queue.
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

		// Remove (Pop) and skip acknowledged items
		if item.Acknowledged {
			continue
		}

		// If item is not visible, add to tempList
		// These are items that have been dequeued but not acknowledged
		// In this case, we don't want to remove them from the heap
		// because they will be re-added to the heap later if they are not acknowledged
		if time.Now().Before(item.visibilityTime) {
			tempList = append(tempList, item)
			continue
		}

		if item.Retries > q.deadLetterQueueRetries {
			// If the item has been retried more than 3 times, move it to the dead letter queue
			q.deadLetterQueue.Enqueue(item)
			continue
		}

		// At this point, we have found a visible item
		// The tempList contains all the items that are not visible yet

		// Item is visible, update visibility timeout and add to in-flight list
		// This item will become invisible for 10 seconds and it is considered in-flight
		// The client will have 10 seconds to acknowledge the item
		// If the item is not acknowledged within 10 seconds, it will be available again
		item.visibilityTime = time.Now().Add(10 * time.Second)
		item.Retries++

		// Add the item to the in-flight list
		// This list is used to keep track of items that have been dequeued but not acknowledged
		q.inFlight = append(q.inFlight, item)

		// Add the item you've just dequeued to the tempList
		// so that it can be re-added to the heap later
		// if it doesn't get acknowledged
		// it will be available again
		tempList = append(tempList, item)

		// In order to find a visible item, we need to skip the items that are not visible
		// These items have been popped and they should be re-added to the heap
		for _, tempItem := range tempList {
			heap.Push(q, tempItem)
		}

		// Return the visible item you have found.
		return &item, nil
	}

	// At this point, we have gone through all the items in the queue and none were visible
	// Re-add the items back to the queue if none were visible.
	// So we have to re-add all the invisible items back to the heap
	// and return an error
	for _, item := range tempList {
		heap.Push(q, item)
	}

	// If all items are invisible, return an error
	return nil, errors.New("queue is empty or all items are invisible")
}

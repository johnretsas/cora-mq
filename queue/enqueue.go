package queue

import (
	"container/heap"
	"time"
)

func (q *Queue) Enqueue(item QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item.index = q.index
	item.Acknowledged = false
	item.visibilityTime = time.Time{}

	// If priority is not set, set it to 1
	if item.Priority == 0 {
		item.Priority = 1
	}

	q.index++

	heap.Push(q, item)
}

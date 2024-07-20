package queue

import (
	"container/heap"
	"time"
)

func (q *Queue) Enqueue(item QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item.index = q.index
	item.acknowledged = false
	item.visibilityTime = time.Time{}

	if item.Priority == 0 {
		item.Priority = 1
	}

	q.index++

	heap.Push(q, item)
}

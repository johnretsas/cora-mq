package queue

import (
	"container/heap"
	"fmt"
	"time"
)

func (q *Queue) Enqueue(item QueueItem) {
	fmt.Println("Enqueue")
	q.mu.Lock()
	defer q.mu.Unlock()

	item.index = q.index
	item.acknowledged = false
	item.visibilityTime = time.Time{}

	q.index++

	heap.Push(q, item)
}

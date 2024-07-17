package queue

import (
	"container/heap"
	"errors"
)

func (q *Queue) Enqueue(item QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item.index = q.index
	q.index++

	heap.Push(q, item)
}

func (q *Queue) Dequeue() (*QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Len() == 0 {
		return nil, errors.New("queue is empty")
	}

	item := heap.Pop(q).(QueueItem)
	return &item, nil
}

package queue

import (
	"container/heap"
	"errors"
	"sync"
)

type QueueItem struct {
	ID       string
	Payload  string
	Priority int
	index    int // The index of the item in the heap
}

type Queue struct {
	items []QueueItem
	mu    sync.Mutex
	index int // The index of the last item in the heap
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]QueueItem, 0),
		index: 0,
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

// Implement the heap.Interface methods
func (q *Queue) Len() int { return len(q.items) }

func (q *Queue) Less(i, j int) bool {
	if q.items[i].Priority == q.items[j].Priority {
		return q.items[i].index < q.items[j].index
	}
	return q.items[i].Priority > q.items[j].Priority
}

func (q *Queue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *Queue) Push(x interface{}) {
	q.items = append(q.items, x.(QueueItem))
}

func (q *Queue) Pop() interface{} {
	old := q.items
	n := len(old)
	item := old[n-1]
	q.items = old[0 : n-1]
	return item
}

// Size returns the number of items in the queue.
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

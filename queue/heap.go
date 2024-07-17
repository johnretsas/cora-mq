package queue

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

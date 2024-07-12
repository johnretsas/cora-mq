package queue

import (
	"errors"
	"fmt"
	"sync"
)

type QueueItem struct {
	ID      string
	Payload string
}

type Queue struct {
	items []interface{}
	mu    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]interface{}, 0),
	}
}

func (q *Queue) Enqueue(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil, errors.New("queue is empty")
	}

	items := q.items[0]
	q.items = q.items[1:]
	return items, nil
}

// Size returns the number of items in the queue.
func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

func main() {
	q := &Queue{}

	q.Enqueue(QueueItem{ID: "1", Payload: "Payload1"})
	q.Enqueue(QueueItem{ID: "2", Payload: "Payload2"})
	q.Enqueue(QueueItem{ID: "3", Payload: "Payload3"})

	fmt.Println(q.Dequeue()) // Outputs: &{1 Payload1}
	fmt.Println(q.Dequeue()) // Outputs: &{2 Payload2}
	fmt.Println(q.Dequeue()) // Outputs: &{3 Payload3}
}

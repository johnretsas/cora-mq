package queue

import (
	"fmt"
	"time"
)

func (q *Queue) Scan() []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.items {
		fmt.Printf("ID: %s, Priority: %d, Visible: %t\n, Acknowledged: %t\n", item.ID, item.Priority, time.Now().After(item.visibilityTime), item.Acknowledged)
	}

	if len(q.items) < 100 {
		return q.items
	}
	return q.items[:100]
}

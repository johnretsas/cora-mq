package queue

import (
	"fmt"
	"time"
)

func (q *Queue) Scan() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.items {
		fmt.Printf("ID: %s, Priority: %d, Visible: %t\n", item.ID, item.Priority, time.Now().After(item.visibilityTime))
	}

}

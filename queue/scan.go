package queue

import (
	"time"
)

func (q *Queue) Scan() ([]QueueItem, []QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.items {
		q.logger.Printf("ID: %s, Priority: %d, Visible: %t\n, Acknowledged: %t\n", item.ID, item.Priority, time.Now().After(item.visibilityTime), item.Acknowledged)
	}

	basicQueueItems := q.items
	deadLetterQueueItems := q.deadLetterQueue.items

	if len(q.items) > 100 {
		basicQueueItems = q.items[:100]
	}

	if len(q.deadLetterQueue.items) > 100 {

		deadLetterQueueItems = q.deadLetterQueue.items[:100]
	}

	return basicQueueItems, deadLetterQueueItems
}

package queue

import (
	"errors"
	"fmt"
)

func (q *Queue) Acknowledge(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	fmt.Println("Items in queue:")
	for i := range q.items {
		if q.items[i].ID == id {
			q.items[i].Acknowledged = true
		}
		fmt.Println("Acknowledge item:")
		q.items[i].PrettyPrint()
	}

	for i, item := range q.inFlight {
		if item.ID == id {
			q.inFlight = append(q.inFlight[:i], q.inFlight[i+1:]...)
			q.ackIndex++
			return nil
		}
	}

	return errors.New("message with id wasn't found in the in-flight queue")
}

package queue

import (
	"errors"
	"fmt"
)

func (q *Queue) Acknowledge(id string) error {
	fmt.Println("Acknowledge")
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range q.items {
		if item.ID == id {
			item.acknowledged = true
		}
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

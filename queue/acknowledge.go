package queue

import (
	"errors"
	"fmt"
)

func (q *Queue) Acknowledge(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	fmt.Println("Items in queue:")
	// Scan the queue items and acknowledge the item with the given id
	// Arguably, not the most efficient way to do this
	// But assuming that the acknowledgement comes from the client fairly quickly after dequeuing
	// The message to be acknowledged should be close to the beginning of the queue
	for i := range q.items {
		if q.items[i].ID == id {
			q.items[i].Acknowledged = true
		}
		fmt.Println("Acknowledge item:")
		q.items[i].PrettyPrint()
	}

	// Finding the id in the in-flight queue
	// means that the item has been dequeued but not acknowledged
	// and we need to remove it from the in-flight queue
	// and increment the ackIndex
	for i, item := range q.inFlight {
		if item.ID == id {
			q.inFlight = append(q.inFlight[:i], q.inFlight[i+1:]...)
			q.ackIndex++
			return nil
		}
	}

	return errors.New("message with id wasn't found in the in-flight queue")
}

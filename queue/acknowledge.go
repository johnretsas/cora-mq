package queue

import (
	"errors"
)

func (q *Queue) Acknowledge(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// If the item is not in the in-flight queue, it means that we should not proceed with the acknowledgements
	var foundItem *QueueItem

	// find item with id in the in-flight queue
	for _, item := range q.inFlight {
		if item.ID == id {
			// Item found in in-flight queue
			foundItem = &item
			break
		}
	}

	if foundItem == nil {
		// q.logger.Printf("Item with id %s not found in the in-flight queue", id)
		return errors.New("message with id wasn't found in the in-flight queue")
	}

	// Scan the queue items and acknowledge the item with the given id
	// Arguably, not the most efficient way to do this
	// But assuming that the acknowledgement comes from the client fairly quickly after dequeuing
	// The message to be acknowledged should be close to the beginning of the queue
	for i := range q.items {
		if q.items[i].ID == id {
			q.items[i].Acknowledged = true
			// q.logger.Printf("Acknowledged item with id %s", id)
			q.items[i].PrettyPrint()
		}
	}

	// Finding the id in the in-flight queue
	// means that the item has been dequeued but not acknowledged
	// and we need to remove it from the in-flight queue

	// Remove the item from the in-flight queue
	// Create a new slice to store items that don't match the ID
	newInFlight := make([]QueueItem, 0, len(q.inFlight))

	// Iterate over the current inFlight items and keep only those that don't match the ID
	for _, item := range q.inFlight {
		if item.ID != id {
			newInFlight = append(newInFlight, item)
		}
	}

	// Replace the old inFlight slice with the new one
	q.inFlight = newInFlight

	return nil
}

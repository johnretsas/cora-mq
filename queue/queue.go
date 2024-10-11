package queue

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type QueueItem struct {
	ID             string    `json:"id"`
	Payload        string    `json:"payload"`
	Priority       int       `json:"priority"`
	index          int       // The index of the item in the heap
	visibilityTime time.Time // The time at which the item becomes visible
	Acknowledged   bool      `json:"acknowledged"`
	Retries        int       `json:"retries"` // Number of times the item has been retried
}

type Queue struct {
	items    []QueueItem
	inFlight []QueueItem
	mu       sync.Mutex
	index    int // The index of the last item in the heap
	ackIndex int // The index of the last acknowledged item

	deadLetterQueue *Queue // Dead letter queue
}

func NewQueue() *Queue {
	return &Queue{
		items:    make([]QueueItem, 0),
		mu:       sync.Mutex{},
		inFlight: make([]QueueItem, 0),
		index:    0,
		ackIndex: 0,
		deadLetterQueue: &Queue{
			items:           make([]QueueItem, 0),
			mu:              sync.Mutex{},
			inFlight:        make([]QueueItem, 0),
			index:           0,
			ackIndex:        0,
			deadLetterQueue: nil,
		},
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

func (qI *QueueItem) PrettyPrint() {
	data, err := json.MarshalIndent(qI, "", "  ")
	if err != nil {
		return
	}
	fmt.Println(string(data))
}

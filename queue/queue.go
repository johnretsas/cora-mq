package queue

import (
	"encoding/json"
	"fmt"
	"log"
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
	name     string
	items    []QueueItem
	inFlight []QueueItem
	mu       sync.Mutex
	index    int // The index of the last item in the heap
	ackIndex int // The index of the last acknowledged item

	deadLetterQueue        *Queue // Dead letter queue
	deadLetterQueueRetries int    // Number of retries for dead letter queue
	logger                 *log.Logger
}

type QueueConfig struct {
	DeadLetterQueueRetries int `json:"deadLetterQueueRetries"`
}

func NewQueue(config QueueConfig, name string) *Queue {
	// default value for dead letter queue retries
	if config.DeadLetterQueueRetries <= 0 {
		config.DeadLetterQueueRetries = 3
	}

	return &Queue{
		name:                   name,
		items:                  make([]QueueItem, 0),
		mu:                     sync.Mutex{},
		inFlight:               make([]QueueItem, 0),
		index:                  0,
		ackIndex:               0,
		deadLetterQueueRetries: config.DeadLetterQueueRetries,
		deadLetterQueue: &Queue{
			items:                  make([]QueueItem, 0),
			mu:                     sync.Mutex{},
			inFlight:               make([]QueueItem, 0),
			index:                  0,
			ackIndex:               0,
			deadLetterQueue:        nil,
			deadLetterQueueRetries: 0,
		},
		logger: log.New(log.Writer(), "Queue - "+time.Now().Format("2006-01-02 15:04:05")+" ", log.LstdFlags), // Initialize logger with timestamp
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

func (q *Queue) Lock() {
	q.mu.Lock()
}

func (q *Queue) Unlock() {
	q.mu.Unlock()
}

func (q *Queue) LenOfVisibleItems() int {
	count := 0
	for _, item := range q.items {
		if time.Now().After(item.visibilityTime) {
			count++
		}
	}
	fmt.Printf("Number of visible items: %d\n", count)
	return count
}

func (q *QueueItem) DeepCopy() QueueItem {
	return QueueItem{
		ID:             q.ID,
		Payload:        q.Payload,
		Priority:       q.Priority,
		index:          q.index,
		visibilityTime: q.visibilityTime,
		Acknowledged:   q.Acknowledged,
		Retries:        q.Retries,
	}
}

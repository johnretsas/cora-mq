package queue_server

import (
	"fmt"
	"go-queue-service/queue"
	"go-queue-service/utils/logger"
	"sync"
	"testing"
)

// Creates queues while other goroutines enqueue to an existing queue, so the
// queues map is read and written concurrently. Run with -race: an unguarded
// map read fails here.
func TestConcurrentCreateWithReads(t *testing.T) {
	log := logger.New("TestServer", logger.ERROR)
	server := NewQueueServer(log, 10)

	if _, err := server.CreateQueue("existing", queue.QueueConfig{}); err != nil {
		t.Fatalf("CreateQueue(existing) error = %v", err)
	}

	const numCreates = 200
	const numEnqueues = 200

	var wg sync.WaitGroup

	for i := 0; i < numCreates; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("queue-%d", i)
			if _, err := server.CreateQueue(name, queue.QueueConfig{}); err != nil {
				t.Errorf("CreateQueue(%s) error = %v", name, err)
			}
		}(i)
	}

	for i := 0; i < numEnqueues; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			item := queue.QueueItem{ID: fmt.Sprintf("item-%d", i), Payload: "test"}
			if _, err := server.Enqueue("existing", item); err != nil {
				t.Errorf("Enqueue(%s) error = %v", item.ID, err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < numCreates; i++ {
		name := fmt.Sprintf("queue-%d", i)
		if _, exists := server.getQueue(name); !exists {
			t.Errorf("queue %s was not created", name)
		}
	}

	q, _ := server.getQueue("existing")
	q.Lock()
	size := q.Len()
	q.Unlock()
	if size != numEnqueues {
		t.Errorf("existing queue size = %d, want %d", size, numEnqueues)
	}
}

// Many goroutines create the same queue at once: exactly one must succeed,
// otherwise the check-and-insert in ProcessCreateQueueRequest is not atomic.
func TestConcurrentDuplicateCreate(t *testing.T) {
	log := logger.New("TestServer", logger.ERROR)
	server := NewQueueServer(log, 10)

	const numCreates = 100

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0

	for i := 0; i < numCreates; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := server.CreateQueue("testQueue", queue.QueueConfig{}); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successes != 1 {
		t.Errorf("successful creates = %d, want 1", successes)
	}
}

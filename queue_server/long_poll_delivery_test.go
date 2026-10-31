package queue_server

import (
	"go-queue-service/queue"
	"go-queue-service/utils/logger"
	"testing"
	"time"
)

// Regression tests for the enqueue->waiter handoff described in
// todos/long-poll-delivery-bugs.md. Each test parks a consumer as a long-poll
// waiter on an empty queue, then enqueues, so the item goes through
// sendItemToOldestWaitingClient instead of a normal dequeue.

type dequeueResult struct {
	item queue.QueueItem
	err  error
}

func newLongPollTestServer(t *testing.T) *QueueServer {
	t.Helper()
	server := NewQueueServer(logger.New("TestServer", logger.ERROR), 3)
	if _, err := server.CreateQueue("lp", queue.QueueConfig{}); err != nil {
		t.Fatalf("CreateQueue error = %v", err)
	}
	return server
}

// startWaiter starts a Dequeue on an empty queue and returns once the consumer
// is registered in waitingListClients.
func startWaiter(t *testing.T, server *QueueServer, queueName string) <-chan dequeueResult {
	t.Helper()

	server.mu.RLock()
	before := len(server.waitingListClients[queueName])
	server.mu.RUnlock()

	resultCh := make(chan dequeueResult, 1)
	go func() {
		item, err := server.Dequeue(queueName)
		resultCh <- dequeueResult{item: item, err: err}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		server.mu.RLock()
		n := len(server.waitingListClients[queueName])
		server.mu.RUnlock()
		if n > before {
			return resultCh
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("consumer never registered as a waiter on %q", queueName)
	return nil
}

func receiveWithin(t *testing.T, resultCh <-chan dequeueResult, d time.Duration) dequeueResult {
	t.Helper()
	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("waiter Dequeue error = %v", res.err)
		}
		return res
	case <-time.After(d):
		t.Fatalf("waiter did not receive an item within %s", d)
		return dequeueResult{}
	}
}

// Bug 1: an item handed to a waiter must be in flight, not still visible for
// another consumer to dequeue.
func TestLongPollDeliveredItemIsNotVisible(t *testing.T) {
	server := newLongPollTestServer(t)
	resultCh := startWaiter(t, server, "lp")

	if _, err := server.Enqueue("lp", queue.NewQueueItem("x", "payload")); err != nil {
		t.Fatalf("Enqueue error = %v", err)
	}
	res := receiveWithin(t, resultCh, 2*time.Second)
	if res.item.ID != "x" {
		t.Fatalf("waiter got item %q, want %q", res.item.ID, "x")
	}

	q, _ := server.getQueue("lp")
	q.Lock()
	visible := q.LenOfVisibleItems()
	q.Unlock()
	if visible != 0 {
		t.Errorf("visible items after waiter delivery = %d, want 0 (item would be delivered twice)", visible)
	}
}

// Bug 2: an item handed to a waiter must be acknowledgeable.
func TestLongPollDeliveredItemCanBeAcknowledged(t *testing.T) {
	server := newLongPollTestServer(t)
	resultCh := startWaiter(t, server, "lp")

	if _, err := server.Enqueue("lp", queue.NewQueueItem("x", "payload")); err != nil {
		t.Fatalf("Enqueue error = %v", err)
	}
	res := receiveWithin(t, resultCh, 2*time.Second)

	if _, err := server.Acknowledge("lp", res.item.ID); err != nil {
		t.Errorf("Acknowledge(%q) error = %v, want nil", res.item.ID, err)
	}
}

// Bug 3: a waiter that timed out must not swallow the wake-up meant for a
// live waiter queued behind it. Slow: waits out the 30s long-poll timeout.
func TestLongPollTimedOutWaiterDoesNotBlockLiveWaiter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 30s long-poll timeout test in short mode")
	}

	server := newLongPollTestServer(t)

	staleCh := startWaiter(t, server, "lp")
	select {
	case res := <-staleCh:
		if res.err == nil {
			t.Fatalf("first waiter got item %q, want timeout", res.item.ID)
		}
	case <-time.After(35 * time.Second):
		t.Fatal("first waiter did not time out")
	}

	liveCh := startWaiter(t, server, "lp")
	if _, err := server.Enqueue("lp", queue.NewQueueItem("x", "payload")); err != nil {
		t.Fatalf("Enqueue error = %v", err)
	}

	res := receiveWithin(t, liveCh, 2*time.Second)
	if res.item.ID != "x" {
		t.Errorf("live waiter got item %q, want %q", res.item.ID, "x")
	}
}

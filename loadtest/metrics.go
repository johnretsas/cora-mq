package loadtest

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Latencies summarises a set of operation durations.
type Latencies struct {
	Count int
	P50   time.Duration
	P90   time.Duration
	P99   time.Duration
	P999  time.Duration
	Max   time.Duration
}

// summarise sorts the samples and extracts percentiles. The input slice is sorted
// in place.
func summarise(samples []time.Duration) Latencies {
	l := Latencies{Count: len(samples)}
	if len(samples) == 0 {
		return l
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
	l.P50 = percentile(samples, 0.50)
	l.P90 = percentile(samples, 0.90)
	l.P99 = percentile(samples, 0.99)
	l.P999 = percentile(samples, 0.999)
	l.Max = samples[len(samples)-1]
	return l
}

// percentile returns the p-th percentile of an already-sorted slice using the
// nearest-rank method.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// Result holds everything the run measured: the correctness invariants and the
// performance numbers.
type Result struct {
	NumMessages int

	// Throughput counters.
	EnqueueSuccess int64
	EnqueueFail    int64
	DequeueSuccess int64
	DequeueFail    int64
	AckSuccess     int64
	AckFail        int64

	// Invariant violations.
	Missing    []string // enqueued but never received
	Duplicates int64     // same ID delivered more than once
	Corrupted  int64     // payload did not match what was enqueued
	Phantoms   int64     // received an ID that was never enqueued

	// Timing.
	EnqueueDuration   time.Duration
	ConsumeDuration   time.Duration
	EnqueueThroughput float64 // msg/sec
	ConsumeThroughput float64 // msg/sec

	EnqueueLatency Latencies
	DequeueLatency Latencies
	AckLatency     Latencies
}

// Clean reports whether every correctness invariant held.
func (r *Result) Clean() bool {
	return len(r.Missing) == 0 &&
		r.Duplicates == 0 &&
		r.Corrupted == 0 &&
		r.Phantoms == 0 &&
		r.AckFail == 0 &&
		r.EnqueueFail == 0
}

// String renders a human-readable report.
func (r *Result) String() string {
	var b strings.Builder

	fmt.Fprintln(&b, "=== Throughput ===")
	fmt.Fprintf(&b, "Enqueue:  %d ok / %d fail in %v (%.0f msg/sec)\n",
		r.EnqueueSuccess, r.EnqueueFail, r.EnqueueDuration.Round(time.Millisecond), r.EnqueueThroughput)
	fmt.Fprintf(&b, "Consume:  %d ok / %d fail in %v (%.0f msg/sec)\n",
		r.DequeueSuccess, r.DequeueFail, r.ConsumeDuration.Round(time.Millisecond), r.ConsumeThroughput)
	fmt.Fprintf(&b, "Ack:      %d ok / %d fail\n", r.AckSuccess, r.AckFail)

	fmt.Fprintln(&b, "\n=== Latency (p50 / p90 / p99 / p999 / max) ===")
	fmt.Fprintf(&b, "Enqueue:  %s\n", latLine(r.EnqueueLatency))
	fmt.Fprintf(&b, "Dequeue:  %s\n", latLine(r.DequeueLatency))
	fmt.Fprintf(&b, "Ack:      %s\n", latLine(r.AckLatency))

	fmt.Fprintln(&b, "\n=== Invariants ===")
	fmt.Fprintf(&b, "Missing (lost):     %d\n", len(r.Missing))
	fmt.Fprintf(&b, "Duplicates:         %d\n", r.Duplicates)
	fmt.Fprintf(&b, "Corrupted payloads: %d\n", r.Corrupted)
	fmt.Fprintf(&b, "Phantom IDs:        %d\n", r.Phantoms)
	if len(r.Missing) > 0 && len(r.Missing) <= 20 {
		fmt.Fprintf(&b, "Missing IDs:        %v\n", r.Missing)
	}

	fmt.Fprintln(&b, "\n=== Result ===")
	if r.Clean() {
		fmt.Fprintln(&b, "✅ PASS — no loss, no duplicates, no corruption, no phantoms")
	} else {
		fmt.Fprintln(&b, "❌ FAIL — invariant violation detected")
	}
	return b.String()
}

func latLine(l Latencies) string {
	if l.Count == 0 {
		return "(no samples)"
	}
	return fmt.Sprintf("%s / %s / %s / %s / %s",
		l.P50.Round(time.Microsecond),
		l.P90.Round(time.Microsecond),
		l.P99.Round(time.Microsecond),
		l.P999.Round(time.Microsecond),
		l.Max.Round(time.Microsecond))
}

package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestAggregatorResult(t *testing.T) {
	aggregator := NewAggregator()
	aggregator.Record(10*time.Millisecond, false)
	aggregator.Record(30*time.Millisecond, true)

	result := aggregator.Result(50 * time.Millisecond)

	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}

	if result.Failed != 1 {
		t.Fatalf("expected failed 1, got %d", result.Failed)
	}

	if result.Duration != 50*time.Millisecond {
		t.Fatalf("expected duration 50ms, got %v", result.Duration)
	}

	if result.Latency.Min != 10*time.Millisecond {
		t.Fatalf("expected min 10ms, got %v", result.Latency.Min)
	}

	if result.Latency.Max != 30*time.Millisecond {
		t.Fatalf("expected max 30ms, got %v", result.Latency.Max)
	}

	if result.Latency.Mean != 20*time.Millisecond {
		t.Fatalf("expected mean 20ms, got %v", result.Latency.Mean)
	}

	if result.Latency.P50 != 10*time.Millisecond {
		t.Fatalf("expected p50 10ms, got %v", result.Latency.P50)
	}
	if result.Latency.P95 != 30*time.Millisecond {
		t.Fatalf("expected p95 30ms, got %v", result.Latency.P95)
	}
	if result.Latency.P99 != 30*time.Millisecond {
		t.Fatalf("expected p99 30ms, got %v", result.Latency.P99)
	}
}

func TestAggregatorConcurrentRecord(t *testing.T) {
	aggregator := NewAggregator()

	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
	}
	failures := []bool{false, true, false, true}

	var wg sync.WaitGroup
	for i := range latencies {
		wg.Add(1)
		go func(latency time.Duration, failed bool) {
			defer wg.Done()
			aggregator.Record(latency, failed)
		}(latencies[i], failures[i])
	}

	wg.Wait()

	result := aggregator.Result(100 * time.Millisecond)

	if result.Total != int64(len(latencies)) {
		t.Fatalf("expected total %d, got %d", len(latencies), result.Total)
	}

	if result.Failed != 2 {
		t.Fatalf("expected failed 2, got %d", result.Failed)
	}

	if result.Duration != 100*time.Millisecond {
		t.Fatalf("expected duration 100ms, got %v", result.Duration)
	}

	if result.Latency.Min != 10*time.Millisecond {
		t.Fatalf("expected min 10ms, got %v", result.Latency.Min)
	}

	if result.Latency.Max != 40*time.Millisecond {
		t.Fatalf("expected max 40ms, got %v", result.Latency.Max)
	}

	if result.Latency.Mean != 25*time.Millisecond {
		t.Fatalf("expected mean 25ms, got %v", result.Latency.Mean)
	}

	if result.Latency.P50 != 20*time.Millisecond {
		t.Fatalf("expected p50 20ms, got %v", result.Latency.P50)
	}
	if result.Latency.P95 != 40*time.Millisecond {
		t.Fatalf("expected p95 40ms, got %v", result.Latency.P95)
	}
	if result.Latency.P99 != 40*time.Millisecond {
		t.Fatalf("expected p99 40ms, got %v", result.Latency.P99)
	}
}

func TestAggregatorRetainsAllLatencies(t *testing.T) {
	a := NewAggregator()
	a.Record(time.Millisecond, false)
	a.Record(2*time.Millisecond, false)
	a.Record(3*time.Millisecond, true)

	if len(a.latencies) != 3 {
		t.Fatalf("expected 3 retained latencies, got %d", len(a.latencies))
	}
	if a.latencies[0] != time.Millisecond || a.latencies[1] != 2*time.Millisecond || a.latencies[2] != 3*time.Millisecond {
		t.Fatalf("unexpected retained order or values: %v", a.latencies)
	}
}

func TestPercentileFromSorted(t *testing.T) {
	s := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		100 * time.Millisecond,
	}
	if got := percentileFromSorted(s, 50); got != 3*time.Millisecond {
		t.Fatalf("p50: want 3ms, got %v", got)
	}
	if got := percentileFromSorted(s, 95); got != 100*time.Millisecond {
		t.Fatalf("p95: want 100ms, got %v", got)
	}
	if got := percentileFromSorted(s, 99); got != 100*time.Millisecond {
		t.Fatalf("p99: want 100ms, got %v", got)
	}
}

func TestAggregatorResultDoesNotMutateRetainedLatencies(t *testing.T) {
	a := NewAggregator()
	a.Record(30*time.Millisecond, false)
	a.Record(10*time.Millisecond, false)
	snapshot := append([]time.Duration(nil), a.latencies...)

	_ = a.Result(time.Second)

	if len(a.latencies) != len(snapshot) {
		t.Fatalf("latencies length: want %d, got %d", len(snapshot), len(a.latencies))
	}
	for i := range snapshot {
		if a.latencies[i] != snapshot[i] {
			t.Fatalf("latencies mutated at %d: want %v, got %v", i, snapshot, a.latencies)
		}
	}
}
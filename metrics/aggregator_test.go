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

func TestAggregatorLatenciesSliceLengthMatchesRecordCalls(t *testing.T) {
	a := NewAggregator()
	const n = 10
	for i := 0; i < n; i++ {
		a.Record(time.Duration(i+1)*time.Millisecond, i%2 == 0)
	}

	if len(a.latencies) != n {
		t.Fatalf("latencies: want len %d, got %d", n, len(a.latencies))
	}

	result := a.Result(time.Second)
	if result.Total != n {
		t.Fatalf("Result.Total: want %d, got %d", n, result.Total)
	}
	for i := 0; i < n; i++ {
		want := time.Duration(i+1) * time.Millisecond
		if a.latencies[i] != want {
			t.Fatalf("latencies[%d]: want %v, got %v", i, want, a.latencies[i])
		}
	}
}

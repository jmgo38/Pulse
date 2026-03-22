package metrics

import (
	"math"
	"sync"
	"time"
)

// Result contains the aggregated execution metrics for a run.
type Result struct {
	Total    int64
	Failed   int64
	Duration time.Duration
	Latency  LatencyStats
}

// Aggregator collects execution metrics for the MVP.
type Aggregator struct {
	mu         sync.Mutex
	total      int64
	failed     int64
	meanNanos  float64
	minLatency time.Duration
	maxLatency time.Duration
	latencies  []time.Duration // retained for future percentile computation
}

// NewAggregator creates an empty metrics aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{
		latencies: make([]time.Duration, 0),
	}
}

// Record stores metrics for a single execution.
func (a *Aggregator) Record(latency time.Duration, failed bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.total++
	if failed {
		a.failed++
	}

	a.meanNanos += (float64(latency.Nanoseconds()) - a.meanNanos) / float64(a.total)
	if a.total == 1 || latency < a.minLatency {
		a.minLatency = latency
	}
	if latency > a.maxLatency {
		a.maxLatency = latency
	}

	a.latencies = append(a.latencies, latency)
}

// Result returns the aggregated metrics snapshot.
func (a *Aggregator) Result(duration time.Duration) Result {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := Result{
		Total:    a.total,
		Failed:   a.failed,
		Duration: duration,
	}

	if a.total == 0 {
		return result
	}

	result.Latency = LatencyStats{
		Min:  a.minLatency,
		Max:  a.maxLatency,
		Mean: time.Duration(math.Round(a.meanNanos)),
	}

	return result
}

package metrics

import "time"

// LatencyStats contains aggregate latency data for a run.
type LatencyStats struct {
	Min  time.Duration
	Max  time.Duration
	Mean time.Duration
	P50  time.Duration
	P90  time.Duration
	P95  time.Duration
	P99  time.Duration
}

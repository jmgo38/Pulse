package pulse

import (
	"fmt"
	"time"

	fabricmetrics "algoryn.io/fabric/metrics"
)

// newRunID returns a simple unique identifier for a Pulse run.
// It avoids external dependencies by using the nanosecond clock.
func newRunID() string {
	return fmt.Sprintf("pulse-%d", time.Now().UnixNano())
}

// ToRunEvent converts a Pulse Result into a fabric.RunEvent,
// making it compatible with other Algoryn ecosystem tools.
// The startedAt parameter should be the time the run began.
// If zero, time.Now() is used as a best-effort approximation.
func ToRunEvent(result Result, passed bool, startedAt time.Time) fabricmetrics.RunEvent {
	if startedAt.IsZero() {
		startedAt = time.Now().Add(-result.Duration)
	}
	endedAt := startedAt.Add(result.Duration)

	// Convert ThresholdOutcomes to []fabricmetrics.ThresholdResult.
	thresholds := make([]fabricmetrics.ThresholdResult, len(result.ThresholdOutcomes))
	for i, t := range result.ThresholdOutcomes {
		thresholds[i] = fabricmetrics.ThresholdResult{
			Description: t.Description,
			Pass:        t.Pass,
		}
	}

	// fabric uses map[int]int64 for StatusCodes — same as Pulse. Copy directly.
	snapshot := fabricmetrics.MetricSnapshot{
		Source:      fabricmetrics.SourcePulse,
		Timestamp:   startedAt,
		Window:      result.Duration,
		Total:       result.Total,
		Failed:      result.Failed,
		RPS:         result.RPS,
		StatusCodes: result.StatusCounts,
		Errors:      result.ErrorCounts,
		Latency: fabricmetrics.LatencyStats{
			Min:  result.Latency.Min,
			Mean: result.Latency.Mean,
			P50:  result.Latency.P50,
			P95:  result.Latency.P95,
			P99:  result.Latency.P99,
			Max:  result.Latency.Max,
		},
	}

	return fabricmetrics.RunEvent{
		ID:         newRunID(),
		Source:     fabricmetrics.SourcePulse,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		Snapshot:   snapshot,
		Thresholds: thresholds,
		Passed:     passed,
	}
}

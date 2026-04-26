package stats

// #cgo CXXFLAGS: -std=c++17
// #cgo LDFLAGS: -lstdc++
//
// #include <stdlib.h>
//
// void* pulse_stats_engine_create(void);
// void pulse_stats_engine_destroy(void* p);
// void pulse_stats_engine_record(void* p, long long nanos);
// double pulse_stats_engine_get_percentile(void* p, double percent);
// void pulse_stats_engine_reset(void* p);
// unsigned long long pulse_stats_engine_total(void* p);
import "C"

import (
	"sync"
	"unsafe"
)

// Engine is a thread-safe view of the C++ logarithmic-histogram engine (1 µs to 60 s, ~3
// significant figures in log10 binning). All methods are safe for concurrent use.
type Engine struct {
	mu     sync.Mutex
	handle unsafe.Pointer
	closed bool
}

// NewEngine creates a new statistics engine. The caller should call Close when the engine
// is no longer needed to release C++ memory (finalizers are not reliable for C++ destructors).
func NewEngine() *Engine {
	return &Engine{handle: C.pulse_stats_engine_create()}
}

// RecordLatency records one latency sample in nanoseconds. Values are clamped to the
// configured histogram range for bucketing: below 1 µs to the first bucket, at or
// above 60 s to the last bucket. Negative values are treated like zero (first bucket).
func (e *Engine) RecordLatency(nanos int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.handle == nil {
		return
	}
	C.pulse_stats_engine_record(e.handle, C.longlong(nanos))
}

// GetPercentile returns the estimated latency in nanoseconds for the given percentile p in
// [0, 100] using a histogram CDF. Returns 0 if there are no samples.
func (e *Engine) GetPercentile(p float64) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.handle == nil {
		return 0
	}
	return float64(C.pulse_stats_engine_get_percentile(e.handle, C.double(p)))
}

// Reset clears all buckets and the total count.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.handle == nil {
		return
	}
	C.pulse_stats_engine_reset(e.handle)
}

// Total returns the number of samples recorded.
func (e *Engine) Total() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.handle == nil {
		return 0
	}
	return uint64(C.pulse_stats_engine_total(e.handle))
}

// Close releases native resources. Safe to call more than once.
func (e *Engine) Close() {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.handle == nil {
		return
	}
	C.pulse_stats_engine_destroy(e.handle)
	e.handle = nil
	e.closed = true
}

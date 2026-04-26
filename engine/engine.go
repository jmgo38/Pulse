package engine

import (
	"context"
	"sync"
	"time"

	"algoryn.io/pulse/internal"
	"algoryn.io/pulse/metrics"
	"algoryn.io/pulse/scheduler"
)

// Engine executes a test definition.
type Engine struct {
	phases         []scheduler.Phase
	scenario       func(context.Context) (int, error)
	maxConcurrency int
}

// New creates an engine for the given execution inputs.
func New(phases []scheduler.Phase, scenario func(context.Context) (int, error), maxConcurrency int) *Engine {
	return &Engine{
		phases:         phases,
		scenario:       scenario,
		maxConcurrency: maxConcurrency,
	}
}

// Run executes each phase in sequence through the scheduler.
// Scenario errors are recorded in metrics and do not stop the run.
// A non-nil error indicates scheduler failure or context cancellation.
func (e *Engine) Run(ctx context.Context) (metrics.Result, error) {
	aggregator := metrics.NewAggregator()
	defer aggregator.Close()
	startedAt := time.Now()
	limiter := internal.NewLimiter(e.maxConcurrency)

	var wg sync.WaitGroup

	wrappedScenario := func(ctx context.Context) error {
		if err := limiter.Acquire(ctx); err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer limiter.Release()

			executionStartedAt := time.Now()
			statusCode, err := e.scenario(ctx)
			aggregator.Record(time.Since(executionStartedAt), statusCode, err)
		}()

		return nil
	}

	for _, phase := range e.phases {
		if err := scheduler.Run(ctx, phase, wrappedScenario); err != nil {
			wg.Wait()
			return aggregator.Result(time.Since(startedAt)), err
		}
	}

	wg.Wait()
	return aggregator.Result(time.Since(startedAt)), nil
}

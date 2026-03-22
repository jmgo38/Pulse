package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmgo38/Pulse/internal"
	"github.com/jmgo38/Pulse/model"
)

var (
	ErrUnsupportedPhaseType   = errors.New("scheduler: unsupported phase type")
	ErrNonPositiveArrivalRate = errors.New("scheduler: phase arrival rate must be positive")
	ErrInvalidRampEndpoints   = errors.New("scheduler: ramp from and to must be positive")
)

// Phase contains the scheduling inputs for a single phase.
type Phase struct {
	Type        model.PhaseType
	Duration    time.Duration
	ArrivalRate int
	From        int
	To          int
}

// Run executes the supported scheduling strategy for a phase.
func Run(ctx context.Context, phase Phase, scenario func(context.Context) error) error {
	switch phase.Type {
	case model.PhaseTypeConstant:
		if phase.ArrivalRate <= 0 {
			return ErrNonPositiveArrivalRate
		}
		return runConstant(ctx, phase, scenario)
	case model.PhaseTypeRamp:
		if phase.From <= 0 || phase.To <= 0 {
			return ErrInvalidRampEndpoints
		}
		return runRamp(ctx, phase, scenario)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedPhaseType, phase.Type)
	}
}

func runConstant(ctx context.Context, phase Phase, scenario func(context.Context) error) error {
	capacity := phase.ArrivalRate
	if capacity < 1 {
		capacity = 1
	}
	// Drained start avoids bursting the whole capacity before the engine limiter
	// can apply backpressure (wrappedScenario returns before work finishes).
	bucket := internal.NewDrainedTokenBucket(capacity, float64(phase.ArrivalRate))
	deadline := time.Now().Add(phase.Duration)
	poll := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		now := time.Now()
		if !now.Before(deadline) {
			return nil
		}

		if bucket.Allow(now) {
			if err := scenario(ctx); err != nil {
				return err
			}
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(poll):
			}
		}
	}
}

func runRamp(ctx context.Context, phase Phase, scenario func(context.Context) error) error {
	start := time.Now()
	deadline := start.Add(phase.Duration)

	capacity := phase.From
	if phase.To > capacity {
		capacity = phase.To
	}
	if capacity < 1 {
		capacity = 1
	}

	initialRate := float64(phase.From)
	if initialRate < 1 {
		initialRate = 1
	}
	bucket := internal.NewDrainedTokenBucket(capacity, initialRate)
	poll := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		now := time.Now()
		if !now.Before(deadline) {
			return nil
		}

		elapsed := now.Sub(start)
		frac := float64(elapsed) / float64(phase.Duration)
		if frac > 1 {
			frac = 1
		}
		rate := float64(phase.From) + float64(phase.To-phase.From)*frac
		if rate < 1 {
			rate = 1
		}

		bucket.SetRefillRate(rate, now)

		if bucket.Allow(now) {
			if err := scenario(ctx); err != nil {
				return err
			}
		} else {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(poll):
			}
		}
	}
}
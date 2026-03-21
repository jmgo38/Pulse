package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	interval := time.Second / time.Duration(phase.ArrivalRate)
	ticker := time.NewTicker(interval)
	timer := time.NewTimer(phase.Duration)
	defer ticker.Stop()
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		case <-ticker.C:
			if err := scenario(ctx); err != nil {
				return err
			}
		}
	}
}

func runRamp(ctx context.Context, phase Phase, scenario func(context.Context) error) error {
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		now := time.Now()
		elapsed := now.Sub(start)
		if elapsed >= phase.Duration {
			return nil
		}

		frac := float64(elapsed) / float64(phase.Duration)
		rate := float64(phase.From) + float64(phase.To-phase.From)*frac
		if rate < 1 {
			rate = 1
		}

		interval := time.Duration(float64(time.Second) / rate)
		if interval <= 0 {
			interval = time.Nanosecond
		}

		remaining := phase.Duration - elapsed
		if interval > remaining {
			interval = remaining
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		if err := scenario(ctx); err != nil {
			return err
		}
	}
}
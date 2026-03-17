package engine

import (
	"context"
)

// Engine executes a test definition.
type Engine struct {
	phaseCount int
	scenario   func(context.Context) error
}

// New creates an engine for the given execution inputs.
func New(phaseCount int, scenario func(context.Context) error) *Engine {
	return &Engine{
		phaseCount: phaseCount,
		scenario:   scenario,
	}
}

// Run executes the scenario once for each phase in sequence.
func (e *Engine) Run(ctx context.Context) error {
	for i := 0; i < e.phaseCount; i++ {
		if err := e.scenario(ctx); err != nil {
			return err
		}
	}

	return nil
}

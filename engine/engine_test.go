package engine

import (
	"context"
	"errors"
	"testing"
)

func TestEngineRunExecutesScenarioForEachPhase(t *testing.T) {
	calls := 0
	engine := New(2, func(context.Context) error {
		calls++
		return nil
	})

	if err := engine.Run(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 scenario calls, got %d", calls)
	}
}

func TestEngineRunUsesProvidedContext(t *testing.T) {
	type ctxKey string

	const key ctxKey = "phase"
	want := "value"

	engine := New(1, func(ctx context.Context) error {
		if got := ctx.Value(key); got != want {
			t.Fatalf("expected context value %q, got %v", want, got)
		}

		return nil
	})

	ctx := context.WithValue(context.Background(), key, want)
	if err := engine.Run(ctx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestEngineRunPropagatesScenarioError(t *testing.T) {
	wantErr := errors.New("scenario failed")
	engine := New(1, func(context.Context) error {
		return wantErr
	})

	if err := engine.Run(context.Background()); err != wantErr {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

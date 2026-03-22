package pulse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmgo38/Pulse/transport"
)

func TestIntegrationRunEndToEndWithHTTPThresholds(t *testing.T) {
	const handlerLatency = 8 * time.Millisecond

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(handlerLatency)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := transport.NewHTTPClient()

	test := Test{
		Config: Config{
			Phases: []Phase{
				{
					Type:        PhaseTypeConstant,
					Duration:    200 * time.Millisecond,
					ArrivalRate: 20,
				},
			},
			MaxConcurrency: 4,
			Thresholds: Thresholds{
				MaxMeanLatency: 50 * time.Millisecond,
				MaxP95Latency:  50 * time.Millisecond,
				MaxP99Latency:  50 * time.Millisecond,
			},
		},
		Scenario: func(ctx context.Context) (int, error) {
			return client.Get(ctx, srv.URL)
		},
	}

	result, err := Run(test)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Total <= 0 {
		t.Fatalf("expected Total > 0, got %d", result.Total)
	}
	if result.Failed != 0 {
		t.Fatalf("expected Failed == 0, got %d", result.Failed)
	}
	if n := result.StatusCounts[200]; n <= 0 {
		t.Fatalf("expected StatusCounts[200] > 0, got %d", n)
	}
	if result.RPS <= 0 {
		t.Fatalf("expected RPS > 0, got %f", result.RPS)
	}

	wantOutcomes := 3
	if len(result.ThresholdOutcomes) != wantOutcomes {
		t.Fatalf("expected %d threshold outcomes, got %+v", wantOutcomes, result.ThresholdOutcomes)
	}
	for i, o := range result.ThresholdOutcomes {
		if !o.Pass {
			t.Fatalf("outcome %d: want Pass true, got %+v", i, o)
		}
	}
}

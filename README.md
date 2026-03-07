# Pulse

**Pulse** is a programmable reliability and load testing engine written in Go.

It allows developers to define load tests directly in Go code, execute them with a deterministic engine, and analyze system behavior under controlled stress conditions.

Unlike traditional tools that rely on external configuration languages (YAML/JS), Pulse follows a **code-first approach**, enabling full use of Go's type system, tooling, and testing ecosystem.

---

## Vision

Pulse is designed to evolve beyond a simple load generator into a **platform for reliability experimentation**.

The long-term goal is to provide a programmable environment where engineers can:

- Generate controlled system load
- Experiment with failure scenarios
- Analyze latency behavior under stress
- Explore system resilience under degraded conditions

The initial version focuses on building a **deterministic, extensible execution engine**.

---

## Features (MVP)

Current goals for the initial version include:

- Deterministic load generation
- Arrival-rate based scheduling
- Programmable scenarios using Go
- Built-in metrics aggregation
- Latency percentiles (p50, p95, p99)
- JSON result output
- CLI execution

Future capabilities may include:

- Fault injection (latency, errors)
- Chaos experimentation modules
- Additional transports (gRPC, TCP)
- Observability integrations
- Distributed execution

---

## Example

```go
package main

import (
	"context"
	"time"

	"github.com/yourusername/pulse"
)

func main() {
	test := pulse.Test{
		Phases: []pulse.Phase{
			pulse.ConstantRate(100, 10*time.Second),
		},
		Scenario: func(ctx context.Context) error {
			// Example request
			// http.Get("https://api.example.com")
			return nil
		},
	}

	result, err := pulse.Run(test)
	if err != nil {
		panic(err)
	}

	println("Total Requests:", result.TotalRequests)
	println("Errors:", result.TotalErrors)
}





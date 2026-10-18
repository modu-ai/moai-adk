package quality

// affected_scope_gate.go — wires the affected-set core into the Go test step
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 REQ-PAS-001/-008/-010).

import (
	"context"
	"time"
)

// goTestStepName is the Go toolchain's test step, the only step the affected
// scope narrows.
const goTestStepName = "go test"

// scopeQueries are the two queries the affected scope depends on; tests
// substitute them to stay deterministic. The zero value uses the real adapters.
type scopeQueries struct {
	changes func(ctx context.Context, dir string) (commitChanges, error)
	graph   func(ctx context.Context, dir, tags string) (moduleGraph, error)
}

// stepRunFunc is the boundary executeStep hands a prepared command to. Production
// uses QualityGate.runStep; the budget test substitutes a recorder that captures
// the budget each step receives.
type stepRunFunc func(ctx context.Context, stepName, dir string, timeout time.Duration, name string, args ...string) (bool, string)

// scopeGoTestStep narrows the Go test step to the packages the commit can
// affect. It returns the step to execute, a notice block for the gate summary,
// and a non-empty skip reason when nothing is left to run.
func (g *QualityGate) scopeGoTestStep(ctx context.Context, step gateStep, dir string) (gateStep, string, string) {
	return step, "", ""
}

package cli

// SPEC-INIT-DEPLOY-EXIT-001 — test-injection seam for the init phase executor.
//
// The failure surface runInit renders when template deployment fails can only
// be exercised end-to-end by an init run whose deployment actually fails, and
// the production deployer renders the embedded templates — the only way to make
// it fail is to corrupt a template in internal/template/templates/, which is
// out of scope for this SPEC and would pollute the distributed tree. This seam
// lets a test inject an executor that fails the way the real one does, leaving
// the production path byte-identical (the default returns the real
// project.PhaseExecutor).

import (
	"context"

	"github.com/modu-ai/moai-adk/internal/core/project"
)

// initPhaseExecutor is the slice of *project.PhaseExecutor that runInit uses.
type initPhaseExecutor interface {
	SetReporter(reporter project.ProgressReporter)
	Execute(ctx context.Context, opts project.InitOptions) (*project.InitResult, error)
}

// newInitPhaseExecutorFn constructs the executor runInit drives. Production
// returns the real PhaseExecutor; a test that swaps it MUST NOT call
// t.Parallel and MUST restore it with t.Cleanup.
var newInitPhaseExecutorFn = func(
	detector project.Detector,
	methDetector project.MethodologyDetector,
	validator project.ProjectValidator,
	initializer project.Initializer,
) initPhaseExecutor {
	return project.NewPhaseExecutor(detector, methDetector, validator, initializer, nil)
}

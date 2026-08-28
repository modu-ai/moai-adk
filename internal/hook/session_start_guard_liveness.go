package hook

import (
	"context"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/guardliveness"
)

// guardLivenessProducer is the seam to the state model
// (SPEC-GUARD-STATE-MODEL-001), which has not landed. Until it does, the wired
// producer reports that it is unwired — reached on every activation, so the
// count at the query layer stays the count of activations rather than dropping
// silently to zero. When the state model lands, this one line names its
// producer instead.
var guardLivenessProducer guardliveness.Producer = guardliveness.Unwired()

// guardLivenessRefresh initiates the guard-liveness refresh for one session
// start (SPEC-GUARD-LIVENESS-001 REQ-GDL-002/003/012, card t333 M1).
//
// Three properties, each of which the SPEC pays for elsewhere if lost:
//
//   - PULL-BASED. The evaluator runs at a moment that already happens for other
//     reasons rather than on a cadence of its own, so it is not itself subject
//     to the defect it watches for and no scheduled workflow is added
//     (REQ-GDL-002).
//   - UNCONDITIONAL. There is no path filter, no changed-file test, and no
//     subject-matter condition here or one frame inward — a condition that
//     stops matching is exactly how the guards of spec.md §A.3 went quiet
//     without being removed. Note what is NOT here: no early return on an empty
//     root, because "nothing to query" is the producer's judgement to make and
//     a caller-side skip is a condition wearing a guard clause's clothes.
//   - NEVER AWAITED. The refresh is initiated and abandoned for this session;
//     session start is latency-bounded and the refresh issues one query per
//     subject, so awaiting it would put the network on the input-lag budget.
//     The render reads a persisted result at a LATER activation, which is M2's
//     half (REQ-GDL-011/012).
func guardLivenessRefresh(ctx context.Context, root string) {
	act := guardliveness.Activation{Root: root}
	evaluator := guardliveness.New(guardLivenessProducer)

	if !deferredScansAsyncEnabled() {
		// Test path (TestMain sets deferredScansAsync=false): run inline so no
		// goroutine outlives the test boundary, matching the binary-lag
		// contributor on this surface and for the same reason.
		if _, err := evaluator.Refresh(ctx, act); err != nil {
			slog.Debug("guard liveness: refresh failed (non-blocking)", "error", err.Error())
		}
		return
	}

	evaluator.OnActivation(ctx, act)
}

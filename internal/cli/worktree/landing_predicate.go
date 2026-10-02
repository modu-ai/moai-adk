package worktree

// landing_predicate.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-A (card t1453),
// REQ-GFD-002 / design D-5. RED-phase scaffold: only the seams the RED tests
// drive are declared here; the layers land in the GREEN commit.

import (
	"context"
	"time"
)

// landingPatchIDCommitCap bounds how many integration-ref commits since the
// merge-base the cumulative patch-id layer compares (design D-5: 500).
var landingPatchIDCommitCap = 500

// landingGHTimeout bounds one `gh` call of the PR-state layer (design D-5: 10 s).
var landingGHTimeout = 10 * time.Second

// landingGH is the `gh` execution seam of the PR-state layer.
var landingGH = func(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return nil, nil
}

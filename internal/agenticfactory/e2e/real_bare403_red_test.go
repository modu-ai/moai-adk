//go:build afred

package e2e

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mirror"
)

// RED on purpose (cross-lane contract, found on the integration tree after
// merging lane-18 f4a426548f with lane-20 5ac3fe5590). forge.Classify now
// returns KindUnclassified for a 403 that has no rate-limit signal and no
// permission wording; forge's own client and effect mapping treat that kind as
// a stop (forge.go: `case KindDenied, KindUnclassified`). mirror.fromStatus maps
// only KindLimited and KindDenied, so the same bare 403 leaves mirror as a PLAIN
// error with no block reason recorded. Lane-18's own tests
// TestFailuresBecomeBlocksWithoutRetry/403 and
// TestLimitVersusDenialFollowsStatusAndHeaders/403_with_no_signal_is_a_denial
// fail on the merged tree for this reason (they pass on the lane-18 branch alone).
// Expected: the bare 403 is a block (permission-denied), recorded and not retried.
// Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_
func TestRED_Mirror_Bare403_IsABlockNotAPlainError(t *testing.T) {
	x := failingHTTP(403, nil, "")
	err := writeOnce(x)
	var be *mirror.BlockedError
	if !errors.As(err, &be) {
		t.Fatalf("a bare 403 left mirror as a plain error: %v", err)
	}
	if be.Reason != core.ReasonPermissionDenied {
		t.Errorf("bare 403 -> %s, want permission-denied", be.Reason)
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls = %d, want 1", len(x.calls))
	}
}

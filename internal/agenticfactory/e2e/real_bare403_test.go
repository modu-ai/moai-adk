package e2e

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/mirror"
)

// Cross-lane contract found on the integration tree after merging lane-18
// f4a426548f with lane-20 5ac3fe5590: forge.Classify answers a 403 with no
// rate-limit signal and no permission wording as KindUnclassified, and mirror
// did not map that kind, so the same 403 left mirror as a plain error with no
// block recorded. This was a RED test until lane-18 80ea4d7e68; promoted once it
// passed, with the reason the owners settled on (unclassified-403, not
// permission-denied: the code does not know it is a permission problem).
func TestMirror_Bare403_IsAnUnclassifiedBlock_NotRetried(t *testing.T) {
	x := failingHTTP(403, nil, "")
	err := writeOnce(x)
	var be *mirror.BlockedError
	if !errors.As(err, &be) {
		t.Fatalf("a bare 403 left mirror as a plain error: %v", err)
	}
	if be.Reason != core.ReasonUnclassified {
		t.Errorf("bare 403 -> %s, want %s", be.Reason, core.ReasonUnclassified)
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls = %d, want 1 (no retry)", len(x.calls))
	}
}

// The wording-based permission reading stays a reason only: a 403 whose body
// says the resource is not accessible is permission-denied, and it is also not
// retried. Both outcomes stop; they differ in the recorded reason.
func TestMirror_NotAccessible403_IsPermissionDenied_NotRetried(t *testing.T) {
	x := failingHTTP(403, nil, `{"message":"Resource not accessible by integration"}`)
	if got := blockedReason(t, writeOnce(x)); got != core.ReasonPermissionDenied {
		t.Errorf("not-accessible 403 -> %s, want permission-denied", got)
	}
	if len(x.calls) != 1 {
		t.Errorf("gh calls = %d, want 1", len(x.calls))
	}
}

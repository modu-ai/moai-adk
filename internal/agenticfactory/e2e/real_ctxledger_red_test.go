//go:build afred

package e2e

import (
	"testing"
)

// RED on purpose (Issue #1948: "a permission-denied 403 is retried 0 times").
// In ctxledger, refusals of Find and Create go through classify and park the
// intent, but a refusal of Update (the edit of an existing comment) is returned
// as it is, so the intent stays "failed" and every Flush tries again. Measured
// with the typed forge.StatusError: an edit refused with 403 (or 404) while the
// comment is still findable costs find+update on every Flush (3 flushes: find 7,
// update 4), while a refused lookup parks after the first Flush (calls stay at 4).
// Expected: after the first refusal later Flushes spend no further remote calls.
// Run with: go test -tags afred ./internal/agenticfactory/e2e/ -run RED_
func TestRED_CtxLedger_EditRefused403_ParksAndStopsSpendingCalls(t *testing.T) {
	inject := func(r *ctxRemote) {
		r.updateErr = httpStatus(403, "Resource not accessible by integration")
	}
	one, _, _ := refusedUpdateScenario(t, inject, 1)
	three, _, _ := refusedUpdateScenario(t, inject, 3)
	if three.calls() != one.calls() {
		t.Errorf("calls grew from %d (1 flush) to %d (3 flushes): a permission 403 on edit is being retried", one.calls(), three.calls())
	}
}

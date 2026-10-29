package e2e

import (
	"testing"
)

// Issue #1948: a permission-denied 403 is retried 0 times. This was a RED test:
// a refused Update of an existing comment was returned unclassified, so the
// intent stayed "failed" and every Flush tried again (measured with the typed
// forge.StatusError: 1 flush = 4 calls, 3 flushes = 10 calls). Promoted after
// lane-20 8558f07fa8 made it pass on the integration tree.
func TestCtxLedger_EditRefused403_ParksAndStopsSpendingCalls(t *testing.T) {
	inject := func(r *ctxRemote) {
		r.updateErr = httpStatus(403, "Resource not accessible by integration")
	}
	one, _, _ := refusedUpdateScenario(t, inject, 1)
	three, _, _ := refusedUpdateScenario(t, inject, 3)
	t.Logf("MEASURED edit refused 403: calls after 1 flush=%d, after 3 flushes=%d", one.calls(), three.calls())
	if three.calls() != one.calls() {
		t.Errorf("calls grew from %d (1 flush) to %d (3 flushes): a permission 403 on edit is being retried", one.calls(), three.calls())
	}
}

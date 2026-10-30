//go:build afred

package core

import "testing"

// P2-3: the composition that lets BUILDING enter MERGING on the queued fact is for a
// card past its spec step, BUILDING/IMPLEMENT, and for nothing else.
func TestNoHandoffEnqueuedIsBuildingImplementOnly(t *testing.T) {
	if err := CheckMove(bImpl, merge, []string{FactQueued}); err != nil {
		t.Errorf("BUILDING/IMPLEMENT: %v", err)
	}
	for name, from := range map[string]Position{"SPEC": bSpec, "FIX": bFix} {
		if err := CheckMove(from, merge, []string{FactQueued, FactReceipt, FactReceiptAccepted}); err == nil {
			t.Errorf("BUILDING/%s must not enter MERGING on the queued fact", name)
		}
	}
}

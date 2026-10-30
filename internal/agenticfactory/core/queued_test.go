package core

import (
	"errors"
	"testing"
)

// The local ledger enters MERGING by queueing, so the edges and the composition that
// reach it stand on the queued fact.
func TestEnteringMergingNeedsTheQueuedFact(t *testing.T) {
	cases := []struct {
		name     string
		from, to Position
	}{
		{"E08 from REVIEWING", review, merge},
		{"a card with no review step (E03, E08)", bImpl, merge},
		{"the same from spec", bSpec, merge},
	}
	for _, c := range cases {
		if err := CheckMove(c.from, c.to, []string{FactQueued}); err != nil {
			t.Errorf("%s with the fact: %v", c.name, err)
		}
		if err := CheckMove(c.from, c.to, nil); !errors.Is(err, ErrGuardMissing) {
			t.Errorf("%s without it: %v", c.name, err)
		}
	}
	if err := CheckMove(bFix, merge, []string{FactQueued, FactReceipt, FactReceiptAccepted}); err == nil {
		t.Error("BUILDING/FIX holds a receipt the leader has not accepted: it does not enter MERGING")
	}
	if err := CheckMove(with(bImpl, ReasonDependency), merge, []string{FactQueued}); err == nil {
		t.Error("a parked position does not enter MERGING")
	}
}

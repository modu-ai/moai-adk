package claim

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestLockWaitBudgetIsTheTriesTimesTheWait(t *testing.T) {
	if b := LockWaitBudget(); b <= 0 || b != lockWait*lockTries {
		t.Errorf("budget %v, want %v tries of %v", b, lockTries, lockWait)
	}
}

// The fixed code of a position error is what a reader switches on: the typed error
// carries its own, an unmappable legacy record gets the unmappable one, anything
// else the catch-all; the cause stays reachable for errors.Is.
func TestPositionCodeIsFixedAndTheCauseStaysReachable(t *testing.T) {
	cause := fmt.Errorf("%w: detail", core.ErrEvidenceIncomplete)
	pe := &positionError{PositionInvalidWait, cause}
	if got := positionCode(fmt.Errorf("wrapped: %w", pe)); got != PositionInvalidWait {
		t.Errorf("typed: %q", got)
	}
	if !errors.Is(pe, core.ErrEvidenceIncomplete) || pe.Error() != cause.Error() {
		t.Errorf("the cause must stay reachable: %v", pe)
	}
	if got := positionCode(fmt.Errorf("%w: x", core.ErrLegacyUnmappable)); got != PositionUnmappable {
		t.Errorf("legacy unmappable: %q", got)
	}
	if got := positionCode(errors.New("other")); got != PositionUnknownOther {
		t.Errorf("other: %q", got)
	}
}

func TestAddEventAppendsAndKeepsTheOlderEntries(t *testing.T) {
	var r Record
	r.AddEvent(Event{Kind: "claim"})
	r.AddEvent(Event{Kind: "handoff", PR: 7})
	if len(r.History) != 2 || r.History[0].Kind != "claim" || r.History[1].PR != 7 {
		t.Errorf("history: %+v", r.History)
	}
}

func TestMarkRepairNeededNeedsItsReasonAndHasOpenHandoffStartsFalse(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	if _, err := s.Acquire(i, "sess-a", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.MarkRepairNeeded(i, "   "); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("a repair signal without a reason: %v", err)
	}
	if open, err := s.HasOpenHandoff("sess-a", i); err != nil || open {
		t.Errorf("a holder with no receipt has no open handoff: %v %v", open, err)
	}
}

func TestPresenceFileIsTheIssuesOwnPath(t *testing.T) {
	s := open(t)
	i := ref(t, 1948)
	if p := s.PresenceFile(i); !strings.HasSuffix(p, i.Resource()+".presence") || p != s.presencePath(i) {
		t.Errorf("presence file: %q", p)
	}
}

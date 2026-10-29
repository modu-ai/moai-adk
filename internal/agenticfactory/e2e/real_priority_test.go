package e2e

import (
	"strconv"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Safety-state writes outrank report comments when the shared call budget runs
// short (Issue #1948). Counts and times are test inputs.

func budgetAtReserve(t *testing.T, reserveCalls, remaining int, resetIn time.Duration) *forge.Budget {
	t.Helper()
	b := &forge.Budget{ReserveCalls: reserveCalls, Now: func() time.Time { return t0 }}
	b.Observe(forge.Response{Status: 200, Header: map[string]string{
		"x-ratelimit-remaining": strconv.Itoa(remaining),
		"x-ratelimit-reset":     strconv.FormatInt(t0.Add(resetIn).Unix(), 10),
	}}, t0)
	return b
}

// At the reserve a report waits and a safety write goes out; nothing is dropped.
func TestPriority_AtTheReserve_SafetyGoesOut_ReportWaits(t *testing.T) {
	b := budgetAtReserve(t, 2, 2, 100*time.Second)
	s := &forge.Scheduler{Gate: b}
	s.Push(forge.Item{Priority: core.PriorityReport, Key: "issue-1948", Payload: "progress"})
	s.Push(forge.Item{Priority: core.PrioritySafety, Key: "claim-1948", Payload: "claim"})

	it, ok, _ := s.Pop(t0.Add(time.Second))
	if !ok || it.Priority != core.PrioritySafety {
		t.Fatalf("first pop = %+v ok=%v, want the safety item", it, ok)
	}
	if it2, ok2, until := s.Pop(t0.Add(2 * time.Second)); ok2 {
		t.Fatalf("a report went out at the reserve: %+v", it2)
	} else if !until.Equal(t0.Add(100 * time.Second)) {
		t.Errorf("retry hint = %v, want the reset time %v", until, t0.Add(100*time.Second))
	}
	if s.Held() != 1 {
		t.Errorf("held = %d, want the report kept (nothing dropped)", s.Held())
	}
	released, ok3, _ := s.Pop(t0.Add(101 * time.Second))
	if !ok3 || released.Priority != core.PriorityReport {
		t.Errorf("after the reset: %+v ok=%v, want the report released", released, ok3)
	}
}

// With budget to spare the order is still safety first, whatever was queued first.
func TestPriority_SafetyFirst_EvenWhenQueuedLater(t *testing.T) {
	s := &forge.Scheduler{Gate: &forge.Budget{ReserveCalls: 1}}
	s.Push(forge.Item{Priority: core.PriorityReport, Key: "r", Payload: 1})
	s.Push(forge.Item{Priority: core.PrioritySafety, Key: "s", Payload: 2})
	it, ok, _ := s.Pop(t0)
	if !ok || it.Priority != core.PrioritySafety {
		t.Errorf("pop = %+v ok=%v, want the safety item first", it, ok)
	}
}

// Reports with the same key are merged, so a backlog of progress updates does
// not turn into one call each once the budget recovers.
func TestPriority_ReportsWithTheSameKey_AreMerged(t *testing.T) {
	s := &forge.Scheduler{Gate: &forge.Budget{}, Merge: func(old, add any) any { return add }}
	for i := range 5 {
		s.Push(forge.Item{Priority: core.PriorityReport, Key: "issue-1948", Payload: i})
	}
	if s.Held() != 1 {
		t.Fatalf("held = %d, want 1 merged report", s.Held())
	}
	it, ok, _ := s.Pop(t0)
	if !ok || it.Payload != 4 {
		t.Errorf("merged payload = %v ok=%v, want the newest (4)", it.Payload, ok)
	}
	t.Logf("MEASURED report merging: 5 progress updates for one key -> 1 write")
}

// Safety items with the same key are NOT merged: each is a distinct state write.
func TestPriority_SafetyItemsAreNeverMerged(t *testing.T) {
	s := &forge.Scheduler{Gate: &forge.Budget{}}
	for i := range 3 {
		s.Push(forge.Item{Priority: core.PrioritySafety, Key: "claim-1948", Payload: i})
	}
	if s.Held() != 3 {
		t.Errorf("held = %d, want 3 distinct safety writes", s.Held())
	}
}

// The budget hook that lane-18's writers call: a report is told to wait when
// the reserve is reached, a safety write never is.
func TestPriority_BudgetHook_ReportWaitsSafetyNever(t *testing.T) {
	b := budgetAtReserve(t, 2, 1, 100*time.Second)
	if b.Reserve(core.PriorityReport) {
		t.Error("a report was allowed below the reserve")
	}
	if !b.Reserve(core.PrioritySafety) {
		t.Error("a safety write was told to wait")
	}
	healthy := &forge.Budget{ReserveCalls: 2, Now: func() time.Time { return t0 }}
	if !healthy.Reserve(core.PriorityReport) {
		t.Error("a report was held with no budget information at all")
	}
}

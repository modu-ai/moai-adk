package forge

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func lowBudget(t *testing.T) *Budget {
	b := &Budget{ReserveCalls: 10}
	b.Observe(resp(200, "", "x-ratelimit-remaining", "10", "x-ratelimit-reset", reset(time.Hour)), t0)
	return b
}

func merge(a, b any) any { return a.(string) + "+" + b.(string) }

func TestSafetyWritesGoBeforeReports(t *testing.T) {
	s := &Scheduler{Gate: &Budget{}, Merge: merge}
	s.Push(Item{Priority: core.PriorityReport, Key: "p1", Payload: "progress"})
	s.Push(Item{Priority: core.PrioritySafety, Key: "w1", Payload: "claim"})
	var order []core.Priority
	for {
		it, ok, _ := s.Pop(t0)
		if !ok {
			break
		}
		order = append(order, it.Priority)
	}
	if len(order) != 2 || order[0] != core.PrioritySafety || order[1] != core.PriorityReport {
		t.Errorf("order = %v", order)
	}
}

func TestReportsNeverSpendTheReservedBudget(t *testing.T) {
	s := &Scheduler{Gate: lowBudget(t), Merge: merge}
	s.Push(Item{Priority: core.PriorityReport, Key: "p", Payload: "report"})
	if _, ok, _ := s.Pop(t0); ok {
		t.Fatal("a report went out on the reserved budget")
	}
	s.Push(Item{Priority: core.PrioritySafety, Key: "w", Payload: "parked"})
	first, ok, _ := s.Pop(t0)
	if !ok || first.Priority != core.PrioritySafety {
		t.Errorf("a safety write must still go under a low budget: %+v ok=%v", first, ok)
	}
	if _, ok, _ := s.Pop(t0); ok {
		t.Error("the report must stay held")
	}
	if s.Held() != 1 {
		t.Errorf("held = %d, want the report kept, not dropped", s.Held())
	}
}

func TestHeldReportsWithTheSameKeyAreMergedIntoOne(t *testing.T) {
	s := &Scheduler{Gate: lowBudget(t), Merge: merge}
	for _, p := range []string{"a", "b", "c"} {
		s.Push(Item{Priority: core.PriorityReport, Key: "issue-7", Payload: p})
	}
	if s.Held() != 1 {
		t.Fatalf("held = %d, want one merged report", s.Held())
	}
	s.Gate = &Budget{} // the budget recovers
	it, ok, _ := s.Pop(t0)
	if !ok || it.Payload != "a+b+c" {
		t.Errorf("merged = %+v", it)
	}
}

func TestSafetyWriteIsNotDelayedBehindQueuedReports(t *testing.T) {
	s := &Scheduler{Gate: &Budget{}, Merge: merge}
	for _, k := range []string{"a", "b", "c"} {
		s.Push(Item{Priority: core.PriorityReport, Key: k, Payload: k})
	}
	s.Push(Item{Priority: core.PrioritySafety, Key: "claim", Payload: "x"})
	if it, ok, _ := s.Pop(t0); !ok || it.Priority != core.PrioritySafety {
		t.Errorf("first out = %+v", it)
	}
}

func TestBlockedBudgetReturnsWhenToLookAgainAndDropsNothing(t *testing.T) {
	b := &Budget{}
	b.Block(t0.Add(45 * time.Second))
	s := &Scheduler{Gate: b, Merge: merge}
	s.Push(Item{Priority: core.PrioritySafety, Key: "w", Payload: "x"})
	if _, ok, until := s.Pop(t0); ok || until.Sub(t0) != 45*time.Second {
		t.Errorf("ok=%v until=%v", ok, until.Sub(t0))
	}
	if s.Held() != 1 {
		t.Error("a safety write was dropped while the budget was blocked")
	}
}

func TestSamePriorityKeepsFirstInFirstOut(t *testing.T) {
	s := &Scheduler{Gate: &Budget{}, Merge: merge}
	s.Push(Item{Priority: core.PrioritySafety, Key: "a", Payload: "1"})
	s.Push(Item{Priority: core.PrioritySafety, Key: "b", Payload: "2"})
	x, _, _ := s.Pop(t0)
	y, _, _ := s.Pop(t0)
	if x.Key != "a" || y.Key != "b" {
		t.Errorf("order %s,%s", x.Key, y.Key)
	}
}

// The priority definition lives in core only. The scheduler must not carry a priority type of its own.
func TestSchedulerUsesCorePriorityNotItsOwnDefinition(t *testing.T) {
	var _ core.Priority = Item{}.Priority
	if core.PrioritySafety >= core.PriorityReport {
		t.Fatal("core ordering assumption broke: a lower value must be more urgent")
	}
}

// core.Budget is the hook lane 18's writers call; forge.Budget answers it.
func TestBudgetAnswersTheCoreHook(t *testing.T) {
	var hook core.Budget = lowBudget(t)
	if hook.Reserve(core.PriorityReport) {
		t.Error("a report must be told to wait when ordinary calls are stopped")
	}
	if !hook.Reserve(core.PrioritySafety) {
		t.Error("a safety write is never held by the budget")
	}
	var open core.Budget = &Budget{}
	if !open.Reserve(core.PriorityReport) {
		t.Error("an unknown budget must let a report through")
	}
	b := &Budget{ReserveCalls: 10, ReportResource: "core", Now: func() time.Time { return t0 }}
	b.Observe(resp(200, "", "x-ratelimit-resource", "core", "x-ratelimit-remaining", "3", "x-ratelimit-reset", reset(time.Hour)), t0)
	b.Observe(resp(200, "", "x-ratelimit-resource", "search", "x-ratelimit-remaining", "0", "x-ratelimit-reset", reset(time.Hour)), t0)
	if b.Reserve(core.PriorityReport) {
		t.Error("core is below the reserve: the report must wait")
	}
	other := &Budget{ReserveCalls: 10, ReportResource: "core", Now: func() time.Time { return t0 }}
	other.Observe(resp(200, "", "x-ratelimit-resource", "search", "x-ratelimit-remaining", "0", "x-ratelimit-reset", reset(time.Hour)), t0)
	if !other.Reserve(core.PriorityReport) {
		t.Error("an exhausted search quota must not hold a report that counts against core")
	}
}

package forge

import (
	"testing"
	"time"
)

func lowBudget(t *testing.T) *Budget {
	b := &Budget{Reserve: 10}
	b.Observe(resp(200, "", "x-ratelimit-remaining", "10", "x-ratelimit-reset", reset(time.Hour)), t0)
	return b
}

func merge(a, b any) any { return a.(string) + "+" + b.(string) }

func TestStateWritesGoBeforeSafetyNoticesBeforeProgress(t *testing.T) {
	s := &Scheduler{Gate: &Budget{}, Merge: merge}
	s.Push(Item{Class: ClassProgress, Key: "p1", Payload: "progress"})
	s.Push(Item{Class: ClassSafetyNotice, Key: "s1", Payload: "security failure"})
	s.Push(Item{Class: ClassStateWrite, Key: "w1", Payload: "claim"})
	var order []Class
	for {
		it, ok, _ := s.Pop(t0)
		if !ok {
			break
		}
		order = append(order, it.Class)
	}
	want := []Class{ClassStateWrite, ClassSafetyNotice, ClassProgress}
	if len(order) != 3 || order[0] != want[0] || order[1] != want[1] || order[2] != want[2] {
		t.Errorf("order = %v, want %v", order, want)
	}
}

func TestProgressNeverSpendsTheReservedBudget(t *testing.T) {
	s := &Scheduler{Gate: lowBudget(t), Merge: merge}
	s.Push(Item{Class: ClassProgress, Key: "p", Payload: "report"})
	if _, ok, _ := s.Pop(t0); ok {
		t.Fatal("a progress comment went out on the reserved budget")
	}
	s.Push(Item{Class: ClassSafetyNotice, Key: "s", Payload: "block"})
	s.Push(Item{Class: ClassStateWrite, Key: "w", Payload: "parked"})
	first, ok1, _ := s.Pop(t0)
	second, ok2, _ := s.Pop(t0)
	if !ok1 || !ok2 || first.Class != ClassStateWrite || second.Class != ClassSafetyNotice {
		t.Errorf("under a low budget state and safety must still go: %+v %+v", first, second)
	}
	if _, ok, _ := s.Pop(t0); ok {
		t.Error("progress still held back expected")
	}
	if s.Held() != 1 {
		t.Errorf("held = %d, want the progress item kept, not dropped", s.Held())
	}
}

func TestDeferredProgressIsMergedIntoOneItem(t *testing.T) {
	s := &Scheduler{Gate: lowBudget(t), Merge: merge}
	s.Push(Item{Class: ClassProgress, Key: "issue-7", Payload: "a"})
	s.Push(Item{Class: ClassProgress, Key: "issue-7", Payload: "b"})
	s.Push(Item{Class: ClassProgress, Key: "issue-7", Payload: "c"})
	if s.Held() != 1 {
		t.Fatalf("held = %d, want one merged progress item", s.Held())
	}
	s.Gate = &Budget{} // budget recovers
	it, ok, _ := s.Pop(t0)
	if !ok || it.Payload != "a+b+c" {
		t.Errorf("merged = %+v", it)
	}
}

func TestSafetyNoticeIsNotDelayedBehindQueuedProgress(t *testing.T) {
	s := &Scheduler{Gate: &Budget{}, Merge: merge}
	for _, k := range []string{"a", "b", "c"} {
		s.Push(Item{Class: ClassProgress, Key: k, Payload: k})
	}
	s.Push(Item{Class: ClassSafetyNotice, Key: "sec", Payload: "approval needed"})
	it, ok, _ := s.Pop(t0)
	if !ok || it.Class != ClassSafetyNotice {
		t.Errorf("first out = %+v", it)
	}
}

func TestBlockedBudgetReturnsWhenToLookAgain(t *testing.T) {
	b := &Budget{}
	b.Block(t0.Add(45 * time.Second))
	s := &Scheduler{Gate: b, Merge: merge}
	s.Push(Item{Class: ClassStateWrite, Key: "w", Payload: "x"})
	if _, ok, until := s.Pop(t0); ok || until.Sub(t0) != 45*time.Second {
		t.Errorf("ok=%v until=%v", ok, until.Sub(t0))
	}
	if s.Held() != 1 {
		t.Error("a state write was dropped while the budget was blocked")
	}
}

func TestSameClassKeepsFirstInFirstOut(t *testing.T) {
	s := &Scheduler{Gate: &Budget{}, Merge: merge}
	s.Push(Item{Class: ClassStateWrite, Key: "a", Payload: "1"})
	s.Push(Item{Class: ClassStateWrite, Key: "b", Payload: "2"})
	x, _, _ := s.Pop(t0)
	y, _, _ := s.Pop(t0)
	if x.Key != "a" || y.Key != "b" {
		t.Errorf("order %s,%s", x.Key, y.Key)
	}
}

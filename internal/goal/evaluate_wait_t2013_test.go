package goal

import (
	"context"
	"strings"
	"testing"
)

// The legitimate-wait exemption: an unmet goal does not block the turn end
// while the caller reports that the session's card holds a valid open wait.
// The evaluator stays pure; the wait is an injected predicate. Every ceiling
// check still runs first and wins over the exemption.

func waitTrue() bool  { return true }
func waitFalse() bool { return false }

// modelOnlyGoal is the goal shape of the long-wait scenario: nothing
// mechanical to run, one model claim pending orchestrator evaluation.
func modelOnlyGoal() *Goal {
	return NewGoal("s", "wait for the predecessor", []Condition{
		{Type: ConditionModel, Claim: "the predecessor landed"},
	})
}

// failingRunner makes every mechanical condition exit 1 with the same output.
func failingRunner() *fakeRunner {
	return &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 1, out: "FAIL: TestX"}}}
}

// (1) An unmet model-only goal with a valid wait does not block, stays armed,
// counts the turn, and records a Progress entry so stagnation keeps working.
func TestWait_ModelOnlyGoalDoesNotBlock(t *testing.T) {
	g := modelOnlyGoal()
	e := &Eval{Runner: &fakeRunner{}, ValidWait: waitTrue}
	v, block := e.Evaluate(context.Background(), g)
	if block || v.Decision != "" {
		t.Fatalf("a valid wait must not block: block=%v verdict=%+v", block, v)
	}
	if g.Status != StatusArmed {
		t.Errorf("goal must stay active, status = %q", g.Status)
	}
	if g.TurnsUsed != 1 {
		t.Errorf("the wait turn must still count, TurnsUsed = %d, want 1", g.TurnsUsed)
	}
	if len(g.Progress) != 1 {
		t.Fatalf("a wait turn appends one Progress entry, got %+v", g.Progress)
	}
	if g.Progress[0].Turn != 1 || g.Progress[0].Fingerprint == "" {
		t.Errorf("the Progress entry must carry the turn and the fingerprint stagnation keys on: %+v", g.Progress[0])
	}
}

// (2) Without a wait (nil predicate, or one that says false) the block is
// exactly the pre-existing model-claim block.
func TestWait_NilOrFalsePredicateBlocksAsBefore(t *testing.T) {
	for name, wait := range map[string]func() bool{"nil": nil, "false": waitFalse} {
		t.Run(name, func(t *testing.T) {
			g := modelOnlyGoal()
			e := &Eval{Runner: &fakeRunner{}, ValidWait: wait}
			v, block := e.Evaluate(context.Background(), g)
			if !block || v.Decision != "block" {
				t.Fatalf("no valid wait must block: block=%v verdict=%+v", block, v)
			}
			if !strings.Contains(v.Reason, "model claim pending orchestrator evaluation") {
				t.Errorf("reason = %q", v.Reason)
			}
			if g.TurnsUsed != 1 || len(g.Progress) != 1 {
				t.Errorf("TurnsUsed=%d progress=%+v", g.TurnsUsed, g.Progress)
			}
		})
	}
}

// (3) The turn ceiling wins over the wait: a wait never extends a goal beyond
// its ceiling.
func TestWait_TurnCeilingStillFires(t *testing.T) {
	g := modelOnlyGoal()
	g.Ceiling.MaxTurns = 2
	g.TurnsUsed = 1
	e := &Eval{Runner: &fakeRunner{}, ValidWait: waitTrue}
	v, block := e.Evaluate(context.Background(), g)
	if block || !v.CeilingExit || v.Verdict == nil {
		t.Fatalf("the ceiling must still fire with a wait present: block=%v verdict=%+v", block, v)
	}
	if g.Status != StatusCeilingExit {
		t.Errorf("status = %q, want ceiling-exit", g.Status)
	}
}

// (3b) The wall-clock bound wins over the wait as well.
func TestWait_WallClockBoundStillFires(t *testing.T) {
	g := modelOnlyGoal()
	g.Ceiling.MaxTurns = 0
	g.Ceiling.MaxDuration = 60
	g.CreatedAt = "2020-01-01T00:00:00Z"
	e := &Eval{Runner: &fakeRunner{}, ValidWait: waitTrue}
	v, block := e.Evaluate(context.Background(), g)
	if block || !v.WallClockExit {
		t.Fatalf("the wall-clock bound must still fire with a wait present: block=%v verdict=%+v", block, v)
	}
}

// (4) Stagnation still stops a waiting goal after three identical turns.
func TestWait_StagnationStillFires(t *testing.T) {
	g := modelOnlyGoal()
	e := &Eval{Runner: &fakeRunner{}, ValidWait: waitTrue}
	for turn := 1; turn <= 3; turn++ {
		v, block := e.Evaluate(context.Background(), g)
		if block || v.Stagnation {
			t.Fatalf("turn %d: want a quiet wait turn, got block=%v verdict=%+v", turn, block, v)
		}
	}
	v, block := e.Evaluate(context.Background(), g)
	if block || !v.Stagnation || v.Verdict == nil {
		t.Fatalf("turn 4: three identical wait turns must end in stagnation: block=%v verdict=%+v", block, v)
	}
	if g.Status != StatusCeilingExit {
		t.Errorf("status = %q, want ceiling-exit", g.Status)
	}
}

// (5) A failing mechanical condition plus a valid wait does not block.
func TestWait_MechanicalFailureDoesNotBlock(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0}})
	e := &Eval{Runner: failingRunner(), ValidWait: waitTrue}
	v, block := e.Evaluate(context.Background(), g)
	if block || v.Decision != "" {
		t.Fatalf("a mechanical failure under a valid wait must not block: block=%v verdict=%+v", block, v)
	}
	if g.Status != StatusArmed || g.TurnsUsed != 1 || len(g.Progress) != 1 {
		t.Errorf("status=%q turns=%d progress=%+v", g.Status, g.TurnsUsed, g.Progress)
	}
}

// (5b) The exit-127 Unsatisfiable path is unaffected by a wait: a condition that
// can never pass is reported, not waited on.
func TestWait_Exit127StillUnsatisfiable(t *testing.T) {
	g := NewGoal("s", "prose", []Condition{{Type: ConditionMechanical, Cmd: "prose claim", ExpectExit: 0}})
	e := &Eval{Runner: exit127Runner{}, ValidWait: waitTrue}
	v, block := e.Evaluate(context.Background(), g)
	if block || !v.Unsatisfiable || v.Verdict == nil {
		t.Fatalf("exit 127 must still be unsatisfiable under a wait: block=%v verdict=%+v", block, v)
	}
	if g.Status != StatusUnsatisfiable {
		t.Errorf("status = %q, want unsatisfiable", g.Status)
	}
}

// (5c) A satisfied goal is satisfied whatever the wait says.
func TestWait_SatisfiedStaysSatisfied(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0}})
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 0, out: "ok"}}}
	e := &Eval{Runner: runner, ValidWait: waitTrue}
	if _, block := e.Evaluate(context.Background(), g); block {
		t.Fatal("a satisfied goal must not block")
	}
	if g.Status != StatusSatisfied {
		t.Errorf("status = %q, want satisfied", g.Status)
	}
}

// (6) The semi-autonomous checkpoint block is suppressed by a valid wait too.
func TestWait_SemiAutonomousCheckpointSuppressed(t *testing.T) {
	g := modelOnlyGoal()
	g.ProgressionMode = ProgressionSemiAutonomous
	e := &Eval{Runner: &fakeRunner{}, ValidWait: waitTrue}
	v, block := e.Evaluate(context.Background(), g)
	if block || v.Decision != "" || v.Mode != "" {
		t.Fatalf("no checkpoint block under a valid wait: block=%v verdict=%+v", block, v)
	}

	// Without the wait the same goal still raises the checkpoint.
	g2 := modelOnlyGoal()
	g2.ProgressionMode = ProgressionSemiAutonomous
	v2, block2 := (&Eval{Runner: &fakeRunner{}}).Evaluate(context.Background(), g2)
	if !block2 || v2.Mode != string(ProgressionSemiAutonomous) {
		t.Fatalf("control: the checkpoint must still fire without a wait: block=%v verdict=%+v", block2, v2)
	}
}

// The native-/goal yield and the inactive-status exits do not consult the wait.
func TestWait_PredicateNotConsultedOnEarlyExits(t *testing.T) {
	called := false
	probe := func() bool { called = true; return true }
	g := modelOnlyGoal()
	g.Status = StatusCleared
	(&Eval{Runner: &fakeRunner{}, ValidWait: probe}).Evaluate(context.Background(), g)
	g2 := modelOnlyGoal()
	(&Eval{Runner: &fakeRunner{}, ValidWait: probe, NativeGoalActive: true}).Evaluate(context.Background(), g2)
	if called {
		t.Error("the wait predicate must be consulted only where an unmet goal would block")
	}
}

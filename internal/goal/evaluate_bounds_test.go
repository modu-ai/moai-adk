package goal

import (
	"context"
	"strings"
	"testing"
	"time"
)

// scriptedReader returns a different output per call index, so a test can drive
// "stagnant" (same output each turn) vs "progressing" (output changes) scenarios.
type scriptedReader struct {
	script []struct {
		exit int
		out  string
	}
	calls int
}

func (s *scriptedReader) Run(_ context.Context, _ string) (int, string, error) {
	if s.calls >= len(s.script) {
		s.calls = len(s.script) - 1 // hold last
	}
	r := s.script[s.calls]
	s.calls++
	return r.exit, r.out, nil
}

// TestAC005_WallClockBoundFiresVerdict asserts that a goal armed with
// --max-turns 0 --max-duration <D> fires a 5-section verdict when the wall-clock
// time since CreatedAt exceeds D, AND the MaxTurns ceiling did NOT fire (it is
// disabled at 0). SPEC-INFINITE-GOAL-001 REQ-4 / OQ-2.
func TestAC005_WallClockBoundFiresVerdict(t *testing.T) {
	g := NewGoal("s", "never converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0               // infinite (AC-001)
	g.Ceiling.MaxDuration = 3600         // 1 hour wall-clock bound
	g.ProgressionMode = ProgressionAutonomous
	// CreatedAt set to 2 hours ago → elapsed(2h) > MaxDuration(1h) → must fire.
	g.CreatedAt = time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)

	e := &Eval{Runner: neverSatisfiedRunner{}, StagnationThreshold: 100}
	v, blocked := e.Evaluate(context.Background(), g)

	if blocked {
		t.Fatalf("AC-005: wall-clock bound must halt (not block); got blocked=true")
	}
	if !v.WallClockExit {
		t.Errorf("AC-005: WallClockExit must be true; got %+v", v)
	}
	if v.CeilingExit {
		t.Errorf("AC-005: MaxTurns ceiling must NOT fire (disabled at 0); got CeilingExit=true")
	}
	if v.Verdict == nil {
		t.Fatalf("AC-005: 5-section verdict must be emitted; got nil")
	}
	// The verdict must carry all 5 sections (indistinguishable in shape from a
	// MaxTurns-ceiling verdict).
	if v.Verdict.Claim == "" || v.Verdict.Evidence == "" ||
		v.Verdict.BaselineAttribution == "" || v.Verdict.Gaps == "" ||
		v.Verdict.ResidualRisk == "" {
		t.Errorf("AC-005: verdict must carry all 5 sections; got %+v", v.Verdict)
	}
	if !strings.Contains(strings.ToLower(v.Verdict.Claim), "wall") &&
		!strings.Contains(strings.ToLower(v.Verdict.Evidence), "wall") &&
		!strings.Contains(strings.ToLower(v.Verdict.Claim), "duration") &&
		!strings.Contains(strings.ToLower(v.Verdict.Evidence), "duration") {
		t.Errorf("AC-005: verdict must indicate the wall-clock/duration bound; got Claim=%q Evidence=%q", v.Verdict.Claim, v.Verdict.Evidence)
	}
	if g.Status != StatusCeilingExit {
		t.Errorf("AC-005: status must be ceiling-exit; got %q", g.Status)
	}
}

// TestAC005_WallClockNotYetElapsedDoesNotFire asserts the bound does NOT fire
// before elapsed exceeds MaxDuration (the goal keeps blocking).
func TestAC005_WallClockNotYetElapsedDoesNotFire(t *testing.T) {
	g := NewGoal("s", "never converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0
	g.Ceiling.MaxDuration = 3600 // 1h
	g.CreatedAt = time.Now().UTC().Format(time.RFC3339) // just now → elapsed ~0
	g.ProgressionMode = ProgressionAutonomous

	e := &Eval{Runner: neverSatisfiedRunner{}, StagnationThreshold: 100}
	v, blocked := e.Evaluate(context.Background(), g)
	if v.WallClockExit {
		t.Errorf("AC-005 negative: wall-clock must not fire immediately; got %+v", v)
	}
	if !blocked {
		t.Errorf("AC-005 negative: goal must still block before the bound fires")
	}
}

// TestAC006_StagnationFiresOnIdenticalMechanicalState asserts the strengthened
// stagnation guard fires after N consecutive turns with IDENTICAL mechanical-
// condition state. SPEC-INFINITE-GOAL-001 REQ-4 / AC-006.
func TestAC006_StagnationFiresOnIdenticalMechanicalState(t *testing.T) {
	g := NewGoal("s", "never converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0 // infinite so the ceiling does not fire first
	g.Ceiling.MaxDuration = 360000
	g.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	g.ProgressionMode = ProgressionAutonomous

	// Identical output every turn → identical mechanical fingerprint.
	runner := &scriptedReader{script: []struct {
		exit int
		out  string
	}{{exit: 1, out: "same output"}}}
	e := &Eval{Runner: runner, StagnationThreshold: 3}

	var fired bool
	for turn := 1; turn <= 5; turn++ {
		v, blocked := e.Evaluate(context.Background(), g)
		if v.Stagnation {
			fired = true
			if v.Verdict == nil {
				t.Fatalf("AC-006: stagnation verdict missing 5-section report")
			}
			_ = turn
			break
		}
		if !blocked {
			t.Fatalf("AC-006: unexpected non-block at turn %d: %+v", turn, v)
		}
	}
	if !fired {
		t.Errorf("AC-006: stagnation did not fire within 5 turns of identical state")
	}
}

// TestAC006_StagnationResetsOnChange asserts a single mechanical-state change
// at turn N-1 resets the counter and the guard does NOT fire at turn N.
func TestAC006_StagnationResetsOnChange(t *testing.T) {
	g := NewGoal("s", "never converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0
	g.Ceiling.MaxDuration = 360000
	g.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	g.ProgressionMode = ProgressionAutonomous

	// Output changes each turn → fingerprint differs → stagnation never fires.
	runner := &scriptedReader{script: []struct {
		exit int
		out  string
	}{
		{exit: 1, out: "turn-1-output"},
		{exit: 1, out: "turn-2-output"},
		{exit: 1, out: "turn-3-output"},
		{exit: 1, out: "turn-4-output"},
		{exit: 1, out: "turn-5-output"},
	}}
	e := &Eval{Runner: runner, StagnationThreshold: 3}

	for turn := 1; turn <= 5; turn++ {
		v, _ := e.Evaluate(context.Background(), g)
		if v.Stagnation {
			t.Errorf("AC-006 negative: stagnation fired at turn %d despite changing output", turn)
		}
	}
}

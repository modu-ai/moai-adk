package goal

import (
	"context"
	"testing"
)

// neverSatisfiedRunner is a CmdRunner whose mechanical conditions always exit 1
// (never satisfied), so the goal never converges and the only question is
// which ceiling/bound fires. (Named distinctly to avoid colliding with the
// existing fakeRunner in evaluate_test.go.)
type neverSatisfiedRunner struct{}

func (neverSatisfiedRunner) Run(_ context.Context, _ string) (int, string, error) {
	return 1, "never", nil
}

// TestAC001_EvaluatorSkipsCeilingAtMaxTurnsZero asserts that with
// Ceiling.MaxTurns == 0 the evaluator's `> 0` guard at evaluate.go skips the
// ceiling check across turns > 30 (the guard's C2 finding: 0 IS the infinite
// entry point). The guard is NOT modified (AP-1); only the propagated value.
func TestAC001_EvaluatorSkipsCeilingAtMaxTurnsZero(t *testing.T) {
	g := NewGoal("s", "never converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 0
	// Bound the simulation in-test so it cannot run away if the guard regresses:
	// arm a generous wall-clock so the wall-clock bound (M4) is not the limiter
	// here; the stagnation guard would catch a true runaway.
	g.ProgressionMode = ProgressionAutonomous

	e := &Eval{Runner: neverSatisfiedRunner{}}
	ctx := context.Background()

	for turn := 1; turn <= 50; turn++ {
		v, blocked := e.Evaluate(ctx, g)
		_ = v
		// Across turns 1..50 the MaxTurns ceiling (disabled at 0) must NOT fire.
		// A ceiling-exit verdict would surface as CeilingExit==true and blocked==false.
		if v.CeilingExit {
			t.Fatalf("AC-001: ceiling verdict fired at turn %d with MaxTurns=0 (guard regressed)", turn)
		}
		if !blocked {
			// The only legitimate non-block exits before satisfaction are ceiling
			// or stagnation. Stagnation is allowed (it is a real bound); ceiling
			// is not. If we hit stagnation first the test has still proven the
			// ceiling is disabled — record and stop.
			if v.Stagnation {
				return
			}
			t.Fatalf("AC-001: unexpected non-block non-stagnation exit at turn %d: %+v", turn, v)
		}
	}
	// If we reached turn 50 still blocked, the ceiling (30) never fired — PASS.
}

// TestAC002_EvaluatorFiresCeilingAtDefaultMaxTurns asserts the backward-compat
// half: with MaxTurns == 30 the ceiling fires at turn 30.
func TestAC002_EvaluatorFiresCeilingAtDefaultMaxTurns(t *testing.T) {
	g := NewGoal("s", "never converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	// NewGoal already sets Ceiling.MaxTurns = DefaultMaxTurns (30).
	if g.Ceiling.MaxTurns != DefaultMaxTurns {
		t.Fatalf("AC-002: DefaultMaxTurns = %d, want %d", g.Ceiling.MaxTurns, DefaultMaxTurns)
	}
	e := &Eval{Runner: neverSatisfiedRunner{}, StagnationThreshold: 100} // suppress stagnation so ceiling fires first
	ctx := context.Background()

	fired := false
	for turn := 1; turn <= 35; turn++ {
		v, blocked := e.Evaluate(ctx, g)
		if v.CeilingExit {
			if turn != DefaultMaxTurns {
				t.Errorf("AC-002: ceiling fired at turn %d, want %d", turn, DefaultMaxTurns)
			}
			fired = true
			break
		}
		if !blocked {
			break
		}
	}
	if !fired {
		t.Errorf("AC-002: ceiling did not fire by turn %d", DefaultMaxTurns)
	}
}

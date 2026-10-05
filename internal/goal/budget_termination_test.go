package goal

import (
	"context"
	"testing"
	"time"
)

// TestGoalBudgetTerminationNotSuccess is AC-HPR-014
// (SPEC-DUAL-HARNESS-HOOK-PARITY-001 REQ-HPR-016): a goal at its turn ceiling,
// its wall-clock bound, or its stagnation limit terminates the loop, carries a
// verdict that persists, and never reads as satisfied.
func TestGoalBudgetTerminationNotSuccess(t *testing.T) {
	cases := []struct {
		name  string
		setup func(g *Goal)
		eval  *Eval
		cause func(v Verdict) bool
	}{
		{
			name:  "turn ceiling",
			setup: func(g *Goal) { g.Ceiling.MaxTurns = 3; g.TurnsUsed = 2 },
			eval:  &Eval{Runner: neverSatisfiedRunner{}, StagnationThreshold: 100},
			cause: func(v Verdict) bool { return v.CeilingExit },
		},
		{
			name: "wall-clock bound",
			setup: func(g *Goal) {
				g.Ceiling.MaxDuration = 60
				g.CreatedAt = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
			},
			eval:  &Eval{Runner: neverSatisfiedRunner{}, StagnationThreshold: 100},
			cause: func(v Verdict) bool { return v.WallClockExit },
		},
		{
			name: "stagnation",
			setup: func(g *Goal) {
				for i := 0; i < DefaultStagnationThreshold; i++ {
					g.Progress = append(g.Progress, ProgressEntry{Turn: i + 1, Note: "blocked", Fingerprint: "same"})
				}
				g.TurnsUsed = DefaultStagnationThreshold
			},
			eval:  &Eval{Runner: neverSatisfiedRunner{}},
			cause: func(v Verdict) bool { return v.Stagnation },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			g := NewGoal("budget-"+tc.name, "never met", []Condition{{Type: ConditionMechanical, Cmd: "false"}})
			g.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			tc.setup(g)

			v, block := tc.eval.Evaluate(context.Background(), g)
			if block {
				t.Fatalf("the loop did not terminate: block = true, verdict %+v", v)
			}
			if !tc.cause(v) {
				t.Fatalf("the termination was not attributed to the %s: %+v", tc.name, v)
			}
			if g.Status == StatusSatisfied {
				t.Fatalf("a %s exit wrote status %q", tc.name, g.Status)
			}
			if g.Status != StatusCeilingExit {
				t.Fatalf("status after a %s exit = %q, want %q", tc.name, g.Status, StatusCeilingExit)
			}
			if v.Verdict == nil {
				t.Fatalf("a %s exit carries no verdict", tc.name)
			}
			if err := SaveVerdict(root, g.SessionID, &v); err != nil {
				t.Fatalf("persist verdict: %v", err)
			}
			got, err := LoadVerdict(root, g.SessionID)
			if err != nil || got == nil || got.Verdict == nil {
				t.Fatalf("persisted verdict not readable: %+v (%v)", got, err)
			}

			// A later Stop on the terminated goal must not flip it to satisfied.
			_, _ = tc.eval.Evaluate(context.Background(), g)
			if g.Status == StatusSatisfied {
				t.Fatalf("a terminated goal became %q on the next Stop", g.Status)
			}
		})
	}
}

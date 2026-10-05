package goal

import (
	"context"
	"strings"
	"testing"
)

// exit127Runner returns the shell's "command not found" status for every
// command, the signal `sh -c` produces when a condition's first word is not a
// command at all — the shape prose takes once it is misclassified mechanical.
type exit127Runner struct{}

func (exit127Runner) Run(_ context.Context, cmd string) (int, string, error) {
	first := cmd
	if i := strings.IndexAny(cmd, " \t\n"); i > 0 {
		first = cmd[:i]
	}
	return 127, "sh: " + first + ": command not found\n", nil
}

// TestEvaluate_Exit127IsUnsatisfiable is the eval-time backstop. A condition
// that cannot run is not a failing condition — it is a condition that can never
// pass, so blocking on it burns every remaining turn. The evaluator must stop
// blocking and say what the string probably was.
//
// This backstop covers goals armed BEFORE the arm-time gate landed, and prose
// whose first word happens to resolve to a real command at arm time.
func TestEvaluate_Exit127IsUnsatisfiable(t *testing.T) {
	g := NewGoal("s", "prose", []Condition{
		{Type: ConditionMechanical, Cmd: "모든 차단 AC가 통과 증거와 함께 표시된다", ExpectExit: 0},
	})
	e := &Eval{Runner: exit127Runner{}}
	v, block := e.Evaluate(context.Background(), g)

	if block {
		t.Fatalf("evaluator blocked on an unrunnable condition; verdict=%+v", v)
	}
	if !v.Unsatisfiable {
		t.Errorf("verdict.Unsatisfiable = false, want true")
	}
	if v.Verdict == nil {
		t.Fatalf("no 5-section verdict emitted")
	}
	blob := v.Verdict.Claim + v.Verdict.Evidence + v.Verdict.Gaps + v.Verdict.ResidualRisk
	if !strings.Contains(blob, "127") {
		t.Errorf("verdict does not cite exit 127:\n%+v", v.Verdict)
	}
	if !strings.Contains(blob, "model:") {
		t.Errorf("verdict does not suggest the model: prefix:\n%+v", v.Verdict)
	}
	if g.Status != StatusUnsatisfiable {
		t.Errorf("goal status = %q, want %q", g.Status, StatusUnsatisfiable)
	}
	if len(v.FailedConditions) != 1 || v.FailedConditions[0].Exit != 127 {
		t.Errorf("failed_conditions = %+v, want the one 127 condition", v.FailedConditions)
	}
}

// TestEvaluate_Expect127IsSatisfiedNotAborted is the false-positive guard: a
// condition that DECLARES exit 127 as its expected status is a legitimate
// assertion ("this command is absent"), and must be satisfied, never aborted.
func TestEvaluate_Expect127IsSatisfiedNotAborted(t *testing.T) {
	g := NewGoal("s", "absence check", []Condition{
		{Type: ConditionMechanical, Cmd: "definitely-not-installed", ExpectExit: 127},
	})
	e := &Eval{Runner: exit127Runner{}}
	v, block := e.Evaluate(context.Background(), g)

	if block {
		t.Fatalf("blocked on a satisfied expect-127 condition; verdict=%+v", v)
	}
	if v.Unsatisfiable {
		t.Errorf("expect-127 condition aborted as unsatisfiable; verdict=%+v", v)
	}
	if g.Status != StatusSatisfied {
		t.Errorf("goal status = %q, want %q", g.Status, StatusSatisfied)
	}
}

// TestEvaluate_OrdinaryFailureStillBlocks pins that the backstop did not turn
// every non-zero exit into an abort: exit 1 is a normal not-yet-converged turn.
func TestEvaluate_OrdinaryFailureStillBlocks(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 1, out: "FAIL"}}}
	v, block := (&Eval{Runner: runner}).Evaluate(context.Background(), g)
	if !block {
		t.Fatalf("ordinary exit-1 failure stopped blocking; verdict=%+v", v)
	}
	if v.Unsatisfiable {
		t.Errorf("ordinary exit-1 failure flagged unsatisfiable")
	}
}

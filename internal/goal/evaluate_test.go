package goal

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// fakeRunner is a deterministic CmdRunner for tests.
type fakeRunner struct {
	results map[string]struct {
		exit int
		out  string
		err  error
	}
}

func (f *fakeRunner) Run(_ context.Context, cmd string) (int, string, error) {
	r, ok := f.results[cmd]
	if !ok {
		return 1, "command not found", nil
	}
	return r.exit, r.out, r.err
}

// TestTier1Block (AC-GLE-010) asserts a failing mechanical condition yields a
// block JSON whose reason carries the failed condition + output tail.
func TestTier1Block(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 1, out: "FAIL: TestX boo\n--- FAIL: TestX (0.00s)\nlong output line"}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("want block on mechanical failure")
	}
	if v.Decision != "block" {
		t.Errorf("decision: want block, got %q", v.Decision)
	}
	if !strings.Contains(v.Reason, "go test ./...") {
		t.Errorf("reason must name the failed cmd: %q", v.Reason)
	}
	if !strings.Contains(v.Reason, "FAIL: TestX") {
		t.Errorf("reason must carry the output tail: %q", v.Reason)
	}
}

// TestTier2Gate (AC-GLE-011) asserts Tier-2 model judgment runs ONLY after all
// mechanical conditions pass AND at least one model condition exists.
func TestTier2Gate(t *testing.T) {
	// All mechanical pass + model condition present → block surfacing the claim.
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
		{Type: ConditionModel, Claim: "all AC rows show PASS"},
	})
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 0, out: "ok"}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("mechanical pass + model condition must block (Tier-2 gate)")
	}
	if !strings.Contains(v.Reason, "all AC rows show PASS") {
		t.Errorf("reason must surface the model claim: %q", v.Reason)
	}
	if !strings.Contains(v.Reason, "orchestrator evaluation") {
		t.Errorf("reason must indicate orchestrator self-eval: %q", v.Reason)
	}
	if g.Status == StatusSatisfied {
		t.Error("must NOT be satisfied while model claim is pending")
	}
}

// TestAllPassNoBlock (AC-GLE-012) asserts all-mechanical-pass + no model
// conditions → no block + status satisfied.
func TestAllPassNoBlock(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 0, out: "ok"}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if block {
		t.Fatalf("all-pass must not block: %+v", v)
	}
	if g.Status != StatusSatisfied {
		t.Errorf("status: want satisfied, got %s", g.Status)
	}
}

// TestCeilingVerdict (AC-GLE-013) asserts reaching max_turns emits a verdict
// carrying the 5 section names and stops blocking.
func TestCeilingVerdict(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 2
	g.TurnsUsed = 1 // next eval increments to 2 → ceiling
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"false": {exit: 1, out: ""}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if block {
		t.Fatal("ceiling must NOT block")
	}
	if !v.CeilingExit {
		t.Error("CeilingExit flag must be set")
	}
	if v.Verdict == nil {
		t.Fatal("ceiling verdict must be populated")
	}
	data, _ := json.Marshal(v.Verdict)
	js := string(data)
	for _, key := range []string{`"Claim"`, `"Evidence"`, `"Baseline-attribution"`, `"Gaps"`, `"Residual-risk"`} {
		if !strings.Contains(js, key) {
			t.Errorf("ceiling verdict missing section %s", key)
		}
	}
	if g.Status != StatusCeilingExit {
		t.Errorf("status: want ceiling-exit, got %s", g.Status)
	}
}

// TestNativeGoalYield (AC-GLE-016) asserts an active native-/goal signal makes
// stop-goal yield (no block).
func TestNativeGoalYield(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"false": {exit: 1, out: ""}}}
	e := &Eval{Runner: runner, NativeGoalActive: true}
	_, block := e.Evaluate(context.Background(), g)
	if block {
		t.Fatal("native /goal active must yield (no block)")
	}
}

// TestStagnationStop (AC-GLE-017) asserts N no-progress iterations trigger a
// stop + an E1/E3 escalation note in the verdict.
func TestStagnationStop(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.Ceiling.MaxTurns = 100 // high so ceiling does not fire first
	// Pre-populate progress with N identical notes → stagnation on next eval.
	for i := 0; i < DefaultStagnationThreshold; i++ {
		g.Progress = append(g.Progress, ProgressEntry{Turn: i + 1, Note: "same"})
	}
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"false": {exit: 1, out: ""}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if block {
		t.Fatal("stagnation must NOT block (it halts)")
	}
	if !v.Stagnation {
		t.Error("Stagnation flag must be set")
	}
	if v.Verdict == nil {
		t.Fatal("stagnation verdict must be populated")
	}
	joined := v.Verdict.Claim + " " + v.Verdict.Gaps + " " + v.Verdict.ResidualRisk
	if !strings.Contains(strings.ToLower(joined), "stagnation") && !strings.Contains(joined, "E1") && !strings.Contains(joined, "E3") {
		t.Errorf("verdict must carry stagnation/E1/E3 note: %q", joined)
	}
}

// TestAutonomousModeNoCheckpoint (AC-GLE-028) asserts autonomous mode emits the
// normal block JSON with NO semi-autonomous checkpoint signal.
func TestAutonomousModeNoCheckpoint(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
	})
	g.ProgressionMode = ProgressionAutonomous
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"false": {exit: 1, out: ""}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("autonomous mode must block on a failed condition")
	}
	if v.Mode == string(ProgressionSemiAutonomous) {
		t.Error("autonomous block must NOT carry the semi-autonomous checkpoint signal")
	}
	if v.Mode != "" {
		t.Errorf("autonomous block should have empty Mode, got %q", v.Mode)
	}
}

// TestSemiAutonomousCheckpointSignal (AC-GLE-029) asserts the semi-autonomous
// checkpoint JSON carries mode, the "semi-autonomous checkpoint" reason prefix,
// and failed_conditions with cmd/exit/tail when a mechanical condition fails.
func TestSemiAutonomousCheckpointSignal(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./internal/goal/...", ExpectExit: 0},
	})
	g.ProgressionMode = ProgressionSemiAutonomous
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./internal/goal/...": {exit: 1, out: "FAIL: TestSemiAutonomousCheckpoint boom"}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("semi-autonomous must block (checkpoint)")
	}
	if v.Decision != "block" {
		t.Errorf("decision: want block, got %q", v.Decision)
	}
	if v.Mode != string(ProgressionSemiAutonomous) {
		t.Errorf("mode: want semi-autonomous, got %q", v.Mode)
	}
	if !strings.Contains(strings.ToLower(v.Reason), "semi-autonomous checkpoint") {
		t.Errorf("reason must carry the checkpoint prefix: %q", v.Reason)
	}
	if len(v.FailedConditions) == 0 {
		t.Fatal("failed_conditions must be populated when a mechanical condition fails")
	}
	fc := v.FailedConditions[0]
	if fc.Cmd != "go test ./internal/goal/..." {
		t.Errorf("failed_conditions[0].cmd: %q", fc.Cmd)
	}
	if fc.Exit != 1 {
		t.Errorf("failed_conditions[0].exit: want 1, got %d", fc.Exit)
	}
	if !strings.Contains(fc.Tail, "FAIL: TestSemiAutonomousCheckpoint") {
		t.Errorf("failed_conditions[0].tail must carry output: %q", fc.Tail)
	}
}

// TestSemiAutonomousCheckpointModelPending asserts the checkpoint fires with
// empty failed_conditions when only a model claim is pending.
func TestSemiAutonomousCheckpointModelPending(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0},
		{Type: ConditionModel, Claim: "all AC rows show PASS"},
	})
	g.ProgressionMode = ProgressionSemiAutonomous
	runner := &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"true": {exit: 0, out: ""}}}
	e := &Eval{Runner: runner}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("semi-autonomous model-pending must block (checkpoint)")
	}
	if v.Mode != string(ProgressionSemiAutonomous) {
		t.Errorf("mode: want semi-autonomous, got %q", v.Mode)
	}
	if len(v.FailedConditions) != 0 {
		t.Errorf("model-pending checkpoint must have empty failed_conditions, got %v", v.FailedConditions)
	}
}

// TestNoKickoffBypass (AC-GLE-015) asserts an armed goal carries no
// run-phase-authorization path — the engine never emits a signal that would
// bypass Implementation Kickoff Approval.
func TestNoKickoffBypass(t *testing.T) {
	cases := []struct {
		name string
		mode ProgressionMode
	}{
		{"autonomous", ProgressionAutonomous},
		{"semi-autonomous", ProgressionSemiAutonomous},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGoal("s", "converge", []Condition{
				{Type: ConditionMechanical, Cmd: "true", ExpectExit: 0},
			})
			g.ProgressionMode = tc.mode
			runner := &fakeRunner{results: map[string]struct {
				exit int
				out  string
				err  error
			}{"true": {exit: 0, out: ""}}}
			e := &Eval{Runner: runner}
			v, _ := e.Evaluate(context.Background(), g)
			data, _ := json.Marshal(v)
			js := strings.ToLower(string(data))
			for _, forbidden := range []string{"kickoff-bypass", "authorize-run", "run-phase-authorized", "skip-kickoff"} {
				if strings.Contains(js, forbidden) {
					t.Errorf("verdict must not carry %q (no-bypass invariant): %s", forbidden, js)
				}
			}
		})
	}
}

// TestKickoffMandatoryBothModes (AC-GLE-034 clause b) asserts NO code path in
// internal/goal/ authorizes run-phase entry regardless of progression_mode. The
// Verdict schema carries no kickoff/authorization field; this test pins that
// invariant by exercising both modes and scanning every serialized field.
func TestKickoffMandatoryBothModes(t *testing.T) {
	for _, mode := range []ProgressionMode{ProgressionAutonomous, ProgressionSemiAutonomous} {
		g := NewGoal("s", "converge", []Condition{
			{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0},
		})
		g.ProgressionMode = mode
		runner := &fakeRunner{results: map[string]struct {
			exit int
			out  string
			err  error
		}{"false": {exit: 1, out: ""}}}
		e := &Eval{Runner: runner}
		v, _ := e.Evaluate(context.Background(), g)
		data, _ := json.Marshal(v)
		js := string(data)
		// The Verdict struct has no field that could authorize run-phase entry.
		// Assert the known-safe field set; any new field that implies authorization
		// would need to be added here deliberately.
		if strings.Contains(strings.ToLower(js), "authoriz") || strings.Contains(strings.ToLower(js), "bypass") {
			t.Errorf("mode=%s verdict must not contain authorization language: %s", mode, js)
		}
	}
}

// TestInactiveGoalNoBlock asserts a cleared/satisfied goal produces no block.
func TestInactiveGoalNoBlock(t *testing.T) {
	for _, status := range []Status{StatusCleared, StatusSatisfied} {
		g := NewGoal("s", "x", []Condition{{Type: ConditionMechanical, Cmd: "false", ExpectExit: 0}})
		g.Status = status
		e := &Eval{Runner: &fakeRunner{}}
		_, block := e.Evaluate(context.Background(), g)
		if block {
			t.Errorf("status=%s must not block", status)
		}
	}
}

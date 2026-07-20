package goal

import (
	"context"
	"strings"
	"testing"
)

// countingRunner wraps fakeRunner and counts Run invocations so tests can
// assert the snapshot-reuse path skips command execution entirely.
type countingRunner struct {
	inner *fakeRunner
	calls int
}

func (c *countingRunner) Run(ctx context.Context, cmd string) (int, string, error) {
	c.calls++
	return c.inner.Run(ctx, cmd)
}

// fakeSnapshotSource is a deterministic SnapshotSource: exact byte-string
// command match against the entries map; anything else is a miss.
type fakeSnapshotSource struct {
	entries map[string]int // command → recorded exit code
	attr    string
	lookups int
}

func (f *fakeSnapshotSource) Lookup(_ context.Context, cmd string) (int, string, bool) {
	f.lookups++
	exit, ok := f.entries[cmd]
	if !ok {
		return 0, "", false
	}
	return exit, f.attr, true
}

// TestEvaluateSnapshotReuse (reuse leg) asserts an exact byte-string command
// match against a fresh snapshot reuses the recorded exit code WITHOUT a
// CmdRunner call, and the verdict payload carries the snapshot attribution.
func TestEvaluateSnapshotReuse(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &countingRunner{inner: &fakeRunner{}}
	src := &fakeSnapshotSource{
		entries: map[string]int{"go test ./...": 0},
		attr:    "snapshot .moai/state/verify/snapshots/abc.json key head:digest cmd \"go test ./...\" exit 0",
	}
	e := &Eval{Runner: runner, Snapshot: src}
	v, block := e.Evaluate(context.Background(), g)
	if block {
		t.Fatalf("recorded exit 0 must satisfy the condition without blocking: %+v", v)
	}
	if runner.calls != 0 {
		t.Fatalf("snapshot hit must NOT execute the command (CmdRunner calls: %d)", runner.calls)
	}
	if g.Status != StatusSatisfied {
		t.Errorf("status: want satisfied, got %s", g.Status)
	}
}

// TestEvaluateSnapshotAttribution asserts the reuse attribution rides the
// verdict payload (a blocking verdict here, so the payload is observable).
func TestEvaluateSnapshotAttribution(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
		{Type: ConditionModel, Claim: "all AC rows show PASS in the transcript"},
	})
	runner := &countingRunner{inner: &fakeRunner{}}
	attr := "snapshot .moai/state/verify/snapshots/abc.json key head:digest cmd \"go test ./...\" exit 0"
	src := &fakeSnapshotSource{entries: map[string]int{"go test ./...": 0}, attr: attr}
	e := &Eval{Runner: runner, Snapshot: src}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("pending model claim must block")
	}
	if runner.calls != 0 {
		t.Fatalf("mechanical condition must be served from the snapshot (CmdRunner calls: %d)", runner.calls)
	}
	if len(v.SnapshotAttribution) != 1 || v.SnapshotAttribution[0] != attr {
		t.Fatalf("verdict must carry the snapshot attribution: %+v", v.SnapshotAttribution)
	}
}

// TestEvaluateSnapshotMissExecutes (miss leg) asserts a stale/missing snapshot
// falls back to the existing CmdRunner execution path unchanged.
func TestEvaluateSnapshotMissExecutes(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &countingRunner{inner: &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 0, out: "ok"}}}}
	src := &fakeSnapshotSource{entries: map[string]int{}} // empty → every lookup misses
	e := &Eval{Runner: runner, Snapshot: src}
	_, block := e.Evaluate(context.Background(), g)
	if block {
		t.Fatal("passing execution must not block")
	}
	if runner.calls != 1 {
		t.Fatalf("snapshot miss must execute the command exactly as today (CmdRunner calls: %d)", runner.calls)
	}
}

// TestEvaluateSnapshotNearMissExecutes (near-miss leg) asserts a command
// variant differing by even one byte (added flag) is NOT served from the
// snapshot — the exact-match contract has no normalization.
func TestEvaluateSnapshotNearMissExecutes(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test -count=1 ./...", ExpectExit: 0},
	})
	runner := &countingRunner{inner: &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test -count=1 ./...": {exit: 0, out: "ok"}}}}
	// Snapshot recorded the UN-flagged variant only.
	src := &fakeSnapshotSource{entries: map[string]int{"go test ./...": 0}}
	e := &Eval{Runner: runner, Snapshot: src}
	_, _ = e.Evaluate(context.Background(), g)
	if runner.calls != 1 {
		t.Fatalf("near-miss variant must execute (CmdRunner calls: %d)", runner.calls)
	}
}

// TestEvaluateSnapshotNilSource asserts a nil Snapshot source preserves the
// pre-existing behavior byte-for-byte (the contract is strictly additive).
func TestEvaluateSnapshotNilSource(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &countingRunner{inner: &fakeRunner{results: map[string]struct {
		exit int
		out  string
		err  error
	}{"go test ./...": {exit: 1, out: "FAIL"}}}}
	e := &Eval{Runner: runner} // no snapshot source
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("failing condition must block")
	}
	if runner.calls != 1 {
		t.Fatalf("nil source must execute (CmdRunner calls: %d)", runner.calls)
	}
	if len(v.SnapshotAttribution) != 0 {
		t.Errorf("nil source must not attribute a snapshot: %+v", v.SnapshotAttribution)
	}
	if !strings.Contains(v.Reason, "FAIL") {
		t.Errorf("execution path must carry the output tail: %q", v.Reason)
	}
}

// TestEvaluateSnapshotReusedFailure asserts a recorded FAILING exit is also
// reused (the contract reuses the recorded exit code, whatever it is) and the
// block reason names the snapshot reuse instead of a fabricated output tail.
func TestEvaluateSnapshotReusedFailure(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
	})
	runner := &countingRunner{inner: &fakeRunner{}}
	src := &fakeSnapshotSource{
		entries: map[string]int{"go test ./...": 1},
		attr:    "snapshot .moai/state/verify/snapshots/abc.json key head:digest cmd \"go test ./...\" exit 1",
	}
	e := &Eval{Runner: runner, Snapshot: src}
	v, block := e.Evaluate(context.Background(), g)
	if !block {
		t.Fatal("recorded failing exit must block")
	}
	if runner.calls != 0 {
		t.Fatalf("recorded failure must be reused without execution (CmdRunner calls: %d)", runner.calls)
	}
	if len(v.SnapshotAttribution) != 1 {
		t.Fatalf("reused failure must carry attribution: %+v", v.SnapshotAttribution)
	}
	if !strings.Contains(v.Reason, "snapshot") {
		t.Errorf("block reason must name the snapshot reuse: %q", v.Reason)
	}
}

// TestEvaluateSnapshotConstantLookups asserts the evaluator performs at most
// one Lookup per mechanical condition per evaluation — the snapshot path adds
// no repeated calls (composes with the Source-side memoized constant cost).
func TestEvaluateSnapshotConstantLookups(t *testing.T) {
	g := NewGoal("s", "converge", []Condition{
		{Type: ConditionMechanical, Cmd: "go test ./...", ExpectExit: 0},
		{Type: ConditionMechanical, Cmd: "go vet ./...", ExpectExit: 0},
	})
	runner := &countingRunner{inner: &fakeRunner{}}
	src := &fakeSnapshotSource{entries: map[string]int{
		"go test ./...": 0,
		"go vet ./...":  0,
	}}
	e := &Eval{Runner: runner, Snapshot: src}
	_, _ = e.Evaluate(context.Background(), g)
	if src.lookups != 2 {
		t.Fatalf("want exactly 1 lookup per mechanical condition (2), got %d", src.lookups)
	}
	if runner.calls != 0 {
		t.Fatalf("all-hit evaluation must not execute (CmdRunner calls: %d)", runner.calls)
	}
}
